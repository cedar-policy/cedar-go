package types

import (
	"bytes"
	"encoding/json"
	"fmt"
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
type Pattern []PatternComponent

func (p Pattern) Cedar() string {
	var buf bytes.Buffer
	buf.WriteRune('"')
	for _, comp := range p {
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

func (p Pattern) Wildcard() Pattern {
	star := PatternComponent{Wildcard: true}
	if len(p) == 0 {
		p = Pattern{star}
		return p
	}

	lastComp := p[len(p)-1]
	if lastComp.Wildcard && lastComp.Literal == "" {
		return p
	}

	p = append(p, star)
	return p
}

func (p Pattern) Literal(s string) Pattern {
	if len(p) == 0 {
		p = Pattern{{}}
	}
	p[len(p)-1].Literal += s
	return p
}

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
	for i, comp := range p {
		lastChunk := i == len(p)-1
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

func (p *Pattern) UnmarshalCedar(b []byte) error {
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
			return err
		}
		comps = append(comps, comp)
	}

	*p = comps
	return nil
}

func (p Pattern) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteRune('[')
	for i, comp := range p {
		if comp.Wildcard {
			buf.WriteString(`"Wildcard"`)
		}

		if comp.Literal != "" {
			if comp.Wildcard {
				buf.WriteString(", ")
			}
			buf.WriteString(`{"Literal":"`)
			buf.WriteString(comp.Literal)
			buf.WriteString(`"}`)
		}

		if i < len(p)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteRune(']')
	return buf.Bytes(), nil
}

func (p *Pattern) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	var comps []any
	if err := dec.Decode(&comps); err != nil {
		return err
	}

	if len(comps) == 0 {
		return fmt.Errorf(`%w: must provide at least one pattern component`, errJSONInvalidPatternComponent)
	}

	newPattern := Pattern{}
	for _, comp := range comps {
		switch v := comp.(type) {
		case string:
			if v != "Wildcard" {
				return fmt.Errorf(`%w: invalid component string "%v"`, errJSONInvalidPatternComponent, v)
			}
			newPattern = newPattern.Wildcard()
		case map[string]any:
			if len(v) != 1 {
				return fmt.Errorf(`%w: too many keys in literal object`, errJSONInvalidPatternComponent)
			}

			literal, ok := v["Literal"]
			if !ok {
				return fmt.Errorf(`%w: missing "Literal" key in literal object`, errJSONInvalidPatternComponent)
			}

			literalStr, ok := literal.(string)
			if !ok {
				return fmt.Errorf(`%w: invalid "Literal" value "%v"`, errJSONInvalidPatternComponent, literal)
			}

			newPattern = newPattern.Literal(literalStr)
		default:
			return fmt.Errorf(`%w: unknown component type`, errJSONInvalidPatternComponent)
		}
	}

	*p = newPattern
	return nil
}
