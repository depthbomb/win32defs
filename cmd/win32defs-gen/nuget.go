package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type serviceIndex struct {
	Resources []serviceResource `json:"resources"`
}

type serviceResource struct {
	ID   string          `json:"@id"`
	Type json.RawMessage `json:"@type"`
}

type versionIndex struct {
	Versions []string `json:"versions"`
}

type numericVersion struct {
	parts      []int
	prerelease string
	raw        string
}

type resolvedPackage struct {
	source  sourcePackage
	archive []byte
}

const (
	nugetServiceIndex      = "https://api.nuget.org/v3/index.json"
	metadataPackageID      = "Microsoft.Windows.SDK.Win32Metadata"
	documentationPackageID = "Microsoft.Windows.SDK.Win32Docs"
)

var httpClient = &http.Client{Timeout: 5 * time.Minute}

func fetch(ctx context.Context, url string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for %s: %w", url, err)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: unexpected HTTP status %s", url, response.Status)
	}

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", url, err)
	}

	return contents, nil
}

func packageBaseAddress(ctx context.Context) (string, error) {
	contents, err := fetch(ctx, nugetServiceIndex)
	if err != nil {
		return "", err
	}

	var index serviceIndex
	if err := json.Unmarshal(contents, &index); err != nil {
		return "", fmt.Errorf("decode NuGet service index: %w", err)
	}

	for _, resource := range index.Resources {
		if resourceHasType(resource.Type, "PackageBaseAddress/3.0.0") {
			return strings.TrimRight(resource.ID, "/") + "/", nil
		}
	}

	return "", errors.New("NuGet service index has no PackageBaseAddress/3.0.0 resource")
}

func resourceHasType(raw json.RawMessage, expected string) bool {
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return single == expected
	}

	var multiple []string
	if err := json.Unmarshal(raw, &multiple); err != nil {
		return false
	}

	for _, value := range multiple {
		if value == expected {
			return true
		}
	}

	return false
}

func discoverLatestVersion(ctx context.Context, baseAddress string, packageID string) (string, error) {
	packageName := strings.ToLower(packageID)
	contents, err := fetch(ctx, baseAddress+packageName+"/index.json")
	if err != nil {
		return "", err
	}

	var index versionIndex
	if err := json.Unmarshal(contents, &index); err != nil {
		return "", fmt.Errorf("decode package version index: %w", err)
	}

	if len(index.Versions) == 0 {
		return "", fmt.Errorf("package %s has no published versions", packageID)
	}

	latest := parseNumericVersion(index.Versions[0])
	for _, version := range index.Versions[1:] {
		candidate := parseNumericVersion(version)
		if compareNumericVersions(candidate, latest) > 0 {
			latest = candidate
		}
	}

	return latest.raw, nil
}

func packageDownloadURL(baseAddress string, packageID string, version string) string {
	packageName := strings.ToLower(packageID)

	return fmt.Sprintf("%s%s/%s/%s.%s.nupkg", baseAddress, packageName, version, packageName, version)
}

func parseNumericVersion(version string) numericVersion {
	components := strings.SplitN(version, "-", 2)
	numeric := components[0]
	pieces := strings.Split(numeric, ".")
	parts := make([]int, len(pieces))

	for index, piece := range pieces {
		value, err := strconv.Atoi(piece)
		if err != nil {
			value = 0
		}

		parts[index] = value
	}

	var prerelease string
	if len(components) == 2 {
		prerelease = components[1]
	}

	return numericVersion{parts: parts, prerelease: prerelease, raw: version}
}

func compareNumericVersions(left numericVersion, right numericVersion) int {
	count := len(left.parts)
	if len(right.parts) > count {
		count = len(right.parts)
	}

	for index := range count {
		var leftPart int
		if index < len(left.parts) {
			leftPart = left.parts[index]
		}

		var rightPart int
		if index < len(right.parts) {
			rightPart = right.parts[index]
		}

		if leftPart < rightPart {
			return -1
		}

		if leftPart > rightPart {
			return 1
		}
	}

	if left.prerelease == right.prerelease {
		return 0
	}

	if left.prerelease == "" {
		return 1
	}

	if right.prerelease == "" {
		return -1
	}

	return strings.Compare(left.prerelease, right.prerelease)
}

func readSourceLock(path string) (sourceLock, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return sourceLock{}, err
	}

	var lock sourceLock
	if err := json.Unmarshal(contents, &lock); err != nil {
		return sourceLock{}, fmt.Errorf("decode source lock: %w", err)
	}

	if !completeSourcePackage(lock.Metadata) || !completeSourcePackage(lock.Documentation) {
		return sourceLock{}, errors.New("source lock is incomplete")
	}

	if lock.Metadata.Package != metadataPackageID {
		return sourceLock{}, fmt.Errorf("metadata source package is %q, want %q", lock.Metadata.Package, metadataPackageID)
	}

	if lock.Documentation.Package != documentationPackageID {
		return sourceLock{}, fmt.Errorf("documentation source package is %q, want %q", lock.Documentation.Package, documentationPackageID)
	}

	return lock, nil
}

func completeSourcePackage(source sourcePackage) bool {
	return source.Package != "" && source.Version != "" && source.SHA256 != "" && source.URL != ""
}

