package ast

import "github.com/cedar-policy/cedar-go/types"

type nodeType uint8

const (
	nodeTypeNone nodeType = iota
	nodeTypeAccess
	nodeTypeAdd
	nodeTypeAll
	nodeTypeAnd
	nodeTypeAnnotation
	nodeTypeBoolean
	nodeTypeContains
	nodeTypeContainsAll
	nodeTypeContainsAny
	nodeTypeEntity
	nodeTypeEntityType
	nodeTypeEquals
	nodeTypeGreater
	nodeTypeGreaterEqual
	nodeTypeLike
	nodeTypeHas
	nodeTypeIf
	nodeTypeIn
	nodeTypeIpAddr
	nodeTypeDecimal
	nodeTypeIs
	nodeTypeIsInRange
	nodeTypeIsIpv4
	nodeTypeIsIpv6
	nodeTypeIsLoopback
	nodeTypeIsMulticast
	nodeTypeLess
	nodeTypeLessEqual
	nodeTypeLong
	nodeTypeMult
	nodeTypeNot
	nodeTypeNegate
	nodeTypeNotEquals
	nodeTypeOr
	nodeTypeRecord
	nodeTypeRecordEntry
	nodeTypeSet
	nodeTypeSub
	nodeTypeString
	nodeTypeVariable
	nodeTypeLessExt
	nodeTypeLessEqualExt
	nodeTypeGreaterExt
	nodeTypeGreaterEqualExt
	nodeTypeWhen
	nodeTypeUnless
	nodeTypeIsIn
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
