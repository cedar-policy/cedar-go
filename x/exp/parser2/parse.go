package parser

import (
	"fmt"
	"strconv"
	"strings"
)

func Parse(tokens []Token) (Policies, error) {
	p := &parser{Tokens: tokens}
	return p.Policies()
}

func ParseEntity(tokens []Token) (Entity, error) {
	p := &parser{Tokens: tokens}
	return p.Entity()
}

type parser struct {
	Tokens []Token
	Pos    int
}

func (p *parser) advance() Token {
	t := p.peek()
	if p.Pos < len(p.Tokens)-1 {
		p.Pos++
	}
	return t
}

func (p *parser) peek() Token {
	return p.Tokens[p.Pos]
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
	if p.Pos < len(p.Tokens) {
		t = p.Tokens[p.Pos]
	}
	err := fmt.Errorf(s, args...)
	return fmt.Errorf("parse error at %v %q: %w", t.Pos, t.Text, err)
}

// Policies := {Policy}

type Policies []Policy

func (c Policies) String() string {
	var sb strings.Builder
	for i, p := range c {
		if i > 0 {
			sb.WriteRune('\n')
		}
		sb.WriteString(p.String())
	}
	return sb.String()
}

func (p *parser) Policies() (Policies, error) {
	var res Policies
	for !p.peek().isEOF() {
		policy, err := p.policy()
		if err != nil {
			return nil, err
		}
		res = append(res, policy)
	}
	return res, nil
}

// Policy := {Annotation} Effect '(' Scope ')' {Conditions} ';'
// Scope := Principal ',' Action ',' Resource

type Policy struct {
	Position    Position
	Annotations []Annotation
	Effect      Effect
	Principal   Principal
	Action      Action
	Resource    Resource
	Conditions  []Condition
}

func (p Policy) String() string {
	var sb strings.Builder
	for i, a := range p.Annotations {
		if i > 0 {
			sb.WriteRune('\n')
		}
		sb.WriteString(a.String())
	}
	sb.WriteString(fmt.Sprintf("%s(\n%s,\n%s,\n%s\n)",
		p.Effect, p.Principal, p.Action, p.Resource,
	))
	for _, c := range p.Conditions {
		sb.WriteRune('\n')
		sb.WriteString(c.String())
	}
	sb.WriteString(";")
	return sb.String()
}

func (p *parser) policy() (Policy, error) {
	var res Policy
	res.Position = p.peek().Pos
	var err error
	if res.Annotations, err = p.annotations(); err != nil {
		return res, err
	}
	if res.Effect, err = p.effect(); err != nil {
		return res, err
	}
	if err := p.exact("("); err != nil {
		return res, err
	}
	if res.Principal, err = p.principal(); err != nil {
		return res, err
	}
	if err := p.exact(","); err != nil {
		return res, err
	}
	if res.Action, err = p.action(); err != nil {
		return res, err
	}
	if err := p.exact(","); err != nil {
		return res, err
	}
	if res.Resource, err = p.resource(); err != nil {
		return res, err
	}
	if err := p.exact(")"); err != nil {
		return res, err
	}
	if res.Conditions, err = p.conditions(); err != nil {
		return res, err
	}
	if err := p.exact(";"); err != nil {
		return res, err
	}
	return res, nil
}

// Annotation := '@'IDENT'('STR')'

type Annotation struct {
	Key   string
	Value string
}

func (a Annotation) String() string {
	return fmt.Sprintf("@%s(%q)", a.Key, a.Value)
}

func (p *parser) annotation() (Annotation, error) {
	var res Annotation
	var err error
	t := p.advance()
	if !t.isIdent() {
		return res, p.errorf("expected ident")
	}
	res.Key = t.Text
	if err := p.exact("("); err != nil {
		return res, err
	}
	t = p.advance()
	if !t.isString() {
		return res, p.errorf("expected string")
	}
	if res.Value, err = t.stringValue(); err != nil {
		return res, err
	}
	if err := p.exact(")"); err != nil {
		return res, err
	}
	return res, nil
}

