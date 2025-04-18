package ast

import (
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

// Annotations allows access to Cedar annotations on a policy
type Annotations ast.Annotations

func (a *Annotations) unwrap() *ast.Annotations {
	return (*ast.Annotations)(a)
}

func wrapAnnotations(a *ast.Annotations) *Annotations {
	return (*Annotations)(a)
}

// Annotation allows AST constructors to make policy in a similar shape to textual Cedar with
// annotations appearing before the actual policy scope:
//
//	ast := Annotation("foo", "bar").
//	    Annotation("baz", "quux").
//		Permit().
//		PrincipalEq(superUser)
func Annotation(key types.Ident, value types.String) *Annotations {
	return wrapAnnotations(ast.Annotation(key, value))
}

// Annotation adds an annotation.  If a previous annotation exists with the same key, this builder will replace it.
func (a *Annotations) Annotation(key types.Ident, value types.String) *Annotations {
	return wrapAnnotations(a.unwrap().Annotation(key, value))
}

// Permit begins a permit policy from the given annotations.
func (a *Annotations) Permit() *Policy {
	return wrapPolicy(a.unwrap().Permit())
}

// Forbid begins a forbid policy from the given annotations.
func (a *Annotations) Forbid() *Policy {
	return wrapPolicy(a.unwrap().Forbid())
}

// Annotate adds an annotation to a Policy. If a previous annotation exists with the same key, this builder will
// replace it.
func (p *Policy) Annotate(key types.Ident, value types.String) *Policy {
	return wrapPolicy(p.unwrap().Annotate(key, value))
}
