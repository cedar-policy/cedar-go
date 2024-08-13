package ast

import (
	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/types"
)

type Annotations struct {
	*ast.Annotations
}

// Annotation allows AST constructors to make policy in a similar shape to textual Cedar with
// annotations appearing before the actual policy scope:
//
//	ast := Annotation("foo", "bar").
//	    Annotation("baz", "quux").
//		Permit().
//		PrincipalEq(superUser)
func Annotation(name, value types.String) *Annotations {
	return &Annotations{ast.Annotation(name, value)}
}

func (a *Annotations) Annotation(name, value types.String) *Annotations {
	return &Annotations{a.Annotations.Annotation(name, value)}
}

func (a *Annotations) Permit() *Policy {
	return &Policy{a.Annotations.Permit()}
}

func (a *Annotations) Forbid() *Policy {
	return &Policy{a.Annotations.Forbid()}
}

func (p *Policy) Annotate(name, value types.String) *Policy {
	return &Policy{p.Policy.Annotate(name, value)}
}
