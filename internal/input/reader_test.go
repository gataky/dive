package input

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	t.Run("valid JSON file", func(t *testing.T) {
		validJSON := `{"name": "test", "value": 123}`
		filePath := filepath.Join(tempDir, "valid.json")
		err := os.WriteFile(filePath, []byte(validJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result, err := ReadFromFile(filePath)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if result != validJSON {
			t.Errorf("Expected %s, got %s", validJSON, result)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "does-not-exist.json")
		_, err := ReadFromFile(filePath)
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("Expected error message to contain 'does not exist', got: %v", err)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "empty.json")
		err := os.WriteFile(filePath, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		_, err = ReadFromFile(filePath)
		if err == nil {
			t.Error("Expected error for empty file, got nil")
		}
		if !strings.Contains(err.Error(), "empty") {
			t.Errorf("Expected error message to contain 'empty', got: %v", err)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		invalidJSON := `{"name": "test", invalid}`
		filePath := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(filePath, []byte(invalidJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		_, err = ReadFromFile(filePath)
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
		if !strings.Contains(err.Error(), "invalid JSON") {
			t.Errorf("Expected error message to contain 'invalid JSON', got: %v", err)
		}
	})

	t.Run("complex valid JSON", func(t *testing.T) {
		complexJSON := `{
			"users": [
				{"id": 1, "name": "Alice"},
				{"id": 2, "name": "Bob"}
			],
			"metadata": {
				"count": 2,
				"active": true
			}
		}`
		filePath := filepath.Join(tempDir, "complex.json")
		err := os.WriteFile(filePath, []byte(complexJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result, err := ReadFromFile(filePath)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if result != complexJSON {
			t.Errorf("Expected %s, got %s", complexJSON, result)
		}
	})

	t.Run("JSON array", func(t *testing.T) {
		arrayJSON := `[1, 2, 3, 4, 5]`
		filePath := filepath.Join(tempDir, "array.json")
		err := os.WriteFile(filePath, []byte(arrayJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result, err := ReadFromFile(filePath)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if result != arrayJSON {
			t.Errorf("Expected %s, got %s", arrayJSON, result)
		}
	})
}

func TestReadFromStdin(t *testing.T) {
	t.Run("valid JSON from stdin", func(t *testing.T) {
		validJSON := `{"name": "test", "value": 123}`

		// Create a temporary file to simulate stdin
		tempFile, err := os.CreateTemp("", "stdin-test-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		_, err = tempFile.WriteString(validJSON)
		if err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}

		// Seek back to the beginning
		_, err = tempFile.Seek(0, 0)
		if err != nil {
			t.Fatalf("Failed to seek temp file: %v", err)
		}

		// Save original stdin and restore after test
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		os.Stdin = tempFile

		result, err := ReadFromStdin()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if result != validJSON {
			t.Errorf("Expected %s, got %s", validJSON, result)
		}
	})

	t.Run("invalid JSON from stdin", func(t *testing.T) {
		invalidJSON := `{"name": "test", invalid}`

		tempFile, err := os.CreateTemp("", "stdin-test-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		_, err = tempFile.WriteString(invalidJSON)
		if err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}

		_, err = tempFile.Seek(0, 0)
		if err != nil {
			t.Fatalf("Failed to seek temp file: %v", err)
		}

		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		os.Stdin = tempFile

		_, err = ReadFromStdin()
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
		if !strings.Contains(err.Error(), "invalid JSON") {
			t.Errorf("Expected error message to contain 'invalid JSON', got: %v", err)
		}
	})

	t.Run("empty stdin", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "stdin-test-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		os.Stdin = tempFile

		_, err = ReadFromStdin()
		if err == nil {
			t.Error("Expected error for empty stdin, got nil")
		}
		if !strings.Contains(err.Error(), "no data") {
			t.Errorf("Expected error message to contain 'no data', got: %v", err)
		}
	})
}
