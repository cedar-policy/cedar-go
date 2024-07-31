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

type annotationNode Node

func (n annotationNode) Key() types.String {
	return n.args[0].value.(types.String)
}

func (n annotationNode) Value() types.String {
	return n.args[1].value.(types.String)
}

func (a *Annotations) Permit() *Policy {
	return newPolicy(effectPermit, a.nodes)
}

func (a *Annotations) Forbid() *Policy {
	return newPolicy(effectForbid, a.nodes)
}

func (p *Policy) Annotate(name, value types.String) *Policy {
	p.annotations = append(p.annotations, newAnnotationNode(name, value))
	return p
}

func newAnnotationNode(name, value types.String) Node {
	return Node{nodeType: nodeTypeAnnotation, args: []Node{String(name), String(value)}}
}
