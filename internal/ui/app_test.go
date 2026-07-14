package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// newTestApp builds an App wired to a SimulationScreen and starts it
// running in a background goroutine, returning the app and a cleanup
// function. Callers must invoke the cleanup (typically via defer) so
// the tview event loop is stopped before the test ends.
func newTestApp(t *testing.T, doc string) (*App, func()) {
	t.Helper()

	s := tcell.NewSimulationScreen("UTF-8")

	app := NewApp(doc)
	// tview.Application.SetScreen calls screen.Init() itself when no
	// screen is set yet (Run() hasn't started), which resets a
	// SimulationScreen's size to its 80x25 default. So the desired
	// size must be applied after SetScreen, not before.
	app.tviewApp.SetScreen(s)
	s.SetSize(100, 30)
	// SetSize alone doesn't trigger a redraw; queue a resize event so
	// tview's event loop re-lays-out against the new dimensions.
	app.tviewApp.QueueEvent(tcell.NewEventResize(100, 30))

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := app.Run(); err != nil {
			t.Errorf("app.Run() returned error: %v", err)
		}
	}()

	cleanup := func() {
		app.Stop()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("app did not stop within timeout")
		}
	}

	// Give the event loop a moment to lay out against the simulated
	// 100x30 screen so GetInnerRect() calls inside the app return the
	// real (wide) dimensions rather than the SimulationScreen's 80x25
	// default (which is in effect immediately after SetScreen, before
	// the queued resize event above is processed).
	waitFor(t, func() bool {
		return syncCall(app, func() bool {
			_, _, w, _ := app.suggestionsBar.GetInnerRect()
			return w > 90
		})
	})

	return app, cleanup
}

// syncCall runs f on the tview Application's own goroutine (the one
// running inside app.Run()) and returns its result. tview widgets are
// only safe to read or write from that goroutine while Run() is active;
// reading them directly from the test goroutine races with the
// concurrent Draw() calls the event loop performs. QueueUpdate blocks
// until the loop has executed f, so the result is always current.
func syncCall[T any](app *App, f func() T) T {
	var result T
	app.tviewApp.QueueUpdate(func() {
		result = f()
	})
	return result
}

// waitFor polls cond until it returns true or a short deadline elapses,
// failing the test if the deadline is hit. Event processing inside the
// tview application happens on its own goroutine, so assertions made
// right after posting an event must synchronize this way.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !cond() {
		t.Fatal("condition not met within deadline")
	}
}

