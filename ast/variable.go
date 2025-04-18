package ast

import "github.com/cedar-policy/cedar-go/x/exp/ast"

// Principal represents the principal in the request
func Principal() Node {
	return wrapNode(ast.Principal())
}

// Action represents the action in the request
func Action() Node {
	return wrapNode(ast.Action())
}

// Resource represents the resource in the request
func Resource() Node {
	return wrapNode(ast.Resource())
}

// Context represents the context in the request
func Context() Node {
	return wrapNode(ast.Context())
}
