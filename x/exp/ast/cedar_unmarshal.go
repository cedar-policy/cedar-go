package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cedar-policy/cedar-go/types"
)

func (p *PolicySet) UnmarshalCedar(b []byte) error {
	tokens, err := Tokenize(b)
	if err != nil {
		return err
	}

	i := 0

	policySet := PolicySet{}
	parser := newParser(tokens)
	for !parser.peek().isEOF() {
		pos := parser.peek().Pos
		policy := Policy{
			principal: scopeTypeAll{},
			action:    scopeTypeAll{},
			resource:  scopeTypeAll{},
		}

		if err = policy.fromCedarWithParser(&parser); err != nil {
			return err
		}

		policyName := fmt.Sprintf("policy%v", i)
		policySet[policyName] = PolicySetEntry{Policy: policy, Position: pos}
		i++
	}

	*p = policySet
	return nil
}

func (p *Policy) UnmarshalCedar(b []byte) error {
	tokens, err := Tokenize(b)
	if err != nil {
		return err
	}

	parser := newParser(tokens)
	return p.fromCedarWithParser(&parser)
}

func (p *Policy) fromCedarWithParser(parser *parser) error {
	annotations, err := parser.annotations()
	if err != nil {
		return err
	}

	newPolicy, err := parser.effect(&annotations)
	if err != nil {
		return err
	}

	if err = parser.exact("("); err != nil {
		return err
	}
	if err = parser.principal(newPolicy); err != nil {
		return err
	}
	if err = parser.exact(","); err != nil {
		return err
	}
	if err = parser.action(newPolicy); err != nil {
		return err
	}
	if err = parser.exact(","); err != nil {
		return err
	}
	if err = parser.resource(newPolicy); err != nil {
		return err
	}
	if err = parser.exact(")"); err != nil {
		return err
	}
	if err = parser.conditions(newPolicy); err != nil {
		return err
	}
	if err = parser.exact(";"); err != nil {
		return err
	}

	*p = *newPolicy
	return nil
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
	known := map[types.String]struct{}{}
	for p.peek().Text == "@" {
		p.advance()
		err := p.annotation(&res, known)
		if err != nil {
			return res, err
		}
	}
	return res, nil

}

func (p *parser) annotation(a *Annotations, known map[types.String]struct{}) error {
	var err error
	t := p.advance()
	if !t.isIdent() {
		return p.errorf("expected ident")
	}
	name := types.String(t.Text)
	if err = p.exact("("); err != nil {
		return err
	}
	if _, ok := known[name]; ok {
		return p.errorf("duplicate annotation: @%s", name)
	}
	known[name] = struct{}{}
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
	t := p.advance()
	if !t.isIdent() {
		return res, p.errorf("expected ident")
	}
	return p.entityFirstPathPreread(t.Text)
}

