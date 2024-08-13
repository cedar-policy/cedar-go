package ast

import "github.com/cedar-policy/cedar-go/internal/ast"

func Principal() Node {
	return wrapNode(ast.Principal())
}

func Action() Node {
	return wrapNode(ast.Action())
}

func Resource() Node {
	return wrapNode(ast.Resource())
}

func Context() Node {
	return wrapNode(ast.Context())
}
