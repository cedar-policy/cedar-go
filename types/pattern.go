package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

var errJSONInvalidPatternComponent = fmt.Errorf("invalid pattern component")

type patternComponent struct {
	Wildcard bool
	Literal  string
}

// Pattern is used to define a string used for the like operator.  It does not
// conform to the Value interface, as it is not one of the Cedar types.
type Pattern struct {
	comps []patternComponent
}

// Wildcard is a type which is used as a component to NewPattern.
type Wildcard struct{}

// NewPattern permits for the programmatic construction of a Pattern out of a slice of pattern components.
// The pattern components may be one of string, types.String, or types.Wildcard.  Any other types will
// cause a panic.
func NewPattern(components ...any) Pattern {
	var comps []patternComponent
	for _, c := range components {
		switch v := c.(type) {
		case string:
			if len(comps) == 0 {
				comps = []patternComponent{{Wildcard: false, Literal: ""}}
			}
			comps[len(comps)-1].Literal += string(v)
		case String:
			if len(comps) == 0 {
				comps = []patternComponent{{Wildcard: false, Literal: ""}}
			}
			comps[len(comps)-1].Literal += string(v)
		case Wildcard:
			if len(comps) == 0 || comps[len(comps)-1].Literal != "" {
				comps = append(comps, patternComponent{Wildcard: true, Literal: ""})
			}
		default:
			panic(fmt.Sprintf("unexpected component type: %T", v))
		}
	}
	return Pattern{comps: comps}
}

func (p Pattern) MarshalCedar() []byte {
	var buf bytes.Buffer
	buf.WriteRune('"')
	for _, comp := range p.comps {
		if comp.Wildcard {
			buf.WriteRune('*')
		}
		// TODO: This is wrong. It needs to escape unicode the Rustic way.
		quotedString := strconv.Quote(comp.Literal)
		quotedString = quotedString[1 : len(quotedString)-1]
		quotedString = strings.ReplaceAll(quotedString, "*", "\\*")
		buf.WriteString(quotedString)
	}
	buf.WriteRune('"')
	return buf.Bytes()
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
func (p Pattern) Match(arg String) (matched bool) {
Pattern:
	for i, comp := range p.comps {
		lastChunk := i == len(p.comps)-1
		if comp.Wildcard && comp.Literal == "" {
			return true
		}
		// Look for Match at current position.
		t, ok := matchChunk(comp.Literal, string(arg))
		// if we're the last chunk, make sure we've exhausted the name
		// otherwise we'll give a false result even if we could still Match
		// using the star
		if ok && (len(t) == 0 || !lastChunk) {
			arg = String(t)
			continue
		}
		if comp.Wildcard {
			// Look for Match skipping i+1 bytes.
			for i := 0; i < len(arg); i++ {
				t, ok := matchChunk(comp.Literal, string(arg[i+1:]))
				if ok {
					// if we're the last chunk, make sure we exhausted the name
					if lastChunk && len(t) > 0 {
						continue
					}
					arg = String(t)
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

func (p Pattern) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteRune('[')
	for i, comp := range p.comps {
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

		if i < len(p.comps)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteRune(']')
	return buf.Bytes(), nil
}

func (p *Pattern) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	var objs []any
	if err := dec.Decode(&objs); err != nil {
		return err
	}

	if len(objs) == 0 {
		return fmt.Errorf(`%w: must provide at least one pattern component`, errJSONInvalidPatternComponent)
	}

	var comps []any
	for _, comp := range objs {
		switch v := comp.(type) {
		case string:
			if v != "Wildcard" {
				return fmt.Errorf(`%w: invalid component string "%v"`, errJSONInvalidPatternComponent, v)
			}
			comps = append(comps, Wildcard{})
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

			comps = append(comps, String(literalStr))
		default:
			return fmt.Errorf(`%w: unknown component type`, errJSONInvalidPatternComponent)
		}
	}

	*p = NewPattern(comps...)
	return nil
}
