package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

type AnnotationType struct {
	Key   types.Ident
	Value types.String
}
type Condition bool

const (
	ConditionWhen   = true
	ConditionUnless = false
)

type ConditionType struct {
	Condition Condition
	Body      IsNode
}

type Effect bool

const (
	EffectPermit Effect = true
	EffectForbid Effect = false
)

// Position is a value that represents a source Position.
// A Position is valid if Line > 0.
type Position struct {
	Filename string // optional name of the source file for the enclosing policy, "" if the source is unknown or not a named file
	Offset   int    // byte offset, starting at 0
	Line     int    // line number, starting at 1
	Column   int    // column number, starting at 1 (character count per line)
}

type Policy struct {
	Effect      Effect
	Annotations []AnnotationType // duplicate keys are prevented via the builders
	Principal   IsPrincipalScopeNode
	Action      IsActionScopeNode
	Resource    IsResourceScopeNode
	Conditions  []ConditionType
	Position    Position
}

func newPolicy(effect Effect, annotations []AnnotationType) *Policy {
	return &Policy{
		Effect:      effect,
		Annotations: annotations,
		Principal:   Scope{}.All(),
		Action:      Scope{}.All(),
		Resource:    Scope{}.All(),
	}
}

func Permit() *Policy {
	return newPolicy(EffectPermit, nil)
}

func Forbid() *Policy {
	return newPolicy(EffectForbid, nil)
}

func (p *Policy) When(node Node) *Policy {
	p.Conditions = append(p.Conditions, ConditionType{Condition: ConditionWhen, Body: node.v})
	return p
}

func (p *Policy) Unless(node Node) *Policy {
	p.Conditions = append(p.Conditions, ConditionType{Condition: ConditionUnless, Body: node.v})
	return p
}
