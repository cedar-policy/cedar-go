package eval

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func ToEval(n ast.IsNode) Evaler {
	switch v := n.(type) {
	case ast.NodeTypeAccess:
		return newAttributeAccessEval(ToEval(v.Arg), v.Value)
	case ast.NodeTypeHas:
		return newHasEval(ToEval(v.Arg), v.Value)
	case ast.NodeTypeGetTag:
		return newGetTagEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeHasTag:
		return newHasTagEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeLike:
		return newLikeEval(ToEval(v.Arg), v.Value)
	case ast.NodeTypeIfThenElse:
		return newIfThenElseEval(ToEval(v.If), ToEval(v.Then), ToEval(v.Else))
	case ast.NodeTypeIs:
		return newIsEval(ToEval(v.Left), v.EntityType)
	case ast.NodeTypeIsIn:
		return newIsInEval(ToEval(v.Left), v.EntityType, ToEval(v.Entity))
	case ast.NodeTypeExtensionCall:
		args := make([]Evaler, len(v.Args))
		for i, a := range v.Args {
			args[i] = ToEval(a)
		}
		return newExtensionEval(v.Name, args)
	case ast.NodeValue:
		return newLiteralEval(v.Value)
	case ast.NodeTypeRecord:
		m := make(map[types.String]Evaler, len(v.Elements))
		for _, e := range v.Elements {
			m[e.Key] = ToEval(e.Value)
		}
		return newRecordLiteralEval(m)
	case ast.NodeTypeSet:
		s := make([]Evaler, len(v.Elements))
		for i, e := range v.Elements {
			s[i] = ToEval(e)
		}
		return newSetLiteralEval(s)
	case ast.NodeTypeNegate:
		return newNegateEval(ToEval(v.Arg))
	case ast.NodeTypeNot:
		return newNotEval(ToEval(v.Arg))
	case ast.NodeTypeVariable:
		switch v.Name {
		case consts.Principal, consts.Action, consts.Resource, consts.Context:
			return newVariableEval(v.Name)
		default:
			panic(fmt.Errorf("unknown variable: %v", v.Name))
		}
	case ast.NodeTypeIn:
		return newInEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeAnd:
		return newAndEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeOr:
		return newOrEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeEquals:
		return newEqualEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeNotEquals:
		return newNotEqualEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeGreaterThan:
		return newComparableValueGreaterThanEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeGreaterThanOrEqual:
		return newComparableValueGreaterThanOrEqualEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeLessThan:
		return newComparableValueLessThanEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeLessThanOrEqual:
		return newComparableValueLessThanOrEqualEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeSub:
		return newSubtractEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeAdd:
		return newAddEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeMult:
		return newMultiplyEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeContains:
		return newContainsEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeContainsAll:
		return newContainsAllEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeContainsAny:
		return newContainsAnyEval(ToEval(v.Left), ToEval(v.Right))
	case ast.NodeTypeIsEmpty:
		return newIsEmptyEval(ToEval(v.Arg))
	default:
		panic(fmt.Sprintf("unknown node type %T", v))
	}
}