func (p *parser) annotations() ([]Annotation, error) {
	var res []Annotation
	for p.peek().Text == "@" {
		p.advance()
		a, err := p.annotation()
		if err != nil {
			return res, err
		}
		for _, aa := range res {
			if aa.Key == a.Key {
				return res, p.errorf("duplicate annotation")
			}
		}
		res = append(res, a)
	}
	return res, nil
}

// Effect := 'permit' | 'forbid'

type Effect string

const (
	EffectPermit = Effect("permit")
	EffectForbid = Effect("forbid")
)

func (p *parser) effect() (Effect, error) {
	next := p.advance()
	res := Effect(next.Text)
	switch res {
	case EffectForbid:
	case EffectPermit:
	default:
		return res, p.errorf("unexpected effect: %v", res)
	}
	return res, nil
}

// MatchType

type MatchType int

const (
	MatchAny = MatchType(iota)
	MatchEquals
	MatchIn
	MatchInList
	MatchIs
	MatchIsIn
)

// Principal := 'principal' [('in' | '==') Entity]

type Principal struct {
	Type   MatchType
	Path   Path
	Entity Entity
}

func (p Principal) String() string {
	var res string
	switch p.Type {
	case MatchAny:
		res = "principal"
	case MatchEquals:
		res = fmt.Sprintf("principal == %s", p.Entity)
	case MatchIs:
		res = fmt.Sprintf("principal is %s", p.Path)
	case MatchIsIn:
		res = fmt.Sprintf("principal is %s in %s", p.Path, p.Entity)
	case MatchIn:
		res = fmt.Sprintf("principal in %s", p.Entity)
	}
	return res
}

func (p *parser) principal() (Principal, error) {
	var res Principal
	if err := p.exact("principal"); err != nil {
		return res, err
	}
	switch p.peek().Text {
	case "==":
		p.advance()
		var err error
		res.Type = MatchEquals
		res.Entity, err = p.Entity()
		return res, err
	case "is":
		p.advance()
		var err error
		res.Type = MatchIs
		res.Path, err = p.Path()
		if err == nil && p.peek().Text == "in" {
			p.advance()
			res.Type = MatchIsIn
			res.Entity, err = p.Entity()
			return res, err
		}
		return res, err
	case "in":
		p.advance()
		var err error
		res.Type = MatchIn
		res.Entity, err = p.Entity()
		return res, err
	default:
		return Principal{
			Type: MatchAny,
		}, nil
	}
}

// Action := 'action' [( '==' Entity | 'in' ('[' EntList ']' | Entity) )]

type Action struct {
	Type     MatchType
	Entities []Entity
}

func (a Action) String() string {
	var sb strings.Builder
	switch a.Type {
	case MatchAny:
		sb.WriteString("action")
	case MatchEquals:
		sb.WriteString(fmt.Sprintf("action == %s", a.Entities[0]))
	case MatchIn:
		sb.WriteString(fmt.Sprintf("action in %s", a.Entities[0]))
	case MatchInList:
		sb.WriteString("action in [")
		for i, e := range a.Entities {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(e.String())
		}
		sb.WriteRune(']')
	}
	return sb.String()
}

func (p *parser) action() (Action, error) {
	var res Action
	var err error
	if err := p.exact("action"); err != nil {
		return res, err
	}
	switch p.peek().Text {
	case "==":
		p.advance()
		res.Type = MatchEquals
		e, err := p.Entity()
		if err != nil {
			return res, err
		}
		res.Entities = append(res.Entities, e)
		return res, nil
	case "in":
		p.advance()
		if p.peek().Text == "[" {
			res.Type = MatchInList
			p.advance()
			res.Entities, err = p.entlist()
			if err != nil {
				return res, err
			}
			p.advance() // entlist guarantees "]"
			return res, nil
		} else {
			res.Type = MatchIn
			e, err := p.Entity()
			if err != nil {
				return res, err
			}
			res.Entities = append(res.Entities, e)
			return res, nil
		}
	default:
		return Action{
			Type: MatchAny,
		}, nil
	}
}

