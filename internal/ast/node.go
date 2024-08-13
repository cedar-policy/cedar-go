package ast

import (
	"bytes"

	"github.com/cedar-policy/cedar-go/types"
)

type Node struct {
	v IsNode // NOTE: not an embed because a `Node` is not a `node`
}

func newNode(v IsNode) Node {
	return Node{v: v}
}

func NewNode(v IsNode) Node {
	return Node{v: v}
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

type nodePrecedenceLevel uint8

const (
	ifPrecedence       nodePrecedenceLevel = 0
	orPrecedence       nodePrecedenceLevel = 1
	andPrecedence      nodePrecedenceLevel = 2
	relationPrecedence nodePrecedenceLevel = 3
	addPrecedence      nodePrecedenceLevel = 4
	multPrecedence     nodePrecedenceLevel = 5
	unaryPrecedence    nodePrecedenceLevel = 6
	accessPrecedence   nodePrecedenceLevel = 7
	primaryPrecedence  nodePrecedenceLevel = 8
)

type NodeTypeIf struct {
	If, Then, Else IsNode
}

func (n NodeTypeIf) precedenceLevel() nodePrecedenceLevel {
	return ifPrecedence
}

func (n NodeTypeIf) isNode() {}

type NodeTypeOr struct{ BinaryNode }

func (n NodeTypeOr) precedenceLevel() nodePrecedenceLevel {
	return orPrecedence
}

type NodeTypeAnd struct {
	BinaryNode
}

func (n NodeTypeAnd) precedenceLevel() nodePrecedenceLevel {
	return andPrecedence
}

type RelationNode struct{}

func (n RelationNode) precedenceLevel() nodePrecedenceLevel {
	return relationPrecedence
}

type NodeTypeLessThan struct {
	BinaryNode
	RelationNode
}
type NodeTypeLessThanOrEqual struct {
	BinaryNode
	RelationNode
}
type NodeTypeGreaterThan struct {
	BinaryNode
	RelationNode
}
type NodeTypeGreaterThanOrEqual struct {
	BinaryNode
	RelationNode
}
type NodeTypeNotEquals struct {
	BinaryNode
	RelationNode
}
type NodeTypeEquals struct {
	BinaryNode
	RelationNode
}
type NodeTypeIn struct {
	BinaryNode
	RelationNode
}

type NodeTypeHas struct {
	StrOpNode
	RelationNode
}

type NodeTypeLike struct {
	Arg   IsNode
	Value types.Pattern
}

func (n NodeTypeLike) precedenceLevel() nodePrecedenceLevel {
	return relationPrecedence
}
func (n NodeTypeLike) isNode() {}

type NodeTypeIs struct {
	Left       IsNode
	EntityType types.Path
}

func (n NodeTypeIs) precedenceLevel() nodePrecedenceLevel {
	return relationPrecedence
}
func (n NodeTypeIs) isNode() {}

type NodeTypeIsIn struct {
	NodeTypeIs
	Entity IsNode
}

func (n NodeTypeIsIn) precedenceLevel() nodePrecedenceLevel {
	return relationPrecedence
}

type AddNode struct{}

func (n AddNode) precedenceLevel() nodePrecedenceLevel {
	return addPrecedence
}

type NodeTypeSub struct {
	BinaryNode
	AddNode
}

type NodeTypeAdd struct {
	BinaryNode
	AddNode
}

type NodeTypeMult struct{ BinaryNode }

func (n NodeTypeMult) precedenceLevel() nodePrecedenceLevel {
	return multPrecedence
}

type UnaryNode struct {
	Arg IsNode
}

func (n UnaryNode) precedenceLevel() nodePrecedenceLevel {
	return unaryPrecedence
}

func (n UnaryNode) isNode() {}

type NodeTypeNegate struct{ UnaryNode }
type NodeTypeNot struct{ UnaryNode }

type NodeTypeAccess struct{ StrOpNode }

func (n NodeTypeAccess) precedenceLevel() nodePrecedenceLevel {
	return accessPrecedence
}

type NodeTypeExtensionCall struct {
	Name types.String // TODO: review type
	Args []IsNode
}

func (n NodeTypeExtensionCall) precedenceLevel() nodePrecedenceLevel {
	return accessPrecedence
}
func (n NodeTypeExtensionCall) isNode() {}

func stripNodes(args []Node) []IsNode {
	res := make([]IsNode, len(args))
	for i, v := range args {
		res[i] = v.v
	}
	return res
}

func newExtensionCall(method types.String, args ...Node) Node {
	return newNode(NodeTypeExtensionCall{
		Name: method,
		Args: stripNodes(args),
	})
}

func newMethodCall(lhs Node, method types.String, args ...Node) Node {
	res := make([]IsNode, 1+len(args))
	res[0] = lhs.v
	for i, v := range args {
		res[i+1] = v.v
	}
	return newNode(NodeTypeExtensionCall{
		Name: method,
		Args: res,
	})
}

type ContainsNode struct{}

func (n ContainsNode) precedenceLevel() nodePrecedenceLevel {
	return accessPrecedence
}

type NodeTypeContains struct {
	BinaryNode
	ContainsNode
}
type NodeTypeContainsAll struct {
	BinaryNode
	ContainsNode
}
type NodeTypeContainsAny struct {
	BinaryNode
	ContainsNode
}

type PrimaryNode struct{}

func (n PrimaryNode) isNode() {}

func (n PrimaryNode) precedenceLevel() nodePrecedenceLevel {
	return primaryPrecedence
}

type NodeValue struct {
	PrimaryNode
	Value types.Value
}

func (n NodeValue) isNode() {}

type RecordElementNode struct {
	Key   types.String
	Value IsNode
}

type NodeTypeRecord struct {
	PrimaryNode
	Elements []RecordElementNode
}

func (n NodeTypeRecord) isNode() {}

type NodeTypeSet struct {
	PrimaryNode
	Elements []IsNode
}

type NodeTypeVariable struct {
	PrimaryNode
	Name types.String // TODO: Review type
}

type IsNode interface {
	isNode()
	marshalCedar(*bytes.Buffer)
	precedenceLevel() nodePrecedenceLevel
}
