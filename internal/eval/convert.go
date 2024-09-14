package eval

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/types"
)

func toEval(n ast.IsNode) Evaler {
	switch v := n.(type) {
	case ast.NodeTypeAccess:
		return newAttributeAccessEval(toEval(v.Arg), v.Value)
	case ast.NodeTypeHas:
		return newHasEval(toEval(v.Arg), v.Value)
	case ast.NodeTypeLike:
		return newLikeEval(toEval(v.Arg), v.Value)
	case ast.NodeTypeIfThenElse:
		return newIfThenElseEval(toEval(v.If), toEval(v.Then), toEval(v.Else))
	case ast.NodeTypeIs:
		return newIsEval(toEval(v.Left), v.EntityType)
	case ast.NodeTypeIsIn:
		return newIsInEval(toEval(v.Left), v.EntityType, toEval(v.Entity))
	case ast.NodeTypeExtensionCall:
		args := make([]Evaler, len(v.Args))
		for i, a := range v.Args {
			args[i] = toEval(a)
		}
		return newExtensionEval(v.Name, args)
	case ast.NodeValue:
		return newLiteralEval(v.Value)
	case ast.NodeTypeRecord:
		m := make(map[types.String]Evaler, len(v.Elements))
		for _, e := range v.Elements {
			m[e.Key] = toEval(e.Value)
		}
		return newRecordLiteralEval(m)
	case ast.NodeTypeSet:
		s := make([]Evaler, len(v.Elements))
		for i, e := range v.Elements {
			s[i] = toEval(e)
		}
		return newSetLiteralEval(s)
	case ast.NodeTypeNegate:
		return newNegateEval(toEval(v.Arg))
	case ast.NodeTypeNot:
		return newNotEval(toEval(v.Arg))
	case ast.NodeTypeVariable:
		switch v.Name {
		case consts.Principal, consts.Action, consts.Resource, consts.Context:
			return newVariableEval(v.Name)
		default:
			panic(fmt.Errorf("unknown variable: %v", v.Name))
		}
	case ast.NodeTypeIn:
		return newInEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeAnd:
		return newAndEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeOr:
		return newOrEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeEquals:
		return newEqualEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeNotEquals:
		return newNotEqualEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeGreaterThan:
		return newComparableValueGreaterThanEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeGreaterThanOrEqual:
		return newComparableValueGreaterThanOrEqualEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeLessThan:
		return newComparableValueLessThanEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeLessThanOrEqual:
		return newComparableValueLessThanOrEqualEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeSub:
		return newSubtractEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeAdd:
		return newAddEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeMult:
		return newMultiplyEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeContains:
		return newContainsEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeContainsAll:
		return newContainsAllEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeContainsAny:
		return newContainsAnyEval(toEval(v.Left), toEval(v.Right))
	default:
		panic(fmt.Sprintf("unknown node type %T", v))
	}
}
