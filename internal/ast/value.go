package ast

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

func Boolean(b types.Boolean) Node {
	return NewValueNode(b)
}

func True() Node {
	return Boolean(true)
}

func False() Node {
	return Boolean(false)
}

func String(s types.String) Node {
	return NewValueNode(s)
}

func Long(l types.Long) Node {
	return NewValueNode(l)
}

// Set is a convenience function that wraps concrete instances of a Cedar Set type
// types in AST value nodes and passes them along to SetNodes.
func Set(s types.Set) Node {
	var nodes []IsNode
	for _, v := range s {
		nodes = append(nodes, valueToNode(v).v)
	}
	return NewNode(NodeTypeSet{Elements: nodes})
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
	return NewNode(NodeTypeSet{Elements: stripNodes(nodes)})
}

// Record is a convenience function that wraps concrete instances of a Cedar Record type
// types in AST value nodes and passes them along to RecordNodes.
func Record(r types.Record) Node {
	// TODO: this results in a double allocation, fix that
	recordNodes := map[types.String]Node{}
	for k, v := range r {
		recordNodes[types.String(k)] = valueToNode(v)
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
//	ast.RecordNodes(map[types.String]Node{
//	    "x": ast.Long(1).Plus(ast.Context().Access("fooCount"))},
//	})
func RecordNodes(entries map[types.String]Node) Node {
	var res NodeTypeRecord
	for k, v := range entries {
		res.Elements = append(res.Elements, RecordElementNode{Key: k, Value: v.v})
	}
	return NewNode(res)
}

type RecordElement struct {
	Key   types.String
	Value Node
}

func RecordElements(elements ...RecordElement) Node {
	var res NodeTypeRecord
	for _, e := range elements {
		res.Elements = append(res.Elements, RecordElementNode{Key: e.Key, Value: e.Value.v})
	}
	return NewNode(res)
}

func EntityUID(e types.EntityUID) Node {
	return NewValueNode(e)
}

func Decimal(d types.Decimal) Node {
	return NewValueNode(d)
}

func IPAddr(i types.IPAddr) Node {
	return NewValueNode(i)
}

func ExtensionCall(name types.String, args ...Node) Node {
	return NewExtensionCall(name, args...)
}

func NewValueNode(v types.Value) Node {
	return NewNode(NodeValue{Value: v})
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
		return EntityUID(x)
	case types.Decimal:
		return Decimal(x)
	case types.IPAddr:
		return IPAddr(x)
	default:
		panic(fmt.Sprintf("unexpected value type: %T(%v)", v, v))
	}
}
