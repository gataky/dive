package export

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveToFileSuccess(t *testing.T) {
	// Create a temp directory
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")

	content := `{"name": "test", "value": 123}`

	err := SaveToFile(content, filePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File was not created")
	}

	// Read the file and verify content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != content {
		t.Errorf("Expected content %q, got %q", content, string(data))
	}
}

func TestSaveToFileEmptyContent(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty.json")

	err := SaveToFile("", filePath)
	if err == nil {
		t.Error("Expected error for empty content, got nil")
	}
}

func TestSaveToFileEmptyPath(t *testing.T) {
	content := `{"test": "data"}`

	err := SaveToFile(content, "")
	if err == nil {
		t.Error("Expected error for empty path, got nil")
	}
}

func TestSaveToFileCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "nested", "directory", "file.json")

	content := `{"created": true}`

	err := SaveToFile(content, nestedPath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("File was not created in nested directory")
	}

	// Read and verify content
	data, err := os.ReadFile(nestedPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != content {
		t.Errorf("Expected content %q, got %q", content, string(data))
	}
}

func TestSaveToFileOverwritesExisting(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "overwrite.json")

	// Create initial file
	initialContent := `{"version": 1}`
	err := SaveToFile(initialContent, filePath)
	if err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	// Overwrite with new content
	newContent := `{"version": 2}`
	err = SaveToFile(newContent, filePath)
	if err != nil {
		t.Fatalf("Failed to overwrite file: %v", err)
	}

	// Verify new content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != newContent {
		t.Errorf("Expected content %q, got %q", newContent, string(data))
	}
}

func TestSaveToFileRelativePath(t *testing.T) {
	tempDir := t.TempDir()
	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	content := `{"relative": true}`
	err := SaveToFile(content, "relative.json")
	if err != nil {
		t.Fatalf("Expected no error for relative path, got: %v", err)
	}

	// Verify file exists in temp directory
	absPath := filepath.Join(tempDir, "relative.json")
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Error("File was not created with relative path")
	}
}

func TestCopyToClipboardEmptyContent(t *testing.T) {
	err := CopyToClipboard("")
	if err == nil {
		t.Error("Expected error for empty content, got nil")
	}
}

func TestCopyToClipboardSuccess(t *testing.T) {
	// Note: This test may fail in CI environments without clipboard support
	// Skip if clipboard is not available
	content := `{"test": "clipboard"}`
	err := CopyToClipboard(content)

	// We just check that the function executes without panicking
	// The actual clipboard functionality depends on the OS and may not work in all environments
	if err != nil {
		t.Logf("Warning: clipboard test failed (may be expected in CI): %v", err)
	}
}
