package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

func Boolean(b types.Boolean) Node {
	return Value(b)
}

func True() Node {
	return Boolean(true)
}

func False() Node {
	return Boolean(false)
}

func String(s types.String) Node {
	return Value(s)
}

func Long(l types.Long) Node {
	return Value(l)
}

// SetDeprecated is a convenience function that wraps concrete instances of a Cedar SetDeprecated type
// types in AST value nodes and passes them along to SetNodes.
func SetDeprecated(s types.Set) Node {
	var nodes []IsNode
	for _, v := range s {
		nodes = append(nodes, Value(v).v)
	}
	return NewNode(NodeTypeSet{Elements: nodes})
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
	return NewNode(NodeTypeSet{Elements: stripNodes(nodes)})
}

type Pair struct {
	Key   types.String
	Value Node
}

type Pairs []Pair

func Record(elements Pairs) Node {
	var res NodeTypeRecord
	for _, e := range elements {
		res.Elements = append(res.Elements, RecordElementNode{Key: e.Key, Value: e.Value.v})
	}
	return NewNode(res)
}

func EntityUID(e types.EntityUID) Node {
	return Value(e)
}

func Decimal(d types.Decimal) Node {
	return Value(d)
}

func IPAddr(i types.IPAddr) Node {
	return Value(i)
}

func ExtensionCall(name types.String, args ...Node) Node {
	return NewExtensionCall(name, args...)
}

func Value(v types.Value) Node {
	return NewNode(NodeValue{Value: v})
}