func resolveSource(ctx context.Context, root string, latest bool) (sourceLock, []byte, []byte, error) {
	lockPath := filepath.Join(root, "internal", "source", "source.lock.json")
	lock, lockErr := readSourceLock(lockPath)

	baseAddress, err := packageBaseAddress(ctx)
	if err != nil {
		return sourceLock{}, nil, nil, err
	}

	if latest || errors.Is(lockErr, os.ErrNotExist) {
		metadataVersion, err := discoverLatestVersion(ctx, baseAddress, metadataPackageID)
		if err != nil {
			return sourceLock{}, nil, nil, err
		}

		documentationVersion, err := discoverLatestVersion(ctx, baseAddress, documentationPackageID)
		if err != nil {
			return sourceLock{}, nil, nil, err
		}

		lock = sourceLock{
			Metadata:      sourcePackage{Package: metadataPackageID, Version: metadataVersion},
			Documentation: sourcePackage{Package: documentationPackageID, Version: documentationVersion},
		}
	} else if lockErr != nil {
		return sourceLock{}, nil, nil, lockErr
	}

	metadata, err := resolvePackage(ctx, baseAddress, lock.Metadata, latest)
	if err != nil {
		return sourceLock{}, nil, nil, err
	}

	documentation, err := resolvePackage(ctx, baseAddress, lock.Documentation, latest)
	if err != nil {
		return sourceLock{}, nil, nil, err
	}

	lock.Metadata = metadata.source
	lock.Documentation = documentation.source

	return lock, metadata.archive, documentation.archive, nil
}

func resolvePackage(ctx context.Context, baseAddress string, source sourcePackage, latest bool) (resolvedPackage, error) {
	source.URL = packageDownloadURL(baseAddress, source.Package, source.Version)

	archive, err := fetch(ctx, source.URL)
	if err != nil {
		return resolvedPackage{}, err
	}

	digest := sha256.Sum256(archive)
	actualHash := hex.EncodeToString(digest[:])
	if source.SHA256 != "" && !latest && source.SHA256 != actualHash {
		return resolvedPackage{}, fmt.Errorf("package %s SHA-256 is %s, want %s", source.Package, actualHash, source.SHA256)
	}

	source.SHA256 = actualHash

	return resolvedPackage{source: source, archive: archive}, nil
}

func verifyPackage(ctx context.Context, packageID string, archive []byte) error {
	temporaryFile, err := os.CreateTemp("", "win32defs-metadata-*.nupkg")
	if err != nil {
		return fmt.Errorf("create package verification file: %w", err)
	}
	path := temporaryFile.Name()
	defer os.Remove(path)

	if _, err := temporaryFile.Write(archive); err != nil {
		temporaryFile.Close()

		return fmt.Errorf("write package verification file: %w", err)
	}

	if err := temporaryFile.Close(); err != nil {
		return fmt.Errorf("close package verification file: %w", err)
	}

	command := exec.CommandContext(ctx, "dotnet", "nuget", "verify", path, "--all")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("verify package %s signature: %w\n%s", packageID, err, output)
	}

	return nil
}

func extractWinMD(archive []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("open metadata package: %w", err)
	}

	for _, file := range reader.File {
		if filepath.Base(file.Name) != "Windows.Win32.winmd" {
			continue
		}

		stream, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open Windows.Win32.winmd: %w", err)
		}
		defer stream.Close()

		contents, err := io.ReadAll(stream)
		if err != nil {
			return nil, fmt.Errorf("read Windows.Win32.winmd: %w", err)
		}

		return contents, nil
	}

	return nil, errors.New("metadata package does not contain Windows.Win32.winmd")
}

func extractDocs(archive []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("open documentation package: %w", err)
	}

	for _, file := range reader.File {
		if filepath.Base(file.Name) != "apidocs.msgpack" {
			continue
		}

		stream, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open apidocs.msgpack: %w", err)
		}
		defer stream.Close()

		contents, err := io.ReadAll(stream)
		if err != nil {
			return nil, fmt.Errorf("read apidocs.msgpack: %w", err)
		}

		return contents, nil
	}

	return nil, errors.New("documentation package does not contain apidocs.msgpack")
}

func exportMetadata(ctx context.Context, root string, winmd []byte, docs []byte) (metadataExport, error) {
	temporaryDirectory, err := os.MkdirTemp("", "win32defs-generate-")
	if err != nil {
		return metadataExport{}, fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(temporaryDirectory)

	winmdPath := filepath.Join(temporaryDirectory, "Windows.Win32.winmd")
	docsPath := filepath.Join(temporaryDirectory, "apidocs.msgpack")
	exportPath := filepath.Join(temporaryDirectory, "metadata.json")
	if err := os.WriteFile(winmdPath, winmd, 0o600); err != nil {
		return metadataExport{}, fmt.Errorf("write temporary winmd: %w", err)
	}
	if err := os.WriteFile(docsPath, docs, 0o600); err != nil {
		return metadataExport{}, fmt.Errorf("write temporary documentation: %w", err)
	}

	projectPath := filepath.Join(root, "tools", "winmd-exporter", "winmd-exporter.csproj")
	command := exec.CommandContext(
		ctx,
		"dotnet",
		"run",
		"--project",
		projectPath,
		"--configuration",
		"Release",
		"--",
		winmdPath,
		docsPath,
		exportPath,
	)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return metadataExport{}, fmt.Errorf("run winmd exporter: %w", err)
	}

	contents, err := os.ReadFile(exportPath)
	if err != nil {
		return metadataExport{}, fmt.Errorf("read metadata export: %w", err)
	}

	var export metadataExport
	if err := json.Unmarshal(contents, &export); err != nil {
		return metadataExport{}, fmt.Errorf("decode metadata export: %w", err)
	}

	return export, nil
}

func writeSourceLock(path string, lock sourceLock) error {
	contents, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("encode source lock: %w", err)
	}

	contents = append(contents, '\n')
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		return fmt.Errorf("write source lock: %w", err)
	}

	return nil
}
