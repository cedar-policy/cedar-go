package ast

type Policy struct {
	effect      effect
	annotations []Node
	principal   Node
	action      Node
	resource    Node
	conditions  []Node
}

func Permit() *Policy {
	return &Policy{effect: effectPermit}
}

func Forbid() *Policy {
	return &Policy{effect: effectForbid}
}

func (p *Policy) When(node Node) *Policy {
	p.conditions = append(p.conditions, node)
	return p
}

func (p *Policy) Unless(node Node) *Policy {
	p.conditions = append(p.conditions, Not(node))
	return p
}

type effect bool

const (
	effectPermit effect = true
	effectForbid effect = false
)
