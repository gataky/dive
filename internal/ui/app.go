package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/gataky/dive/internal/autocomplete"
	"github.com/gataky/dive/internal/export"
	"github.com/gataky/dive/internal/query"
	"github.com/gataky/dive/internal/render"
	"github.com/gataky/dive/internal/ui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const footerText = "[white::b]Tab[::-]: Complete | [white::b]F1[::-]: Help | [white::b]Ctrl+O[::-]: Output | [white::b]Ctrl+Y[::-]: Copy | [white::b]Ctrl+S[::-]: Save | [white::b]Ctrl+Q[::-]: Quit"

func init() {
	tview.Borders.HorizontalFocus = tview.Borders.Horizontal
	tview.Borders.VerticalFocus = tview.Borders.Vertical
	tview.Borders.TopLeftFocus = tview.Borders.TopLeft
	tview.Borders.TopRightFocus = tview.Borders.TopRight
	tview.Borders.BottomLeftFocus = tview.Borders.BottomLeft
	tview.Borders.BottomRightFocus = tview.Borders.BottomRight
}

// App wires the fixed layout to the query/suggest/render pipeline.
type App struct {
	tviewApp *tview.Application
	pages    *tview.Pages
	outer    *tview.Flex // main column + help panel
	theme    *theme.Theme

	inputField     *tview.InputField
	suggestionsBar *tview.TextView
	outputPanel    *tview.TextView
	footer         *tview.TextView
	helpPanel      *tview.TextView
	saveInput      *tview.InputField

	data        string                    // the JSON document
	suggestions []autocomplete.Suggestion // candidates for the current text
	cycle       *autocomplete.CycleState  // nil when not Tab-cycling
	settingText bool                      // guards cycle-driven SetText calls
	helpVisible bool
	plainOutput string // uncolored JSON currently displayed, for copy/save

	stopped  chan struct{} // closed by Stop(); unblocks pending flash timers
	stopOnce sync.Once
}

// NewApp creates the application for the given JSON document.
func NewApp(jsonData string) *App {
	a := &App{
		tviewApp: tview.NewApplication(),
		theme:    theme.DefaultTheme(),
		data:     jsonData,
		stopped:  make(chan struct{}),
	}

	a.inputField = createInputField(a.theme)
	a.suggestionsBar = createSuggestionsBar(a.theme)
	a.outputPanel = createOutputPanel(a.theme)
	a.footer = createFooter(a.theme)
	a.helpPanel = createHelpPanel(a.theme)
	a.saveInput = createSaveInput(a.theme)

	a.buildLayout()
	a.bindKeys()

	a.refreshSuggestions()
	a.refreshOutput()

	// The suggestions bar's real width isn't known until it has been
	// through a layout pass, which only happens once Run() performs its
	// first draw against the actual screen; until then GetInnerRect
	// returns tview's pre-layout default box size, not the terminal
	// width. Queue a one-shot redraw for right after that first draw so
	// the initial suggestions list reflects the real terminal width
	// instead of being spuriously truncated. Note this goroutine blocks
	// until Run() services the update queue, so constructing an App
	// without ever calling Run() leaks it; acceptable for this CLI's
	// NewApp→Run lifecycle.
	go a.tviewApp.QueueUpdateDraw(a.drawSuggestions)

	return a
}

// buildLayout assembles the fixed layout. Nothing is ever added to or
// removed from it afterwards: the help panel hides via zero-width resize
// and the save dialog is a Pages overlay.
func (a *App) buildLayout() {
	main := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.inputField, 3, 0, true).
		AddItem(a.suggestionsBar, 1, 0, false).
		AddItem(a.outputPanel, 0, 1, false).
		AddItem(a.footer, 1, 0, false)

	a.outer = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(main, 0, 1, true).
		AddItem(a.helpPanel, 0, 0, false) // hidden until F1

	saveOverlay := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(a.saveInput, 60, 0, true).
			AddItem(nil, 0, 1, false), 3, 0, true).
		AddItem(nil, 0, 1, false)

	a.pages = tview.NewPages().
		AddPage("main", a.outer, true, true).
		AddPage("save", saveOverlay, true, false)

	a.tviewApp.SetRoot(a.pages, true)
	a.tviewApp.SetFocus(a.inputField)
}

