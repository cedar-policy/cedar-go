package parser

import "github.com/cedar-policy/cedar-go/x/exp/ast"

type Policy ast.Policy

func (p *Policy) unwrap() *ast.Policy {
	return (*ast.Policy)(p)
}

type PolicySlice struct {
	StaticPolicies []*Policy
	Templates      []*Template
}
