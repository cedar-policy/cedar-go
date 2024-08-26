package parser

import (
	"github.com/cedar-policy/cedar-go/internal/rust"
	"github.com/cedar-policy/cedar-go/types"
)

// ParsePattern will parse an unquoted rust-style string with \*'s in it.
func ParsePattern(v string) (types.Pattern, error) {
	b := []byte(v)
	var comps []any
	for len(b) > 0 {
		for len(b) > 0 && b[0] == '*' {
			b = b[1:]
			comps = append(comps, types.Wildcard{})
		}
		var err error
		var literal string
		literal, b, err = rust.Unquote(b, true)
		if err != nil {
			return types.Pattern{}, err
		}
		comps = append(comps, types.String(literal))
	}
	return types.NewPattern(comps...), nil
}
