package ast

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
