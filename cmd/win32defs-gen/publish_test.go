package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPublishRollback(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	stage := t.TempDir()
	for _, name := range []string{"a.go", "b.go"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(stage, name), []byte("new"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	calls := 0
	err := publishFiles(stage, root, func(path string, contents []byte) error {
		calls++
		if calls == 2 {
			return errors.New("injected I/O failure")
		}

		return replaceFile(path, contents)
	})
	if err == nil {
		t.Fatal("publication unexpectedly succeeded")
	}

	for _, name := range []string{"a.go", "b.go"} {
		contents, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || string(contents) != "old" {
			t.Fatalf("%s was not restored: %q, %v", name, contents, err)
		}
	}
}

func TestPublishPreflightAndUnchangedFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	stage := t.TempDir()
	if err := replaceFile(filepath.Join(root, "a.go"), []byte("same")); err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(filepath.Join(stage, "a.go"), []byte("same")); err != nil {
		t.Fatal(err)
	}

	err := publishFiles(stage, root, func(string, []byte) error {
		t.Fatal("unchanged file should not be replaced")

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(filepath.Join(root, "b.go"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(filepath.Join(stage, "b.go"), []byte("new")); err != nil {
		t.Fatal(err)
	}

	err = publishFiles(stage, root, func(string, []byte) error {
		t.Fatal("preflight failure must prevent every replacement")

		return nil
	})
	if err == nil {
		t.Fatal("expected preflight error")
	}
}
