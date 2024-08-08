package ast

import (
	"bytes"

	"github.com/cedar-policy/cedar-go/types"
)

type Node struct {
	v node // NOTE: not an embed because a `Node` is not a `node`
}

func newNode(v node) Node {
	return Node{v: v}
}

type strOpNode struct {
	Arg   node
	Value types.String
}

func (n strOpNode) isNode() {}

type binaryNode struct {
	Left, Right node
}

func (n binaryNode) isNode() {}

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

type nodeTypeIf struct {
	If, Then, Else node
}

func (n nodeTypeIf) precedenceLevel() nodePrecedenceLevel {
	return ifPrecedence
}

func (n nodeTypeIf) isNode() {}

type nodeTypeOr struct{ binaryNode }

func (n nodeTypeOr) precedenceLevel() nodePrecedenceLevel {
	return orPrecedence
}

type nodeTypeAnd struct {
	binaryNode
}

func (n nodeTypeAnd) precedenceLevel() nodePrecedenceLevel {
	return andPrecedence
}

type relationNode struct{}

func (n relationNode) precedenceLevel() nodePrecedenceLevel {
	return relationPrecedence
}

type nodeTypeLessThan struct {
	binaryNode
	relationNode
}
type nodeTypeLessThanOrEqual struct {
	binaryNode
	relationNode
}
type nodeTypeGreaterThan struct {
	binaryNode
	relationNode
}
type nodeTypeGreaterThanOrEqual struct {
	binaryNode
	relationNode
}
type nodeTypeNotEquals struct {
	binaryNode
	relationNode
}
type nodeTypeEquals struct {
	binaryNode
	relationNode
}
type nodeTypeIn struct {
	binaryNode
	relationNode
}

type nodeTypeHas struct {
	strOpNode
	relationNode
}

type nodeTypeLike struct {
	Arg   node
	Value types.Pattern
}

func (n nodeTypeLike) precedenceLevel() nodePrecedenceLevel {
	return relationPrecedence
}
func (n nodeTypeLike) isNode() {}

type nodeTypeIs struct {
	Left       node
	EntityType types.Path
}

func (n nodeTypeIs) precedenceLevel() nodePrecedenceLevel {
	return relationPrecedence
}
func (n nodeTypeIs) isNode() {}

type nodeTypeIsIn struct {
	nodeTypeIs
	Entity node
}

func (n nodeTypeIsIn) precedenceLevel() nodePrecedenceLevel {
	return relationPrecedence
}

type addNode struct{}

func (n addNode) precedenceLevel() nodePrecedenceLevel {
	return addPrecedence
}

type nodeTypeSub struct {
	binaryNode
	addNode
}

type nodeTypeAdd struct {
	binaryNode
	addNode
}

type nodeTypeMult struct{ binaryNode }

func (n nodeTypeMult) precedenceLevel() nodePrecedenceLevel {
	return multPrecedence
}

type unaryNode struct {
	Arg node
}

func (n unaryNode) precedenceLevel() nodePrecedenceLevel {
	return unaryPrecedence
}

func (n unaryNode) isNode() {}

type nodeTypeNegate struct{ unaryNode }
type nodeTypeNot struct{ unaryNode }

type nodeTypeAccess struct{ strOpNode }

func (n nodeTypeAccess) precedenceLevel() nodePrecedenceLevel {
	return accessPrecedence
}

type nodeTypeExtensionCall struct {
	Name types.String // TODO: review type
	Args []node
}

func (n nodeTypeExtensionCall) precedenceLevel() nodePrecedenceLevel {
	return accessPrecedence
}
func (n nodeTypeExtensionCall) isNode() {}

func stripNodes(args []Node) []node {
	res := make([]node, len(args))
	for i, v := range args {
		res[i] = v.v
	}
	return res
}

func newExtensionCall(method types.String, args ...Node) Node {
	return newNode(nodeTypeExtensionCall{
		Name: method,
		Args: stripNodes(args),
	})
}

func newMethodCall(lhs Node, method types.String, args ...Node) Node {
	res := make([]node, 1+len(args))
	res[0] = lhs.v
	for i, v := range args {
		res[i+1] = v.v
	}
	return newNode(nodeTypeExtensionCall{
		Name: method,
		Args: res,
	})
}

type containsNode struct{}

func (n containsNode) precedenceLevel() nodePrecedenceLevel {
	return accessPrecedence
}

type nodeTypeContains struct {
	binaryNode
	containsNode
}
type nodeTypeContainsAll struct {
	binaryNode
	containsNode
}
type nodeTypeContainsAny struct {
	binaryNode
	containsNode
}

type primaryNode struct{}

func (n primaryNode) isNode() {}

func (n primaryNode) precedenceLevel() nodePrecedenceLevel {
	return primaryPrecedence
}

type nodeValue struct {
	primaryNode
	Value types.Value
}

func (n nodeValue) isNode() {}

type recordElement struct {
	Key   types.String
	Value node
}

type nodeTypeRecord struct {
	primaryNode
	Elements []recordElement
}

func (n nodeTypeRecord) isNode() {}

type nodeTypeSet struct {
	primaryNode
	Elements []node
}

type nodeTypeVariable struct {
	primaryNode
	Name types.String // TODO: Review type
}

type node interface {
	isNode()
	marshalCedar(*bytes.Buffer)
	precedenceLevel() nodePrecedenceLevel
}