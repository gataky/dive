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
		// If the last segment is a simple key on an array result, verify it's not empty.
		// gjson returns an empty array when accessing non-existent fields on arrays,
		// but we should treat that as unmatched.
		if i > 0 && segs[i-1].Kind == path.KindKey && r.IsArray() && r.String() == "[]" {
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
