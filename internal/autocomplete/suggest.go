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
// Note that a key containing a literal dot (e.g. "fav.movie") is only
// completed while the user types its escaped form ("fav\."); typing the
// unescaped dot parses as a separator and suggestions go blank.
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

	type candidate struct {
		key     string // escaped path segment
		display string // unescaped label
	}
	var cands []candidate
	switch {
	case parent.IsObject():
		parent.ForEach(func(key, _ gjson.Result) bool {
			raw := key.String()
			cands = append(cands, candidate{key: path.EscapeKey(raw), display: raw})
			return true
		})
	case parent.IsArray():
		for i := range parent.Array() {
			idx := strconv.Itoa(i)
			cands = append(cands, candidate{key: idx, display: idx})
		}
	default:
		return nil
	}

	// Indexing into a query/fan-out result only works through gjson's
	// pipe: "a.#(q)#.0" yields an empty artifact, "a.#(q)#|0" the element.
	// Fan-out persists through trailing plain key/index segments, so
	// anchoring is decided over the whole parent tail via path.FanoutRun.
	sep := "."
	if _, anchored := path.FanoutRun(parentSegs); anchored && parent.IsArray() {
		sep = "|"
	}

	var out []Suggestion
	for _, c := range cands {
		if !strings.HasPrefix(c.key, partial) {
			continue
		}
		full := c.key
		if parentPath != "" {
			full = parentPath + sep + c.key
		}
		out = append(out, Suggestion{Full: full, Display: c.display})
	}
	return out
}
