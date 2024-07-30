package ast

import "github.com/cedar-policy/cedar-go/types"

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
	nodeTypeRecordEntry
	nodeTypeSet
	nodeTypeSub
	nodeTypeString
	nodeTypeVariable
)

type Node struct {
	nodeType nodeType
	args     []Node      // For inner nodes like operators, records, etc
	value    types.Value // For leaf nodes like String, Long, EntityUID
}
