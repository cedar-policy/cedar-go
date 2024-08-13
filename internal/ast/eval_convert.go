package ast

import (
	"fmt"
)

func toEval(n IsNode) Evaler {
	switch v := n.(type) {
	case NodeTypeAccess:
		return newAttributeAccessEval(toEval(v.Arg), string(v.Value))
	case NodeTypeHas:
		return newHasEval(toEval(v.Arg), string(v.Value))
	case NodeTypeLike:
		return newLikeEval(toEval(v.Arg), v.Value)
	case NodeTypeIf:
		return newIfThenElseEval(toEval(v.If), toEval(v.Then), toEval(v.Else))
	case NodeTypeIs:
		return newIsEval(toEval(v.Left), newLiteralEval(v.EntityType))
	case NodeTypeIsIn:
		obj := toEval(v.Left)
		lhs := newIsEval(obj, newLiteralEval(v.EntityType))
		rhs := newInEval(obj, toEval(v.Entity))
		return newAndEval(lhs, rhs)
	case NodeTypeExtensionCall:
		i, ok := extMap[v.Name]
		if !ok {
			return newErrorEval(fmt.Errorf("%w: %s", errUnknownExtensionFunction, v.Name))
		}
		if i.Args != len(v.Args) {
			return newErrorEval(fmt.Errorf("%w: %s takes %d parameter(s)", errArity, v.Name, i.Args))
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
	case NodeValue:
		return newLiteralEval(v.Value)
	case NodeTypeRecord:
		m := make(map[string]Evaler, len(v.Elements))
		for _, e := range v.Elements {
			m[string(e.Key)] = toEval(e.Value)
		}
		return newRecordLiteralEval(m)
	case NodeTypeSet:
		s := make([]Evaler, len(v.Elements))
		for i, e := range v.Elements {
			s[i] = toEval(e)
		}
		return newSetLiteralEval(s)
	case NodeTypeNegate:
		return newNegateEval(toEval(v.Arg))
	case NodeTypeNot:
		return newNotEval(toEval(v.Arg))
	case NodeTypeVariable:
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
	case NodeTypeIn:
		return newInEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeAnd:
		return newAndEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeEquals:
		return newEqualEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeGreaterThan:
		return newLongGreaterThanEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeGreaterThanOrEqual:
		return newLongGreaterThanOrEqualEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeLessThan:
		return newLongLessThanEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeLessThanOrEqual:
		return newLongLessThanOrEqualEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeSub:
		return newSubtractEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeAdd:
		return newAddEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeContains:
		return newContainsEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeContainsAll:
		return newContainsAllEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeContainsAny:
		return newContainsAnyEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeMult:
		return newMultiplyEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeNotEquals:
		return newNotEqualEval(toEval(v.Left), toEval(v.Right))
	case NodeTypeOr:
		return newOrNode(toEval(v.Left), toEval(v.Right))
	default:
		panic(fmt.Sprintf("unknown node type %T", v))
	}
}