// onTextChanged handles user edits (not cycle-driven SetText calls):
// any real edit ends the cycle and recomputes suggestions and output.
func (a *App) onTextChanged(text string) {
	if a.settingText {
		return
	}
	a.cycle = nil
	a.refreshSuggestions()
	a.refreshOutput()
}

// refreshSuggestions recomputes candidates for the current text.
func (a *App) refreshSuggestions() {
	a.suggestions = autocomplete.Suggest(a.data, a.inputField.GetText())
	a.drawSuggestions()
}

// drawSuggestions redraws the bar from current candidates + cycle state.
func (a *App) drawSuggestions() {
	selected := -1
	if a.cycle != nil {
		selected = a.cycle.Selected()
	}
	_, _, width, _ := a.suggestionsBar.GetInnerRect()
	if width <= 0 {
		width = 80 // before first draw
	}
	a.suggestionsBar.SetText(formatSuggestions(a.suggestions, selected, width, a.theme))
}

// refreshOutput resolves the current path to its deepest valid ancestor
// and renders it. The output panel title names what is shown; the input
// border goes red when part of the path didn't match.
func (a *App) refreshOutput() {
	res := query.Resolve(a.data, a.inputField.GetText())
	a.plainOutput = render.Pretty(res.Result)
	a.outputPanel.SetText(render.Colorize(res.Result, a.theme))
	a.outputPanel.ScrollToBeginning()

	shown := res.ResolvedPath
	if shown == "" {
		shown = "(root)"
	}
	if res.UnmatchedSuffix == "" {
		a.outputPanel.SetTitle(" " + render.EscapeContent(shown) + " ")
		a.inputField.SetBorderColor(a.theme.BorderFocused)
	} else {
		a.outputPanel.SetTitle(fmt.Sprintf(" %s ✗ %s ", render.EscapeContent(shown), render.EscapeContent(res.UnmatchedSuffix)))
		a.inputField.SetBorderColor(a.theme.BorderInvalid)
	}
}

// setTextFromCycle updates the input without resetting the cycle.
func (a *App) setTextFromCycle(text string) {
	a.settingText = true
	a.inputField.SetText(text)
	a.settingText = false
}

// cycleSuggestion advances (or rewinds) Tab-cycling. The originally
// typed text is position 0 of the cycle.
func (a *App) cycleSuggestion(forward bool) {
	if a.cycle == nil {
		if len(a.suggestions) == 0 {
			return
		}
		a.cycle = autocomplete.NewCycleState(a.inputField.GetText(), a.suggestions)
	}
	var text string
	if forward {
		text = a.cycle.Next()
	} else {
		text = a.cycle.Prev()
	}
	a.setTextFromCycle(text)
	a.drawSuggestions()
	a.refreshOutput()
}

