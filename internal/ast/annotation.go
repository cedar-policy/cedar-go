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
func Annotation(key types.Ident, value types.String) *Annotations {
	return &Annotations{nodes: []AnnotationType{newAnnotation(key, value)}}
}

func addAnnotation(in []AnnotationType, key types.Ident, value types.String) []AnnotationType {
	for i, aa := range in {
		if aa.Key == key {
			in[i] = newAnnotation(key, value)
			return in
		}
	}
	return append(in, newAnnotation(key, value))
}

func (a *Annotations) Annotation(key types.Ident, value types.String) *Annotations {
	a.nodes = addAnnotation(a.nodes, key, value)
	return a
}

func (a *Annotations) Permit() *Policy {
	return newPolicy(EffectPermit, a.nodes)
}

func (a *Annotations) Forbid() *Policy {
	return newPolicy(EffectForbid, a.nodes)
}

func (p *Policy) Annotate(key types.Ident, value types.String) *Policy {
	p.Annotations = addAnnotation(p.Annotations, key, value)
	return p
}

func newAnnotation(key types.Ident, value types.String) AnnotationType {
	return AnnotationType{Key: key, Value: value}
}