func (p *parser) entityFirstPathPreread(firstPath string) (types.EntityUID, error) {
	var res types.EntityUID
	var err error
	res.Type = firstPath
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

func (p *parser) pathFirstPathPreread(firstPath string) (types.Path, error) {
	res := types.Path(firstPath)
	for {
		if p.peek().Text != "::" {
			return res, nil
		}
		p.advance()
		t := p.advance()
		switch {
		case t.isIdent():
			res = types.Path(fmt.Sprintf("%v::%v", res, t.Text))
		default:
			return res, p.errorf("unexpected token")
		}
	}
}

func (p *parser) path() (types.Path, error) {
	t := p.advance()
	if !t.isIdent() {
		return "", p.errorf("expected ident")
	}
	return p.pathFirstPathPreread(t.Text)
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
			policy.ActionInSet(entities...)
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

func (p *parser) conditions(policy *Policy) error {
	for {
		switch p.peek().Text {
		case "when":
			p.advance()
			expr, err := p.condition()
			if err != nil {
				return err
			}
			policy.When(expr)
		case "unless":
			p.advance()
			expr, err := p.condition()
			if err != nil {
				return err
			}
			policy.Unless(expr)
		default:
			return nil
		}
	}
}

func (p *parser) condition() (Node, error) {
	var res Node
	var err error
	if err := p.exact("{"); err != nil {
		return res, err
	}
	if res, err = p.expression(); err != nil {
		return res, err
	}
	if err := p.exact("}"); err != nil {
		return res, err
	}
	return res, nil
}

func (p *parser) expression() (Node, error) {
	t := p.peek()
	if t.Text == "if" {
		p.advance()

		condition, err := p.expression()
		if err != nil {
			return Node{}, err
		}

		if err = p.exact("then"); err != nil {
			return Node{}, err
		}
		ifTrue, err := p.expression()
		if err != nil {
			return Node{}, err
		}

		if err = p.exact("else"); err != nil {
			return Node{}, err
		}
		ifFalse, err := p.expression()
		if err != nil {
			return Node{}, err
		}

		return If(condition, ifTrue, ifFalse), nil
	}

	return p.or()
}

func (p *parser) or() (Node, error) {
	lhs, err := p.and()
	if err != nil {
		return Node{}, err
	}

	for p.peek().Text == "||" {
		p.advance()
		rhs, err := p.and()
		if err != nil {
			return Node{}, err
		}
		lhs = lhs.Or(rhs)
	}

	return lhs, nil
}

func (p *parser) and() (Node, error) {
	lhs, err := p.relation()
	if err != nil {
		return Node{}, err
	}

	for p.peek().Text == "&&" {
		p.advance()
		rhs, err := p.relation()
		if err != nil {
			return Node{}, err
		}
		lhs = lhs.And(rhs)
	}

	return lhs, nil
}

func (p *parser) relation() (Node, error) {
	lhs, err := p.add()
	if err != nil {
		return Node{}, err
	}

	t := p.peek()

	if t.Text == "has" {
		p.advance()
		return p.has(lhs)
	} else if t.Text == "like" {
		p.advance()
		return p.like(lhs)
	} else if t.Text == "is" {
		p.advance()
		return p.is(lhs)
	}

	// RELOP
	var operator func(Node, Node) Node
	switch t.Text {
	case "<":
		operator = Node.LessThan
	case "<=":
		operator = Node.LessThanOrEqual
	case ">":
		operator = Node.GreaterThan
	case ">=":
		operator = Node.GreaterThanOrEqual
	case "!=":
		operator = Node.NotEquals
	case "==":
		operator = Node.Equals
	case "in":
		operator = Node.In
	default:
		return lhs, nil

	}

	p.advance()
	rhs, err := p.add()
	if err != nil {
		return Node{}, err
	}
	return operator(lhs, rhs), nil
}

func (p *parser) has(lhs Node) (Node, error) {
	t := p.advance()
	if t.isIdent() {
		return lhs.Has(t.Text), nil
	} else if t.isString() {
		str, err := t.stringValue()
		if err != nil {
			return Node{}, err
		}
		return lhs.Has(str), nil
	}
	return Node{}, p.errorf("expected ident or string")
}

func (p *parser) like(lhs Node) (Node, error) {
	t := p.advance()
	if !t.isString() {
		return Node{}, p.errorf("expected string literal")
	}
	patternRaw := t.Text
	patternRaw = strings.TrimPrefix(patternRaw, "\"")
	patternRaw = strings.TrimSuffix(patternRaw, "\"")
	pattern, err := types.ParsePattern(patternRaw)
	if err != nil {
		return Node{}, err
	}
	return lhs.Like(pattern), nil
}

func (p *parser) is(lhs Node) (Node, error) {
	entityType, err := p.path()
	if err != nil {
		return Node{}, err
	}
	if p.peek().Text == "in" {
		p.advance()
		inEntity, err := p.add()
		if err != nil {
			return Node{}, err
		}
		return lhs.IsIn(entityType, inEntity), nil
	}
	return lhs.Is(entityType), nil
}

func (p *parser) add() (Node, error) {
	lhs, err := p.mult()
	if err != nil {
		return Node{}, err
	}

	for {
		t := p.peek()
		var operator func(Node, Node) Node
		switch t.Text {
		case "+":
			operator = Node.Plus
		case "-":
			operator = Node.Minus
		}

		if operator == nil {
			break
		}

		p.advance()
		rhs, err := p.mult()
		if err != nil {
			return Node{}, err
		}
		lhs = operator(lhs, rhs)
	}

	return lhs, nil
}

func (p *parser) mult() (Node, error) {
	lhs, err := p.unary()
	if err != nil {
		return Node{}, err
	}

	for p.peek().Text == "*" {
		p.advance()
		rhs, err := p.unary()
		if err != nil {
			return Node{}, err
		}
		lhs = lhs.Times(rhs)
	}

	return lhs, nil
}

func (p *parser) unary() (Node, error) {
	var ops []bool
	for {
		opToken := p.peek()
		if opToken.Text != "-" && opToken.Text != "!" {
			break
		}
		p.advance()
		ops = append(ops, opToken.Text == "-")
	}

	var res Node

	// special case for max negative long
	tok := p.peek()
	if len(ops) > 0 && ops[len(ops)-1] && tok.isInt() {
		p.advance()
		i, err := strconv.ParseInt("-"+tok.Text, 10, 64)
		if err != nil {
			return Node{}, err
		}
		res = Long(types.Long(i))
		ops = ops[:len(ops)-1]
	} else {
		var err error
		res, err = p.member()
		if err != nil {
			return res, err
		}
	}

	for i := len(ops) - 1; i >= 0; i-- {
		if ops[i] {
			res = Negate(res)
		} else {
			res = Not(res)
		}
	}
	return res, nil
}

func (p *parser) member() (Node, error) {
	res, err := p.primary()
	if err != nil {
		return res, err
	}
	for {
		var ok bool
		res, ok, err = p.access(res)
		if !ok {
			return res, err
		}
	}
}

func (p *parser) primary() (Node, error) {
	var res Node
	t := p.advance()
	switch {
	case t.isInt():
		i, err := t.intValue()
		if err != nil {
			return res, err
		}
		res = Long(types.Long(i))
	case t.isString():
		str, err := t.stringValue()
		if err != nil {
			return res, err
		}
		res = String(types.String(str))
	case t.Text == "true":
		res = True()
	case t.Text == "false":
		res = False()
	case t.Text == "principal":
		res = Principal()
	case t.Text == "action":
		res = Action()
	case t.Text == "resource":
		res = Resource()
	case t.Text == "context":
		res = Context()
	case t.isIdent():
		return p.entityOrExtFun(t.Text)
	case t.Text == "(":
		expr, err := p.expression()
		if err != nil {
			return res, err
		}
		if err := p.exact(")"); err != nil {
			return res, err
		}
		res = expr
	case t.Text == "[":
		set, err := p.expressions("]")
		if err != nil {
			return res, err
		}
		p.advance() // expressions guarantees "]"
		res = SetNodes(set...)
	case t.Text == "{":
		record, err := p.record()
		if err != nil {
			return res, err
		}
		res = record
	default:
		return res, p.errorf("invalid primary")
	}
	return res, nil
}

func (p *parser) entityOrExtFun(ident string) (Node, error) {
	var res types.EntityUID
	var err error
	res.Type = ident
	for {
		t := p.advance()
		switch t.Text {
		case "::":
			t := p.advance()
			switch {
			case t.isIdent():
				res.Type = fmt.Sprintf("%v::%v", res.Type, t.Text)
			case t.isString():
				res.ID, err = t.stringValue()
				if err != nil {
					return Node{}, err
				}
				return EntityUID(res), nil
			default:
				return Node{}, p.errorf("unexpected token")
			}
		case "(":
			args, err := p.expressions(")")
			if err != nil {
				return Node{}, err
			}
			p.advance()
			// i, ok := extMap[types.String(res.Type)]
			// if !ok {
			// 	return Node{}, p.errorf("`%v` is not a function", res.Type)
			// }
			// if i.IsMethod {
			// 	return Node{}, p.errorf("`%v` is a method, not a function", res.Type)
			// }
			return ExtensionCall(types.String(res.Type), args...), nil
		default:
			return Node{}, p.errorf("unexpected token")
		}
	}
}

func (p *parser) expressions(endOfListMarker string) ([]Node, error) {
	var res []Node
	for p.peek().Text != endOfListMarker {
		if len(res) > 0 {
			if err := p.exact(","); err != nil {
				return res, err
			}
		}
		e, err := p.expression()
		if err != nil {
			return res, err
		}
		res = append(res, e)
	}
	return res, nil
}

func (p *parser) record() (Node, error) {
	var res Node
	entries := map[types.String]Node{}
	for {
		t := p.peek()
		if t.Text == "}" {
			p.advance()
			return RecordNodes(entries), nil
		}
		if len(entries) > 0 {
			if err := p.exact(","); err != nil {
				return res, err
			}
		}
		k, v, err := p.recordEntry()
		if err != nil {
			return res, err
		}

		if _, ok := entries[k]; ok {
			return res, p.errorf("duplicate key: %v", k)
		}
		entries[k] = v
	}
}

func (p *parser) recordEntry() (types.String, Node, error) {
	var key types.String
	var value Node
	var err error
	t := p.advance()
	switch {
	case t.isIdent():
		key = types.String(t.Text)
	case t.isString():
		str, err := t.stringValue()
		if err != nil {
			return key, value, err
		}
		key = types.String(str)
	default:
		return key, value, p.errorf("unexpected token")
	}
	if err := p.exact(":"); err != nil {
		return key, value, err
	}
	value, err = p.expression()
	if err != nil {
		return key, value, err
	}
	return key, value, nil
}

func (p *parser) access(lhs Node) (Node, bool, error) {
	t := p.peek()
	switch t.Text {
	case ".":
		p.advance()
		t := p.advance()
		if !t.isIdent() {
			return Node{}, false, p.errorf("unexpected token")
		}
		if p.peek().Text == "(" {
			methodName := t.Text
			p.advance()
			exprs, err := p.expressions(")")
			if err != nil {
				return Node{}, false, err
			}
			p.advance() // expressions guarantees ")"

			var knownMethod func(Node, Node) Node
			switch methodName {
			case "contains":
				knownMethod = Node.Contains
			case "containsAll":
				knownMethod = Node.ContainsAll
			case "containsAny":
				knownMethod = Node.ContainsAny
			default:
				// i, ok := extMap[types.String(methodName)]
				// if !ok {
				// 	return Node{}, false, p.errorf("not a valid method name: `%v`", methodName)
				// }
				// if !i.IsMethod {
				// 	return Node{}, false, p.errorf("`%v` is a function, not a method", methodName)
				// }
				return newMethodCall(lhs, types.String(methodName), exprs...), true, nil
			}

			if len(exprs) != 1 {
				return Node{}, false, p.errorf("%v expects one argument", methodName)
			}
			return knownMethod(lhs, exprs[0]), true, nil
		} else {
			return lhs.Access(t.Text), true, nil
		}
	case "[":
		p.advance()
		t := p.advance()
		if !t.isString() {
			return Node{}, false, p.errorf("unexpected token")
		}
		name, err := t.stringValue()
		if err != nil {
			return Node{}, false, err
		}
		if err := p.exact("]"); err != nil {
			return Node{}, false, err
		}
		return lhs.Access(name), true, nil
	default:
		return lhs, false, nil
	}
}
