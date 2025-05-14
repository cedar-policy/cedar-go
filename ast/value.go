package ast

import (
	"net/netip"
	"time"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

// Boolean creates a value node containing a Boolean.
func Boolean[T ~bool](b T) Node {
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
func String[T ~string](s T) Node {
	return wrapNode(ast.String(types.String(s)))
}

// Long creates a value node containing a Long.
func Long[T ~int | ~int64](l T) Node {
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

// Pairs is a slice of Pair elements
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

// IPAddr creates a value node containing an IPAddr.
func IPAddr[T netip.Prefix | types.IPAddr](i T) Node {
	return wrapNode(ast.IPAddr(types.IPAddr(i)))
}

// Datetime creates a value node containing a timestamp
func Datetime(t time.Time) Node {
	return Value(types.NewDatetime(t))
}

// Duration creates a value node containing a duration
func Duration(d time.Duration) Node {
	return Value(types.NewDuration(d))
}

// Value creates a value node from any value.
func Value(v types.Value) Node {
	return wrapNode(ast.Value(v))
}

// DecimalExtensionCall wraps a node with the cedar `decimal()` extension call
func DecimalExtensionCall(rhs Node) Node {
	return wrapNode(ast.ExtensionCall("decimal", rhs.Node))
}

// IPExtensionCall wraps a node with the cedar `ip()` extension call
func IPExtensionCall(rhs Node) Node {
	return wrapNode(ast.ExtensionCall("ip", rhs.Node))
}

// DatetimeExtensionCall wraps a node with the cedar `datetime()` extension call
func DatetimeExtensionCall(rhs Node) Node {
	return wrapNode(ast.ExtensionCall("datetime", rhs.Node))
}

// DurationExtensionCall wraps a node with the cedar `duration()` extension call
func DurationExtensionCall(rhs Node) Node {
	return wrapNode(ast.ExtensionCall("duration", rhs.Node))
}
