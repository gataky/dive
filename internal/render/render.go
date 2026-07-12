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
