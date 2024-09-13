package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

type Node struct {
	v IsNode // NOTE: not an embed because a `Node` is not a `node`
}

func NewNode(v IsNode) Node {
	return Node{v: v}
}

func (n Node) AsIsNode() IsNode {
	return n.v
}

type StrOpNode struct {
	Arg   IsNode
	Value types.String
}

func (n StrOpNode) isNode() {}

type BinaryNode struct {
	Left, Right IsNode
}

func (n BinaryNode) isNode() {}

type NodeTypeIfThenElse struct {
	If, Then, Else IsNode
}

func (n NodeTypeIfThenElse) isNode() {}

type NodeTypeOr struct{ BinaryNode }

type NodeTypeAnd struct {
	BinaryNode
}

type NodeTypeLessThan struct {
	BinaryNode
}
type NodeTypeLessThanOrEqual struct {
	BinaryNode
}
type NodeTypeGreaterThan struct {
	BinaryNode
}
type NodeTypeGreaterThanOrEqual struct {
	BinaryNode
}
type NodeTypeNotEquals struct {
	BinaryNode
}
type NodeTypeEquals struct {
	BinaryNode
}
type NodeTypeIn struct {
	BinaryNode
}

type NodeTypeHas struct {
	StrOpNode
}

type NodeTypeLike struct {
	Arg   IsNode
	Value types.Pattern
}

func (n NodeTypeLike) isNode() {}

type NodeTypeIs struct {
	Left       IsNode
	EntityType types.EntityType
}

func (n NodeTypeIs) isNode() {}

type NodeTypeIsIn struct {
	NodeTypeIs
	Entity IsNode
}

type AddNode struct{}

type NodeTypeSub struct {
	BinaryNode
	AddNode
}

type NodeTypeAdd struct {
	BinaryNode
	AddNode
}

type NodeTypeMult struct{ BinaryNode }

type UnaryNode struct {
	Arg IsNode
}

func (n UnaryNode) isNode() {}

type NodeTypeNegate struct{ UnaryNode }
type NodeTypeNot struct{ UnaryNode }

type NodeTypeAccess struct{ StrOpNode }

type NodeTypeExtensionCall struct {
	Name types.Path
	Args []IsNode
}

func (n NodeTypeExtensionCall) isNode() {}

func stripNodes(args []Node) []IsNode {
	if args == nil {
		return nil
	}
	res := make([]IsNode, len(args))
	for i, v := range args {
		res[i] = v.v
	}
	return res
}

func NewExtensionCall(method types.Path, args ...Node) Node {
	return NewNode(NodeTypeExtensionCall{
		Name: method,
		Args: stripNodes(args),
	})
}

func NewMethodCall(lhs Node, method types.Path, args ...Node) Node {
	res := make([]IsNode, 1+len(args))
	res[0] = lhs.v
	for i, v := range args {
		res[i+1] = v.v
	}
	return NewNode(NodeTypeExtensionCall{
		Name: method,
		Args: res,
	})
}

type NodeTypeContains struct {
	BinaryNode
}
type NodeTypeContainsAll struct {
	BinaryNode
}
type NodeTypeContainsAny struct {
	BinaryNode
}

type NodeValue struct {
	Value types.Value
}

func (n NodeValue) isNode() {}

type RecordElementNode struct {
	Key   types.String
	Value IsNode
}

type NodeTypeRecord struct {
	Elements []RecordElementNode
}

func (n NodeTypeRecord) isNode() {}

type NodeTypeSet struct {
	Elements []IsNode
}

func (n NodeTypeSet) isNode() {}

type NodeTypeVariable struct {
	Name types.String
}

func (n NodeTypeVariable) isNode() {}

type IsNode interface {
	isNode()
}
