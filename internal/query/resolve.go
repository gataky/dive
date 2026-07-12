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
// count as unmatched. data must be valid JSON — the application validates
// input at startup. With empty or invalid data the fallback Result will
// not Exist.
func Resolve(data, p string) Resolution {
	segs := path.Split(p)
	if n := len(segs); n > 0 && segs[n-1].Raw == "" {
		segs = segs[:n-1]
	}
	for i := len(segs); i > 0; i-- {
		candidate := path.Join(segs[:i])
		r := gjson.Get(data, candidate)
		if !r.Exists() {
			continue
		}
		if isEmptyFanoutArtifact(segs, i, r) {
			continue
		}
		return Resolution{
			ResolvedPath:    candidate,
			UnmatchedSuffix: path.Join(segs[i:]),
			Result:          r,
		}
	}
	return Resolution{
		ResolvedPath:    "",
		UnmatchedSuffix: path.Join(segs),
		Result:          gjson.Parse(data),
	}
}

// isEmptyFanoutArtifact reports whether candidate segs[:i] ends in a run
// of plain keys and indexes following an advanced segment and r is an
// empty array. gjson fans plain keys and indexes out over array-query
// results, yielding an empty array even when the key exists nowhere, so
// such a result is treated as unmatched rather than a valid resolution.
// A genuinely empty array (no fan-out context) still resolves. The tail
// analysis is delegated to path.FanoutRun.
func isEmptyFanoutArtifact(segs []path.Segment, i int, r gjson.Result) bool {
	runLen, anchored := path.FanoutRun(segs[:i])
	return anchored && runLen >= 1 && r.IsArray() && len(r.Array()) == 0
}