// Resource := 'resource' [('in' | '==') Entity)]

type Resource struct {
	Type   MatchType
	Path   Path
	Entity Entity
}

func (r Resource) String() string {
	var res string
	switch r.Type {
	case MatchAny:
		res = "resource"
	case MatchEquals:
		res = fmt.Sprintf("resource == %s", r.Entity)
	case MatchIs:
		res = fmt.Sprintf("resource is %s", r.Path)
	case MatchIsIn:
		res = fmt.Sprintf("resource is %s in %s", r.Path, r.Entity)
	case MatchIn:
		res = fmt.Sprintf("resource in %s", r.Entity)
	}
	return res
}

func (p *parser) resource() (Resource, error) {
	var res Resource
	if err := p.exact("resource"); err != nil {
		return res, err
	}
	switch p.peek().Text {
	case "==":
		p.advance()
		var err error
		res.Type = MatchEquals
		res.Entity, err = p.Entity()
		return res, err
	case "is":
		p.advance()
		var err error
		res.Type = MatchIs
		res.Path, err = p.Path()
		if err == nil && p.peek().Text == "in" {
			p.advance()
			res.Type = MatchIsIn
			res.Entity, err = p.Entity()
			return res, err
		}
		return res, err
	case "in":
		p.advance()
		var err error
		res.Type = MatchIn
		res.Entity, err = p.Entity()
		return res, err
	default:
		return Resource{
			Type: MatchAny,
		}, nil
	}
}

// Entity := Path '::' STR

type Entity struct {
	Path []string
}

func (e Entity) String() string {
	return fmt.Sprintf(
		"%s::%q",
		strings.Join(e.Path[0:len(e.Path)-1], "::"),
		e.Path[len(e.Path)-1],
	)
}

func (p *parser) Entity() (Entity, error) {
	var res Entity
	t := p.advance()
	if !t.isIdent() {
		return res, p.errorf("expected ident")
	}
	res.Path = append(res.Path, t.Text)
	for {
		if err := p.exact("::"); err != nil {
			return res, err
		}
		t := p.advance()
		switch {
		case t.isIdent():
			res.Path = append(res.Path, t.Text)
		case t.isString():
			component, err := t.stringValue()
			if err != nil {
				return res, err
			}
			res.Path = append(res.Path, component)
			return res, nil
		default:
			return res, p.errorf("unexpected token")
		}
	}
}

// Path ::= IDENT {'::' IDENT}

type Path struct {
	Path []string
}

func (e Path) String() string {
	return strings.Join(e.Path, "::")
}

func (p *parser) Path() (Path, error) {
	var res Path
	t := p.advance()
	if !t.isIdent() {
		return res, p.errorf("expected ident")
	}
	res.Path = append(res.Path, t.Text)
	for {
		if p.peek().Text != "::" {
			return res, nil
		}
		p.advance()
		t := p.advance()
		switch {
		case t.isIdent():
			res.Path = append(res.Path, t.Text)
		default:
			return res, p.errorf("unexpected token")
		}
	}
}

// EntList := Entity {',' Entity}

func (p *parser) entlist() ([]Entity, error) {
	var res []Entity
	for p.peek().Text != "]" {
		if len(res) > 0 {
			if err := p.exact(","); err != nil {
				return res, err
			}
		}
		e, err := p.Entity()
		if err != nil {
			return res, err
		}
		res = append(res, e)
	}
	return res, nil
}

// Condition := ('when' | 'unless') '{' Expr '}'

type ConditionType string

const (
	ConditionWhen   ConditionType = "when"
	ConditionUnless ConditionType = "unless"
)

type Condition struct {
	Type       ConditionType
	Expression Expression
}

func (c Condition) String() string {
	var res string
	switch c.Type {
	case ConditionWhen:
		res = fmt.Sprintf("when {\n%s\n}", c.Expression)
	case ConditionUnless:
		res = fmt.Sprintf("unless {\n%s\n}", c.Expression)
	}
	return res
}

