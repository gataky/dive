package ui

import (
	"github.com/gataky/dive/internal/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// createHeader creates the header component displaying the app name and instructions
func createHeader(th *theme.Theme) *tview.TextView {
	header := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true).
		SetText("[yellow::b]dive - Interactive JSON Viewer[::-]\n[gray]Navigate JSON with gjson paths | Press Tab for autocomplete[-]")

	header.SetBorder(true).
		SetBorderColor(th.BorderUnfocused).
		SetBackgroundColor(th.Background)

	return header
}

// createInputField creates the path input field component
func createInputField(th *theme.Theme) *tview.InputField {
	style := tcell.Style{}
	style.Background(th.FieldBackground)

	inputField := tview.NewInputField().
		SetLabel("> ").
		SetFieldWidth(0). // Use all available width
		SetPlaceholder("Enter gjson path (e.g., users.0.name)").
		SetPlaceholderStyle(style)

	inputField.SetBorder(true).
		SetTitle(" Query ").
		SetBorderColor(th.BorderUnfocused).
		SetBackgroundColor(th.Background)

	return inputField
}

// createOutputPanel creates the output panel component for displaying query results
func createOutputPanel(th *theme.Theme) *tview.TextView {
	outputPanel := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			// Auto-scroll to the end when content changes
			// This can be disabled if we want to maintain scroll position
		})

	outputPanel.SetBorder(true).
		SetTitle(" Output ").
		SetBorderColor(th.BorderUnfocused).
		SetBackgroundColor(th.Background)

	// Set initial message
	outputPanel.SetText("[gray]Enter a gjson path to query the JSON data...[-]")

	return outputPanel
}

// createFooter creates the footer component showing keybindings
func createFooter(th *theme.Theme) *tview.TextView {
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[white::b]Tab[::-]: Autocomplete | [white::b]Ctrl+C[::-]: Copy | [white::b]Ctrl+S[::-]: Save | [white::b]Ctrl+Q[::-]: Quit")

	footer.SetBackgroundColor(th.Background)

	return footer
}

// createAutocompleteDropdown creates the autocomplete dropdown using tview.List
func createAutocompleteDropdown(th *theme.Theme) *tview.List {
	dropdown := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)

	dropdown.SetBorder(true).
		SetTitle(" Suggestions ").
		SetBorderColor(th.BorderUnfocused).
		SetBackgroundColor(th.Background)

	return dropdown
}
