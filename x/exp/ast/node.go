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

func (n StrOpNode) isNode() { _ = 0 } // No-op statement injected for code coverage instrumentation
func (n StrOpNode) inspect(fn func(IsNode) bool) {
	inspectNode(n.Arg, fn)
}

type BinaryNode struct {
	Left, Right IsNode
}

func (n BinaryNode) isNode() { _ = 0 } // No-op statement injected for code coverage instrumentation
func (n BinaryNode) inspect(fn func(IsNode) bool) {
	inspectNode(n.Left, fn)
	inspectNode(n.Right, fn)
}

type NodeTypeIfThenElse struct {
	If, Then, Else IsNode
}

func (n NodeTypeIfThenElse) isNode() { _ = 0 } // No-op statement injected for code coverage instrumentation
func (n NodeTypeIfThenElse) inspect(fn func(IsNode) bool) {
	inspectNode(n.If, fn)
	inspectNode(n.Then, fn)
	inspectNode(n.Else, fn)
}

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

type NodeTypeHasTag struct{ BinaryNode }

type NodeTypeLike struct {
	Arg   IsNode
	Value types.Pattern
}

func (n NodeTypeLike) isNode() { _ = 0 } // No-op statement injected for code coverage instrumentation
func (n NodeTypeLike) inspect(fn func(IsNode) bool) {
	inspectNode(n.Arg, fn)
}

type NodeTypeIs struct {
	Left       IsNode
	EntityType types.EntityType
}

func (n NodeTypeIs) isNode() { _ = 0 } // No-op statement injected for code coverage instrumentation
func (n NodeTypeIs) inspect(fn func(IsNode) bool) {
	inspectNode(n.Left, fn)
}

type NodeTypeIsIn struct {
	NodeTypeIs
	Entity IsNode
}

func (n NodeTypeIsIn) inspect(fn func(IsNode) bool) {
	n.NodeTypeIs.inspect(fn)
	inspectNode(n.Entity, fn)
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

func (n UnaryNode) isNode() { _ = 0 } // No-op statement injected for code coverage instrumentation
func (n UnaryNode) inspect(fn func(IsNode) bool) {
	inspectNode(n.Arg, fn)
}

type NodeTypeNegate struct{ UnaryNode }
type NodeTypeNot struct{ UnaryNode }

type NodeTypeAccess struct{ StrOpNode }

type NodeTypeGetTag struct{ BinaryNode }

type NodeTypeExtensionCall struct {
	Name types.Path
	Args []IsNode
}

func (n NodeTypeExtensionCall) isNode() { _ = 0 } // No-op statement injected for code coverage instrumentation
func (n NodeTypeExtensionCall) inspect(fn func(IsNode) bool) {
	for _, a := range n.Args {
		inspectNode(a, fn)
	}
}

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

type NodeTypeIsEmpty struct {
	UnaryNode
}

type NodeValue struct {
	Value types.Value
}

func (n NodeValue) isNode()                   { _ = 0 } // No-op statement injected for code coverage instrumentation
func (n NodeValue) inspect(func(IsNode) bool) { _ = 0 } // No-op statements injected for code coverage instrumentation

type RecordElementNode struct {
	Key   types.String
	Value IsNode
}

type NodeTypeRecord struct {
	Elements []RecordElementNode
}

func (n NodeTypeRecord) isNode() { _ = 0 } // No-op statement injected for code coverage instrumentation
func (n NodeTypeRecord) inspect(fn func(IsNode) bool) {
	for _, e := range n.Elements {
		inspectNode(e.Value, fn)
	}
}

type NodeTypeSet struct {
	Elements []IsNode
}

func (n NodeTypeSet) isNode() { _ = 0 } // No-op statement injected for code coverage instrumentation
func (n NodeTypeSet) inspect(fn func(IsNode) bool) {
	for _, e := range n.Elements {
		inspectNode(e, fn)
	}
}

type NodeTypeVariable struct {
	Name types.String
}

func (n NodeTypeVariable) isNode()                   { _ = 0 } // No-op statement injected for code coverage instrumentation
func (n NodeTypeVariable) inspect(func(IsNode) bool) { _ = 0 } // No-op statements injected for code coverage instrumentation

type IsNode interface {
	isNode()
	inspect(func(IsNode) bool)
}
