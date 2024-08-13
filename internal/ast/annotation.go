package ast

import "github.com/cedar-policy/cedar-go/types"

type Annotations struct {
	nodes []AnnotationType
}

// Annotation allows AST constructors to make policy in a similar shape to textual Cedar with
// annotations appearing before the actual policy scope:
//
//	ast := Annotation("foo", "bar").
//	    Annotation("baz", "quux").
//		Permit().
//		PrincipalEq(superUser)
func Annotation(name, value types.String) *Annotations {
	return &Annotations{nodes: []AnnotationType{newAnnotation(name, value)}}
}

func (a *Annotations) Annotation(name, value types.String) *Annotations {
	a.nodes = append(a.nodes, newAnnotation(name, value))
	return a
}

func (a *Annotations) Permit() *Policy {
	return newPolicy(EffectPermit, a.nodes)
}

func (a *Annotations) Forbid() *Policy {
	return newPolicy(EffectForbid, a.nodes)
}

func (p *Policy) Annotate(name, value types.String) *Policy {
	p.Annotations = append(p.Annotations, AnnotationType{Key: name, Value: value})
	return p
}

func newAnnotation(name, value types.String) AnnotationType {
	return AnnotationType{Key: name, Value: value}
}
