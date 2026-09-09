package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func generate(ctx context.Context, root string, options generationOptions) error {
	lock, metadataArchive, documentationArchive, err := resolveSource(ctx, root, options)
	if err != nil {
		return err
	}

	winmd, err := extractWinMD(metadataArchive)
	if err != nil {
		return err
	}

	docs, err := extractDocs(documentationArchive)
	if err != nil {
		return err
	}

	export, err := exportMetadata(ctx, root, winmd, docs, options.Offline)
	if err != nil {
		return err
	}

	sourceRoot := root
	root, err = os.MkdirTemp("", "win32defs-stage-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(root)

	derived, err := deriveConstants(export.Constants)
	if err != nil {
		return err
	}

	classifiedExport := export
	classifiedExport.Constants = append(classifiedExport.Constants, derived...)
	packages, collisions, skipped, rejected := collectConstants(classifiedExport)
	constantMethods, skippedMethods := collectConstantMethods(export.ConstantMethods)
	coverage := measureCoverage(export)
	packageCounts := make(map[string]int, len(packageSpecs))
	documented := 0

	for _, spec := range packageSpecs {
		methods := []generatedConstantMethod(nil)
		if spec.Name == "process" {
			methods = constantMethods
		}

		contents, err := renderConstants(spec, packages[spec.Name], methods, lock)
		if err != nil {
			return err
		}

		path := filepath.Join(root, spec.Name, "zz_generated.go")
		if err := writeFormatted(path, contents); err != nil {
			return err
		}

		packageCounts[spec.Name] = len(packages[spec.Name])
		for _, constant := range packages[spec.Name] {
			if normalizeComment(constant.Comment) != "" {
				documented++
			}
		}
	}

	guids, guidCollisions, skippedGUIDs := collectGUIDs(export.GUIDs)
	guidContents, err := renderGUIDs(guids, lock)
	if err != nil {
		return err
	}

	if err := writeFormatted(filepath.Join(root, "guid", "zz_generated.go"), guidContents); err != nil {
		return err
	}

	propertyKeys, skippedPropertyKeys := collectKeys(export.Initializers, ".PROPERTYKEY")
	propertyContents, err := renderKeys("propertykey", propertyKeys, lock)
	if err != nil {
		return err
	}

	if err := writeFormatted(filepath.Join(root, "propertykey", "zz_generated.go"), propertyContents); err != nil {
		return err
	}

	deviceKeys, skippedDeviceKeys := collectKeys(export.Initializers, ".DEVPROPKEY")
	deviceContents, err := renderKeys("devpropkey", deviceKeys, lock)
	if err != nil {
		return err
	}

	if err := writeFormatted(filepath.Join(root, "devpropkey", "zz_generated.go"), deviceContents); err != nil {
		return err
	}

	authorities, skippedAuthorities := collectAuthorities(export.Initializers)
	authorityContents, err := renderAuthorities(authorities, lock)
	if err != nil {
		return err
	}

	if err := writeFormatted(filepath.Join(root, "security", "zz_generated_authorities.go"), authorityContents); err != nil {
		return err
	}

	syncInitializers, skippedSyncInitializers := collectSyncInitializers(export.Initializers)
	syncContents, err := renderSyncInitializers(syncInitializers, lock)
	if err != nil {
		return err
	}

	if err := writeFormatted(filepath.Join(root, "syncinit", "zz_generated.go"), syncContents); err != nil {
		return err
	}

	provenanceContents, err := renderProvenance(lock)
	if err != nil {
		return err
	}

	if err := writeFormatted(filepath.Join(root, "metadata", "zz_generated.go"), provenanceContents); err != nil {
		return err
	}

	catalogContents, err := renderCatalog(packages, lock)
	if err != nil {
		return err
	}

	if err := writeFormatted(filepath.Join(root, "catalog", "zz_generated.go"), catalogContents); err != nil {
		return err
	}

	skipped["guid"] += skippedGUIDs
	skipped["property_key"] += skippedPropertyKeys
	skipped["device_property_key"] += skippedDeviceKeys
	skipped["constant_method"] += skippedMethods
	skipped["sid_identifier_authority"] += skippedAuthorities
	skipped["synchronization_initializer"] += skippedSyncInitializers
	collisions = append(collisions, guidCollisions...)

	report := generationReport{
		Source:               lock,
		PackageCount:         packageCounts,
		GUIDCount:            len(guids),
		GUIDSourceCount:      len(export.GUIDs),
		PropertyKeys:         len(propertyKeys),
		DeviceKeys:           len(deviceKeys),
		Documented:           documented,
		Coverage:             coverage,
		MethodCount:          len(constantMethods),
		AuthorityCount:       len(authorities),
		SyncInitializerCount: len(syncInitializers),
		Collisions:           collisions,
		Skipped:              skipped,
		Rejected:             append(rejected, rejectedStructuredDefinitions(export)...),
		Symbols:              emittedSymbols(packages),
		DerivedConstants:     derived,
	}
	for _, item := range guids {
		report.Symbols["guid"] = append(report.Symbols["guid"], item.Identifier)
	}
	for _, item := range propertyKeys {
		report.Symbols["propertykey"] = append(report.Symbols["propertykey"], item.Identifier)
	}
	for _, item := range deviceKeys {
		report.Symbols["devpropkey"] = append(report.Symbols["devpropkey"], item.Identifier)
	}
	for _, item := range constantMethods {
		report.Symbols["process"] = append(report.Symbols["process"], item.Name)
	}
	for _, item := range authorities {
		report.Symbols["security"] = append(report.Symbols["security"], item.Name)
	}
	for _, item := range syncInitializers {
		report.Symbols["syncinit"] = append(report.Symbols["syncinit"], item.Name)
	}
	for packageName := range report.Symbols {
		sort.Strings(report.Symbols[packageName])
	}

	if err := validateGeneration(sourceRoot, report, options.AcceptChanges); err != nil {
		return err
	}

	if err := writeReport(filepath.Join(root, "internal", "source", "report.json"), report); err != nil {
		return err
	}

	if err := writeSourceLock(filepath.Join(root, "internal", "source", "source.lock.json"), lock); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return publishGeneration(root, sourceRoot)
}

func run() error {
	root := flag.String("root", ".", "repository root")
	latest := flag.Bool("latest", false, "discover and use the newest metadata package")
	offline := flag.Bool("offline", false, "use cached, signature-verified source packages without network access")
	cacheDir := flag.String("cache-dir", "", "source package cache directory (defaults to the user cache)")
	acceptChanges := flag.Bool("accept-changes", false, "accept reviewed new omissions, rejected definitions, or collisions")
	timeout := flag.Duration("timeout", 15*time.Minute, "overall generation timeout")

	flag.Parse()
	if *cacheDir == "" {
		directory, err := os.UserCacheDir()
		if err != nil {
			return err
		}

		*cacheDir = filepath.Join(directory, "win32defs")
	}

	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}

	if _, err := os.Stat(filepath.Join(absoluteRoot, "go.mod")); err != nil {
		return fmt.Errorf("repository root is invalid: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	options := generationOptions{
		Latest:        *latest,
		Offline:       *offline,
		CacheDir:      *cacheDir,
		AcceptChanges: *acceptChanges,
	}
	if err := generate(ctx, absoluteRoot, options); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("generation timed out after %s: %w", *timeout, err)
		}

		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
