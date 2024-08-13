package ast

import "github.com/cedar-policy/cedar-go/internal/ast"

func Principal() Node {
	return newNode(ast.Principal())
}

func Action() Node {
	return newNode(ast.Action())
}

func Resource() Node {
	return newNode(ast.Resource())
}

func Context() Node {
	return newNode(ast.Context())
}
