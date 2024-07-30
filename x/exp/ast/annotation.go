package ast

import "github.com/cedar-policy/cedar-go/types"

type Annotations struct {
	nodes []Node
}

// Annotation allows AST constructors to make policy in a similar shape to textual Cedar with
// annotations appearing before the actual policy scope:
//
//	ast := Annotation("foo", "bar").
//	    Annotation("baz", "quux").
//		Permit().
//		PrincipalEq(superUser)
func Annotation(name, value types.String) *Annotations {
	return &Annotations{nodes: []Node{newAnnotationNode(name, value)}}
}

func (a *Annotations) Annotation(name, value types.String) *Annotations {
	a.nodes = append(a.nodes, newAnnotationNode(name, value))
	return a
}

func (a *Annotations) Permit() *Policy {
	p := Permit()
	p.annotations = a.nodes
	return p
}

func (a *Annotations) Forbid() *Policy {
	p := Forbid()
	p.annotations = a.nodes
	return p
}

func (p *Policy) Annotate(name, value types.String) *Policy {
	p.annotations = append(p.annotations, newAnnotationNode(name, value))
}

func newAnnotationNode(name, value types.String) Node {
	return Node{nodeType: nodeTypeAnnotation, args: []Node{String(name), String(value)}}
}
