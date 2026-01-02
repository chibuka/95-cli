package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chibuka/95-cli/client"
)

// ExecuteSetup performs setup operations before running a test
func ExecuteSetup(setup *client.TestSetup) error {
	if setup == nil {
		return nil
	}

	// Delete files first (ensure clean slate from previous runs)
	for _, file := range setup.DeleteFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file %s: %w", file, err)
		}
	}

	// Create directories
	for _, dir := range setup.CreateDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create files with content
	for path, content := range setup.CreateFiles {
		// Ensure parent directory exists
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", path, err)
		}

		// Write file
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", path, err)
		}
	}

	return nil
}

// ExecuteCleanup performs cleanup operations after running a test
func ExecuteCleanup(cleanup *client.TestCleanup) error {
	if cleanup == nil {
		return nil
	}

	var errors []error

	// Delete files
	for _, file := range cleanup.DeleteFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("failed to delete file %s: %w", file, err))
		}
	}

	// Delete directories (including all contents)
	for _, dir := range cleanup.DeleteDirs {
		if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("failed to delete directory %s: %w", dir, err))
		}
	}

	// Return combined errors if any
	if len(errors) > 0 {
		errMsg := "cleanup errors:"
		for _, err := range errors {
			errMsg += "\n  - " + err.Error()
		}
		return fmt.Errorf(errMsg)
	}

	return nil
}