func (a *App) bindKeys() {
	a.inputField.SetChangedFunc(a.onTextChanged)

	a.inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyBacktab:
			a.cycleSuggestion(event.Key() == tcell.KeyTab)
			return nil
		case tcell.KeyEscape:
			if a.cycle != nil {
				a.setTextFromCycle(a.cycle.Base())
				a.cycle = nil
				a.drawSuggestions()
				a.refreshOutput()
			}
			return nil
		case tcell.KeyLeft, tcell.KeyRight, tcell.KeyHome, tcell.KeyEnd:
			if a.cycle != nil {
				a.cycle = nil
				a.refreshSuggestions()
			}
			return event
		}
		return event
	})

	a.outputPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape ||
			(event.Key() == tcell.KeyRune && (event.Rune() == 'i' || event.Rune() == 'I')) {
			a.tviewApp.SetFocus(a.inputField)
			return nil
		}
		return event
	})
	a.outputPanel.SetFocusFunc(func() { a.outputPanel.SetBorderColor(a.theme.BorderFocused) })
	a.outputPanel.SetBlurFunc(func() { a.outputPanel.SetBorderColor(a.theme.BorderUnfocused) })

	a.helpPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.toggleHelp()
			return nil
		}
		return event
	})

	a.saveInput.SetDoneFunc(func(key tcell.Key) {
		// tview fires the done func for Tab/Backtab too; those must not
		// dismiss the dialog.
		if key == tcell.KeyTab || key == tcell.KeyBacktab {
			return
		}
		if key == tcell.KeyEnter {
			if name := a.saveInput.GetText(); name != "" {
				if err := export.SaveToFile(a.plainOutput, name); err != nil {
					a.flashMessage(fmt.Sprintf("Error: %v", err), true)
				} else {
					a.flashMessage("Saved to "+name, false)
				}
			}
		}
		a.pages.HidePage("save")
		a.tviewApp.SetFocus(a.inputField)
	})

	a.tviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Quit always works, even with the save dialog up.
		switch event.Key() {
		case tcell.KeyCtrlQ, tcell.KeyCtrlC:
			a.Stop()
			return nil
		}

		// While the save dialog is up, the other global hotkeys must not
		// fire (Ctrl+O/F1 would move focus or resize panels underneath
		// the still-visible dialog, leaving it stuck with nothing focused
		// to dismiss it). Pass events through so only the save input's
		// own handling applies.
		if name, _ := a.pages.GetFrontPage(); name == "save" {
			return event
		}

		switch event.Key() {
		case tcell.KeyCtrlY:
			a.copyOutput()
			return nil
		case tcell.KeyCtrlS:
			a.pages.ShowPage("save")
			a.tviewApp.SetFocus(a.saveInput)
			return nil
		case tcell.KeyF1:
			a.toggleHelp()
			return nil
		case tcell.KeyCtrlO:
			a.tviewApp.SetFocus(a.outputPanel)
			return nil
		}
		return event
	})
}

// toggleHelp shows/hides the help panel by resizing it, never rebuilding.
func (a *App) toggleHelp() {
	a.helpVisible = !a.helpVisible
	if a.helpVisible {
		a.outer.ResizeItem(a.helpPanel, 0, 1)
		a.tviewApp.SetFocus(a.helpPanel)
	} else {
		a.outer.ResizeItem(a.helpPanel, 0, 0)
		a.helpPanel.ScrollToBeginning()
		a.tviewApp.SetFocus(a.inputField)
	}
}

// copyOutput copies the plain (uncolored) JSON to the clipboard.
func (a *App) copyOutput() {
	if err := export.CopyToClipboard(a.plainOutput); err != nil {
		a.flashMessage(fmt.Sprintf("Error: %v", err), true)
	} else {
		a.flashMessage("Copied to clipboard!", false)
	}
}

// flashMessage shows a temporary footer message for three seconds.
func (a *App) flashMessage(msg string, isErr bool) {
	color := "green"
	if isErr {
		color = "red"
	}
	a.footer.SetText(fmt.Sprintf("[%s]%s[-]", color, render.EscapeContent(msg)))
	go func() {
		select {
		case <-time.After(3 * time.Second):
		case <-a.stopped:
			return // app stopped; QueueUpdateDraw would block forever
		}
		// Re-check after the timer: Stop may have raced it. A residual
		// microsecond window remains (Stop between this check and the
		// queue send); process exit reclaims the goroutine in that case.
		select {
		case <-a.stopped:
			return
		default:
		}
		a.tviewApp.QueueUpdateDraw(func() {
			a.footer.SetText(footerText)
		})
	}()
}

// Run starts the application.
func (a *App) Run() error {
	return a.tviewApp.Run()
}

// Stop stops the application and releases any pending flash timers.
func (a *App) Stop() {
	a.stopOnce.Do(func() { close(a.stopped) })
	a.tviewApp.Stop()
}
