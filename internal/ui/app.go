package ui

import (
	"fmt"
	"time"

	"github.com/gataky/dive/internal/autocomplete"
	"github.com/gataky/dive/internal/export"
	"github.com/gataky/dive/internal/query"
	"github.com/gataky/dive/internal/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FocusableComponent represents which component currently has focus
type FocusableComponent int

const (
	FocusNone FocusableComponent = iota
	FocusInputField
	FocusDropdown
	FocusOutputPanel
)

func init() {
	tview.Borders.HorizontalFocus = tview.Borders.Horizontal
	tview.Borders.VerticalFocus= tview.Borders.Vertical
	tview.Borders.TopLeftFocus = tview.Borders.TopLeft
	tview.Borders.TopRightFocus = tview.Borders.TopRight
	tview.Borders.BottomLeftFocus = tview.Borders.BottomLeft
	tview.Borders.BottomRightFocus = tview.Borders.BottomRight
}


// App represents the main application UI
type App struct {
	tviewApp             *tview.Application
	layout               *tview.Flex
	inputField           *tview.InputField
	outputPanel          *tview.TextView
	footer               *tview.TextView
	autocompleteDropdown *tview.List
	saveModal            *tview.InputField
	theme                *theme.Theme
	focusedComponent     FocusableComponent
	jsonData             string
	currentQuery         string
	queryEngine          *query.Engine
	dropdownVisible      bool
	originalFooterText   string
}

// NewApp creates and initializes a new tview application with all UI components
func NewApp(jsonData string) *App {
	app := &App{
		tviewApp:           tview.NewApplication(),
		theme:              theme.DefaultTheme(),
		jsonData:           jsonData,
		queryEngine:        query.NewEngine(jsonData),
		originalFooterText: "[white::b]Tab[::-]: Autocomplete | [white::b]Ctrl+C[::-]: Copy | [white::b]Ctrl+S[::-]: Save | [white::b]Ctrl+Q[::-]: Quit",
	}

	app.initComponents()
	app.setupLayout()
	app.setupKeyBindings()
	app.setupInputFieldKeyBindings()
	app.setupQueryCallbacks()
	app.setupFocusHandlers()

	// Set the json data so it shows up on startup
	result := app.queryEngine.Query(jsonData)
	app.outputPanel.SetText(result.Value)

	// Set initial focus state to match the initially focused component
	app.focusedComponent = FocusInputField
	app.inputField.SetBorderColor(app.theme.BorderFocused)

	return app
}

// initComponents initializes all UI components
func (a *App) initComponents() {
	a.inputField = createInputField(a.theme)
	a.outputPanel = createOutputPanel(a.theme)
	a.footer = createFooter(a.theme)
	a.autocompleteDropdown = createAutocompleteDropdown(a.theme)
}

// setupLayout arranges all components in a vertical flex layout
func (a *App) setupLayout() {
	a.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.inputField, 3, 0, true).
		AddItem(a.outputPanel, 0, 1, false).
		AddItem(a.footer, 1, 0, false)

	a.tviewApp.SetRoot(a.layout, true)

	// Set initial focus to the input field
	a.tviewApp.SetFocus(a.inputField)
}

// showDropdown displays the autocomplete dropdown below the input field
func (a *App) showDropdown(suggestions []string) {
	if len(suggestions) == 0 {
		a.hideDropdown()
		return
	}

	// Clear existing items
	a.autocompleteDropdown.Clear()

	// Add suggestions to the dropdown
	for _, suggestion := range suggestions {
		// Capture suggestion in the closure
		s := suggestion
		a.autocompleteDropdown.AddItem(s, "", 0, func() {
			// Handle Enter key to select a suggestion (task 5.11)
			a.selectSuggestion(s)
		})
	}

	// Set up input capture for the dropdown to handle Escape and navigation
	a.autocompleteDropdown.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentItem := a.autocompleteDropdown.GetCurrentItem()
		itemCount := a.autocompleteDropdown.GetItemCount()

		switch event.Key() {
		case tcell.KeyEscape:
			// Hide dropdown and return focus to input field
			a.hideDropdown()
			return nil
		case tcell.KeyTab:
			// Move to next item, wrap to top if at bottom
			if currentItem < itemCount-1 {
				a.autocompleteDropdown.SetCurrentItem(currentItem + 1)
			} else {
				a.autocompleteDropdown.SetCurrentItem(0)
			}
			return nil
		case tcell.KeyBacktab:
			// Move to previous item (Shift+Tab), wrap to bottom if at top
			if currentItem > 0 {
				a.autocompleteDropdown.SetCurrentItem(currentItem - 1)
			} else {
				a.autocompleteDropdown.SetCurrentItem(itemCount - 1)
			}
			return nil
		case tcell.KeyUp:
			// If at the top of the list, return to input field
			if currentItem == 0 {
				a.tviewApp.SetFocus(a.inputField)
				return nil
			}
		}
		return event
	})

	// Update layout to show dropdown if not already visible
	if !a.dropdownVisible {
		a.dropdownVisible = true
		// Rebuild layout with dropdown
		a.layout.Clear()
		a.layout.AddItem(a.inputField, 3, 0, true)
		a.layout.AddItem(a.autocompleteDropdown, 8, 0, false) // Show dropdown with height of 8
		a.layout.AddItem(a.outputPanel, 0, 1, false)
		a.layout.AddItem(a.footer, 1, 0, false)
	}
}

