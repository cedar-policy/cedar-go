package types

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/cedar-policy/cedar-go/internal/rust"
)

type PatternComponent struct {
	Wildcard bool
	Literal  string
}

// Pattern is used to define a string used for the like operator.  It does not
// conform to the Value interface, as it is not one of the Cedar types.
type Pattern struct {
	Components []PatternComponent
}

func (p Pattern) Cedar() string {
	var buf bytes.Buffer
	buf.WriteRune('"')
	for _, comp := range p.Components {
		if comp.Wildcard {
			buf.WriteRune('*')
		}
		// TODO: This is wrong. It needs to escape unicode the Rustic way.
		quotedString := strconv.Quote(comp.Literal)
		quotedString = quotedString[1 : len(quotedString)-1]
		quotedString = strings.Replace(quotedString, "*", "\\*", -1)
		buf.WriteString(quotedString)
	}
	buf.WriteRune('"')
	return buf.String()
}

func (p *Pattern) AddWildcard() *Pattern {
	star := PatternComponent{Wildcard: true}
	if len(p.Components) == 0 {
		p.Components = []PatternComponent{star}
		return p
	}

	lastComp := p.Components[len(p.Components)-1]
	if lastComp.Wildcard && lastComp.Literal == "" {
		return p
	}

	p.Components = append(p.Components, star)
	return p
}

func (p *Pattern) AddLiteral(s string) *Pattern {
	if len(p.Components) == 0 {
		p.Components = []PatternComponent{{}}
	}

	lastComp := &p.Components[len(p.Components)-1]
	lastComp.Literal = lastComp.Literal + s
	return p
}

// TODO: move this into the types package

// ported from Go's stdlib and reduced to our scope.
// https://golang.org/src/path/filepath/match.go?s=1226:1284#L34

// Match reports whether name matches the shell file name pattern.
// The pattern syntax is:
//
//	pattern:
//		{ term }
//	term:
//		'*'         matches any sequence of non-Separator characters
//		c           matches character c (c != '*')
func (p Pattern) Match(arg string) (matched bool) {
Pattern:
	for i, comp := range p.Components {
		lastChunk := i == len(p.Components)-1
		if comp.Wildcard && comp.Literal == "" {
			return true
		}
		// Look for Match at current position.
		t, ok := matchChunk(comp.Literal, arg)
		// if we're the last chunk, make sure we've exhausted the name
		// otherwise we'll give a false result even if we could still Match
		// using the star
		if ok && (len(t) == 0 || !lastChunk) {
			arg = t
			continue
		}
		if comp.Wildcard {
			// Look for Match skipping i+1 bytes.
			for i := 0; i < len(arg); i++ {
				t, ok := matchChunk(comp.Literal, arg[i+1:])
				if ok {
					// if we're the last chunk, make sure we exhausted the name
					if lastChunk && len(t) > 0 {
						continue
					}
					arg = t
					continue Pattern
				}
			}
		}
		return false
	}
	return len(arg) == 0
}

// matchChunk checks whether chunk matches the beginning of s.
// If so, it returns the remainder of s (after the Match).
// Chunk is all single-character operators: literals, char classes, and ?.
func matchChunk(chunk, s string) (rest string, ok bool) {
	for len(chunk) > 0 {
		if len(s) == 0 {
			return
		}
		if chunk[0] != s[0] {
			return
		}
		s = s[1:]
		chunk = chunk[1:]
	}
	return s, true
}

func ParsePattern(s string) (Pattern, error) {
	b := []byte(s)

	var comps []PatternComponent
	for len(b) > 0 {
		var comp PatternComponent
		var err error
		for len(b) > 0 && b[0] == '*' {
			b = b[1:]
			comp.Wildcard = true
		}
		comp.Literal, b, err = rust.Unquote(b, true)
		if err != nil {
			return Pattern{}, err
		}
		comps = append(comps, comp)
	}
	return Pattern{
		Components: comps,
	}, nil
}
