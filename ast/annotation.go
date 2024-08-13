package ast

import (
	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/types"
)

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
func Annotation(name, value types.String) *Annotations {
	return wrapAnnotations(ast.Annotation(name, value))
}

func (a *Annotations) Annotation(name, value types.String) *Annotations {
	return wrapAnnotations(a.unwrap().Annotation(name, value))
}

func (a *Annotations) Permit() *Policy {
	return wrapPolicy(a.unwrap().Permit())
}

func (a *Annotations) Forbid() *Policy {
	return wrapPolicy(a.unwrap().Forbid())
}

func (p *Policy) Annotate(name, value types.String) *Policy {
	return wrapPolicy(p.unwrap().Annotate(name, value))
}
