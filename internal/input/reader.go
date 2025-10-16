package input

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ReadFromFile reads JSON data from a file at the given path.
// It returns the raw JSON as a string and any error encountered.
// The JSON is validated before being returned.
func ReadFromFile(path string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", path)
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	// Check for empty file
	if len(data) == 0 {
		return "", fmt.Errorf("file is empty: %s", path)
	}

	// Validate JSON
	if !json.Valid(data) {
		return "", fmt.Errorf("invalid JSON in file: %s", path)
	}

	return string(data), nil
}

// ReadFromStdin reads JSON data from standard input.
// It returns the raw JSON as a string and any error encountered.
// The JSON is validated before being returned.
func ReadFromStdin() (string, error) {
	// Read all data from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("error reading from stdin: %w", err)
	}

	// Check for empty input
	if len(data) == 0 {
		return "", fmt.Errorf("no data received from stdin")
	}

	// Validate JSON
	if !json.Valid(data) {
		return "", fmt.Errorf("invalid JSON from stdin")
	}

	return string(data), nil
}
