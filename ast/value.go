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

// Set is a convenience function that wraps concrete instances of a Cedar Set type
// types in AST value nodes and passes them along to SetNodes.
func Set(s types.Set) Node {
	return wrapNode(ast.Set(s))
}

// SetNodes allows for a complex set definition with values potentially
// being Cedar expressions of their own. For example, this Cedar text:
//
//	[1, 2 + 3, context.fooCount]
//
// could be expressed in Golang as:
//
//	ast.SetNodes(
//	    ast.Long(1),
//	    ast.Long(2).Plus(ast.Long(3)),
//	    ast.Context().Access("fooCount"),
//	)
func SetNodes(nodes ...Node) Node {
	var astNodes []ast.Node
	for _, n := range nodes {
		astNodes = append(astNodes, n.Node)
	}
	return wrapNode(ast.SetNodes(astNodes...))
}

// Record is a convenience function that wraps concrete instances of a Cedar Record type
// types in AST value nodes and passes them along to RecordNodes.
func Record(r types.Record) Node {
	return wrapNode(ast.Record(r))
}

// RecordNodes allows for a complex record definition with values potentially
// being Cedar expressions of their own. For example, this Cedar text:
//
//	{"x": 1 + context.fooCount}
//
// could be expressed in Golang as:
//
//	ast.RecordNodes(map[types.String]Node{
//	    "x": ast.Long(1).Plus(ast.Context().Access("fooCount"))},
//	})
func RecordNodes(entries map[types.String]Node) Node {
	astNodes := map[types.String]ast.Node{}
	for k, v := range entries {
		astNodes[k] = v.Node
	}
	return wrapNode(ast.RecordNodes(astNodes))
}

type RecordElement struct {
	Key   types.String
	Value Node
}

func RecordElements(elements ...RecordElement) Node {
	var astNodes []ast.RecordElement
	for _, v := range elements {
		astNodes = append(astNodes, ast.RecordElement{Key: v.Key, Value: v.Value.Node})
	}
	return wrapNode(ast.RecordElements(astNodes...))
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
