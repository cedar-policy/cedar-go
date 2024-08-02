package ast

type PolicySet []Policy

type Policy struct {
	effect      effect
	annotations []nodeTypeAnnotation
	principal   isScopeNode
	action      isScopeNode
	resource    isScopeNode
	conditions  []nodeTypeCondition
}

func newPolicy(effect effect, annotations []nodeTypeAnnotation) *Policy {
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
	p.conditions = append(p.conditions, nodeTypeCondition{Condition: conditionWhen, Body: node.v})
	return p
}

func (p *Policy) Unless(node Node) *Policy {
	p.conditions = append(p.conditions, nodeTypeCondition{Condition: conditionUnless, Body: node.v})
	return p
}

type effect bool

const (
	effectPermit effect = true
	effectForbid effect = false
)
