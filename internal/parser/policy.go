package parser

import "github.com/cedar-policy/cedar-go/internal/ast"

type PolicyID string

type PolicySet map[PolicyID]PolicySetEntry

type PolicySetEntry struct {
	Policy   Policy
	Position Position
}

type Policy struct {
	ast.Policy
}
