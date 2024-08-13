package ast

import "github.com/cedar-policy/cedar-go/internal/ast"

func Principal() Node {
	return Node{ast.Principal()}
}

func Action() Node {
	return Node{ast.Action()}
}

func Resource() Node {
	return Node{ast.Resource()}
}

func Context() Node {
	return Node{ast.Context()}
}
