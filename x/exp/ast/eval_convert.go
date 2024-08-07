package ast

import (
	"fmt"
)

func toEval(n node) Evaler {
	switch v := n.(type) {
	case nodeTypeAccess:
		return newAttributeAccessEval(toEval(v.Arg), string(v.Value))
	case nodeTypeHas:
		return newHasEval(toEval(v.Arg), string(v.Value))
	case nodeTypeLike:
		return newLikeEval(toEval(v.Arg), v.Value)
	case nodeTypeIf:
		return newIfThenElseEval(toEval(v.If), toEval(v.Then), toEval(v.Else))
	case nodeTypeIs:
		return newIsEval(toEval(v.Left), newLiteralEval(v.EntityType))
	case nodeTypeIsIn:
		obj := toEval(v.Left)
		lhs := newIsEval(obj, newLiteralEval(v.EntityType))
		rhs := newInEval(obj, toEval(v.Entity))
		return newAndEval(lhs, rhs)
	case nodeTypeExtensionCall:
		i, ok := extMap[v.Name]
		if !ok {
			return newErrorEval(fmt.Errorf("%w: %s", errUnknownMethod, v.Name))
		}
		if i.Args != len(v.Args) {
			return newErrorEval(fmt.Errorf("%w: %s takes 1 parameter", errArity, v.Name))
		}
		switch {
		case v.Name == "ip":
			return newIPLiteralEval(toEval(v.Args[0]))
		case v.Name == "decimal":
			return newDecimalLiteralEval(toEval(v.Args[0]))

		case v.Name == "lessThan":
			return newDecimalLessThanEval(toEval(v.Args[0]), toEval(v.Args[1]))
		case v.Name == "lessThanOrEqual":
			return newDecimalLessThanOrEqualEval(toEval(v.Args[0]), toEval(v.Args[1]))
		case v.Name == "greaterThan":
			return newDecimalGreaterThanEval(toEval(v.Args[0]), toEval(v.Args[1]))
		case v.Name == "greaterThanOrEqual":
			return newDecimalGreaterThanOrEqualEval(toEval(v.Args[0]), toEval(v.Args[1]))

		case v.Name == "isIpv4":
			return newIPTestEval(toEval(v.Args[0]), ipTestIPv4)
		case v.Name == "isIpv6":
			return newIPTestEval(toEval(v.Args[0]), ipTestIPv6)
		case v.Name == "isLoopback":
			return newIPTestEval(toEval(v.Args[0]), ipTestLoopback)
		case v.Name == "isMulticast":
			return newIPTestEval(toEval(v.Args[0]), ipTestMulticast)
		case v.Name == "isInRange":
			return newIPIsInRangeEval(toEval(v.Args[0]), toEval(v.Args[1]))
		default:
			panic(fmt.Errorf("unknown extension: %v", v.Name))
		}
	case nodeValue:
		return newLiteralEval(v.Value)
	case nodeTypeRecord:
		m := make(map[string]Evaler, len(v.Elements))
		for _, e := range v.Elements {
			m[string(e.Key)] = toEval(e.Value)
		}
		return newRecordLiteralEval(m)
	case nodeTypeSet:
		s := make([]Evaler, len(v.Elements))
		for i, e := range v.Elements {
			s[i] = toEval(e)
		}
		return newSetLiteralEval(s)
	case nodeTypeNegate:
		return newNegateEval(toEval(v.Arg))
	case nodeTypeNot:
		return newNotEval(toEval(v.Arg))
	case nodeTypeVariable:
		switch v.Name {
		case "principal":
			return newVariableEval(variableNamePrincipal)
		case "action":
			return newVariableEval(variableNameAction)
		case "resource":
			return newVariableEval(variableNameResource)
		case "context":
			return newVariableEval(variableNameContext)
		default:
			panic(fmt.Errorf("unknown variable: %v", v.Name))
		}
	case nodeTypeIn:
		return newInEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeAnd:
		return newAndEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeEquals:
		return newEqualEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeGreaterThan:
		return newLongGreaterThanEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeGreaterThanOrEqual:
		return newLongGreaterThanOrEqualEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeLessThan:
		return newLongLessThanEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeLessThanOrEqual:
		return newLongLessThanOrEqualEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeSub:
		return newSubtractEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeAdd:
		return newAddEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeContains:
		return newContainsEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeContainsAll:
		return newContainsAllEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeContainsAny:
		return newContainsAnyEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeMult:
		return newMultiplyEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeNotEquals:
		return newNotEqualEval(toEval(v.Left), toEval(v.Right))
	case nodeTypeOr:
		return newOrNode(toEval(v.Left), toEval(v.Right))
	default:
		panic(fmt.Sprintf("unknown node type %T", v))
	}
}
