package export

import (
	"fmt"
	"os"
	"path/filepath"
)

// SaveToFile writes the provided content to the specified file path
func SaveToFile(content string, filePath string) error {
	if content == "" {
		return fmt.Errorf("cannot save empty content to file")
	}

	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Get the absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to resolve file path: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the file with proper permissions (rw-r--r--)
	err = os.WriteFile(absPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