// selectSuggestion updates the input field with the selected suggestion
func (a *App) selectSuggestion(suggestion string) {
	a.inputField.SetText(suggestion)
	a.hideDropdown()
}

// hideDropdown hides the autocomplete dropdown
func (a *App) hideDropdown() {
	if !a.dropdownVisible {
		return
	}

	a.dropdownVisible = false
	// Rebuild layout without dropdown
	a.layout.Clear()
	a.layout.AddItem(a.inputField, 3, 0, true)
	a.layout.AddItem(a.outputPanel, 0, 1, false)
	a.layout.AddItem(a.footer, 1, 0, false)

	// Restore focus to input field
	a.tviewApp.SetFocus(a.inputField)
}

// updateSuggestions gets autocomplete suggestions for the current path and shows dropdown
func (a *App) updateSuggestions() {
	currentPath := a.inputField.GetText()
	suggestions := autocomplete.GetSuggestions(a.jsonData, currentPath)
	a.showDropdown(suggestions)
}

// setupKeyBindings configures global key bindings
func (a *App) setupKeyBindings() {
	a.tviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			// Quit the application
			a.tviewApp.Stop()
			return nil
		case tcell.KeyCtrlC:
			// Copy current output to clipboard (task 6.7)
			a.copyToClipboard()
			return nil
		case tcell.KeyCtrlS:
			// Open save dialog (task 6.8)
			a.showSaveDialog()
			return nil
		}
		return event
	})
}

// setupInputFieldKeyBindings configures key bindings for the input field
func (a *App) setupInputFieldKeyBindings() {
	a.inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			// Handle Tab key press to show autocomplete dropdown (task 5.9)
			a.updateSuggestions()
			// Automatically focus on the dropdown if it has suggestions
			if a.dropdownVisible {
				a.tviewApp.SetFocus(a.autocompleteDropdown)
			}
			return nil // Consume the Tab key event
		case tcell.KeyEscape:
			// Hide dropdown when Escape is pressed
			a.hideDropdown()
			return nil
		case tcell.KeyDown:
			// Handle arrow keys to navigate dropdown selections (task 5.10)
			if a.dropdownVisible {
				a.tviewApp.SetFocus(a.autocompleteDropdown)
				return nil
			}
		case tcell.KeyUp:
			// Handle arrow keys to navigate dropdown selections (task 5.10)
			if a.dropdownVisible {
				a.tviewApp.SetFocus(a.autocompleteDropdown)
				return nil
			}
		}
		return event
	})
}

// Run starts the tview application
func (a *App) Run() error {
	return a.tviewApp.Run()
}

// Stop stops the tview application
func (a *App) Stop() {
	a.tviewApp.Stop()
}

// GetApplication returns the underlying tview application
func (a *App) GetApplication() *tview.Application {
	return a.tviewApp
}

// setupQueryCallbacks wires up the input field to call the query engine on each keystroke
func (a *App) setupQueryCallbacks() {
	a.inputField.SetChangedFunc(func(text string) {
		// Store the current query
		a.currentQuery = text

		// Call the query engine with the current path
		result := a.queryEngine.Query(text)

		// Update output panel with query results in real-time (task 4.8)
		a.outputPanel.SetText(result.Value)

		// Implement visual feedback for invalid paths (task 4.9 & 4.10)
		if result.IsValid {
			// Restore normal color when path becomes valid (task 4.10)
			a.inputField.SetBorderColor(a.theme.BorderValid)
		} else {
			// Change border color to red when path is invalid (task 4.9)
			a.inputField.SetBorderColor(a.theme.BorderInvalid)
		}
	})
}