// typeText posts one KeyRune event per rune of s.
func typeText(app *App, s string) {
	for _, r := range s {
		app.tviewApp.QueueEvent(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
}

func pressKey(app *App, key tcell.Key) {
	app.tviewApp.QueueEvent(tcell.NewEventKey(key, 0, tcell.ModNone))
}

// inputText, outputTitle, plainOutput, and suggestionsText are small
// syncCall wrappers used pervasively by the scenarios below to read
// app state without racing the running event loop.

func inputText(app *App) string {
	return syncCall(app, func() string { return app.inputField.GetText() })
}

func outputTitle(app *App) string {
	return syncCall(app, func() string { return app.outputPanel.GetTitle() })
}

func plainOutput(app *App) string {
	return syncCall(app, func() string { return app.plainOutput })
}

func suggestionsText(app *App) string {
	return syncCall(app, func() string { return app.suggestionsBar.GetText(true) })
}

func suggestionCount(app *App) int {
	return syncCall(app, func() int { return len(app.suggestions) })
}

func frontPage(app *App) string {
	return syncCall(app, func() string {
		name, _ := app.pages.GetFrontPage()
		return name
	})
}

// focusIs reports whether the application focus is exactly p.
func focusIs(app *App, p tview.Primitive) bool {
	return syncCall(app, func() bool { return app.tviewApp.GetFocus() == p })
}

func footerTextNow(app *App) string {
	return syncCall(app, func() string { return app.footer.GetText(false) })
}

const testDoc = `{"users":[{"name":"Ada","nationality":"US"}],"empty":{}}`

func TestApp_Startup(t *testing.T) {
	app, cleanup := newTestApp(t, testDoc)
	defer cleanup()

	waitFor(t, func() bool {
		return outputTitle(app) == " (root) "
	})

	if got := plainOutput(app); !strings.Contains(got, "users") {
		t.Errorf("expected plainOutput to contain %q, got %q", "users", got)
	}

	barText := suggestionsText(app)
	if !strings.Contains(barText, "users") || !strings.Contains(barText, "empty") {
		t.Errorf("expected suggestions bar to contain %q and %q, got %q", "users", "empty", barText)
	}
}

func TestApp_TypingShowsUnmatchedSuffix(t *testing.T) {
	app, cleanup := newTestApp(t, testDoc)
	defer cleanup()

	typeText(app, "users.0.na")

	waitFor(t, func() bool {
		return outputTitle(app) == " users.0 ✗ na "
	})

	barText := suggestionsText(app)
	if !strings.Contains(barText, "name") || !strings.Contains(barText, "nationality") {
		t.Errorf("expected suggestions bar to contain %q and %q, got %q", "name", "nationality", barText)
	}

	if got := plainOutput(app); !strings.Contains(got, "Ada") {
		t.Errorf("expected plainOutput to contain %q, got %q", "Ada", got)
	}
}

func TestApp_TabCyclingAndBacktab(t *testing.T) {
	app, cleanup := newTestApp(t, testDoc)
	defer cleanup()

	typeText(app, "users.0.na")
	waitFor(t, func() bool {
		return outputTitle(app) == " users.0 ✗ na "
	})

	// First Tab: "users.0.name"
	pressKey(app, tcell.KeyTab)
	waitFor(t, func() bool {
		return inputText(app) == "users.0.name"
	})
	if got := outputTitle(app); got != " users.0.name " {
		t.Errorf("after first Tab: expected title %q, got %q", " users.0.name ", got)
	}
	if got := plainOutput(app); got != `"Ada"` {
		t.Errorf("after first Tab: expected plainOutput %q, got %q", `"Ada"`, got)
	}

	// Second Tab: "users.0.nationality"
	pressKey(app, tcell.KeyTab)
	waitFor(t, func() bool {
		return inputText(app) == "users.0.nationality"
	})

	// Third Tab: wraps back to the originally typed text "users.0.na"
	pressKey(app, tcell.KeyTab)
	waitFor(t, func() bool {
		return inputText(app) == "users.0.na"
	})

	// Shift+Tab (Backtab) from position 0 rewinds to the last candidate.
	pressKey(app, tcell.KeyBacktab)
	waitFor(t, func() bool {
		return inputText(app) == "users.0.nationality"
	})
}

func TestApp_EscapeMidCycleRestoresTypedText(t *testing.T) {
	app, cleanup := newTestApp(t, testDoc)
	defer cleanup()

	typeText(app, "users.0.na")
	waitFor(t, func() bool {
		return outputTitle(app) == " users.0 ✗ na "
	})

	pressKey(app, tcell.KeyTab)
	waitFor(t, func() bool {
		return inputText(app) == "users.0.name"
	})

	pressKey(app, tcell.KeyEscape)
	waitFor(t, func() bool {
		return inputText(app) == "users.0.na"
	})
}

func TestApp_TypoResilienceShowsAncestor(t *testing.T) {
	app, cleanup := newTestApp(t, testDoc)
	defer cleanup()

	typeText(app, "users.0.nmae")

	waitFor(t, func() bool {
		return outputTitle(app) == " users.0 ✗ nmae "
	})

	if got := plainOutput(app); !strings.Contains(got, "Ada") {
		t.Errorf("expected ancestor view plainOutput to contain %q, got %q", "Ada", got)
	}
}

func TestApp_TabNoOpWithoutSuggestions(t *testing.T) {
	app, cleanup := newTestApp(t, testDoc)
	defer cleanup()

	typeText(app, "zzz")
	waitFor(t, func() bool {
		return inputText(app) == "zzz"
	})

	// No suggestions exist for "zzz"; give the app a moment to settle,
	// then confirm Tab does not alter the input text.
	waitFor(t, func() bool {
		return suggestionCount(app) == 0
	})

	pressKey(app, tcell.KeyTab)

	// There is no positive event to wait on for "nothing happened", so
	// poll briefly and assert the text never changes.
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if got := inputText(app); got != "zzz" {
			t.Fatalf("expected Tab to be a no-op, input changed to %q", got)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestApp_SaveOverlayGatesGlobalHotkeys(t *testing.T) {
	app, cleanup := newTestApp(t, testDoc)
	defer cleanup()

	// Open the save dialog.
	pressKey(app, tcell.KeyCtrlS)
	waitFor(t, func() bool {
		return frontPage(app) == "save" && focusIs(app, app.saveInput)
	})

	// Ctrl+O must NOT steal focus to the output panel while the save
	// dialog is up — the dialog would be stuck with no way to dismiss it.
	pressKey(app, tcell.KeyCtrlO)

	// Negative assertion: poll briefly, the overlay and focus must not move.
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if got := frontPage(app); got != "save" {
			t.Fatalf("Ctrl+O while save dialog open: front page changed to %q", got)
		}
		if !focusIs(app, app.saveInput) {
			t.Fatal("Ctrl+O while save dialog open: focus left the save input")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Escape dismisses the dialog and returns focus to the query input.
	pressKey(app, tcell.KeyEscape)
	waitFor(t, func() bool {
		return frontPage(app) == "main" && focusIs(app, app.inputField)
	})
}

func TestApp_SaveOverlayTabDoesNotDismiss(t *testing.T) {
	app, cleanup := newTestApp(t, testDoc)
	defer cleanup()

	pressKey(app, tcell.KeyCtrlS)
	waitFor(t, func() bool {
		return frontPage(app) == "save"
	})

	// Tab in the filename field must not cancel the dialog (tview fires
	// an InputField's done func for Tab/Backtab as well as Enter/Escape).
	pressKey(app, tcell.KeyTab)

	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if got := frontPage(app); got != "save" {
			t.Fatalf("Tab in save dialog dismissed it: front page is %q", got)
		}
		time.Sleep(20 * time.Millisecond)
	}

	pressKey(app, tcell.KeyEscape)
	waitFor(t, func() bool {
		return frontPage(app) == "main"
	})
}

func TestApp_HelpToggle(t *testing.T) {
	app, cleanup := newTestApp(t, testDoc)
	defer cleanup()

	// F1 shows the help panel (nonzero width after relayout) and
	// focuses it.
	pressKey(app, tcell.KeyF1)
	waitFor(t, func() bool {
		return syncCall(app, func() bool {
			if !app.helpVisible || app.tviewApp.GetFocus() != app.helpPanel {
				return false
			}
			_, _, w, _ := app.helpPanel.GetRect()
			return w > 0
		})
	})

	// Escape hides it (zero width) and returns focus to the input field.
	pressKey(app, tcell.KeyEscape)
	waitFor(t, func() bool {
		return syncCall(app, func() bool {
			if app.helpVisible || app.tviewApp.GetFocus() != app.inputField {
				return false
			}
			_, _, w, _ := app.helpPanel.GetRect()
			return w == 0
		})
	})
}

func TestApp_CtrlYFlashesFooter(t *testing.T) {
	app, cleanup := newTestApp(t, testDoc)
	defer cleanup()

	if got := footerTextNow(app); got != footerText {
		t.Fatalf("precondition: footer should show key hints, got %q", got)
	}

	// Ctrl+Y attempts a clipboard copy. In a headless environment the
	// copy may fail, so assert only that the footer flashes SOMETHING —
	// either "Copied to clipboard!" or an error — i.e. it differs from
	// the default key-hint text.
	pressKey(app, tcell.KeyCtrlY)
	waitFor(t, func() bool {
		return footerTextNow(app) != footerText
	})
}
