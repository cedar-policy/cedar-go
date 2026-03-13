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

func (n StrOpNode) inspect(fn func(IsNode) bool) {
	inspectNode(n.Arg, fn)
}

type BinaryNode struct {
	Left, Right IsNode
}

func (n BinaryNode) inspect(fn func(IsNode) bool) {
	inspectNode(n.Left, fn)
	inspectNode(n.Right, fn)
}

type NodeTypeIfThenElse struct {
	If, Then, Else IsNode
}

func (n NodeTypeIfThenElse) isNode() { _ = "hack for code coverage" }
func (n NodeTypeIfThenElse) inspect(fn func(IsNode) bool) {
	inspectNode(n.If, fn)
	inspectNode(n.Then, fn)
	inspectNode(n.Else, fn)
}

type NodeTypeOr struct{ BinaryNode }

func (n NodeTypeOr) isNode() { _ = "hack for code coverage" }

type NodeTypeAnd struct {
	BinaryNode
}

func (n NodeTypeAnd) isNode() { _ = "hack for code coverage" }

type NodeTypeLessThan struct {
	BinaryNode
}

func (n NodeTypeLessThan) isNode() { _ = "hack for code coverage" }

type NodeTypeLessThanOrEqual struct {
	BinaryNode
}

func (n NodeTypeLessThanOrEqual) isNode() { _ = "hack for code coverage" }

type NodeTypeGreaterThan struct {
	BinaryNode
}

func (n NodeTypeGreaterThan) isNode() { _ = "hack for code coverage" }

type NodeTypeGreaterThanOrEqual struct {
	BinaryNode
}

func (n NodeTypeGreaterThanOrEqual) isNode() { _ = "hack for code coverage" }

type NodeTypeNotEquals struct {
	BinaryNode
}

func (n NodeTypeNotEquals) isNode() { _ = "hack for code coverage" }

type NodeTypeEquals struct {
	BinaryNode
}

func (n NodeTypeEquals) isNode() { _ = "hack for code coverage" }

type NodeTypeIn struct {
	BinaryNode
}

func (n NodeTypeIn) isNode() { _ = "hack for code coverage" }

type NodeTypeHas struct {
	StrOpNode
}

func (n NodeTypeHas) isNode() { _ = "hack for code coverage" }

type NodeTypeHasTag struct{ BinaryNode }

func (n NodeTypeHasTag) isNode() { _ = "hack for code coverage" }

type NodeTypeLike struct {
	Arg   IsNode
	Value types.Pattern
}

func (n NodeTypeLike) isNode() { _ = "hack for code coverage" }
func (n NodeTypeLike) inspect(fn func(IsNode) bool) {
	inspectNode(n.Arg, fn)
}

type NodeTypeIs struct {
	Left       IsNode
	EntityType types.EntityType
}

func (n NodeTypeIs) isNode() { _ = "hack for code coverage" }
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

func (n NodeTypeSub) isNode() { _ = "hack for code coverage" }

type NodeTypeAdd struct {
	BinaryNode
	AddNode
}

func (n NodeTypeAdd) isNode() { _ = "hack for code coverage" }

type NodeTypeMult struct{ BinaryNode }

func (n NodeTypeMult) isNode() { _ = "hack for code coverage" }

type UnaryNode struct {
	Arg IsNode
}

func (n UnaryNode) inspect(fn func(IsNode) bool) {
	inspectNode(n.Arg, fn)
}

type NodeTypeNegate struct{ UnaryNode }

func (n NodeTypeNegate) isNode() { _ = "hack for code coverage" }

type NodeTypeNot struct{ UnaryNode }

func (n NodeTypeNot) isNode() { _ = "hack for code coverage" }

type NodeTypeAccess struct{ StrOpNode }

func (n NodeTypeAccess) isNode() { _ = "hack for code coverage" }

type NodeTypeGetTag struct{ BinaryNode }

func (n NodeTypeGetTag) isNode() { _ = "hack for code coverage" }

type NodeTypeExtensionCall struct {
	Name types.Path
	Args []IsNode
}

func (n NodeTypeExtensionCall) isNode() { _ = "hack for code coverage" }
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

func (n NodeTypeContains) isNode() { _ = "hack for code coverage" }

type NodeTypeContainsAll struct {
	BinaryNode
}

func (n NodeTypeContainsAll) isNode() { _ = "hack for code coverage" }

type NodeTypeContainsAny struct {
	BinaryNode
}

func (n NodeTypeContainsAny) isNode() { _ = "hack for code coverage" }

type NodeTypeIsEmpty struct {
	UnaryNode
}

func (n NodeTypeIsEmpty) isNode() { _ = "hack for code coverage" }

type NodeValue struct {
	Value types.Value
}

func (n NodeValue) isNode()                   { _ = "hack for code coverage" }
func (n NodeValue) inspect(func(IsNode) bool) { _ = "hack for code coverage" }

type RecordElementNode struct {
	Key   types.String
	Value IsNode
}

type NodeTypeRecord struct {
	Elements []RecordElementNode
}

func (n NodeTypeRecord) isNode() { _ = "hack for code coverage" }
func (n NodeTypeRecord) inspect(fn func(IsNode) bool) {
	for _, e := range n.Elements {
		inspectNode(e.Value, fn)
	}
}

type NodeTypeSet struct {
	Elements []IsNode
}

func (n NodeTypeSet) isNode() { _ = "hack for code coverage" }
func (n NodeTypeSet) inspect(fn func(IsNode) bool) {
	for _, e := range n.Elements {
		inspectNode(e, fn)
	}
}

type NodeTypeVariable struct {
	Name types.String
}

func (n NodeTypeVariable) isNode()                   { _ = "hack for code coverage" }
func (n NodeTypeVariable) inspect(func(IsNode) bool) { _ = "hack for code coverage" }

//sumtype:decl
type IsNode interface {
	isNode()
	inspect(func(IsNode) bool)
}
