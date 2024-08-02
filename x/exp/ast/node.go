package ast

import "github.com/cedar-policy/cedar-go/types"

type strOpNode struct {
	node
	Arg   node
	Value types.String
}

type nodeTypeAccess struct{ strOpNode }
type nodeTypeHas struct{ strOpNode }

type nodeTypeLike struct {
	node
	Arg   node
	Value Pattern
}

type nodeTypeAnnotation struct {
	node
	Key   types.String // TODO: review type
	Value types.String
}

type nodeTypeIf struct {
	node
	If, Then, Else node
}

type nodeTypeIs struct {
	node
	Left       node
	EntityType types.String // TODO: review type
}

type nodeTypeIsIn struct {
	nodeTypeIs
	Entity node
}

type nodeTypeExtMethodCall struct {
	node
	Left   node
	Method types.String // TODO: review type
	Args   []node
}

func stripNodes(args []Node) []node {
	res := make([]node, len(args))
	for i, v := range args {
		res[i] = v.v
	}
	return res
}

func newExtMethodCallNode(left Node, method types.String, args ...Node) Node {
	return newNode(nodeTypeExtMethodCall{
		Left:   left.v,
		Method: method,
		Args:   stripNodes(args),
	})
}

type nodeValue struct {
	node
	Value types.Value
}

type recordElement struct {
	Key   types.String
	Value node
}
type nodeTypeRecord struct {
	node
	Elements []recordElement
}

type nodeTypeSet struct {
	node
	Elements []node
}

type unaryNode struct {
	node
	Arg node
}

type nodeTypeNegate struct{ unaryNode }
type nodeTypeNot struct{ unaryNode }

type condition bool

const (
	conditionWhen   = true
	conditionUnless = false
)

type nodeTypeCondition struct {
	node
	Condition condition
	Body      node
}

type nodeTypeVariable struct {
	node
	Name types.String // TODO: Review type
}

type binaryNode struct {
	node
	Left, Right node
}

type nodeTypeIn struct{ binaryNode }
type nodeTypeAnd struct{ binaryNode }
type nodeTypeEquals struct{ binaryNode }
type nodeTypeGreaterThan struct{ binaryNode }
type nodeTypeGreaterThanOrEqual struct{ binaryNode }
type nodeTypeLessThan struct{ binaryNode }
type nodeTypeLessThanOrEqual struct{ binaryNode }
type nodeTypeSub struct{ binaryNode }
type nodeTypeAdd struct{ binaryNode }
type nodeTypeContains struct{ binaryNode }
type nodeTypeContainsAll struct{ binaryNode }
type nodeTypeContainsAny struct{ binaryNode }
type nodeTypeMult struct{ binaryNode }
type nodeTypeNotEquals struct{ binaryNode }
type nodeTypeOr struct{ binaryNode }

type node interface {
	isNode()
}

type Node struct {
	v node // NOTE: not an embed because a `Node` is not a `node`
}

func newNode(v node) Node {
	return Node{v: v}
}
