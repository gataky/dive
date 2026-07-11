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
		r := gjson.Get(data, candidate)
		if !r.Exists() {
			continue
		}
		// A plain key fanned out over an array query yields an empty array
		// even when the key exists nowhere, so an empty fan-out result is
		// treated as unmatched; a genuinely empty array (plain-key parent)
		// is a valid resolution.
		if i >= 2 && segs[i-1].Kind == path.KindKey && segs[i-2].Kind == path.KindAdvanced &&
			r.IsArray() && len(r.Array()) == 0 {
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
