package ast

type Policy struct {
	effect      effect
	annotations []Node
	principal   Node
	action      Node
	resource    Node
	conditions  []Node
}

func newPolicy(effect effect, annotations []Node) *Policy {
	return &Policy{
		effect:      effect,
		annotations: annotations,
		principal:   Node{nodeType: nodeTypeAll},
		action:      Node{nodeType: nodeTypeAll},
		resource:    Node{nodeType: nodeTypeAll},
	}
}

func Permit() *Policy {
	return newPolicy(effectPermit, nil)
}

func Forbid() *Policy {
	return newPolicy(effectForbid, nil)
}

func (p *Policy) When(node Node) *Policy {
	p.conditions = append(p.conditions, Node{nodeType: nodeTypeWhen, args: []Node{node}})
	return p
}

func (p *Policy) Unless(node Node) *Policy {
	p.conditions = append(p.conditions, Node{nodeType: nodeTypeUnless, args: []Node{node}})
	return p
}

type effect bool

const (
	effectPermit effect = true
	effectForbid effect = false
)
