package parser

import "github.com/cedar-policy/cedar-go/internal/ast"

type PolicySet map[string]PolicySetEntry

type PolicySetEntry struct {
	Policy   Policy
	Position Position
}

type Policy struct {
	ast.Policy
}
