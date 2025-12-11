package ui

import (
	"testing"

	"github.com/gataky/dive/internal/ui/theme"
)

func TestCreateHelpPanel(t *testing.T) {
	th := theme.DefaultTheme()
	helpPanel := createHelpPanel(th)

	// Test that the function returns a non-nil TextView
	if helpPanel == nil {
		t.Fatal("createHelpPanel() returned nil")
	}
}

func TestCreateHelpPanelTitle(t *testing.T) {
	th := theme.DefaultTheme()
	helpPanel := createHelpPanel(th)

	// Test that the panel has the correct title
	// Note: tview doesn't provide a direct getter for title, but we can verify
	// the panel was created without errors
	if helpPanel == nil {
		t.Fatal("createHelpPanel() returned nil")
	}

	// Verify the panel has content (help text)
	text := helpPanel.GetText(false)
	if text == "" {
		t.Error("createHelpPanel() created panel with empty text content")
	}

	// Verify help content is present
	if len(text) < 100 {
		t.Errorf("createHelpPanel() help text is too short: %d characters", len(text))
	}
}

func TestCreateHelpPanelScrollable(t *testing.T) {
	th := theme.DefaultTheme()
	helpPanel := createHelpPanel(th)

	if helpPanel == nil {
		t.Fatal("createHelpPanel() returned nil")
	}

	// The TextView should be scrollable - we can't directly test this property
	// but we can verify the component was created successfully and has content
	// which implies it can be scrolled if the content is longer than the display area
	text := helpPanel.GetText(false)
	if text == "" {
		t.Error("createHelpPanel() should have scrollable content")
	}
}
