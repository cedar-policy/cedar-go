package scanner

import "strings"

// IsPath reports whether s is a sequence of one or more identifiers, none of
// them reserved keywords, separated by "::". Note that this is purposely more
// strict than what the Cedar text grammar allows since it's used by the JSON
// and programmatic interfaces where whitespace and comments between path
// elements aren't allowed.
func IsPath(s string) bool {
	for part := range strings.SplitSeq(s, "::") {
		if !IsIdent(part) {
			return false
		}
	}
	return true
}

func IsActionType(s string) bool {
	if i := strings.LastIndex(s, "::"); i >= 0 {
		s = s[i+2:]
	}
	return s == "Action"
}
