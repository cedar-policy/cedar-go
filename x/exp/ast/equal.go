package ast

import (
	"fmt"
	"slices"
)

// Equal reports whether a and b are equivalent expressions.
func Equal(a, b Node) bool {
	return equalNodes(a.v, b.v)
}

func equalNodes(a, b IsNode) bool {
	switch a := a.(type) {
	case NodeTypeIfThenElse:
		b, ok := b.(NodeTypeIfThenElse)
		return ok && equalNodes(a.If, b.If) && equalNodes(a.Then, b.Then) && equalNodes(a.Else, b.Else)
	case NodeTypeOr:
		b, ok := b.(NodeTypeOr)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeAnd:
		b, ok := b.(NodeTypeAnd)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeLessThan:
		b, ok := b.(NodeTypeLessThan)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeLessThanOrEqual:
		b, ok := b.(NodeTypeLessThanOrEqual)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeGreaterThan:
		b, ok := b.(NodeTypeGreaterThan)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeGreaterThanOrEqual:
		b, ok := b.(NodeTypeGreaterThanOrEqual)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeNotEquals:
		b, ok := b.(NodeTypeNotEquals)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeEquals:
		b, ok := b.(NodeTypeEquals)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeIn:
		b, ok := b.(NodeTypeIn)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeHas:
		b, ok := b.(NodeTypeHas)
		return ok && equalStrOpNodes(a.StrOpNode, b.StrOpNode)
	case NodeTypeHasTag:
		b, ok := b.(NodeTypeHasTag)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeLike:
		b, ok := b.(NodeTypeLike)
		return ok && equalNodes(a.Arg, b.Arg) && a.Value.Equal(b.Value)
	case NodeTypeIs:
		b, ok := b.(NodeTypeIs)
		return ok && equalIsNodes(a, b)
	case NodeTypeIsIn:
		b, ok := b.(NodeTypeIsIn)
		return ok && equalIsNodes(a.NodeTypeIs, b.NodeTypeIs) && equalNodes(a.Entity, b.Entity)
	case NodeTypeSub:
		b, ok := b.(NodeTypeSub)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeAdd:
		b, ok := b.(NodeTypeAdd)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeMult:
		b, ok := b.(NodeTypeMult)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeNegate:
		b, ok := b.(NodeTypeNegate)
		return ok && equalNodes(a.Arg, b.Arg)
	case NodeTypeNot:
		b, ok := b.(NodeTypeNot)
		return ok && equalNodes(a.Arg, b.Arg)
	case NodeTypeAccess:
		b, ok := b.(NodeTypeAccess)
		return ok && equalStrOpNodes(a.StrOpNode, b.StrOpNode)
	case NodeTypeGetTag:
		b, ok := b.(NodeTypeGetTag)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeExtensionCall:
		b, ok := b.(NodeTypeExtensionCall)
		return ok && a.Name == b.Name && slices.EqualFunc(a.Args, b.Args, equalNodes)
	case NodeTypeContains:
		b, ok := b.(NodeTypeContains)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeContainsAll:
		b, ok := b.(NodeTypeContainsAll)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeContainsAny:
		b, ok := b.(NodeTypeContainsAny)
		return ok && equalBinaryNodes(a.BinaryNode, b.BinaryNode)
	case NodeTypeIsEmpty:
		b, ok := b.(NodeTypeIsEmpty)
		return ok && equalNodes(a.Arg, b.Arg)
	case NodeValue:
		b, ok := b.(NodeValue)
		return ok && a.Value.Equal(b.Value)
	case NodeTypeRecord:
		b, ok := b.(NodeTypeRecord)
		return ok && equalRecordNodes(a, b)
	case NodeTypeSet:
		b, ok := b.(NodeTypeSet)
		return ok && slices.EqualFunc(a.Elements, b.Elements, equalNodes)
	case NodeTypeVariable:
		b, ok := b.(NodeTypeVariable)
		return ok && a.Name == b.Name
	default:
		panic(fmt.Sprintf("unknown node type %T", a))
	}
}

func equalBinaryNodes(a, b BinaryNode) bool {
	return equalNodes(a.Left, b.Left) && equalNodes(a.Right, b.Right)
}

func equalStrOpNodes(a, b StrOpNode) bool {
	return a.Value == b.Value && equalNodes(a.Arg, b.Arg)
}

func equalIsNodes(a, b NodeTypeIs) bool {
	return a.EntityType == b.EntityType && equalNodes(a.Left, b.Left)
}

func equalRecordNodes(a, b NodeTypeRecord) bool {
	if len(a.Elements) != len(b.Elements) {
		return false
	}
	bValues := make(map[string]IsNode, len(b.Elements))
	for _, e := range b.Elements {
		bValues[string(e.Key)] = e.Value
	}
	for _, e := range a.Elements {
		bValue, ok := bValues[string(e.Key)]
		if !ok || !equalNodes(e.Value, bValue) {
			return false
		}
	}
	return true
}