func (p *parser) condition() (Condition, error) {
	var res Condition
	var err error
	res.Type = ConditionType(p.advance().Text)
	if err := p.exact("{"); err != nil {
		return res, err
	}
	if res.Expression, err = p.expression(); err != nil {
		return res, err
	}
	if err := p.exact("}"); err != nil {
		return res, err
	}
	return res, nil
}

func (p *parser) conditions() ([]Condition, error) {
	var res []Condition
	for {
		switch p.peek().Text {
		case "when", "unless":
			c, err := p.condition()
			if err != nil {
				return res, err
			}
			res = append(res, c)
		default:
			return res, nil
		}
	}
}

// Expr := Or | If

type ExpressionType int

const (
	ExpressionOr ExpressionType = iota
	ExpressionIf
)

type Expression struct {
	Type ExpressionType
	Or   Or
	If   *If
}

func (e Expression) String() string {
	var res string
	switch e.Type {
	case ExpressionOr:
		res = e.Or.String()
	case ExpressionIf:
		res = e.If.String()
	}
	return res
}

func (p *parser) expression() (Expression, error) {
	var res Expression
	var err error
	if p.peek().Text == "if" {
		p.advance()
		res.Type = ExpressionIf
		i, err := p.ifExpr()
		if err != nil {
			return res, err
		}
		res.If = &i
		return res, nil
	} else {
		res.Type = ExpressionOr
		if res.Or, err = p.or(); err != nil {
			return res, err
		}
		return res, nil
	}
}

// If := 'if' Expr 'then' Expr 'else' Expr

type If struct {
	If   Expression
	Then Expression
	Else Expression
}

func (i If) String() string {
	return fmt.Sprintf("if %s then %s else %s", i.If, i.Then, i.Else)
}

func (p *parser) ifExpr() (If, error) {
	var res If
	var err error
	if res.If, err = p.expression(); err != nil {
		return res, err
	}
	if err = p.exact("then"); err != nil {
		return res, err
	}
	if res.Then, err = p.expression(); err != nil {
		return res, err
	}
	if err = p.exact("else"); err != nil {
		return res, err
	}
	if res.Else, err = p.expression(); err != nil {
		return res, err
	}
	return res, err
}

// Or := And {'||' And}

type Or struct {
	Ands []And
}

func (o Or) String() string {
	var sb strings.Builder
	for i, and := range o.Ands {
		if i > 0 {
			sb.WriteString(" || ")
		}
		sb.WriteString(and.String())
	}
	return sb.String()
}

func (p *parser) or() (Or, error) {
	var res Or
	for {
		a, err := p.and()
		if err != nil {
			return res, err
		}
		res.Ands = append(res.Ands, a)
		if p.peek().Text != "||" {
			return res, nil
		}
		p.advance()
	}
}

// And := Relation {'&&' Relation}

type And struct {
	Relations []Relation
}

func (a And) String() string {
	var sb strings.Builder
	for i, rel := range a.Relations {
		if i > 0 {
			sb.WriteString(" && ")
		}
		sb.WriteString(rel.String())
	}
	return sb.String()
}

func (p *parser) and() (And, error) {
	var res And
	for {
		r, err := p.relation()
		if err != nil {
			return res, err
		}
		res.Relations = append(res.Relations, r)
		if p.peek().Text != "&&" {
			return res, nil
		}
		p.advance()
	}
}

// Relation := Add [RELOP Add] | Add 'has' (IDENT | STR) | Add 'like' PAT

type RelationType string

const (
	RelationNone       RelationType = "none"
	RelationRelOp      RelationType = "relop"
	RelationHasIdent   RelationType = "hasident"
	RelationHasLiteral RelationType = "hasliteral"
	RelationLike       RelationType = "like"
	RelationIs         RelationType = "is"
	RelationIsIn       RelationType = "isIn"
)

