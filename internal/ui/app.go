package ui

import (
	"github.com/gataky/dive/internal/query"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the main application UI
type App struct {
	tviewApp     *tview.Application
	layout       *tview.Flex
	header       *tview.TextView
	inputField   *tview.InputField
	outputPanel  *tview.TextView
	footer       *tview.TextView
	jsonData     string
	currentQuery string
	queryEngine  *query.Engine
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
	app.setupQueryCallbacks()

	return app
}

// initComponents initializes all UI components
func (a *App) initComponents() {
	a.header = createHeader()
	a.inputField = createInputField()
	a.outputPanel = createOutputPanel()
	a.footer = createFooter()
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
