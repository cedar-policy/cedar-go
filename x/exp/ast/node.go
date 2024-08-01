package ast

import "github.com/cedar-policy/cedar-go/types"

type nodeType uint8

const (
	nodeTypeAccess = iota
	nodeTypeAdd
	nodeTypeAll
	nodeTypeAnd
	nodeTypeAnnotation
	nodeTypeBoolean
	nodeTypeContains
	nodeTypeContainsAll
	nodeTypeContainsAny
	nodeTypeDecimal
	nodeTypeEntity
	nodeTypeEntityType
	nodeTypeEquals
	nodeTypeGreater
	nodeTypeGreaterEqual
	nodeTypeHas
	nodeTypeIf
	nodeTypeIn
	nodeTypeIpAddr
	nodeTypeIs
	nodeTypeIsIn
	nodeTypeLess
	nodeTypeLessEqual
	nodeTypeLike
	nodeTypeLong
	nodeTypeExtMethodCall
	nodeTypeMult
	nodeTypeNegate
	nodeTypeNot
	nodeTypeNotEquals
	nodeTypeOr
	nodeTypeRecord
	nodeTypeRecordEntry
	nodeTypeSet
	nodeTypeString
	nodeTypeSub
	nodeTypeUnless
	nodeTypeVariable
	nodeTypeWhen
)

type Node struct {
	nodeType nodeType
	args     []Node      // For inner nodes like operators, records, etc
	value    types.Value // For leaf nodes like String, Long, EntityUID
}

func newUnaryNode(op nodeType, arg Node) Node {
	return Node{nodeType: op, args: []Node{arg}}
}

type unaryNode Node

func (n unaryNode) Arg() Node { return n.args[0] }

func newBinaryNode(op nodeType, arg1, arg2 Node) Node {
	return Node{nodeType: op, args: []Node{arg1, arg2}}
}

type binaryNode Node

func (n binaryNode) Left() Node  { return n.args[0] }
func (n binaryNode) Right() Node { return n.args[1] }

func newTrinaryNode(op nodeType, arg1, arg2, arg3 Node) Node {
	return Node{nodeType: op, args: []Node{arg1, arg2, arg3}}
}

type trinaryNode Node

func (n trinaryNode) A() Node { return n.args[0] }
func (n trinaryNode) B() Node { return n.args[1] }
func (n trinaryNode) C() Node { return n.args[2] }

func newExtMethodCallNode(object Node, methodName string, args ...Node) Node {
	nodes := []Node{object, String(types.String(methodName))}
	return Node{
		nodeType: nodeTypeExtMethodCall,
		args:     append(nodes, args...),
	}
}

type extMethodCallNode Node

func (n extMethodCallNode) Object() Node {
	return n.args[0]
}

func (n extMethodCallNode) Name() string {
	return string(n.args[1].value.(types.String))
}

func (n extMethodCallNode) Args() []Node {
	return n.args[2:]
}
