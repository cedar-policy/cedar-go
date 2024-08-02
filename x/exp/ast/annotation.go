package ast

import "github.com/cedar-policy/cedar-go/types"

type Annotations struct {
	nodes []nodeTypeAnnotation
}

// Annotation allows AST constructors to make policy in a similar shape to textual Cedar with
// annotations appearing before the actual policy scope:
//
//	ast := Annotation("foo", "bar").
//	    Annotation("baz", "quux").
//		Permit().
//		PrincipalEq(superUser)
func Annotation(name, value types.String) *Annotations {
	return &Annotations{nodes: []nodeTypeAnnotation{newAnnotation(name, value)}}
}

func (a *Annotations) Annotation(name, value types.String) *Annotations {
	a.nodes = append(a.nodes, newAnnotation(name, value))
	return a
}

func (a *Annotations) Permit() *Policy {
	return newPolicy(effectPermit, a.nodes)
}

func (a *Annotations) Forbid() *Policy {
	return newPolicy(effectForbid, a.nodes)
}

func (p *Policy) Annotate(name, value types.String) *Policy {
	p.annotations = append(p.annotations, nodeTypeAnnotation{Key: name, Value: value})
	return p
}

func newAnnotation(name, value types.String) nodeTypeAnnotation {
	return nodeTypeAnnotation{Key: name, Value: value}
}
