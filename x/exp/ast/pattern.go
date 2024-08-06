package ast

import (
	"bytes"
	"strconv"
	"strings"
)

type PatternComponent struct {
	Star  bool
	Chunk string
}

type Pattern struct {
	Comps []PatternComponent
}

func PatternFromCedar(cedar string) (Pattern, error) {
	b := []byte(cedar)

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
	}, nil
}

func (p Pattern) MarshalCedar(buf *bytes.Buffer) {
	buf.WriteRune('"')
	for _, comp := range p.Comps {
		if comp.Star {
			buf.WriteRune('*')
		}
		// TODO: This is wrong. It needs to escape unicode the Rustic way.
		quotedString := strconv.Quote(comp.Chunk)
		quotedString = quotedString[1 : len(quotedString)-1]
		quotedString = strings.Replace(quotedString, "*", "\\*", -1)
		buf.WriteString(quotedString)
	}
	buf.WriteRune('"')
}

func (p *Pattern) AddWildcard() *Pattern {
	star := PatternComponent{Star: true}
	if len(p.Comps) == 0 {
		p.Comps = []PatternComponent{star}
		return p
	}

	lastComp := p.Comps[len(p.Comps)-1]
	if lastComp.Star && lastComp.Chunk == "" {
		return p
	}

	p.Comps = append(p.Comps, star)
	return p
}

func (p *Pattern) AddLiteral(s string) *Pattern {
	if len(p.Comps) == 0 {
		p.Comps = []PatternComponent{{}}
	}

	lastComp := &p.Comps[len(p.Comps)-1]
	lastComp.Chunk = lastComp.Chunk + s
	return p
}
