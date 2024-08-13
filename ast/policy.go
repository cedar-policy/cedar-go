package ast

import "github.com/cedar-policy/cedar-go/internal/ast"

type Policy struct {
	*ast.Policy
}

func Permit() *Policy {
	return &Policy{ast.Permit()}
}

func Forbid() *Policy {
	return &Policy{ast.Forbid()}
}

func (p *Policy) When(node Node) *Policy {
	return &Policy{p.Policy.When(node.Node)}
}

func (p *Policy) Unless(node Node) *Policy {
	return &Policy{p.Policy.Unless(node.Node)}
}
