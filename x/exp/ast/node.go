package ast

type nodeType uint8

const (
	nodeTypeAccess nodeType = iota
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
	nodeTypeMult
	nodeTypeNot
	nodeTypeNotEquals
	nodeTypeOr
	nodeTypeRecord
	nodeTypeSet
	nodeTypeSub
	nodeTypeString
	nodeTypeVariable
)

type Node struct {
	nodeType nodeType
	// TODO: Should we just have `value any`?
	args  []Node
	value any
}