type Relation struct {
	Add      Add
	Type     RelationType
	RelOp    RelOp
	RelOpRhs Add
	Str      string
	Pat      Pattern
	Path     Path
	Entity   Add
}

func (r Relation) String() string {
	var sb strings.Builder
	sb.WriteString(r.Add.String())
	switch r.Type {
	case RelationNone:
	case RelationRelOp:
		sb.WriteString(" ")
		sb.WriteString(string(r.RelOp))
		sb.WriteString(" ")
		sb.WriteString(r.RelOpRhs.String())
	case RelationHasIdent:
		sb.WriteString(" has ")
		sb.WriteString(r.Str)
	case RelationHasLiteral:
		sb.WriteString(" has ")
		sb.WriteString(strconv.Quote(r.Str))
	case RelationLike:
		sb.WriteString(" like ")
		sb.WriteString(r.Pat.String())
	case RelationIs:
		sb.WriteString(" is ")
		sb.WriteString(r.Path.String())
	case RelationIsIn:
		sb.WriteString(" is ")
		sb.WriteString(r.Path.String())
		sb.WriteString(" in ")
		sb.WriteString(r.Entity.String())
	}
	return sb.String()
}

func (p *parser) relation() (Relation, error) {
	var res Relation
	var err error
	if res.Add, err = p.add(); err != nil {
		return res, err
	}

	t := p.peek()
	switch t.Text {
	case "<", "<=", ">=", ">", "!=", "==", "in":
		p.advance()
		res.Type = RelationRelOp
		res.RelOp = RelOp(t.Text)
		if res.RelOpRhs, err = p.add(); err != nil {
			return res, err
		}
	case "has":
		p.advance()
		t := p.advance()
		switch {
		case t.isIdent():
			res.Type = RelationHasIdent
			res.Str = t.Text
		case t.isString():
			res.Type = RelationHasLiteral
			if res.Str, err = t.stringValue(); err != nil {
				return res, err
			}
		default:
			return res, p.errorf("unexpected token")
		}
	case "like":
		p.advance()
		res.Type = RelationLike
		t := p.advance()
		if !t.isString() {
			return res, p.errorf("unexpected token")
		}
		if res.Pat, err = t.patternValue(); err != nil {
			return res, err
		}
	case "is":
		p.advance()
		var err error
		res.Type = RelationIs
		res.Path, err = p.Path()
		if err == nil && p.peek().Text == "in" {
			p.advance()
			res.Type = RelationIsIn
			res.Entity, err = p.add()
			return res, err
		}
		return res, err
	default:
		res.Type = RelationNone
	}
	return res, nil
}

// RELOP := '<' | '<=' | '>=' | '>' | '!=' | '==' | 'in'

type RelOp string

const (
	RelOpLt RelOp = "<"
	RelOpLe RelOp = "<="
	RelOpGe RelOp = ">="
	RelOpGt RelOp = ">"
	RelOpNe RelOp = "!="
	RelOpEq RelOp = "=="
	RelOpIn RelOp = "in"
)

// Add := Mult {ADDOP Mult}

type Add struct {
	Mults  []Mult
	AddOps []AddOp
}

func (a Add) String() string {
	var sb strings.Builder
	sb.WriteString(a.Mults[0].String())
	for i, op := range a.AddOps {
		sb.WriteString(fmt.Sprintf(" %s %s", op, a.Mults[i+1].String()))
	}
	return sb.String()
}

func (p *parser) add() (Add, error) {
	var res Add
	var err error
	mult, err := p.mult()
	if err != nil {
		return res, err
	}
	res.Mults = append(res.Mults, mult)
	for {
		op := AddOp(p.peek().Text)
		switch op {
		case AddOpAdd, AddOpSub:
		default:
			return res, nil
		}
		p.advance()
		mult, err := p.mult()
		if err != nil {
			return res, err
		}
		res.AddOps = append(res.AddOps, op)
		res.Mults = append(res.Mults, mult)
	}
}

// ADDOP := '+' | '-'

type AddOp string

