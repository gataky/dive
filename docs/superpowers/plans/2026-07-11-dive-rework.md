# dive Rework Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rework dive's interaction and rendering: a fixed never-rebuilt layout with an always-visible live suggestions bar, Tab-cycling completion (typed text = hidden index 0), deepest-valid-ancestor output that never blanks, and syntax-highlighted JSON.

**Architecture:** New pure packages `internal/path` (gjson path segmentation) and `internal/render` (colorizer); `internal/query` and `internal/autocomplete` gain new stateless APIs (`Resolve`, `Suggest`, `CycleState`) built alongside the old code so the build stays green; `internal/ui` is rewritten last to wire everything into one keystroke pipeline, at which point the old engine/suggester files are deleted. Spec: `docs/superpowers/specs/2026-07-11-dive-rework-design.md`.

**Tech Stack:** Go, tview, tcell, gjson, tidwall/pretty (already an indirect dep, becomes direct).

**Conventions for every task:** run commands from the repo root `/Users/jeff/Documents/dive`. Test only `./internal/...` (the `.direnv` directory contains a vendored Go module cache — never run `go test ./...` or `gofmt .` on the repo root). The build must compile after every task.

---

### Task 1: Bump dependency versions

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Update the three direct dependencies to latest**

Run:
```bash
go get github.com/gdamore/tcell/v2@latest github.com/rivo/tview@latest github.com/tidwall/gjson@latest && go mod tidy
```

- [ ] **Step 2: Verify everything still builds and passes**

Run: `go build ./... && go test ./internal/...`
Expected: builds clean; all existing tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: bump tcell, tview, gjson to latest"
```

---

### Task 2: `internal/path` — segment splitting, classification, joining, escaping

Pure path logic. A gjson path splits on dots, except: a dot after `\` is part of the key, and dots inside parentheses (queries like `#(a.b=="x")`) don't split. Segments classify as plain key, array index, or "advanced" (anything gjson-special: `#`, `@modifier`, wildcards, slices, queries, multipath).

**Files:**
- Create: `internal/path/path.go`
- Test: `internal/path/path_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/path/path_test.go`:

