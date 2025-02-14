package ast

import "github.com/cedar-policy/cedar-go/x/exp/ast"

// Node is a wrapper type for all the Cedar language operators.  See the [Cedar operators documentation] for details.
//
// [Cedar operators documentation]: https://docs.cedarpolicy.com/policies/syntax-operators.html
type Node struct {
	ast.Node
}

func wrapNode(n ast.Node) Node {
	return Node{n}
}
