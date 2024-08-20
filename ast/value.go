package ast

import (
	"net/netip"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/types"
)

func Boolean[T bool | types.Boolean](b T) Node {
	return wrapNode(ast.Boolean(types.Boolean(b)))
}

func True() Node {
	return Boolean(true)
}

func False() Node {
	return Boolean(false)
}

func String[T string | types.String](s T) Node {
	return wrapNode(ast.String(types.String(s)))
}

func Long[T int | int64 | types.Long](l T) Node {
	return wrapNode(ast.Long(types.Long(l)))
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
	Key   string
	Value Node
}

type Pairs []Pair

// In the case where duplicate keys exist, the latter value will be preserved.
func Record(elements Pairs) Node {
	var astNodes []ast.Pair
	for _, v := range elements {
		astNodes = append(astNodes, ast.Pair{Key: v.Key, Value: v.Value.Node})
	}
	return wrapNode(ast.Record(astNodes))
}

func EntityUID(typ, id string) Node {
	return wrapNode(ast.EntityUID(typ, id))
}

func IPAddr[T netip.Prefix | types.IPAddr](i T) Node {
	return wrapNode(ast.IPAddr(types.IPAddr(i)))
}

func Value(v types.Value) Node {
	return wrapNode(ast.Value(v))
}
