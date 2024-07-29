package ast

type opType uint8

const (
	nodeTypeAccess opType = iota
	nodeTypeAdd
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
	nodeTypeHas
	nodeTypeIf
	nodeTypeIn
	nodeTypeIpAddr
	nodeTypeIs
	nodeTypeIsInRange
	nodeTypeIsIpv4
	nodeTypeIsIpv6
	nodeTypeIsLoopback
	nodeTypeIsMulticast
	nodeTypeLess
	nodeTypeLessEqual
	nodeTypeLong
	nodeTypeMap
	nodeTypeMult
	nodeTypeNot
	nodeTypeNotEquals
	nodeTypeOr
	nodeTypeSet
	nodeTypeSub
	nodeTypeString
	nodeTypeVariable
)

type Node struct {
	op opType
	// TODO: Should we just have `value any`?
	args  []Node
	value any
}
