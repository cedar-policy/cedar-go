package ast

import "github.com/cedar-policy/cedar-go/internal/ast"

type Policy struct {
	*ast.Policy
}

func newPolicy(p *ast.Policy) *Policy {
	return &Policy{p}
}

func Permit() *Policy {
	return newPolicy(ast.Permit())
}

func Forbid() *Policy {
	return newPolicy(ast.Forbid())
}

func (p *Policy) When(node Node) *Policy {
	return newPolicy(p.Policy.When(node.Node))
}

func (p *Policy) Unless(node Node) *Policy {
	return newPolicy(p.Policy.Unless(node.Node))
}
