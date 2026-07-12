package autocomplete

// CycleState tracks Tab-cycling through suggestions. Position 0 is the
// originally typed text; positions 1..len(candidates) are the candidates.
type CycleState struct {
	base       string
	candidates []string
	index      int
}

// NewCycleState starts a cycle from the typed text and its candidates.
func NewCycleState(base string, suggestions []Suggestion) *CycleState {
	c := &CycleState{base: base}
	for _, s := range suggestions {
		c.candidates = append(c.candidates, s.Full)
	}
	return c
}

// Next advances the cycle (wrapping) and returns the text to show.
func (c *CycleState) Next() string {
	c.index = (c.index + 1) % (len(c.candidates) + 1)
	return c.Current()
}

// Prev rewinds the cycle (wrapping) and returns the text to show.
func (c *CycleState) Prev() string {
	n := len(c.candidates) + 1
	c.index = (c.index - 1 + n) % n
	return c.Current()
}

// Current returns the text at the current cycle position.
func (c *CycleState) Current() string {
	if c.index == 0 {
		return c.base
	}
	return c.candidates[c.index-1]
}

// Base returns the originally typed text (position 0).
func (c *CycleState) Base() string { return c.base }

// Selected returns the highlighted candidate index, or -1 at position 0.
func (c *CycleState) Selected() int { return c.index - 1 }