const (
	AddOpAdd AddOp = "+"
	AddOpSub AddOp = "-"
)

// Mult := Unary { '*' Unary}

type Mult struct {
	Unaries []Unary
}

func (m Mult) String() string {
	var sb strings.Builder
	for i, u := range m.Unaries {
		if i > 0 {
			sb.WriteString(" * ")
		}
		sb.WriteString(u.String())
	}
	return sb.String()
}

func (p *parser) mult() (Mult, error) {
	var res Mult
	for {
		u, err := p.unary()
		if err != nil {
			return res, err
		}
		res.Unaries = append(res.Unaries, u)
		if p.peek().Text != "*" {
			return res, nil
		}
		p.advance()
	}
}

// Unary := [UNARYOP]x4 Member

type Unary struct {
	Ops    []UnaryOp
	Member Member
}

func (u Unary) String() string {
	var sb strings.Builder
	for _, o := range u.Ops {
		sb.WriteString(string(o))
	}
	sb.WriteString(u.Member.String())
	return sb.String()
}

func (p *parser) unary() (Unary, error) {
	var res Unary
	for {
		o := UnaryOp(p.peek().Text)
		switch o {
		case UnaryOpNot:
			p.advance()
			res.Ops = append(res.Ops, o)
		case UnaryOpMinus:
			p.advance()
			if p.peek().isInt() {
				t := p.advance()
				i, err := strconv.ParseInt("-"+t.Text, 10, 64)
				if err != nil {
					return res, err
				}
				res.Member = Member{
					Primary: Primary{
						Type: PrimaryLiteral,
						Literal: Literal{
							Type: LiteralInt,
							Long: i,
						},
					},
				}
				return res, nil
			}
			res.Ops = append(res.Ops, o)
		default:
			var err error
			res.Member, err = p.member()
			if err != nil {
				return res, err
			}
			return res, nil
		}
	}
}

// UNARYOP := '!' | '-'

type UnaryOp string

const (
	UnaryOpNot   UnaryOp = "!"
	UnaryOpMinus UnaryOp = "-"
)

// Member := Primary {Access}

type Member struct {
	Primary  Primary
	Accesses []Access
}

func (m Member) String() string {
	var sb strings.Builder
	sb.WriteString(m.Primary.String())
	for _, a := range m.Accesses {
		sb.WriteString(a.String())
	}
	return sb.String()
}

func (p *parser) member() (Member, error) {
	var res Member
	var err error
	if res.Primary, err = p.primary(); err != nil {
		return res, err
	}
	for {
		a, ok, err := p.access()
		if !ok {
			return res, err
		} else {
			res.Accesses = append(res.Accesses, a)
		}
	}
}

// Primary := LITERAL
// 			| VAR
// 			| Entity
// 			| ExtFun '(' [ExprList] ')'
// 			| '(' Expr ')'
// 			| '[' [ExprList] ']'
// 			| '{' [RecInits] '}'

type PrimaryType int

const (
	PrimaryLiteral PrimaryType = iota
	PrimaryVar
	PrimaryEntity
	PrimaryExtFun
	PrimaryExpr
	PrimaryExprList
	PrimaryRecInits
)

type Primary struct {
	Type        PrimaryType
	Literal     Literal
	Var         Var
	Entity      Entity
	ExtFun      ExtFun
	Expression  Expression
	Expressions []Expression
	RecInits    []RecInit
}

func (p Primary) String() string {
	var res string
	switch p.Type {
	case PrimaryLiteral:
		res = p.Literal.String()
	case PrimaryVar:
		res = p.Var.String()
	case PrimaryEntity:
		res = p.Entity.String()
	case PrimaryExtFun:
		res = p.ExtFun.String()
	case PrimaryExpr:
		res = fmt.Sprintf("(%s)", p.Expression)
	case PrimaryExprList:
		var sb strings.Builder
		sb.WriteRune('[')
		for i, e := range p.Expressions {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(e.String())
		}
		sb.WriteRune(']')
		res = sb.String()
	case PrimaryRecInits:
		var sb strings.Builder
		sb.WriteRune('{')
		for i, r := range p.RecInits {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(r.String())
		}
		sb.WriteRune('}')
		res = sb.String()
	}
	return res
}

