package ast

import "github.com/cedar-policy/cedar-go/internal/ast"

type Policy ast.Policy

func wrapPolicy(p *ast.Policy) *Policy {
	return (*Policy)(p)
}

func (p *Policy) unwrap() *ast.Policy {
	return (*ast.Policy)(p)
}

func Permit() *Policy {
	return wrapPolicy(ast.Permit())
}

func Forbid() *Policy {
	return wrapPolicy(ast.Forbid())
}

func (p *Policy) When(node Node) *Policy {
	return wrapPolicy(p.unwrap().When(node.Node))
}

func (p *Policy) Unless(node Node) *Policy {
	return wrapPolicy(p.unwrap().Unless(node.Node))
}