```go
package path

import (
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []Segment
	}{
		{"empty", "", nil},
		{"single key", "users", []Segment{{"users", KindKey}}},
		{"nested keys", "a.b.c", []Segment{{"a", KindKey}, {"b", KindKey}, {"c", KindKey}}},
		{"trailing dot yields empty segment", "users.", []Segment{{"users", KindKey}, {"", KindKey}}},
		{"array index", "users.0.name", []Segment{{"users", KindKey}, {"0", KindIndex}, {"name", KindKey}}},
		{"negative index", "users.-1", []Segment{{"users", KindKey}, {"-1", KindIndex}}},
		{"escaped dot stays in key", `fav\.movie.title`, []Segment{{`fav\.movie`, KindKey}, {"title", KindKey}}},
		{"count segment", "users.#", []Segment{{"users", KindKey}, {"#", KindAdvanced}}},
		{"query with inner dots", `users.#(a.b=="x")#.name`, []Segment{{"users", KindKey}, {`#(a.b=="x")#`, KindAdvanced}, {"name", KindKey}}},
		{"modifier", "users.@reverse", []Segment{{"users", KindKey}, {"@reverse", KindAdvanced}}},
		{"wildcard", "data.*.value", []Segment{{"data", KindKey}, {"*", KindAdvanced}, {"value", KindKey}}},
		{"slice", "users.0:3", []Segment{{"users", KindKey}, {"0:3", KindAdvanced}}},
		{"pipe", "users.#.age|@reverse", []Segment{{"users", KindKey}, {"#", KindAdvanced}, {"age|@reverse", KindAdvanced}}},
		{"multipath", "{name,age}", []Segment{{"{name,age}", KindAdvanced}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Split(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Split(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	paths := []string{"", "users", "a.b.c", `fav\.movie.title`, `users.#(a.b=="x")#.name`, "users."}
	for _, p := range paths {
		if got := Join(Split(p)); got != p {
			t.Errorf("Join(Split(%q)) = %q, want round-trip", p, got)
		}
	}
}

func TestJoinPrefix(t *testing.T) {
	segs := Split("a.b.c")
	if got := Join(segs[:2]); got != "a.b" {
		t.Errorf("Join(segs[:2]) = %q, want %q", got, "a.b")
	}
	if got := Join(segs[:0]); got != "" {
		t.Errorf("Join(segs[:0]) = %q, want empty", got)
	}
}

func TestEscapeKey(t *testing.T) {
	tests := []struct{ in, want string }{
		{"name", "name"},
		{"fav.movie", `fav\.movie`},
		{"a*b", `a\*b`},
		{"a?b", `a\?b`},
		{"a|b", `a\|b`},
		{`back\slash`, `back\\slash`},
	}
	for _, tt := range tests {
		if got := EscapeKey(tt.in); got != tt.want {
			t.Errorf("EscapeKey(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/path/ -v`
Expected: FAIL — package doesn't compile (`Segment`, `Split` undefined).

- [ ] **Step 3: Implement `internal/path/path.go`**

```go
// Package path splits gjson paths into segments so the rest of the app
// can reason about how much of a typed path is plain navigation.
package path

import "strings"

// Kind classifies a path segment.
type Kind int

const (
	KindKey      Kind = iota // plain object key (possibly with escapes)
	KindIndex                // numeric array index, e.g. "0" or "-1"
	KindAdvanced             // gjson-special: #, @modifier, wildcard, slice, query, multipath, pipe
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

// EscapeKey escapes a raw object key for use as a path segment.
func EscapeKey(key string) string {
	var b strings.Builder
	for _, r := range key {
		switch r {
		case '\\', '.', '*', '?', '|':
			b.WriteRune('\\')
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
		case r == '*' || r == '?' || r == '|' || r == ':' || r == '(':
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/path/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/path/
git commit -m "feat: add internal/path for gjson path segmentation"
```

---

### Task 3: `internal/query` — deepest-valid-ancestor `Resolve`

New stateless API added **alongside** the old `Engine` (which `app.go` still uses until Task 9 — do not delete `engine.go` yet).

**Files:**
- Create: `internal/query/resolve.go`
- Test: `internal/query/resolve_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/query/resolve_test.go`:

```go
package query

import "testing"

const doc = `{
	"users": [
		{"name": "Alice", "age": 28, "tags": ["a", "b"]},
		{"name": "Bob", "age": 35}
	],
	"company": {"name": "Acme", "founded": 2010}
}`

func TestResolve(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantResolved string
		wantSuffix   string
		wantValue    string // Result.String() spot check, "" to skip
	}{
		{"empty path is whole doc", "", "", "", ""},
		{"full match", "company.name", "company.name", "", "Acme"},
		{"array index", "users.1.name", "users.1.name", "", "Bob"},
		{"typo on leaf", "company.nmae", "company", "nmae", ""},
		{"typo mid-path", "users.9.name", "users", "9.name", ""},
		{"nothing matches", "zzz.yyy", "", "zzz.yyy", ""},
		{"trailing dot not unmatched", "company.", "company", "", ""},
		{"bare trailing dot", ".", "", "", ""},
		{"count query", "users.#", "users.#", "", "2"},
		{"filter query", `users.#(age>30)#.name`, `users.#(age>30)#.name`, "", ""},
		{"typo after query", `users.#(age>30)#.nmae`, `users.#(age>30)#`, "nmae", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(doc, tt.path)
			if got.ResolvedPath != tt.wantResolved {
				t.Errorf("ResolvedPath = %q, want %q", got.ResolvedPath, tt.wantResolved)
			}
			if got.UnmatchedSuffix != tt.wantSuffix {
				t.Errorf("UnmatchedSuffix = %q, want %q", got.UnmatchedSuffix, tt.wantSuffix)
			}
			if tt.wantValue != "" && got.Result.String() != tt.wantValue {
				t.Errorf("Result = %q, want %q", got.Result.String(), tt.wantValue)
			}
			if !got.Result.Exists() {
				t.Error("Result must always exist (never blank)")
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/query/ -run TestResolve -v`
Expected: FAIL — `Resolve` undefined.

- [ ] **Step 3: Implement `internal/query/resolve.go`**

```go
package query

import (
	"github.com/gataky/dive/internal/path"
	"github.com/tidwall/gjson"
)

// Resolution describes how much of a typed path matched the document.
type Resolution struct {
	ResolvedPath    string       // longest prefix that exists; "" means whole document
	UnmatchedSuffix string       // remainder of the typed path; "" when fully matched
	Result          gjson.Result // value at ResolvedPath; always exists
}

// Resolve finds the deepest valid ancestor of p in data. A trailing empty
// segment (path ends in ".") means the user is mid-typing and does not
// count as unmatched.
func Resolve(data, p string) Resolution {
	segs := path.Split(p)
	if n := len(segs); n > 0 && segs[n-1].Raw == "" {
		segs = segs[:n-1]
	}
	for i := len(segs); i > 0; i-- {
		candidate := path.Join(segs[:i])
		if r := gjson.Get(data, candidate); r.Exists() {
			return Resolution{
				ResolvedPath:    candidate,
				UnmatchedSuffix: path.Join(segs[i:]),
				Result:          r,
			}
		}
	}
	return Resolution{
		ResolvedPath:    "",
		UnmatchedSuffix: path.Join(segs),
		Result:          gjson.Parse(data),
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/query/ -v`
Expected: `TestResolve` subtests all PASS; old `Engine` tests still PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/query/resolve.go internal/query/resolve_test.go
git commit -m "feat: add query.Resolve with deepest-valid-ancestor fallback"
```

---

### Task 4: Theme JSON colors + `internal/render` colorizer

`render.Pretty` gives plain indented JSON (key order preserved via `tidwall/pretty` — do NOT use `json.Unmarshal` into a map, that reorders keys). `render.Colorize` walks the gjson tree and emits tview markup; all document content passes through `tview.Escape` so a `[red]` inside the JSON can't corrupt the display.

**Files:**
- Modify: `internal/ui/theme/colors.go`
- Create: `internal/render/render.go`
- Test: `internal/render/render_test.go`

- [ ] **Step 1: Add JSON/suggestion colors to the theme**

In `internal/ui/theme/colors.go`, add to the `Theme` struct (after the `ColorSuccess`/`ColorError` fields):

```go
	// JSON syntax highlighting (tview markup color names)
	JSONKey    string // object keys
	JSONString string // string values
	JSONNumber string // numbers
	JSONBool   string // true/false
	JSONNull   string // null

	// SuggestHighlight is the fg:bg tview markup for the selected suggestion.
	SuggestHighlight string
```

And to `DefaultTheme()` (after the `ColorError` line):

```go
		// JSON syntax highlighting
		JSONKey:    "aqua",
		JSONString: "green",
		JSONNumber: "yellow",
		JSONBool:   "orange",
		JSONNull:   "gray",

		SuggestHighlight: "black:aqua",
```

- [ ] **Step 2: Write the failing tests**

Create `internal/render/render_test.go`:

```go
package render

import (
	"strings"
	"testing"

	"github.com/gataky/dive/internal/ui/theme"
	"github.com/tidwall/gjson"
)

func TestPrettyPreservesKeyOrder(t *testing.T) {
	r := gjson.Parse(`{"zebra":1,"apple":2}`)
	got := Pretty(r)
	if strings.Index(got, "zebra") > strings.Index(got, "apple") {
		t.Errorf("Pretty reordered keys:\n%s", got)
	}
	if !strings.Contains(got, "  \"zebra\"") {
		t.Errorf("Pretty output not indented:\n%s", got)
	}
}

func TestPrettyPrimitive(t *testing.T) {
	r := gjson.Get(`{"name":"Ada"}`, "name")
	if got := Pretty(r); got != `"Ada"` {
		t.Errorf("Pretty(string) = %q, want %q", got, `"Ada"`)
	}
}

func TestColorizeTypes(t *testing.T) {
	th := theme.DefaultTheme()
	r := gjson.Parse(`{"s":"x","n":1.5,"b":true,"z":null,"arr":[1],"obj":{"k":2}}`)
	got := Colorize(r, th)

	for _, want := range []string{
		"[" + th.JSONKey + `]"s"[-]`,
		"[" + th.JSONString + `]"x"[-]`,
		"[" + th.JSONNumber + "]1.5[-]",
		"[" + th.JSONBool + "]true[-]",
		"[" + th.JSONNull + "]null[-]",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Colorize missing %q in:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "{") || !strings.Contains(got, "[\n") {
		t.Errorf("Colorize missing structural punctuation:\n%s", got)
	}
}

func TestColorizeEscapesMarkupInContent(t *testing.T) {
	th := theme.DefaultTheme()
	r := gjson.Parse(`{"note":"[red]danger[-]"}`)
	got := Colorize(r, th)
	// tview.Escape turns [tag] into [tag[] so it renders literally
	if !strings.Contains(got, "[red[]") || !strings.Contains(got, "[-[]") {
		t.Errorf("document content not escaped:\n%s", got)
	}
}

func TestColorizeEmptyContainers(t *testing.T) {
	th := theme.DefaultTheme()
	if got := Colorize(gjson.Parse(`{}`), th); got != "{}" {
		t.Errorf("empty object = %q, want {}", got)
	}
	if got := Colorize(gjson.Parse(`[]`), th); got != "[]" {
		t.Errorf("empty array = %q, want []", got)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/render/ -v`
Expected: FAIL — package doesn't exist yet.

- [ ] **Step 4: Implement `internal/render/render.go`**

```go
// Package render turns gjson results into display text: plain indented
// JSON for export, and tview-markup-colored JSON for the output panel.
package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gataky/dive/internal/ui/theme"
	"github.com/rivo/tview"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
)

const indentStep = "  "

// Pretty returns plain indented JSON for result, preserving key order.
func Pretty(result gjson.Result) string {
	if result.IsObject() || result.IsArray() {
		opts := &pretty.Options{Indent: indentStep, SortKeys: false}
		return strings.TrimRight(string(pretty.PrettyOptions([]byte(result.Raw), opts)), "\n")
	}
	return result.Raw
}

// Colorize returns result as indented JSON with tview color markup.
// Document content is escaped so it cannot inject markup.
func Colorize(result gjson.Result, th *theme.Theme) string {
	var b strings.Builder
	writeValue(&b, result, th, 0)
	return b.String()
}

func writeValue(b *strings.Builder, v gjson.Result, th *theme.Theme, depth int) {
	indent := strings.Repeat(indentStep, depth)
	switch {
	case v.IsObject():
		b.WriteString("{")
		first := true
		v.ForEach(func(key, value gjson.Result) bool {
			if !first {
				b.WriteString(",")
			}
			first = false
			b.WriteString("\n" + indent + indentStep)
			fmt.Fprintf(b, "[%s]%s[-]: ", th.JSONKey, tview.Escape(strconv.Quote(key.String())))
			writeValue(b, value, th, depth+1)
			return true
		})
		if !first {
			b.WriteString("\n" + indent)
		}
		b.WriteString("}")
	case v.IsArray():
		b.WriteString("[")
		first := true
		v.ForEach(func(_, value gjson.Result) bool {
			if !first {
				b.WriteString(",")
			}
			first = false
			b.WriteString("\n" + indent + indentStep)
			writeValue(b, value, th, depth+1)
			return true
		})
		if !first {
			b.WriteString("\n" + indent)
		}
		b.WriteString("]")
	default:
		var color string
		switch v.Type {
		case gjson.String:
			color = th.JSONString
		case gjson.Number:
			color = th.JSONNumber
		case gjson.True, gjson.False:
			color = th.JSONBool
		default:
			color = th.JSONNull
		}
		fmt.Fprintf(b, "[%s]%s[-]", color, tview.Escape(v.Raw))
	}
}
```

Then make `tidwall/pretty` a direct dependency:

```bash
go mod tidy
```

(`go.mod` should move `github.com/tidwall/pretty` out of the `// indirect` block.)

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/render/ ./internal/ui/... -v`
Expected: render tests PASS; existing ui/theme-dependent tests still PASS.

Note: a bare `[` we emit for arrays is always followed by `\n` or `]`, which tview renders literally — only *content* needs escaping, which `tview.Escape` handles.

- [ ] **Step 6: Commit**

```bash
git add internal/render/ internal/ui/theme/colors.go go.mod go.sum
git commit -m "feat: add render package with JSON colorizer and theme colors"
```

---

### Task 5: `internal/autocomplete` — new `Suggest` API

New file **alongside** the old `suggester.go` (still used by `app.go` until Task 9). Fixes the ≥10-index bug (`strconv.Itoa`, not `rune('0'+i)`), escapes dots in keys, and returns full replacement texts plus display labels.

**Files:**
- Create: `internal/autocomplete/suggest.go`
- Test: `internal/autocomplete/suggest_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/autocomplete/suggest_test.go`:

```go
package autocomplete

import (
	"reflect"
	"testing"

	"github.com/tidwall/gjson"
)

const doc = `{
	"users": [
		{"name": "Alice", "nationality": "US", "age": 28},
		{"name": "Bob"}, {"name": "C"}, {"name": "D"}, {"name": "E"},
		{"name": "F"}, {"name": "G"}, {"name": "H"}, {"name": "I"},
		{"name": "J"}, {"name": "K"}, {"name": "L"}
	],
	"fav.movie": "Dune",
	"company": {"name": "Acme"}
}`

func fulls(s []Suggestion) []string {
	var out []string
	for _, x := range s {
		out = append(out, x.Full)
	}
	return out
}

func TestSuggestTopLevel(t *testing.T) {
	got := Suggest(doc, "")
	want := []string{"users", `fav\.movie`, "company"}
	if !reflect.DeepEqual(fulls(got), want) {
		t.Errorf("Suggest(\"\") = %v, want %v", fulls(got), want)
	}
}

func TestSuggestPrefixFilter(t *testing.T) {
	got := Suggest(doc, "users.0.na")
	want := []string{"users.0.name", "users.0.nationality"}
	if !reflect.DeepEqual(fulls(got), want) {
		t.Errorf("got %v, want %v", fulls(got), want)
	}
	if got[0].Display != "name" || got[1].Display != "nationality" {
		t.Errorf("Display labels = %q, %q; want key names", got[0].Display, got[1].Display)
	}
}

func TestSuggestTrailingDot(t *testing.T) {
	got := Suggest(doc, "company.")
	want := []string{"company.name"}
	if !reflect.DeepEqual(fulls(got), want) {
		t.Errorf("got %v, want %v", fulls(got), want)
	}
}

func TestSuggestArrayIndicesPast10(t *testing.T) {
	got := Suggest(doc, "users.")
	if len(got) != 12 {
		t.Fatalf("got %d suggestions, want 12", len(got))
	}
	if got[10].Display != "10" || got[11].Display != "11" {
		t.Errorf("indices past 9 wrong: %q, %q", got[10].Display, got[11].Display)
	}
	// index prefix filtering
	got = Suggest(doc, "users.1")
	want := []string{"users.1", "users.10", "users.11"}
	if !reflect.DeepEqual(fulls(got), want) {
		t.Errorf("got %v, want %v", fulls(got), want)
	}
}

func TestSuggestionsResolve(t *testing.T) {
	// every suggestion must be a path gjson can actually resolve
	for _, s := range Suggest(doc, "") {
		if !gjson.Get(doc, s.Full).Exists() {
			t.Errorf("suggestion %q does not resolve", s.Full)
		}
	}
}

func TestSuggestInvalidParent(t *testing.T) {
	if got := Suggest(doc, "zzz.a"); got != nil {
		t.Errorf("invalid parent should give nil, got %v", got)
	}
}

func TestSuggestAdvancedSegmentDegrades(t *testing.T) {
	if got := Suggest(doc, "users.#(age>2"); got != nil {
		t.Errorf("advanced partial should give nil, got %v", got)
	}
	// but a *parent* that is an advanced query still enumerates its result
	got := Suggest(doc, "users.#(age>20)#.")
	if len(got) == 0 {
		t.Error("expected index suggestions under advanced parent")
	}
}

func TestSuggestPrimitiveParent(t *testing.T) {
	if got := Suggest(doc, "company.name."); got != nil {
		t.Errorf("primitive parent should give nil, got %v", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/autocomplete/ -run TestSuggest -v`
Expected: FAIL — `Suggestion`, `Suggest` undefined.

- [ ] **Step 3: Implement `internal/autocomplete/suggest.go`**

```go
package autocomplete

import (
	"strconv"
	"strings"

	"github.com/gataky/dive/internal/path"
	"github.com/tidwall/gjson"
)

// Suggestion is one completion candidate for the query input.
type Suggestion struct {
	Full    string // full input text this candidate completes to
	Display string // the key or index shown in the suggestions bar
}

// Suggest returns completions for the last path segment of text. It
// returns nil when the parent path is invalid, the parent is a
// primitive, or the segment being typed uses advanced gjson syntax.
func Suggest(data, text string) []Suggestion {
	segs := path.Split(text)
	partial := ""
	var parentSegs []path.Segment
	if len(segs) > 0 {
		last := segs[len(segs)-1]
		if last.Kind == path.KindAdvanced {
			return nil
		}
		partial = last.Raw
		parentSegs = segs[:len(segs)-1]
	}

	parentPath := path.Join(parentSegs)
	var parent gjson.Result
	if parentPath == "" {
		parent = gjson.Parse(data)
	} else {
		parent = gjson.Get(data, parentPath)
	}
	if !parent.Exists() {
		return nil
	}

	var keys []string
	switch {
	case parent.IsObject():
		parent.ForEach(func(key, _ gjson.Result) bool {
			keys = append(keys, path.EscapeKey(key.String()))
			return true
		})
	case parent.IsArray():
		for i := range parent.Array() {
			keys = append(keys, strconv.Itoa(i))
		}
	default:
		return nil
	}

	var out []Suggestion
	for _, k := range keys {
		if !strings.HasPrefix(k, partial) {
			continue
		}
		full := k
		if parentPath != "" {
			full = parentPath + "." + k
		}
		out = append(out, Suggestion{Full: full, Display: k})
	}
	return out
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/autocomplete/ -v`
Expected: new `TestSuggest*` tests PASS; old suggester tests still PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/autocomplete/suggest.go internal/autocomplete/suggest_test.go
git commit -m "feat: add autocomplete.Suggest with escaping and correct array indices"
```

---

### Task 6: `internal/autocomplete` — `CycleState`

Pure Tab-cycle state machine: position 0 is the originally typed text, positions 1..n are candidates; Tab advances, Shift+Tab rewinds, both wrap.

**Files:**
- Create: `internal/autocomplete/cycle.go`
- Test: `internal/autocomplete/cycle_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/autocomplete/cycle_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/autocomplete/ -run TestCycle -v`
Expected: FAIL — `CycleState` undefined.

- [ ] **Step 3: Implement `internal/autocomplete/cycle.go`**

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/autocomplete/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/autocomplete/cycle.go internal/autocomplete/cycle_test.go
git commit -m "feat: add CycleState for Tab-cycling completions"
```

---

### Task 7: `internal/ui` — suggestions bar formatting

Pure function that lays out candidate labels in one line: two-space separators, selected candidate highlighted, window shifts so the selection is always visible, `… +N more` overflow marker.

**Files:**
- Create: `internal/ui/suggestions.go`
- Test: `internal/ui/suggestions_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/ui/suggestions_test.go`:

```go
package ui

import (
	"strings"
	"testing"

	"github.com/gataky/dive/internal/autocomplete"
	"github.com/gataky/dive/internal/ui/theme"
)

func suggs(names ...string) []autocomplete.Suggestion {
	var out []autocomplete.Suggestion
	for _, n := range names {
		out = append(out, autocomplete.Suggestion{Full: n, Display: n})
	}
	return out
}

func TestFormatSuggestionsEmpty(t *testing.T) {
	th := theme.DefaultTheme()
	if got := formatSuggestions(nil, -1, 80, th); got != "" {
		t.Errorf("empty input should give empty string, got %q", got)
	}
}

func TestFormatSuggestionsPlain(t *testing.T) {
	th := theme.DefaultTheme()
	got := formatSuggestions(suggs("name", "nationality"), -1, 80, th)
	if got != "name  nationality" {
		t.Errorf("got %q, want %q", got, "name  nationality")
	}
}

func TestFormatSuggestionsHighlight(t *testing.T) {
	th := theme.DefaultTheme()
	got := formatSuggestions(suggs("name", "nationality"), 1, 80, th)
	want := "[" + th.SuggestHighlight + "] nationality [-:-]"
	if !strings.Contains(got, want) {
		t.Errorf("selected candidate not highlighted: %q", got)
	}
	if strings.Contains(got, "] name [") {
		t.Errorf("unselected candidate should not be highlighted: %q", got)
	}
}

func TestFormatSuggestionsOverflow(t *testing.T) {
	th := theme.DefaultTheme()
	many := suggs("alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel")
	got := formatSuggestions(many, -1, 30, th)
	if !strings.Contains(got, "more") {
		t.Errorf("expected overflow marker in %q", got)
	}
	if strings.Contains(got, "hotel") {
		t.Errorf("overflowing candidate should not render: %q", got)
	}
}

func TestFormatSuggestionsSelectedAlwaysVisible(t *testing.T) {
	th := theme.DefaultTheme()
	many := suggs("alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel")
	got := formatSuggestions(many, 7, 30, th)
	if !strings.Contains(got, "hotel") {
		t.Errorf("selected candidate must be visible: %q", got)
	}
	if !strings.HasPrefix(got, "… ") {
		t.Errorf("shifted window should show leading ellipsis: %q", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/ -run TestFormatSuggestions -v`
Expected: FAIL — `formatSuggestions` undefined.

- [ ] **Step 3: Implement `internal/ui/suggestions.go`**

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/gataky/dive/internal/autocomplete"
	"github.com/gataky/dive/internal/ui/theme"
	"github.com/rivo/tview"
)

// formatSuggestions renders the suggestions bar: candidates separated by
// two spaces, the selected one highlighted, windowed so the selection is
// always visible, with "… +N more" when candidates overflow the width.
func formatSuggestions(suggs []autocomplete.Suggestion, selected, width int, th *theme.Theme) string {
	if len(suggs) == 0 {
		return ""
	}
	start := 0
	if selected >= 0 && fitCount(suggs, 0, width) <= selected {
		start = selected
	}
	n := fitCount(suggs, start, width)

	var b strings.Builder
	if start > 0 {
		b.WriteString("… ")
	}
	for i := start; i < start+n && i < len(suggs); i++ {
		if i > start {
			b.WriteString("  ")
		}
		if i == selected {
			b.WriteString("[" + th.SuggestHighlight + "] " + tview.Escape(suggs[i].Display) + " [-:-]")
		} else {
			b.WriteString(tview.Escape(suggs[i].Display))
		}
	}
	if rest := len(suggs) - (start + n); rest > 0 {
		fmt.Fprintf(&b, "  … +%d more", rest)
	}
	return b.String()
}

// fitCount returns how many suggestions starting at start fit in width
// display cells, reserving 12 cells for overflow markers. Always >= 1 so
// at least the selected candidate renders.
func fitCount(suggs []autocomplete.Suggestion, start, width int) int {
	budget := width - 12
	used, n := 0, 0
	for i := start; i < len(suggs); i++ {
		cells := len([]rune(suggs[i].Display))
		if n > 0 {
			cells += 2
		}
		if used+cells > budget && n > 0 {
			break
		}
		used += cells
		n++
	}
	return n
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/ui/ -v`
Expected: all PASS (including the pre-existing help panel tests).

- [ ] **Step 5: Commit**

```bash
git add internal/ui/suggestions.go internal/ui/suggestions_test.go
git commit -m "feat: add suggestions bar formatting with highlight and overflow"
```

---

### Task 8: `internal/ui` — component constructors

Rewrite `components.go`: drop the dropdown, add the suggestions bar and the save input, turn word wrap off on the output panel. The pre-existing help panel tests in `components_test.go` keep passing unchanged.

**Files:**
- Modify: `internal/ui/components.go` (full replacement below)

- [ ] **Step 1: Replace `internal/ui/components.go` entirely with:**

```go
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
```

Note: this references `footerText`, a constant that Task 9 defines in `app.go`. Do Tasks 8 and 9 in one working session — the package won't compile between them; that's fine because they're committed together in Task 9's final step. If you must commit Task 8 alone, temporarily keep the old literal string.

- [ ] **Step 2: Continue directly to Task 9** (single commit covers both).

---

### Task 9: `internal/ui` — rewrite `app.go` (fixed layout, pipeline, keys)

The heart of the rework. Fixed layout built once (`Pages` for the save overlay, zero-width flex item for hidden help). One pipeline per keystroke. Tab cycling per spec. Old `Engine`/`GetSuggestions` code deleted at the end.

**Files:**
- Modify: `internal/ui/app.go` (full replacement below)
- Delete: `internal/query/engine.go`, `internal/query/engine_test.go`, `internal/autocomplete/suggester.go`, `internal/autocomplete/suggester_test.go`

- [ ] **Step 1: Replace `internal/ui/app.go` entirely with:**

```go
package ui

import (
	"fmt"
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
}

// NewApp creates the application for the given JSON document.
func NewApp(jsonData string) *App {
	a := &App{
		tviewApp: tview.NewApplication(),
		theme:    theme.DefaultTheme(),
		data:     jsonData,
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
		a.outputPanel.SetTitle(" " + tview.Escape(shown) + " ")
		a.inputField.SetBorderColor(a.theme.BorderFocused)
	} else {
		a.outputPanel.SetTitle(fmt.Sprintf(" %s ✗ %s ", tview.Escape(shown), tview.Escape(res.UnmatchedSuffix)))
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
		switch event.Key() {
		case tcell.KeyCtrlQ, tcell.KeyCtrlC:
			a.tviewApp.Stop()
			return nil
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
	a.footer.SetText(fmt.Sprintf("[%s]%s[-]", color, tview.Escape(msg)))
	go func() {
		time.Sleep(3 * time.Second)
		a.tviewApp.QueueUpdateDraw(func() {
			a.footer.SetText(footerText)
		})
	}()
}

// Run starts the application.
func (a *App) Run() error {
	return a.tviewApp.Run()
}

// Stop stops the application.
func (a *App) Stop() {
	a.tviewApp.Stop()
}
```

- [ ] **Step 2: Delete the superseded files**

```bash
git rm internal/query/engine.go internal/query/engine_test.go internal/autocomplete/suggester.go internal/autocomplete/suggester_test.go
```

- [ ] **Step 3: Verify build, vet, and all tests**

Run: `go build ./... && go vet ./internal/... && go test ./internal/...`
Expected: builds clean, vet clean, all tests PASS. If anything still references `query.NewEngine` or `autocomplete.GetSuggestions`, the compiler will point at it — the only caller was old `app.go`.

- [ ] **Step 4: Manual smoke test**

Run: `go build -o dive . && ./dive test.json`

Check each of these:
1. Startup shows the full document colorized, title ` (root) `, top-level keys (`users  projects  company  metadata`) in the bar.
2. Type `users.0.` — suggestions show the user's keys; output shows user 0.
3. Type `na` after it — bar filters to `name  nationality`... (test.json user 0 has `name`; verify prefix filtering).
4. Press Tab repeatedly — input cycles `users.0.name` → next candidate → back to `users.0.na`; the highlighted chip in the bar follows; output updates each press; Shift+Tab goes backward; Escape mid-cycle restores `users.0.na`.
5. Type a typo `users.0.nmae` — output shows `users.0` content, title ` users.0 ✗ nmae `, red input border; delete the typo and the border recovers.
6. Type `users.#(age>30)#.name` — advanced query works, empty suggestions bar, no crash.
7. Type `users.` — the bar shows indices `0  1  2` as proper digits (the old build printed garbage runes for indices ≥ 10; `test.json` only has 3 users, so the unit test in Task 5 covers the ≥ 10 case).
8. F1 opens help split; Esc closes it; layout never jumps.
9. Ctrl+O focuses output (border highlights); arrows scroll; `i` returns to input.
10. Ctrl+Y flashes "Copied to clipboard!"; paste somewhere — plain JSON, no `[aqua]` markup.
11. Ctrl+S, accept `output.json` — footer confirms; `cat output.json` shows plain JSON. Delete it after: `rm -f output.json`.
12. Ctrl+Q quits cleanly.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/components.go internal/ui/app.go
git commit -m "feat: rework UI with fixed layout, live suggestions, Tab cycling"
```

---

### Task 10: README update and final verification

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update README.md**

Replace the **Keyboard Shortcuts** table with:

```markdown
| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Cycle through completions (back to your typed text at the end) |
| `Esc` | Cancel completion cycling / leave output panel |
| `F1` | Toggle gjson syntax help |
| `Ctrl+O` | Focus output panel for scrolling |
| `Ctrl+Y` | Copy current output to clipboard |
| `Ctrl+S` | Save output to file |
| `Ctrl+C` / `Ctrl+Q` | Quit application |
```

Replace the **Autocomplete** subsection under "Features in Detail" with:

```markdown
### Autocomplete

The suggestions bar under the query box is always visible and updates as you
type, showing the keys (or array indices) available at your current path
level. Press `Tab` to cycle through the candidates — your originally typed
text stays in the cycle, so tabbing past the last candidate brings it back.
`Shift+Tab` cycles backward and `Esc` cancels the cycle.
```

Replace the **Visual Feedback** subsection with:

```markdown
### Visual Feedback

The output panel always shows the deepest part of your path that matches the
document — a typo never blanks the view. Its title names the path being
shown, and when part of your path doesn't match (e.g. `users.0.nmae`), the
title shows where matching stopped (`users.0 ✗ nmae`) and the query border
turns red. Output JSON is syntax-highlighted.
```

Update the copy line in **Features** (`📋 **Clipboard Support** - Copy results with Ctrl+C` → `Ctrl+Y`) and the **Export Options** heading `**Copy to Clipboard (Ctrl+C)**` → `**(Ctrl+Y)**`.

Update the **Architecture** tree to the new layout:

```markdown
dive/
├── main.go                          # Entry point
├── internal/
│   ├── input/                       # JSON input handling
│   ├── path/                        # gjson path segmentation
│   ├── query/                       # deepest-valid-ancestor resolution
│   ├── autocomplete/                # suggestions + Tab-cycle state
│   ├── render/                      # pretty-printing and colorizing
│   ├── export/                      # clipboard / file export
│   └── ui/                          # tview layout and key bindings
└── test.json                        # Sample data
```

- [ ] **Step 2: Final full verification**

Run: `gofmt -l main.go internal && go vet ./internal/... && go test ./internal/...`
Expected: gofmt lists nothing; vet clean; all tests PASS.

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: update README for reworked interaction model"
```
