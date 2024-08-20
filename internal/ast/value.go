package ast

import (
	"net/netip"

	"github.com/cedar-policy/cedar-go/types"
)

func Boolean[T bool | types.Boolean](b T) Node {
	return Value(types.Boolean(b))
}

func True() Node {
	return Boolean(true)
}

func False() Node {
	return Boolean(false)
}

func String[T string | types.String](s T) Node {
	return Value(types.String(s))
}

func Long[T int | int64 | types.Long](l T) Node {
	return Value(types.Long(l))
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

// In the case where duplicate keys exist, the latter value will be preserved.
func Record(elements Pairs) Node {
	var res NodeTypeRecord
	m := make(map[types.String]int, len(elements))
	for _, v := range elements {
		if i, ok := m[v.Key]; ok {
			res.Elements[i] = RecordElementNode{Key: types.String(v.Key), Value: v.Value.v}
			continue
		}
		m[v.Key] = len(res.Elements)
		res.Elements = append(res.Elements, RecordElementNode{Key: types.String(v.Key), Value: v.Value.v})
	}
	return NewNode(res)
}

func EntityUID(typ types.Ident, id types.String) Node {
	e := types.NewEntityUID(types.Path(typ), types.String(id))
	return Value(e)
}

func IPAddr[T netip.Prefix | types.IPAddr](i T) Node {
	return Value(types.IPAddr(i))
}

func ExtensionCall(name types.Path, args ...Node) Node {
	return NewExtensionCall(name, args...)
}

func Value(v types.Value) Node {
	return NewNode(NodeValue{Value: v})
}
