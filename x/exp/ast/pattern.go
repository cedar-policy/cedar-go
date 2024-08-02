package ast

import "strings"

type PatternComponent struct {
	Star  bool
	Chunk string
}

type Pattern struct {
	Comps []PatternComponent
	Raw   string
}

func (p Pattern) String() string {
	return p.Raw
}

func NewPattern(literal string) (Pattern, error) {
	rawPat := literal

	literal = strings.TrimPrefix(literal, "\"")
	literal = strings.TrimSuffix(literal, "\"")

	b := []byte(literal)

	var comps []PatternComponent
	for len(b) > 0 {
		var comp PatternComponent
		var err error
		for len(b) > 0 && b[0] == '*' {
			b = b[1:]
			comp.Star = true
		}
		comp.Chunk, b, err = rustUnquote(b, true)
		if err != nil {
			return Pattern{}, err
		}
		comps = append(comps, comp)
	}
	return Pattern{
		Comps: comps,
		Raw:   rawPat,
	}, nil
}