// setupFocusHandlers wires up focus change handlers for all focusable components
func (a *App) setupFocusHandlers() {
	// Input field focus handler
	a.inputField.SetFocusFunc(func() {
		a.setComponentFocus(FocusInputField)
	})

	// Autocomplete dropdown focus handler
	a.autocompleteDropdown.SetFocusFunc(func() {
		a.setComponentFocus(FocusDropdown)
	})

	// Output panel focus handler (for scrolling)
	a.outputPanel.SetFocusFunc(func() {
		a.setComponentFocus(FocusOutputPanel)
	})
}

// showMessage displays a temporary message in the footer (task 6.9)
func (a *App) showMessage(message string, isError bool) {
	// Use theme colors: "green" matches theme.ColorSuccess, "red" matches theme.ColorError
	// tview's text markup requires string color names, not tcell.Color types
	color := a.theme.ColorSuccess
	if isError {
		color = a.theme.ColorError
	}
	a.footer.SetText(fmt.Sprintf("[%s]%s[-]", color, message))

	// Restore original footer text after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)
		a.tviewApp.QueueUpdateDraw(func() {
			a.footer.SetText(a.originalFooterText)
		})
	}()
}

// copyToClipboard copies the current output to the clipboard (task 6.7)
func (a *App) copyToClipboard() {
	content := a.outputPanel.GetText(false)
	err := export.CopyToClipboard(content)
	if err != nil {
		a.showMessage(fmt.Sprintf("Error: %v", err), true)
	} else {
		a.showMessage("Copied to clipboard!", false)
	}
}

// showSaveDialog displays a modal dialog to prompt for filename (task 6.6 & 6.8)
func (a *App) showSaveDialog() {
	// Create a modal input field for filename
	modal := tview.NewInputField().
		SetLabel("Save to file: ").
		SetFieldWidth(40).
		SetText("output.json")

	modal.SetBorder(true).
		SetTitle(" Save Output ").
		SetBorderColor(a.theme.BorderFocused)

	// Create a frame to center the modal
	frame := tview.NewFrame(modal).
		SetBorders(2, 2, 2, 2, 4, 4)

	// Handle input
	modal.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			filename := modal.GetText()
			if filename != "" {
				content := a.outputPanel.GetText(false)
				err := export.SaveToFile(content, filename)
				if err != nil {
					a.showMessage(fmt.Sprintf("Error: %v", err), true)
				} else {
					a.showMessage(fmt.Sprintf("Saved to %s", filename), false)
				}
			}
			// Restore original layout
			a.restoreLayout()
		} else if key == tcell.KeyEscape {
			// Cancel and restore layout
			a.restoreLayout()
		}
	})

	// Show the modal
	a.tviewApp.SetRoot(frame, true)
	a.tviewApp.SetFocus(modal)
}

// restoreLayout restores the main application layout
func (a *App) restoreLayout() {
	if a.dropdownVisible {
		// Rebuild layout with dropdown
		a.layout.Clear()
		a.layout.AddItem(a.inputField, 3, 0, true)
		a.layout.AddItem(a.autocompleteDropdown, 8, 0, false)
		a.layout.AddItem(a.outputPanel, 0, 1, false)
		a.layout.AddItem(a.footer, 1, 0, false)
	} else {
		// Rebuild layout without dropdown
		a.layout.Clear()
		a.layout.AddItem(a.inputField, 3, 0, true)
		a.layout.AddItem(a.outputPanel, 0, 1, false)
		a.layout.AddItem(a.footer, 1, 0, false)
	}
	a.tviewApp.SetRoot(a.layout, true)
	a.tviewApp.SetFocus(a.inputField)
}

// setComponentFocus updates border colors when focus changes between components
func (a *App) setComponentFocus(newFocus FocusableComponent) {
	// If focus hasn't changed, nothing to do
	if a.focusedComponent == newFocus {
		return
	}

	// Remove focus styling from previously focused component
	switch a.focusedComponent {
	case FocusInputField:
		// Input field might have validation state, only change if no validation
		// Let validation states be handled by setupQueryCallbacks
		a.inputField.SetBorderColor(a.theme.BorderUnfocused)
	case FocusDropdown:
		a.autocompleteDropdown.SetBorderColor(a.theme.BorderUnfocused)
	case FocusOutputPanel:
		a.outputPanel.SetBorderColor(a.theme.BorderUnfocused)
	}

	// Apply focus styling to newly focused component
	switch newFocus {
	case FocusInputField:
		a.inputField.SetBorderColor(a.theme.BorderFocused)
	case FocusDropdown:
		a.autocompleteDropdown.SetBorderColor(a.theme.BorderFocused)
	case FocusOutputPanel:
		a.outputPanel.SetBorderColor(a.theme.BorderFocused)
	}

	// Update tracked focus
	a.focusedComponent = newFocus
}
