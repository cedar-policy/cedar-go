package types

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/go-json-experiment/json"

	"github.com/cedar-policy/cedar-go/internal/rust"
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
			comps = appendLiteral(comps, v)
		case String:
			comps = appendLiteral(comps, string(v))
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

func appendLiteral(comps []patternComponent, s string) []patternComponent {
	if s == "" {
		return comps
	}
	if len(comps) == 0 {
		return []patternComponent{{Literal: s}}
	}
	comps[len(comps)-1].Literal += s
	return comps
}

// Equal reports whether p and other are the same pattern.
func (p Pattern) Equal(other Pattern) bool {
	return slices.Equal(p.comps, other.comps)
}

func (p Pattern) MarshalCedar() []byte {
	var buf bytes.Buffer
	buf.WriteRune('"')
	for _, comp := range p.comps {
		if comp.Wildcard {
			buf.WriteRune('*')
		}
		// EscapeCharAll (not EscapeString) because Rust Cedar escapes each
		// pattern character independently via char::escape_debug(), so grapheme
		// extend chars are always escaped rather than passed through raw in
		// continuation positions.
		quotedString := rust.EscapeCharAll(comp.Literal)
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
func (p Pattern) Match(arg String) bool {
	s := string(arg)
	for i, comp := range p.comps {
		var ok bool
		switch {
		case !comp.Wildcard:
			// Only a pattern's first component has no wildcard before it.
			s, ok = strings.CutPrefix(s, comp.Literal)
		case comp.Literal == "":
			// A trailing wildcard matches whatever is left.
			return true
		case i == len(p.comps)-1:
			return strings.HasSuffix(s, comp.Literal)
		default:
			_, s, ok = strings.Cut(s, comp.Literal)
		}
		if !ok {
			return false
		}
	}
	return len(s) == 0
}

func (p Pattern) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteRune('[')
	for i, comp := range p.comps {
		if comp.Wildcard {
			buf.WriteString(`"Wildcard"`)
		}

		if !comp.Wildcard || comp.Literal != "" {
			if comp.Wildcard {
				buf.WriteString(", ")
			}
			kb, err := json.Marshal(comp.Literal)
			if err != nil {
				return nil, err
			}
			buf.WriteString(`{"Literal":`)
			buf.Write(kb)
			buf.WriteByte('}')
		}

		if i < len(p.comps)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteRune(']')
	return buf.Bytes(), nil
}

func (p *Pattern) UnmarshalJSON(b []byte) error {
	var objs []any
	if err := json.Unmarshal(b, &objs); err != nil {
		return err
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
