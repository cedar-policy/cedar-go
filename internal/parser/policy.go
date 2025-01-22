package parser

import "github.com/cedar-policy/cedar-go/internal/ast"

//type PolicySlice []*Policy
type Policy ast.Policy

func (p *Policy) unwrap() *ast.Policy {
    return (*ast.Policy)(p)
}

//todo: fix
type PolicySlice struct {
    StaticPolicies []*Policy
    Templates      []*Template
}