func (p *parser) primary() (Primary, error) {
	var res Primary
	var err error
	t := p.advance()
	switch {
	case t.isInt():
		i, err := t.intValue()
		if err != nil {
			return res, err
		}
		res.Type = PrimaryLiteral
		res.Literal = Literal{
			Type: LiteralInt,
			Long: i,
		}
	case t.isString():
		res.Type = PrimaryLiteral
		res.Literal.Type = LiteralString
		if res.Literal.Str, err = t.stringValue(); err != nil {
			return res, err
		}
	case t.Text == "true", t.Text == "false":
		res.Type = PrimaryLiteral
		res.Literal = Literal{
			Type: LiteralBool,
			Bool: t.Text == "true",
		}
	case t.Text == string(VarPrincipal),
		t.Text == string(VarAction),
		t.Text == string(VarResource),
		t.Text == string(VarContext):
		res.Type = PrimaryVar
		res.Var = Var{
			Type: VarType(t.Text),
		}
	case t.isIdent():
		e, f, err := p.entityOrExtFun(t.Text)
		switch {
		case e != nil:
			res.Type = PrimaryEntity
			res.Entity = *e
		case f != nil:
			res.Type = PrimaryExtFun
			res.ExtFun = *f
		default:
			return res, err
		}
	case t.Text == "(":
		res.Type = PrimaryExpr
		if res.Expression, err = p.expression(); err != nil {
			return res, err
		}
		if err := p.exact(")"); err != nil {
			return res, err
		}
	case t.Text == "[":
		res.Type = PrimaryExprList
		if res.Expressions, err = p.expressions("]"); err != nil {
			return res, err
		}
		p.advance() // expressions guarantees "]"
		return res, err
	case t.Text == "{":
		res.Type = PrimaryRecInits
		if res.RecInits, err = p.recInits(); err != nil {
			return res, err
		}
		return res, err
	default:
		return res, p.errorf("invalid primary")
	}
	return res, nil
}

func (p *parser) entityOrExtFun(first string) (*Entity, *ExtFun, error) {
	path := []string{first}
	for {
		if p.peek().Text != "::" {
			f, err := p.extFun(path)
			if err != nil {
				return nil, nil, err
			}
			return nil, &f, err
		}
		p.advance()
		t := p.advance()
		switch {
		case t.isIdent():
			path = append(path, t.Text)
		case t.isString():
			component, err := t.stringValue()
			if err != nil {
				return nil, nil, err
			}
			path = append(path, component)
			return &Entity{Path: path}, nil, nil
		default:
			return nil, nil, p.errorf("unexpected token")
		}
	}
}

