package parser

import "github.com/cedar-policy/cedar-go/internal/ast"

type PolicySlice []*Policy
type Policy ast.Policy

func (p *Policy) unwrap() *ast.Policy {
    return (*ast.Policy)(p)
}

type PolicySlice2 struct {
    StaticPolicies []*Policy
    Templates      []*Template
}
