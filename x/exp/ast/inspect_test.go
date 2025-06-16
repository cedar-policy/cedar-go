package ast

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestInspectCounts(t *testing.T) {
	t.Parallel()
	leaf1 := NodeValue{Value: types.Long(1)}
	leaf2 := NodeValue{Value: types.Long(2)}
	cases := []struct {
		name string
		node Node
		want int
	}{
		{"IfThenElse", NewNode(NodeTypeIfThenElse{If: leaf1, Then: leaf1, Else: leaf1}), 4},
		{"Or", NewNode(NodeTypeOr{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"And", NewNode(NodeTypeAnd{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"LessThan", NewNode(NodeTypeLessThan{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"LessThanOrEqual", NewNode(NodeTypeLessThanOrEqual{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"GreaterThan", NewNode(NodeTypeGreaterThan{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"GreaterThanOrEqual", NewNode(NodeTypeGreaterThanOrEqual{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"NotEquals", NewNode(NodeTypeNotEquals{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"Equals", NewNode(NodeTypeEquals{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"In", NewNode(NodeTypeIn{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"HasTag", NewNode(NodeTypeHasTag{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"GetTag", NewNode(NodeTypeGetTag{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"Contains", NewNode(NodeTypeContains{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"ContainsAll", NewNode(NodeTypeContainsAll{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"ContainsAny", NewNode(NodeTypeContainsAny{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"Add", NewNode(NodeTypeAdd{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"Sub", NewNode(NodeTypeSub{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"Mult", NewNode(NodeTypeMult{BinaryNode: BinaryNode{Left: leaf1, Right: leaf2}}), 3},
		{"Has", NewNode(NodeTypeHas{StrOpNode: StrOpNode{Arg: leaf1, Value: "a"}}), 2},
		{"Access", NewNode(NodeTypeAccess{StrOpNode: StrOpNode{Arg: leaf1, Value: "a"}}), 2},
		{"Like", NewNode(NodeTypeLike{Arg: leaf1, Value: types.NewPattern(types.Wildcard{})}), 2},
		{"Is", NewNode(NodeTypeIs{Left: leaf1, EntityType: "T"}), 2},
		{"IsIn", NewNode(NodeTypeIsIn{NodeTypeIs: NodeTypeIs{Left: leaf1, EntityType: "T"}, Entity: leaf2}), 3},
		{"Negate", NewNode(NodeTypeNegate{UnaryNode: UnaryNode{Arg: leaf1}}), 2},
		{"Not", NewNode(NodeTypeNot{UnaryNode: UnaryNode{Arg: leaf1}}), 2},
		{"IsEmpty", NewNode(NodeTypeIsEmpty{UnaryNode: UnaryNode{Arg: leaf1}}), 2},
		{"ExtensionCall", NewNode(NodeTypeExtensionCall{Name: "f", Args: []IsNode{leaf1, leaf2}}), 3},
		{"Record", NewNode(NodeTypeRecord{Elements: []RecordElementNode{{Key: "k", Value: leaf1}}}), 2},
		{"Set", NewNode(NodeTypeSet{Elements: []IsNode{leaf1, leaf2}}), 3},
		{"Variable", NewNode(NodeTypeVariable{Name: "v"}), 1},
		{"StrOpNode", NewNode(StrOpNode{Arg: leaf1, Value: "a"}), 2},
		{"BinaryNode", NewNode(BinaryNode{Left: leaf1, Right: leaf2}), 3},
		{"UnaryNode", NewNode(UnaryNode{Arg: leaf1}), 2},
		{"Value", NewNode(leaf1), 1},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			count := 0
			Inspect(tt.node, func(IsNode) bool { count++; return true })
			testutil.Equals(t, count, tt.want)
		})
	}
}

func TestInspectSkipChildren(t *testing.T) {
	t.Parallel()
	leaf := NewNode(NodeValue{Value: types.Long(1)})
	root := NewNode(NodeTypeAnd{BinaryNode: BinaryNode{Left: leaf.v, Right: leaf.v}})
	var count int
	Inspect(root, func(n IsNode) bool {
		count++
		if _, ok := n.(NodeTypeAnd); ok {
			return false
		}
		return true
	})
	testutil.Equals(t, count, 1)
}

func TestInspectNil(t *testing.T) {
	t.Parallel()
	var c int
	Inspect(Node{}, func(IsNode) bool { c++; return true })
	testutil.Equals(t, c, 0)
}
