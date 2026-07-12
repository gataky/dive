package autocomplete

import "testing"

func newTestCycle() *CycleState {
	return NewCycleState("users.0.na", []Suggestion{
		{Full: "users.0.name", Display: "name"},
		{Full: "users.0.nationality", Display: "nationality"},
	})
}

func TestCycleForwardWraps(t *testing.T) {
	c := newTestCycle()
	want := []string{"users.0.name", "users.0.nationality", "users.0.na", "users.0.name"}
	for i, w := range want {
		if got := c.Next(); got != w {
			t.Fatalf("Next() #%d = %q, want %q", i+1, got, w)
		}
	}
}

func TestCycleBackwardWraps(t *testing.T) {
	c := newTestCycle()
	want := []string{"users.0.nationality", "users.0.name", "users.0.na"}
	for i, w := range want {
		if got := c.Prev(); got != w {
			t.Fatalf("Prev() #%d = %q, want %q", i+1, got, w)
		}
	}
}

func TestCycleSelected(t *testing.T) {
	c := newTestCycle()
	if c.Selected() != -1 {
		t.Errorf("initial Selected() = %d, want -1", c.Selected())
	}
	c.Next()
	if c.Selected() != 0 {
		t.Errorf("after one Next, Selected() = %d, want 0", c.Selected())
	}
	c.Next()
	c.Next() // back to base
	if c.Selected() != -1 {
		t.Errorf("wrapped to base, Selected() = %d, want -1", c.Selected())
	}
}

func TestCycleBase(t *testing.T) {
	c := newTestCycle()
	c.Next()
	if c.Base() != "users.0.na" {
		t.Errorf("Base() = %q, want original text", c.Base())
	}
}

func TestCycleNoCandidates(t *testing.T) {
	c := NewCycleState("plain", nil)
	if got := c.Next(); got != "plain" {
		t.Errorf("Next() with no candidates = %q, want base", got)
	}
	if got := c.Prev(); got != "plain" {
		t.Errorf("Prev() with no candidates = %q, want base", got)
	}
	if c.Selected() != -1 {
		t.Errorf("Selected() with no candidates = %d, want -1", c.Selected())
	}
}
