// Package path splits gjson paths into segments so the rest of the app
// can reason about how much of a typed path is plain navigation.
package path

import "strings"

// Kind classifies a path segment.
type Kind int

const (
	KindKey      Kind = iota // plain object key (possibly with escapes)
	KindIndex                // numeric array index, e.g. "0" or "-1"
	KindAdvanced             // gjson-special: #, @modifier, wildcard, query, multipath, pipe
)

// Segment is one dot-separated component of a gjson path, escapes intact.
type Segment struct {
	Raw  string
	Kind Kind
}

// Split divides a gjson path into segments on unescaped dots. Dots that
// follow a backslash or sit inside parentheses (queries) do not split.
// A trailing dot yields a trailing empty segment.
func Split(p string) []Segment {
	if p == "" {
		return nil
	}
	var segs []Segment
	var cur strings.Builder
	depth := 0
	escaped := false
	for _, r := range p {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case r == '\\':
			cur.WriteRune(r)
			escaped = true
		case r == '(':
			depth++
			cur.WriteRune(r)
		case r == ')':
			if depth > 0 {
				depth--
			}
			cur.WriteRune(r)
		case r == '.' && depth == 0:
			segs = append(segs, newSegment(cur.String()))
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	segs = append(segs, newSegment(cur.String()))
	return segs
}

// Join reassembles segments into a path string.
func Join(segs []Segment) string {
	parts := make([]string, len(segs))
	for i, s := range segs {
		parts[i] = s.Raw
	}
	return strings.Join(parts, ".")
}

// EscapeKey escapes a raw object key for use as a path segment. The
// characters \ . * ? | are escaped everywhere; # @ { [ are only special to
// gjson at the start of a segment, so they are escaped in leading position
// only.
func EscapeKey(key string) string {
	var b strings.Builder
	for i, r := range key {
		switch r {
		case '\\', '.', '*', '?', '|':
			b.WriteRune('\\')
		case '#', '@', '{', '[':
			if i == 0 {
				b.WriteRune('\\')
			}
		}
		b.WriteRune(r)
	}
	return b.String()
}

func newSegment(raw string) Segment {
	return Segment{Raw: raw, Kind: classify(raw)}
}

func classify(raw string) Kind {
	if raw == "" {
		return KindKey
	}
	if isIndex(raw) {
		return KindIndex
	}
	if strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "@") ||
		strings.HasPrefix(raw, "{") || strings.HasPrefix(raw, "[") {
		return KindAdvanced
	}
	escaped := false
	for _, r := range raw {
		switch {
		case escaped:
			escaped = false
		case r == '\\':
			escaped = true
		case r == '*' || r == '?' || r == '|':
			return KindAdvanced
		}
	}
	return KindKey
}

func isIndex(raw string) bool {
	s := strings.TrimPrefix(raw, "-")
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
