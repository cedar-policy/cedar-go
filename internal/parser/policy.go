package parser

import "github.com/cedar-policy/cedar-go/internal/ast"

type PolicySet []Policy

type Policy struct {
	ast.Policy
}