func (p *parser) expressions(endOfListMarker string) ([]Expression, error) {
	var res []Expression
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

func (p *parser) recInits() ([]RecInit, error) {
	var res []RecInit
	for {
		t := p.peek()
		if t.Text == "}" {
			p.advance()
			return res, nil
		}
		if len(res) > 0 {
			if err := p.exact(","); err != nil {
				return res, err
			}
		}
		e, err := p.recInit()
		if err != nil {
			return res, err
		}
		res = append(res, e)
	}
}

// LITERAL := BOOL | INT | STR

type LiteralType int

const (
	LiteralBool LiteralType = iota
	LiteralInt
	LiteralString
)

type Literal struct {
	Type LiteralType
	Bool bool
	Long int64
	Str  string
}

func (l Literal) String() string {
	var res string
	switch l.Type {
	case LiteralBool:
		res = strconv.FormatBool(l.Bool)
	case LiteralInt:
		res = strconv.FormatInt(l.Long, 10)
	case LiteralString:
		res = strconv.Quote(l.Str)
	}
	return res
}

// VAR := 'principal' | 'action' | 'resource' | 'context'

type VarType string

const (
	VarPrincipal VarType = "principal"
	VarAction    VarType = "action"
	VarResource  VarType = "resource"
	VarContext   VarType = "context"
)

type Var struct {
	Type VarType
}

func (v Var) String() string {
	return string(v.Type)
}

// ExtFun := [Path '::'] IDENT

type ExtFun struct {
	Path        []string
	Expressions []Expression
}

func (f ExtFun) String() string {
	var sb strings.Builder
	sb.WriteString(strings.Join(f.Path, "::"))
	sb.WriteRune('(')
	for i, e := range f.Expressions {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(e.String())
	}
	sb.WriteRune(')')
	return sb.String()
}

func (p *parser) extFun(path []string) (ExtFun, error) {
	res := ExtFun{Path: path}
	if err := p.exact("("); err != nil {
		return res, err
	}
	var err error
	if res.Expressions, err = p.expressions(")"); err != nil {
		return res, err
	}
	p.advance() // expressions guarantees ")"
	return res, err
}

// Access := '.' IDENT ['(' [ExprList] ')'] | '[' STR ']'

type AccessType int

const (
	AccessField AccessType = iota
	AccessCall
	AccessIndex
)

type Access struct {
	Type        AccessType
	Name        string
	Expressions []Expression
}

func (a Access) String() string {
	var sb strings.Builder
	switch a.Type {
	case AccessField:
		sb.WriteRune('.')
		sb.WriteString(a.Name)
	case AccessCall:
		sb.WriteRune('.')
		sb.WriteString(a.Name)
		sb.WriteRune('(')
		for i, e := range a.Expressions {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(e.String())
		}
		sb.WriteRune(')')
	case AccessIndex:
		sb.WriteRune('[')
		sb.WriteString(strconv.Quote(a.Name))
		sb.WriteRune(']')
	}
	return sb.String()
}

func (p *parser) access() (Access, bool, error) {
	var res Access
	var err error
	t := p.peek()
	switch t.Text {
	case ".":
		p.advance()
		t := p.advance()
		if !t.isIdent() {
			return res, false, p.errorf("unexpected token")
		}
		res.Name = t.Text
		if p.peek().Text == "(" {
			p.advance()
			res.Type = AccessCall
			if res.Expressions, err = p.expressions(")"); err != nil {
				return res, false, err
			}
			p.advance() // expressions guarantees ")"
		} else {
			res.Type = AccessField
		}
	case "[":
		p.advance()
		res.Type = AccessIndex
		t := p.advance()
		if !t.isString() {
			return res, false, p.errorf("unexpected token")
		}
		if res.Name, err = t.stringValue(); err != nil {
			return res, false, err
		}
		if err := p.exact("]"); err != nil {
			return res, false, err
		}
	default:
		return res, false, nil
	}
	return res, true, nil
}

// RecInits := (IDENT | STR) ':' Expr {',' (IDENT | STR) ':' Expr}

type RecKeyType int

const (
	RecKeyIdent RecKeyType = iota
	RecKeyString
)

type RecInit struct {
	KeyType RecKeyType
	Key     string
	Value   Expression
}

func (r RecInit) String() string {
	var sb strings.Builder
	switch r.KeyType {
	case RecKeyIdent:
		sb.WriteString(r.Key)
	case RecKeyString:
		sb.WriteString(strconv.Quote(r.Key))
	}
	sb.WriteString(": ")
	sb.WriteString(r.Value.String())
	return sb.String()
}

func (p *parser) recInit() (RecInit, error) {
	var res RecInit
	var err error
	t := p.advance()
	switch {
	case t.isIdent():
		res.KeyType = RecKeyIdent
		res.Key = t.Text
	case t.isString():
		res.KeyType = RecKeyString
		if res.Key, err = t.stringValue(); err != nil {
			return res, err
		}
	default:
		return res, p.errorf("unexpected token")
	}
	if err := p.exact(":"); err != nil {
		return res, err
	}
	if res.Value, err = p.expression(); err != nil {
		return res, err
	}
	return res, nil
}
