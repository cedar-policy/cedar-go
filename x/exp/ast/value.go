package ast

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

func Boolean(b types.Boolean) Node {
	return newValueNode(nodeTypeBoolean, b)
}

func True() Node {
	return Boolean(true)
}

func False() Node {
	return Boolean(false)
}

func String(s types.String) Node {
	return newValueNode(nodeTypeString, s)
}

func Long(l types.Long) Node {
	return newValueNode(nodeTypeLong, l)
}

// Set is a convenience function that wraps concrete instances of a Cedar Set type
// types in AST value nodes and passes them along to SetNodes.
func Set(s types.Set) Node {
	var nodes []Node
	for _, v := range s {
		nodes = append(nodes, valueToNode(v))
	}
	return SetNodes(nodes)
}

// SetNodes allows for a complex set definition with values potentially
// being Cedar expressions of their own. For example, this Cedar text:
//
//	[1, 2 + 3, context.fooCount]
//
// could be expressed in Golang as:
//
//	ast.SetNodes([]ast.Node{
//	    ast.Long(1),
//	    ast.Long(2).Plus(ast.Long(3)),
//	    ast.Context().Access("fooCount"),
//	})
func SetNodes(nodes []Node) Node {
	return newValueNode(nodeTypeSet, nodes)
}

// Record is a convenience function that wraps concrete instances of a Cedar Record type
// types in AST value nodes and passes them along to RecordNodes.
func Record(r types.Record) Node {
	recordNodes := map[string]Node{}
	for k, v := range r {
		recordNodes[k] = valueToNode(v)
	}
	return RecordNodes(recordNodes)
}

// RecordNodes allows for a complex record definition with values potentially
// being Cedar expressions of their own. For example, this Cedar text:
//
//	{"x": 1 + context.fooCount}
//
// could be expressed in Golang as:
//
//		ast.RecordNodes([]ast.RecordNode{
//	     {Key: "x", Value: ast.Long(1).Plus(ast.Context().Access("resourceField"))},
//	 })
func RecordNodes(nodes map[string]Node) Node {
	return newValueNode(nodeTypeRecord, nodes)
}

func EntityType(e string) Node {
	return newValueNode(nodeTypeEntityType, e)
}

func Entity(e types.EntityUID) Node {
	return newValueNode(nodeTypeEntity, e)
}

func Decimal(d types.Decimal) Node {
	return newValueNode(nodeTypeEntity, d)
}

func IPAddr(i types.IPAddr) Node {
	return newValueNode(nodeTypeIpAddr, i)
}

func newValueNode(nodeType nodeType, v any) Node {
	return Node{nodeType: nodeType, value: v}
}

func valueToNode(v types.Value) Node {
	switch x := v.(type) {
	case types.Boolean:
		return Boolean(x)
	case types.String:
		return String(x)
	case types.Long:
		return Long(x)
	case types.Set:
		return Set(x)
	case types.Record:
		return Record(x)
	case types.EntityUID:
		return Entity(x)
	case types.Decimal:
		return Decimal(x)
	case types.IPAddr:
		return IPAddr(x)
	default:
		panic(fmt.Sprintf("unexpected value type: %T(%v)", v, v))
	}
}
