package ast

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

func policyFromCedar(p *parser) (*Policy, error) {
	annotations, err := p.annotations()
	if err != nil {
		return nil, err
	}

	policy, err := p.effect(&annotations)
	if err != nil {
		return nil, err
	}

	if err = p.exact("("); err != nil {
		return nil, err
	}
	if err = p.principal(policy); err != nil {
		return nil, err
	}
	if err = p.exact(","); err != nil {
		return nil, err
	}
	if err = p.action(policy); err != nil {
		return nil, err
	}
	if err = p.exact(","); err != nil {
		return nil, err
	}
	if err = p.resource(policy); err != nil {
		return nil, err
	}
	if err = p.exact(")"); err != nil {
		return nil, err
	}
	// if res.Conditions, err = p.conditions(); err != nil {
	// 	return res, err
	// }
	if err = p.exact(";"); err != nil {
		return nil, err
	}

	return policy, nil
}

type parser struct {
	tokens []Token
	pos    int
}

func newParser(tokens []Token) parser {
	return parser{tokens: tokens, pos: 0}
}

func (p *parser) advance() Token {
	t := p.peek()
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}
	return t
}

func (p *parser) peek() Token {
	return p.tokens[p.pos]
}

func (p *parser) exact(tok string) error {
	t := p.advance()
	if t.Text != tok {
		return p.errorf("exact got %v want %v", t.Text, tok)
	}
	return nil
}

func (p *parser) errorf(s string, args ...interface{}) error {
	var t Token
	if p.pos < len(p.tokens) {
		t = p.tokens[p.pos]
	}
	err := fmt.Errorf(s, args...)
	return fmt.Errorf("parse error at %v %q: %w", t.Pos, t.Text, err)
}

func (p *parser) annotations() (Annotations, error) {
	var res Annotations
	for p.peek().Text == "@" {
		p.advance()
		err := p.annotation(&res)
		if err != nil {
			return res, err
		}
	}
	return res, nil

}

func (p *parser) annotation(a *Annotations) error {
	var err error
	t := p.advance()
	if !t.isIdent() {
		return p.errorf("expected ident")
	}
	name := types.String(t.Text)
	if err = p.exact("("); err != nil {
		return err
	}
	t = p.advance()
	if !t.isString() {
		return p.errorf("expected string")
	}
	value, err := t.stringValue()
	if err != nil {
		return err
	}
	if err = p.exact(")"); err != nil {
		return err
	}

	a.Annotation(name, types.String(value))
	return nil
}

func (p *parser) effect(a *Annotations) (*Policy, error) {
	next := p.advance()
	if next.Text == "permit" {
		return a.Permit(), nil
	} else if next.Text == "forbid" {
		return a.Forbid(), nil
	}

	return nil, p.errorf("unexpected effect: %v", next.Text)
}

func (p *parser) principal(policy *Policy) error {
	if err := p.exact("principal"); err != nil {
		return err
	}
	switch p.peek().Text {
	case "==":
		p.advance()
		entity, err := p.entity()
		if err != nil {
			return err
		}
		policy.PrincipalEq(entity)
		return nil
	case "is":
		p.advance()
		path, err := p.path()
		if err != nil {
			return err
		}
		if p.peek().Text == "in" {
			p.advance()
			entity, err := p.entity()
			if err != nil {
				return err
			}
			policy.PrincipalIsIn(path, entity)
			return nil
		}

		policy.PrincipalIs(path)
		return nil
	case "in":
		p.advance()
		entity, err := p.entity()
		if err != nil {
			return err
		}
		policy.PrincipalIn(entity)
		return nil
	}

	return nil
}

func (p *parser) entity() (types.EntityUID, error) {
	var res types.EntityUID
	var err error
	t := p.advance()
	if !t.isIdent() {
		return res, p.errorf("expected ident")
	}
	res.Type = t.Text
	for {
		if err := p.exact("::"); err != nil {
			return res, err
		}
		t := p.advance()
		switch {
		case t.isIdent():
			res.Type = fmt.Sprintf("%v::%v", res.Type, t.Text)
		case t.isString():
			res.ID, err = t.stringValue()
			if err != nil {
				return res, err
			}
			return res, nil
		default:
			return res, p.errorf("unexpected token")
		}
	}
}

func (p *parser) path() (types.String, error) {
	var res types.String
	t := p.advance()
	if !t.isIdent() {
		return res, p.errorf("expected ident")
	}
	res = types.String(t.Text)
	for {
		if p.peek().Text != "::" {
			return res, nil
		}
		p.advance()
		t := p.advance()
		switch {
		case t.isIdent():
			res = types.String(fmt.Sprintf("%v::%v", res, t.Text))
		default:
			return res, p.errorf("unexpected token")
		}
	}
}

func (p *parser) action(policy *Policy) error {
	if err := p.exact("action"); err != nil {
		return err
	}
	switch p.peek().Text {
	case "==":
		p.advance()
		entity, err := p.entity()
		if err != nil {
			return err
		}
		policy.ActionEq(entity)
		return nil
	case "in":
		p.advance()
		if p.peek().Text == "[" {
			p.advance()
			entities, err := p.entlist()
			if err != nil {
				return err
			}
			policy.ActionIn(entities...)
			p.advance() // entlist guarantees "]"
			return nil
		} else {
			entity, err := p.entity()
			if err != nil {
				return err
			}
			policy.ActionIn(entity)
			return nil
		}
	}

	return nil
}

func (p *parser) entlist() ([]types.EntityUID, error) {
	var res []types.EntityUID
	for p.peek().Text != "]" {
		if len(res) > 0 {
			if err := p.exact(","); err != nil {
				return nil, err
			}
		}
		e, err := p.entity()
		if err != nil {
			return nil, err
		}
		res = append(res, e)
	}
	return res, nil
}

func (p *parser) resource(policy *Policy) error {
	if err := p.exact("resource"); err != nil {
		return err
	}
	switch p.peek().Text {
	case "==":
		p.advance()
		entity, err := p.entity()
		if err != nil {
			return err
		}
		policy.ResourceEq(entity)
		return nil
	case "is":
		p.advance()
		path, err := p.path()
		if err != nil {
			return err
		}
		if p.peek().Text == "in" {
			p.advance()
			entity, err := p.entity()
			if err != nil {
				return err
			}
			policy.ResourceIsIn(path, entity)
			return nil
		}

		policy.ResourceIs(path)
		return nil
	case "in":
		p.advance()
		entity, err := p.entity()
		if err != nil {
			return err
		}
		policy.ResourceIn(entity)
		return nil
	}

	return nil
}
