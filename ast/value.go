package ast

import (
	"net/netip"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/types"
)

// Boolean creates a value node containing a Boolean.
func Boolean[T bool | types.Boolean](b T) Node {
	return wrapNode(ast.Boolean(types.Boolean(b)))
}

// True creates a value node containing True.
func True() Node {
	return Boolean(true)
}

// False creates a value node containing False.
func False() Node {
	return Boolean(false)
}

// String creates a value node containing a String.
func String[T string | types.String](s T) Node {
	return wrapNode(ast.String(types.String(s)))
}

// Long creates a value node containing a Long.
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

// Pair is map of Key string to Value node.
type Pair struct {
	Key   types.String
	Value Node
}

type Pairs []Pair

// Record creates a record node.  In the case where duplicate keys exist, the latter value will be preserved.
func Record(elements Pairs) Node {
	var astNodes []ast.Pair
	for _, v := range elements {
		astNodes = append(astNodes, ast.Pair{Key: v.Key, Value: v.Value.Node})
	}
	return wrapNode(ast.Record(astNodes))
}

// EntityUID creates a value node containing an EntityUID.
func EntityUID(typ types.Ident, id types.String) Node {
	return wrapNode(ast.EntityUID(typ, id))
}

// IPAddr creates an value node containing an IPAddr.
func IPAddr[T netip.Prefix | types.IPAddr](i T) Node {
	return wrapNode(ast.IPAddr(types.IPAddr(i)))
}

// Value creates a value node from any value.
func Value(v types.Value) Node {
	return wrapNode(ast.Value(v))
}
