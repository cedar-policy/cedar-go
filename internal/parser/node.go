package parser

import (
	"bytes"

	"github.com/cedar-policy/cedar-go/internal/ast"
)

type NodeTypeIf struct{ ast.NodeTypeIfThenElse }

func (n NodeTypeIf) precedenceLevel() nodePrecedenceLevel {
	return ifPrecedence
}

type NodeTypeOr struct{ ast.NodeTypeOr }

func (n NodeTypeOr) precedenceLevel() nodePrecedenceLevel {
	return orPrecedence
}

type NodeTypeAnd struct{ ast.NodeTypeAnd }

func (n NodeTypeAnd) precedenceLevel() nodePrecedenceLevel {
	return andPrecedence
}

type RelationNode struct{}

func (n RelationNode) precedenceLevel() nodePrecedenceLevel {
	return relationPrecedence
}

type NodeTypeLessThan struct {
	ast.NodeTypeLessThan
	RelationNode
}

type NodeTypeLessThanOrEqual struct {
	ast.NodeTypeLessThanOrEqual
	RelationNode
}
type NodeTypeGreaterThan struct {
	ast.NodeTypeGreaterThan
	RelationNode
}
type NodeTypeGreaterThanOrEqual struct {
	ast.NodeTypeGreaterThanOrEqual
	RelationNode
}
type NodeTypeNotEquals struct {
	ast.NodeTypeNotEquals
	RelationNode
}
type NodeTypeEquals struct {
	ast.NodeTypeEquals
	RelationNode
}
type NodeTypeIn struct {
	ast.NodeTypeIn
	RelationNode
}

type NodeTypeHas struct {
	ast.NodeTypeHas
	RelationNode
}

type NodeTypeLike struct {
	ast.NodeTypeLike
	RelationNode
}

type NodeTypeIs struct {
	ast.NodeTypeIs
	RelationNode
}

type NodeTypeIsIn struct {
	ast.NodeTypeIsIn
	RelationNode
}

type AddNode struct{}

func (n AddNode) precedenceLevel() nodePrecedenceLevel {
	return addPrecedence
}

type NodeTypeSub struct {
	ast.NodeTypeSub
	AddNode
}

type NodeTypeAdd struct {
	ast.NodeTypeAdd
	AddNode
}

type NodeTypeMult struct{ ast.NodeTypeMult }

func (n NodeTypeMult) precedenceLevel() nodePrecedenceLevel {
	return multPrecedence
}

type UnaryNode struct{ ast.UnaryNode }

func (n UnaryNode) precedenceLevel() nodePrecedenceLevel {
	return unaryPrecedence
}

type NodeTypeNegate struct {
	ast.NodeTypeNegate
	UnaryNode
}
type NodeTypeNot struct {
	ast.NodeTypeNot
	UnaryNode
}

type NodeTypeAccess struct{ ast.NodeTypeAccess }

func (n NodeTypeAccess) precedenceLevel() nodePrecedenceLevel {
	return accessPrecedence
}

type NodeTypeExtensionCall struct{ ast.NodeTypeExtensionCall }

func (n NodeTypeExtensionCall) precedenceLevel() nodePrecedenceLevel {
	return accessPrecedence
}

type ContainsNode struct{}

func (n ContainsNode) precedenceLevel() nodePrecedenceLevel {
	return accessPrecedence
}

type NodeTypeContains struct {
	ast.NodeTypeContains
	ContainsNode
}
type NodeTypeContainsAll struct {
	ast.NodeTypeContainsAll
	ContainsNode
}
type NodeTypeContainsAny struct {
	ast.NodeTypeContainsAny
	ContainsNode
}

type PrimaryNode struct{}

func (n PrimaryNode) precedenceLevel() nodePrecedenceLevel {
	return primaryPrecedence
}

type NodeValue struct {
	ast.NodeValue
	PrimaryNode
}

type NodeTypeRecord struct {
	ast.NodeTypeRecord
	PrimaryNode
}

type NodeTypeSet struct {
	ast.NodeTypeSet
	PrimaryNode
}

type NodeTypeVariable struct {
	ast.NodeTypeVariable
	PrimaryNode
}

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

type IsNode interface {
	precedenceLevel() nodePrecedenceLevel
	marshalCedar(*bytes.Buffer)
}
