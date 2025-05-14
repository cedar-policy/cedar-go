package parser

import (
	"bytes"

	"github.com/cedar-policy/cedar-go/x/exp/ast"
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

type relationPrecedenceNode struct{}

func (n relationPrecedenceNode) precedenceLevel() nodePrecedenceLevel {
	return relationPrecedence
}

type NodeTypeLessThan struct {
	ast.NodeTypeLessThan
	relationPrecedenceNode
}

type NodeTypeLessThanOrEqual struct {
	ast.NodeTypeLessThanOrEqual
	relationPrecedenceNode
}
type NodeTypeGreaterThan struct {
	ast.NodeTypeGreaterThan
	relationPrecedenceNode
}
type NodeTypeGreaterThanOrEqual struct {
	ast.NodeTypeGreaterThanOrEqual
	relationPrecedenceNode
}
type NodeTypeNotEquals struct {
	ast.NodeTypeNotEquals
	relationPrecedenceNode
}
type NodeTypeEquals struct {
	ast.NodeTypeEquals
	relationPrecedenceNode
}
type NodeTypeIn struct {
	ast.NodeTypeIn
	relationPrecedenceNode
}

type NodeTypeHas struct {
	ast.NodeTypeHas
	relationPrecedenceNode
}

type NodeTypeLike struct {
	ast.NodeTypeLike
	relationPrecedenceNode
}

type NodeTypeIs struct {
	ast.NodeTypeIs
	relationPrecedenceNode
}

type NodeTypeIsIn struct {
	ast.NodeTypeIsIn
	relationPrecedenceNode
}

type addPrecedenceNode struct{}

func (n addPrecedenceNode) precedenceLevel() nodePrecedenceLevel {
	return addPrecedence
}

type NodeTypeSub struct {
	ast.NodeTypeSub
	addPrecedenceNode
}

type NodeTypeAdd struct {
	ast.NodeTypeAdd
	addPrecedenceNode
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

type accessPrecedenceNode struct{}

func (n accessPrecedenceNode) precedenceLevel() nodePrecedenceLevel {
	return accessPrecedence
}

type NodeTypeAccess struct {
	ast.NodeTypeAccess
	accessPrecedenceNode
}

type NodeTypeHasTag struct {
	ast.NodeTypeHasTag
	accessPrecedenceNode
}

type NodeTypeGetTag struct {
	ast.NodeTypeGetTag
	accessPrecedenceNode
}

type NodeTypeExtensionCall struct {
	ast.NodeTypeExtensionCall
	accessPrecedenceNode
}
type NodeTypeContains struct {
	ast.NodeTypeContains
	accessPrecedenceNode
}
type NodeTypeContainsAll struct {
	ast.NodeTypeContainsAll
	accessPrecedenceNode
}
type NodeTypeContainsAny struct {
	ast.NodeTypeContainsAny
	accessPrecedenceNode
}

type NodeTypeIsEmpty struct {
	ast.NodeTypeIsEmpty
	accessPrecedenceNode
}

type primaryPrecedenceNode struct{}

func (n primaryPrecedenceNode) precedenceLevel() nodePrecedenceLevel {
	return primaryPrecedence
}

type NodeValue struct {
	ast.NodeValue
	primaryPrecedenceNode
}

type NodeTypeRecord struct {
	ast.NodeTypeRecord
	primaryPrecedenceNode
}

type NodeTypeSet struct {
	ast.NodeTypeSet
	primaryPrecedenceNode
}

type NodeTypeVariable struct {
	ast.NodeTypeVariable
	primaryPrecedenceNode
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
