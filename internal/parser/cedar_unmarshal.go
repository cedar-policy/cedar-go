package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/internal/extensions"
	"github.com/cedar-policy/cedar-go/types"
)

func (p *PolicySlice) UnmarshalCedar(b []byte) error {
	tokens, err := Tokenize(b)
	if err != nil {
		return err
	}

	var policySet PolicySlice
	parser := newParser(tokens)
	for !parser.peek().isEOF() {
		var policy Policy
		if err = policy.fromCedar(&parser); err != nil {
			return err
		}

		policySet = append(policySet, &policy)
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
	return p.fromCedar(&parser)
}

func (p *Policy) fromCedar(parser *parser) error {
	pos := parser.peek().Pos
	annotations, err := parser.annotations()
	if err != nil {
		return err
	}

	newPolicy, err := parser.effect(&annotations)
	if err != nil {
		return err
	}
	newPolicy.Position = (ast.Position)(pos)

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

	*p = *(*Policy)(newPolicy)
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

func (p *parser) annotations() (ast.Annotations, error) {
	var res ast.Annotations
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

func (p *parser) annotation(a *ast.Annotations, known map[types.String]struct{}) error {
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

func (p *parser) effect(a *ast.Annotations) (*ast.Policy, error) {
	next := p.advance()
	if next.Text == "permit" {
		return a.Permit(), nil
	} else if next.Text == "forbid" {
		return a.Forbid(), nil
	}

	return nil, p.errorf("unexpected effect: %v", next.Text)
}

func (p *parser) principal(policy *ast.Policy) error {
	if err := p.exact(consts.Principal); err != nil {
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
	return p.entityFirstPathPreread(types.EntityType(t.Text))
}

func (p *parser) entityFirstPathPreread(firstPath types.EntityType) (types.EntityUID, error) {
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
			res.Type = types.EntityType(res.Type.String() + "::" + t.Text)
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

func (p *parser) pathFirstPathPreread(firstPath string) (types.EntityType, error) {
	res := types.EntityType(firstPath)
	for {
		if p.peek().Text != "::" {
			return res, nil
		}
		p.advance()
		t := p.advance()
		switch {
		case t.isIdent():
			res = types.EntityType(fmt.Sprintf("%v::%v", res, t.Text))
		default:
			return res, p.errorf("unexpected token")
		}
	}
}

func (p *parser) path() (types.EntityType, error) {
	t := p.advance()
	if !t.isIdent() {
		return "", p.errorf("expected ident")
	}
	return p.pathFirstPathPreread(t.Text)
}

func (p *parser) action(policy *ast.Policy) error {
	if err := p.exact(consts.Action); err != nil {
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

func (p *parser) resource(policy *ast.Policy) error {
	if err := p.exact(consts.Resource); err != nil {
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

func (p *parser) conditions(policy *ast.Policy) error {
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

func (p *parser) condition() (ast.Node, error) {
	var res ast.Node
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

func (p *parser) expression() (ast.Node, error) {
	t := p.peek()
	if t.Text == "if" {
		p.advance()

		condition, err := p.expression()
		if err != nil {
			return ast.Node{}, err
		}

		if err = p.exact("then"); err != nil {
			return ast.Node{}, err
		}
		ifTrue, err := p.expression()
		if err != nil {
			return ast.Node{}, err
		}

		if err = p.exact("else"); err != nil {
			return ast.Node{}, err
		}
		ifFalse, err := p.expression()
		if err != nil {
			return ast.Node{}, err
		}

		return ast.If(condition, ifTrue, ifFalse), nil
	}

	return p.or()
}

func (p *parser) or() (ast.Node, error) {
	lhs, err := p.and()
	if err != nil {
		return ast.Node{}, err
	}

	for p.peek().Text == "||" {
		p.advance()
		rhs, err := p.and()
		if err != nil {
			return ast.Node{}, err
		}
		lhs = lhs.Or(rhs)
	}

	return lhs, nil
}

func (p *parser) and() (ast.Node, error) {
	lhs, err := p.relation()
	if err != nil {
		return ast.Node{}, err
	}

	for p.peek().Text == "&&" {
		p.advance()
		rhs, err := p.relation()
		if err != nil {
			return ast.Node{}, err
		}
		lhs = lhs.And(rhs)
	}

	return lhs, nil
}

func (p *parser) relation() (ast.Node, error) {
	lhs, err := p.add()
	if err != nil {
		return ast.Node{}, err
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
	var operator func(ast.Node, ast.Node) ast.Node
	switch t.Text {
	case "<":
		operator = ast.Node.LessThan
	case "<=":
		operator = ast.Node.LessThanOrEqual
	case ">":
		operator = ast.Node.GreaterThan
	case ">=":
		operator = ast.Node.GreaterThanOrEqual
	case "!=":
		operator = ast.Node.NotEquals
	case "==":
		operator = ast.Node.Equals
	case "in":
		operator = ast.Node.In
	default:
		return lhs, nil

	}

	p.advance()
	rhs, err := p.add()
	if err != nil {
		return ast.Node{}, err
	}
	return operator(lhs, rhs), nil
}

func (p *parser) has(lhs ast.Node) (ast.Node, error) {
	t := p.advance()
	if t.isIdent() {
		return lhs.Has(t.Text), nil
	} else if t.isString() {
		str, err := t.stringValue()
		if err != nil {
			return ast.Node{}, err
		}
		return lhs.Has(str), nil
	}
	return ast.Node{}, p.errorf("expected ident or string")
}

func (p *parser) like(lhs ast.Node) (ast.Node, error) {
	t := p.advance()
	if !t.isString() {
		return ast.Node{}, p.errorf("expected string literal")
	}
	patternRaw := t.Text
	patternRaw = strings.TrimPrefix(patternRaw, "\"")
	patternRaw = strings.TrimSuffix(patternRaw, "\"")
	var pattern types.Pattern
	if err := pattern.UnmarshalCedar([]byte(patternRaw)); err != nil {
		return ast.Node{}, err
	}
	return lhs.Like(pattern), nil
}

func (p *parser) is(lhs ast.Node) (ast.Node, error) {
	entityType, err := p.path()
	if err != nil {
		return ast.Node{}, err
	}
	if p.peek().Text == "in" {
		p.advance()
		inEntity, err := p.add()
		if err != nil {
			return ast.Node{}, err
		}
		return lhs.IsIn(entityType, inEntity), nil
	}
	return lhs.Is(entityType), nil
}

func (p *parser) add() (ast.Node, error) {
	lhs, err := p.mult()
	if err != nil {
		return ast.Node{}, err
	}

	for {
		t := p.peek()
		var operator func(ast.Node, ast.Node) ast.Node
		switch t.Text {
		case "+":
			operator = ast.Node.Plus
		case "-":
			operator = ast.Node.Minus
		}

		if operator == nil {
			break
		}

		p.advance()
		rhs, err := p.mult()
		if err != nil {
			return ast.Node{}, err
		}
		lhs = operator(lhs, rhs)
	}

	return lhs, nil
}

func (p *parser) mult() (ast.Node, error) {
	lhs, err := p.unary()
	if err != nil {
		return ast.Node{}, err
	}

	for p.peek().Text == "*" {
		p.advance()
		rhs, err := p.unary()
		if err != nil {
			return ast.Node{}, err
		}
		lhs = lhs.Times(rhs)
	}

	return lhs, nil
}

func (p *parser) unary() (ast.Node, error) {
	var ops []bool
	for {
		opToken := p.peek()
		if opToken.Text != "-" && opToken.Text != "!" {
			break
		}
		p.advance()
		ops = append(ops, opToken.Text == "-")
	}

	var res ast.Node

	// special case for max negative long
	tok := p.peek()
	if len(ops) > 0 && ops[len(ops)-1] && tok.isInt() {
		p.advance()
		i, err := strconv.ParseInt("-"+tok.Text, 10, 64)
		if err != nil {
			return ast.Node{}, err
		}
		res = ast.Long(i)
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
			res = ast.Negate(res)
		} else {
			res = ast.Not(res)
		}
	}
	return res, nil
}

func (p *parser) member() (ast.Node, error) {
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

func (p *parser) primary() (ast.Node, error) {
	var res ast.Node
	t := p.advance()
	switch {
	case t.isInt():
		i, err := t.intValue()
		if err != nil {
			return res, err
		}
		res = ast.Long(i)
	case t.isString():
		str, err := t.stringValue()
		if err != nil {
			return res, err
		}
		res = ast.String(str)
	case t.Text == "true":
		res = ast.True()
	case t.Text == "false":
		res = ast.False()
	case t.Text == consts.Principal:
		res = ast.Principal()
	case t.Text == consts.Action:
		res = ast.Action()
	case t.Text == consts.Resource:
		res = ast.Resource()
	case t.Text == consts.Context:
		res = ast.Context()
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
		res = ast.Set(set...)
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

func (p *parser) entityOrExtFun(prefix string) (ast.Node, error) {
	for {
		t := p.advance()
		switch t.Text {
		case "::":
			t := p.advance()
			switch {
			case t.isIdent():
				prefix = prefix + "::" + t.Text
			case t.isString():
				id, err := t.stringValue()
				if err != nil {
					return ast.Node{}, err
				}
				return ast.EntityUID(prefix, id), nil
			default:
				return ast.Node{}, p.errorf("unexpected token")
			}
		case "(":
			// Although the Cedar grammar says that any name can be provided here, the reference implementation actually
			// checks at parse time whether the name corresponds to a known extension function.
			i, ok := extensions.ExtMap[types.String(prefix)]
			if !ok {
				return ast.Node{}, p.errorf("`%v` is not a function", prefix)
			}
			if i.IsMethod {
				return ast.Node{}, p.errorf("`%v` is a method, not a function", prefix)
			}

			args, err := p.expressions(")")
			if err != nil {
				return ast.Node{}, err
			}
			p.advance()
			return ast.ExtensionCall(types.String(prefix), args...), nil
		default:
			return ast.Node{}, p.errorf("unexpected token")
		}
	}
}

func (p *parser) expressions(endOfListMarker string) ([]ast.Node, error) {
	var res []ast.Node
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

func (p *parser) record() (ast.Node, error) {
	var res ast.Node
	var elements ast.Pairs
	known := map[string]struct{}{}
	for {
		t := p.peek()
		if t.Text == "}" {
			p.advance()
			return ast.Record(elements), nil
		}
		if len(elements) > 0 {
			if err := p.exact(","); err != nil {
				return res, err
			}
		}
		k, v, err := p.recordEntry()
		if err != nil {
			return res, err
		}

		if _, ok := known[k]; ok {
			return res, p.errorf("duplicate key: %v", k)
		}
		known[k] = struct{}{}
		elements = append(elements, ast.Pair{Key: k, Value: v})
	}
}

func (p *parser) recordEntry() (string, ast.Node, error) {
	var key string
	var value ast.Node
	var err error
	t := p.advance()
	switch {
	case t.isIdent():
		key = t.Text
	case t.isString():
		str, err := t.stringValue()
		if err != nil {
			return key, value, err
		}
		key = str
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

func (p *parser) access(lhs ast.Node) (ast.Node, bool, error) {
	t := p.peek()
	switch t.Text {
	case ".":
		p.advance()
		t := p.advance()
		if !t.isIdent() {
			return ast.Node{}, false, p.errorf("unexpected token")
		}
		if p.peek().Text == "(" {
			methodName := t.Text
			p.advance()
			exprs, err := p.expressions(")")
			if err != nil {
				return ast.Node{}, false, err
			}
			p.advance() // expressions guarantees ")"

			var knownMethod func(ast.Node, ast.Node) ast.Node
			switch methodName {
			case "contains":
				knownMethod = ast.Node.Contains
			case "containsAll":
				knownMethod = ast.Node.ContainsAll
			case "containsAny":
				knownMethod = ast.Node.ContainsAny
			default:
				// Although the Cedar grammar says that any name can be provided here, the reference implementation
				// actually checks at parse time whether the name corresponds to a known extension method.
				i, ok := extensions.ExtMap[types.String(methodName)]
				if !ok {
					return ast.Node{}, false, p.errorf("`%v` is not a method", methodName)
				}
				if !i.IsMethod {
					return ast.Node{}, false, p.errorf("`%v` is a function, not a method", methodName)
				}
				return ast.NewMethodCall(lhs, types.String(methodName), exprs...), true, nil
			}

			if len(exprs) != 1 {
				return ast.Node{}, false, p.errorf("%v expects one argument", methodName)
			}
			return knownMethod(lhs, exprs[0]), true, nil
		} else {
			return lhs.Access(t.Text), true, nil
		}
	case "[":
		p.advance()
		t := p.advance()
		if !t.isString() {
			return ast.Node{}, false, p.errorf("unexpected token")
		}
		name, err := t.stringValue()
		if err != nil {
			return ast.Node{}, false, err
		}
		if err := p.exact("]"); err != nil {
			return ast.Node{}, false, err
		}
		return lhs.Access(name), true, nil
	default:
		return lhs, false, nil
	}
}
