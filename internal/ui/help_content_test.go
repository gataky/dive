package ui

import (
	"strings"
	"testing"
)

func TestGetHelpContent(t *testing.T) {
	content := getHelpContent()

	// Test that content is not empty
	if content == "" {
		t.Error("getHelpContent() returned empty string")
	}

	// Test that content has reasonable length (at least 500 characters)
	if len(content) < 500 {
		t.Errorf("getHelpContent() returned content that is too short: %d characters", len(content))
	}
}

func TestGetHelpContentSections(t *testing.T) {
	content := getHelpContent()

	// Test that all three complexity level sections are present
	requiredSections := []string{
		"BEGINNER",
		"INTERMEDIATE",
		"ADVANCED",
	}

	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			t.Errorf("getHelpContent() missing required section: %s", section)
		}
	}
}

func TestGetHelpContentSyntaxExamples(t *testing.T) {
	content := getHelpContent()

	// Test that expected syntax examples are present
	expectedExamples := []string{
		"users.0",           // Array access
		"users.#",           // Array count
		"users.#.name",      // Wildcard query
		"users.#(age>21)#",  // Conditional query
		"@reverse",          // Modifier
		"{name,age}",        // Multi-path query
	}

	for _, example := range expectedExamples {
		if !strings.Contains(content, example) {
			t.Errorf("getHelpContent() missing expected syntax example: %s", example)
		}
	}
}
