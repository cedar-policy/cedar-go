package ast

import "github.com/cedar-policy/cedar-go/types"

type PolicySet map[string]PolicySetEntry

type PolicySetEntry struct {
	Policy   Policy
	Position Position
}

type annotationType struct {
	Key   types.String // TODO: review type
	Value types.String
}
type condition bool

const (
	conditionWhen   = true
	conditionUnless = false
)

type conditionType struct {
	Condition condition
	Body      node
}

type effect bool

const (
	effectPermit effect = true
	effectForbid effect = false
)

type Policy struct {
	effect      effect
	annotations []annotationType
	principal   isScopeNode
	action      isScopeNode
	resource    isScopeNode
	conditions  []conditionType
}

func newPolicy(effect effect, annotations []annotationType) *Policy {
	return &Policy{
		effect:      effect,
		annotations: annotations,
		principal:   scope(rawPrincipalNode()).All(),
		action:      scope(rawActionNode()).All(),
		resource:    scope(rawResourceNode()).All(),
	}
}

func Permit() *Policy {
	return newPolicy(effectPermit, nil)
}

func Forbid() *Policy {
	return newPolicy(effectForbid, nil)
}

func (p *Policy) When(node Node) *Policy {
	p.conditions = append(p.conditions, conditionType{Condition: conditionWhen, Body: node.v})
	return p
}

func (p *Policy) Unless(node Node) *Policy {
	p.conditions = append(p.conditions, conditionType{Condition: conditionUnless, Body: node.v})
	return p
}
