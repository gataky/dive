package ui

import (
	"github.com/gataky/dive/internal/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// createInputField creates the path input field component
func createInputField(th *theme.Theme) *tview.InputField {
	style := (tcell.Style{}).
	Background(th.Background)

	inputField := tview.NewInputField().
		SetFieldWidth(0). // Use all available width
		SetPlaceholder("Enter gjson path (e.g., users.0.name)").
		SetFieldBackgroundColor(th.FieldBackground).
		SetPlaceholderStyle(style)

	inputField.SetBorder(true).
		SetBorderColor(th.BorderUnfocused).
		SetBackgroundColor(th.Background)

	return inputField
}

// createOutputPanel creates the output panel component for displaying query results
func createOutputPanel(th *theme.Theme) *tview.TextView {
	style := (tcell.Style{}).
	Background(th.Background)

	outputPanel := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetTextColor(th.TextDefault).
		SetChangedFunc(func() {
			// Auto-scroll to the end when content changes
			// This can be disabled if we want to maintain scroll position
		}).
		SetTextStyle(style)

	outputPanel.SetBorder(true).
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
		SetTextColor(th.TextDefault).
		SetText("[white::b]Tab[::-]: Autocomplete | [white::b]Ctrl+O[::-]: Focus Output | [white::b]Ctrl+C[::-]: Copy | [white::b]Ctrl+S[::-]: Save | [white::b]Ctrl+Q[::-]: Quit")

	footer.SetBackgroundColor(th.Background)

	return footer
}

// createAutocompleteDropdown creates the autocomplete dropdown using tview.List
func createAutocompleteDropdown(th *theme.Theme) *tview.List {
	dropdown := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetMainTextColor(th.TextDefault)

	dropdown.SetBorder(true).
		SetBorderColor(th.BorderUnfocused).
		SetBackgroundColor(th.Background)

	return dropdown
}

// createHelpPanel creates the help panel component for displaying gjson syntax help
func createHelpPanel(th *theme.Theme) *tview.TextView {
	style := (tcell.Style{}).
		Background(th.Background)

	helpPanel := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetTextColor(th.TextDefault).
		SetText(getHelpContent()).
		SetTextStyle(style)

	helpPanel.SetBorder(true).
		SetTitle(" gjson Syntax Help ").
		SetBorderColor(th.BorderFocused).
		SetBackgroundColor(th.Background)

	return helpPanel
}
