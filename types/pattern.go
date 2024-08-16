package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cedar-policy/cedar-go/internal/rust"
)

type patternComponent struct {
	Wildcard bool
	Literal  string
}

// Pattern is used to define a string used for the like operator.  It does not
// conform to the Value interface, as it is not one of the Cedar types.
type Pattern struct {
	comps []patternComponent
}

var WildcardPattern = newPattern([]patternComponent{{Wildcard: true}})

func newPattern(comps []patternComponent) Pattern {
	return Pattern{comps: comps}
}

func LiteralPattern(literal string) Pattern {
	if literal == "" {
		return newPattern(nil)
	}
	return newPattern([]patternComponent{{Wildcard: false, Literal: literal}})
}

func (p Pattern) Cedar() string {
	var buf bytes.Buffer
	buf.WriteRune('"')
	for _, comp := range p.comps {
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
	for i, comp := range p.comps {
		lastChunk := i == len(p.comps)-1
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
	var comps []patternComponent
	for len(b) > 0 {
		var comp patternComponent
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

	*p = Pattern{comps: comps}
	return nil
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
	var comps []any
	if err := dec.Decode(&comps); err != nil {
		return err
	}

	if len(comps) == 0 {
		return fmt.Errorf(`%w: must provide at least one pattern component`, errJSONInvalidPatternComponent)
	}

	pb := PatternBuilder{}
	for _, comp := range comps {
		switch v := comp.(type) {
		case string:
			if v != "Wildcard" {
				return fmt.Errorf(`%w: invalid component string "%v"`, errJSONInvalidPatternComponent, v)
			}
			pb = pb.AddWildcard()
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

			pb = pb.AddLiteral(literalStr)
		default:
			return fmt.Errorf(`%w: unknown component type`, errJSONInvalidPatternComponent)
		}
	}

	*p = pb.Build()
	return nil
}

// PatternBuilder can be used to programmatically build a Cedar pattern string, like so:
// PatternBuilder{}.AddWildcard().AddLiteral("foo").AddWildcard().Build()
type PatternBuilder []patternComponent

func (p PatternBuilder) AddWildcard() PatternBuilder {
	star := patternComponent{Wildcard: true}
	if len(p) == 0 {
		return PatternBuilder{star}
	}

	lastComp := (p)[len(p)-1]
	if lastComp.Wildcard && lastComp.Literal == "" {
		return p
	}

	return append(p, star)
}

func (p PatternBuilder) AddLiteral(s string) PatternBuilder {
	if len(p) == 0 {
		p = PatternBuilder{patternComponent{}}
	}
	p[len(p)-1].Literal += s
	return p
}

func (p PatternBuilder) Build() Pattern {
	return newPattern(p)
}
