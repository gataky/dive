package ui

import (
	"github.com/gataky/dive/internal/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// createInputField creates the query input field.
func createInputField(th *theme.Theme) *tview.InputField {
	style := (tcell.Style{}).Background(th.Background)

	inputField := tview.NewInputField().
		SetFieldWidth(0).
		SetPlaceholder("Enter gjson path (e.g., users.0.name)").
		SetFieldBackgroundColor(th.FieldBackground).
		SetPlaceholderStyle(style)

	inputField.SetBorder(true).
		SetTitle(" Query ").
		SetBorderColor(th.BorderFocused).
		SetBackgroundColor(th.Background)

	return inputField
}

// createSuggestionsBar creates the always-visible completions row.
func createSuggestionsBar(th *theme.Theme) *tview.TextView {
	bar := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetTextColor(th.TextDefault)

	bar.SetBackgroundColor(th.Background)

	return bar
}

// createOutputPanel creates the panel showing the resolved JSON value.
func createOutputPanel(th *theme.Theme) *tview.TextView {
	style := (tcell.Style{}).Background(th.Background)

	outputPanel := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false).
		SetTextColor(th.TextDefault).
		SetTextStyle(style)

	outputPanel.SetBorder(true).
		SetTitle(" (root) ").
		SetBorderColor(th.BorderUnfocused).
		SetBackgroundColor(th.Background)

	return outputPanel
}

// createFooter creates the key-hint footer.
func createFooter(th *theme.Theme) *tview.TextView {
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(th.TextDefault).
		SetText(footerText)

	footer.SetBackgroundColor(th.Background)

	return footer
}

// createHelpPanel creates the gjson syntax help panel.
func createHelpPanel(th *theme.Theme) *tview.TextView {
	style := (tcell.Style{}).Background(th.Background)

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

// createSaveInput creates the filename prompt shown by the save overlay.
func createSaveInput(th *theme.Theme) *tview.InputField {
	saveInput := tview.NewInputField().
		SetLabel("Save to file: ").
		SetFieldWidth(0).
		SetText("output.json")

	saveInput.SetBorder(true).
		SetTitle(" Save Output ").
		SetBorderColor(th.BorderFocused).
		SetBackgroundColor(th.Background)

	return saveInput
}
