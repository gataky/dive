package ui

import (
	"github.com/gataky/dive/internal/autocomplete"
	"github.com/gataky/dive/internal/query"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the main application UI
type App struct {
	tviewApp            *tview.Application
	layout              *tview.Flex
	header              *tview.TextView
	inputField          *tview.InputField
	outputPanel         *tview.TextView
	footer              *tview.TextView
	autocompleteDropdown *tview.List
	jsonData            string
	currentQuery        string
	queryEngine         *query.Engine
	dropdownVisible     bool
}

// NewApp creates and initializes a new tview application with all UI components
func NewApp(jsonData string) *App {
	app := &App{
		tviewApp:    tview.NewApplication(),
		jsonData:    jsonData,
		queryEngine: query.NewEngine(jsonData),
	}

	app.initComponents()
	app.setupLayout()
	app.setupKeyBindings()
	app.setupInputFieldKeyBindings()
	app.setupQueryCallbacks()

	return app
}

// initComponents initializes all UI components
func (a *App) initComponents() {
	a.header = createHeader()
	a.inputField = createInputField()
	a.outputPanel = createOutputPanel()
	a.footer = createFooter()
	a.autocompleteDropdown = createAutocompleteDropdown()
}

// setupLayout arranges all components in a vertical flex layout
func (a *App) setupLayout() {
	a.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.header, 3, 0, false).
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
		switch event.Key() {
		case tcell.KeyEscape:
			// Hide dropdown and return focus to input field
			a.hideDropdown()
			return nil
		case tcell.KeyUp:
			// If at the top of the list, return to input field
			if a.autocompleteDropdown.GetCurrentItem() == 0 {
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
		a.layout.AddItem(a.header, 3, 0, false)
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
	a.layout.AddItem(a.header, 3, 0, false)
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
			// Copy to clipboard (to be implemented)
			// TODO: Wire up clipboard copy functionality
			return nil
		case tcell.KeyCtrlS:
			// Save to file (to be implemented)
			// TODO: Wire up save to file functionality
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
			a.inputField.SetBorderColor(tcell.ColorGreen)
		} else {
			// Change border color to red when path is invalid (task 4.9)
			a.inputField.SetBorderColor(tcell.ColorRed)
		}
	})
}
