package ast

import "github.com/cedar-policy/cedar-go/internal/ast"

type Policy ast.Policy

func wrapPolicy(p *ast.Policy) *Policy {
	return (*Policy)(p)
}

func (p *Policy) unwrap() *ast.Policy {
	return (*ast.Policy)(p)
}

// Permit creates a new Permit policy.
func Permit() *Policy {
	return wrapPolicy(ast.Permit())
}

// Forbid creates a new Forbid policy.
func Forbid() *Policy {
	return wrapPolicy(ast.Forbid())
}

// When adds a conditional clause.
func (p *Policy) When(node Node) *Policy {
	return wrapPolicy(p.unwrap().When(node.Node))
}

// Unless adds a conditional clause.
func (p *Policy) Unless(node Node) *Policy {
	return wrapPolicy(p.unwrap().Unless(node.Node))
}
