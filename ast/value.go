package ast

import (
	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/types"
)

func Boolean(b types.Boolean) Node {
	return wrapNode(ast.Boolean(b))
}

func True() Node {
	return Boolean(true)
}

func False() Node {
	return Boolean(false)
}

func String(s types.String) Node {
	return wrapNode(ast.String(s))
}

func Long(l types.Long) Node {
	return wrapNode(ast.Long(l))
}

// Set allows for a complex set definition with values potentially
// being Cedar expressions of their own. For example, this Cedar text:
//
//	[1, 2 + 3, context.fooCount]
//
// could be expressed in Golang as:
//
//	ast.Set(
//	    ast.Long(1),
//	    ast.Long(2).Plus(ast.Long(3)),
//	    ast.Context().Access("fooCount"),
//	)
func Set(nodes ...Node) Node {
	var astNodes []ast.Node
	for _, n := range nodes {
		astNodes = append(astNodes, n.Node)
	}
	return wrapNode(ast.Set(astNodes...))
}

type Pair struct {
	Key   types.String
	Value Node
}

type Pairs []Pair

func Record(elements Pairs) Node {
	var astNodes []ast.Pair
	for _, v := range elements {
		astNodes = append(astNodes, ast.Pair{Key: v.Key, Value: v.Value.Node})
	}
	return wrapNode(ast.Record(astNodes))
}

func EntityUID(e types.EntityUID) Node {
	return wrapNode(ast.EntityUID(e))
}

func Decimal(d types.Decimal) Node {
	return wrapNode(ast.Decimal(d))
}

func IPAddr(i types.IPAddr) Node {
	return wrapNode(ast.IPAddr(i))
}

func Value(v types.Value) Node {
	return wrapNode(ast.Value(v))
}
