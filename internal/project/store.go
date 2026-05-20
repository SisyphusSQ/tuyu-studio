package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func CreateBaselineProject(root string, manifest Manifest) error {
	if root == "" {
		return errors.New("project root is required")
	}

	report := ValidateManifest(manifest)
	if report.HasBlocking() {
		return ValidationError{Report: report}
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("create project root: %w", err)
	}

	for _, dir := range RequiredProjectDirectories() {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
			return fmt.Errorf("create project directory %q: %w", dir, err)
		}
	}

	data, err := EncodeManifest(manifest)
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(root, ManifestFileName)
	file, err := os.OpenFile(manifestPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create project manifest: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write project manifest: %w", err)
	}

	return nil
}
