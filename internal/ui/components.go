package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// createHeader creates the header component displaying the app name and instructions
func createHeader() *tview.TextView {
	header := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true).
		SetText("[yellow::b]dive - Interactive JSON Viewer[::-]\n[gray]Navigate JSON with gjson paths | Press Tab for autocomplete[-]")

	header.SetBorder(true).
		SetBorderColor(tcell.ColorBlue).
		SetBackgroundColor(tcell.ColorDefault)

	return header
}

// createInputField creates the path input field component
func createInputField() *tview.InputField {
	inputField := tview.NewInputField().
		SetLabel("> ").
		SetFieldWidth(0). // Use all available width
		SetPlaceholder("Enter gjson path (e.g., users.0.name)")

	inputField.SetBorder(true).
		SetTitle(" Query ").
		SetBorderColor(tcell.ColorGreen).
		SetBackgroundColor(tcell.ColorDefault)

	// TODO: Wire up real-time query updates
	// inputField.SetChangedFunc(func(text string) {
	//     // Call query engine and update output panel
	// })

	// TODO: Wire up autocomplete on Tab key
	// inputField.SetInputCapture(func(event *tview.EventKey) *tview.EventKey {
	//     if event.Key() == tview.KeyTab {
	//         // Show autocomplete dropdown
	//     }
	//     return event
	// })

	return inputField
}

// createOutputPanel creates the output panel component for displaying query results
func createOutputPanel() *tview.TextView {
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
		SetBorderColor(tcell.ColorAqua).
		SetBackgroundColor(tcell.ColorDefault)

	// Set initial message
	outputPanel.SetText("[gray]Enter a gjson path to query the JSON data...[-]")

	return outputPanel
}

// createFooter creates the footer component showing keybindings
func createFooter() *tview.TextView {
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[white::b]Tab[::-]: Autocomplete | [white::b]Ctrl+C[::-]: Copy | [white::b]Ctrl+S[::-]: Save | [white::b]Ctrl+Q[::-]: Quit")

	footer.SetBackgroundColor(tcell.ColorDefault)

	return footer
}

// createAutocompleteDropdown creates the autocomplete dropdown using tview.List
func createAutocompleteDropdown() *tview.List {
	dropdown := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)

	dropdown.SetBorder(true).
		SetTitle(" Suggestions ").
		SetBorderColor(tcell.ColorYellow).
		SetBackgroundColor(tcell.ColorDefault)

	return dropdown
}
