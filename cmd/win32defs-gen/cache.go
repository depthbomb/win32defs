package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func packageCachePath(directory string, digest string) (string, error) {
	decoded, err := hex.DecodeString(digest)
	if err != nil || len(decoded) != sha256.Size {
		return "", fmt.Errorf("invalid package SHA-256 %q", digest)
	}

	return filepath.Join(directory, "packages", strings.ToLower(digest)+".nupkg"), nil
}

func cachedPackage(source sourcePackage, directory string) ([]byte, bool, error) {
	if source.SHA256 == "" {
		return nil, false, nil
	}

	path, err := packageCachePath(directory, source.SHA256)
	if err != nil {
		return nil, false, err
	}

	marker, err := os.ReadFile(path + ".verified")
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	if string(marker) != strings.ToLower(source.SHA256) {
		return nil, false, fmt.Errorf("invalid signature-verification record for %s", source.Package)
	}

	archive, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	digest := sha256.Sum256(archive)
	if !strings.EqualFold(hex.EncodeToString(digest[:]), source.SHA256) {
		return nil, false, fmt.Errorf("cached package %s failed SHA-256 verification: %s", source.Package, path)
	}

	return archive, true, nil
}

func cachePackage(source sourcePackage, archive []byte, directory string) error {
	path, err := packageCachePath(directory, source.SHA256)
	if err != nil {
		return err
	}

	if err := replaceFile(path, archive); err != nil {
		return err
	}

	return replaceFile(path+".verified", []byte(strings.ToLower(source.SHA256)))
}
