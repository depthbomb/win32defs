package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestOfflineVerifiedCache(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	cache := t.TempDir()
	archive := []byte("test archive already signature-verified by the caller")
	digest := sha256.Sum256(archive)
	source := sourcePackage{
		Package: metadataPackageID,
		Version: "1.0.0",
		SHA256:  hex.EncodeToString(digest[:]),
		URL:     "https://invalid.invalid/metadata.nupkg",
	}
	lock := sourceLock{
		Metadata:      source,
		Documentation: source,
	}
	lock.Documentation.Package = documentationPackageID
	if err := os.MkdirAll(filepath.Join(root, "internal", "source"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := writeSourceLock(filepath.Join(root, "internal", "source", "source.lock.json"), lock); err != nil {
		t.Fatal(err)
	}

	options := generationOptions{
		Offline:  true,
		CacheDir: cache,
	}
	if _, _, _, err := resolveSource(context.Background(), root, options); err == nil {
		t.Fatal("offline generation must reject a cold cache")
	}

	if err := cachePackage(source, archive, cache); err != nil {
		t.Fatal(err)
	}

	_, got, _, err := resolveSource(context.Background(), root, options)
	if err != nil || string(got) != string(archive) {
		t.Fatalf("offline cache failed: %q, %v", got, err)
	}

	path, err := packageCachePath(cache, source.SHA256)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, _, err := resolveSource(context.Background(), root, options); err == nil {
		t.Fatal("tampered cache was accepted")
	}

	options.Latest = true
	if _, _, _, err := resolveSource(context.Background(), root, options); err == nil {
		t.Fatal("offline latest discovery was accepted")
	}
}

func TestCacheRequiresVerificationRecord(t *testing.T) {
	t.Parallel()

	cache := t.TempDir()
	archive := []byte("unverified")
	digest := sha256.Sum256(archive)
	source := sourcePackage{
		SHA256: hex.EncodeToString(digest[:]),
	}
	path, err := packageCachePath(cache, source.SHA256)
	if err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(path, archive); err != nil {
		t.Fatal(err)
	}

	if _, cached, err := cachedPackage(source, cache); err != nil || cached {
		t.Fatalf("unverified cache = %v, %v", cached, err)
	}

	if _, err := packageCachePath(cache, "../escape"); err == nil {
		t.Fatal("invalid digest accepted as a cache path")
	}
}
