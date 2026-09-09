package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type generatedFile struct {
	path     string
	contents []byte
	previous []byte
	existed  bool
}

func replaceFile(path string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	temporary, err := os.CreateTemp(filepath.Dir(path), ".win32defs-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())

	if _, err := temporary.Write(contents); err != nil {
		temporary.Close()

		return err
	}

	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()

		return err
	}

	if err := temporary.Close(); err != nil {
		return err
	}

	return os.Rename(temporary.Name(), path)
}

func publishGeneration(stage string, root string) error {
	return publishFiles(stage, root, replaceFile)
}

func publishFiles(stage string, root string, replace func(string, []byte) error) error {
	var files []generatedFile
	err := filepath.WalkDir(stage, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			return nil
		}

		if !entry.Type().IsRegular() {
			return fmt.Errorf("unexpected staged file type: %s", path)
		}

		relative, err := filepath.Rel(stage, path)
		if err != nil {
			return err
		}

		target := filepath.Join(root, relative)
		previous, err := os.ReadFile(target)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}

		existed := err == nil
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if existed && bytes.Equal(contents, previous) {
			return nil
		}

		files = append(files, generatedFile{
			path:     target,
			contents: contents,
			previous: previous,
			existed:  existed,
		})

		return nil
	})
	if err != nil {
		return fmt.Errorf("preflight generated output: %w", err)
	}

	for index, file := range files {
		if err := replace(file.path, file.contents); err != nil {
			failures := []error{fmt.Errorf("publish %s: %w", file.path, err)}
			for rollback := index; rollback >= 0; rollback-- {
				original := files[rollback]
				var restoreErr error
				if original.existed {
					restoreErr = replaceFile(original.path, original.previous)
				} else {
					restoreErr = os.Remove(original.path)
				}

				if restoreErr != nil && !errors.Is(restoreErr, os.ErrNotExist) {
					failures = append(failures, fmt.Errorf("restore %s: %w", original.path, restoreErr))
				}
			}

			return errors.Join(failures...)
		}
	}

	return nil
}
