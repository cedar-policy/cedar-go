package ast

import (
	"net/netip"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/types"
)

func Boolean(b bool) Node {
	return wrapNode(ast.Boolean(b))
}

func True() Node {
	return Boolean(true)
}

func False() Node {
	return Boolean(false)
}

func String(s string) Node {
	return wrapNode(ast.String(s))
}

func Long(l int64) Node {
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
	Key   string
	Value Node
}

type Pairs []Pair

// Record, TODO: document how duplicate keys might not really get handled in a meaningful way
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

func IPAddr(i netip.Prefix) Node {
	return wrapNode(ast.IPAddr(i))
}

func Value(v types.Value) Node {
	return wrapNode(ast.Value(v))
}
