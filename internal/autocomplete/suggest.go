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
