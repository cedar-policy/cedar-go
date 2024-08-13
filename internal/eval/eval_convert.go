package eval

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/extensions"
	"github.com/cedar-policy/cedar-go/types"
)

func toEval(n ast.IsNode) Evaler {
	switch v := n.(type) {
	case ast.NodeTypeAccess:
		return newAttributeAccessEval(toEval(v.Arg), string(v.Value))
	case ast.NodeTypeHas:
		return newHasEval(toEval(v.Arg), string(v.Value))
	case ast.NodeTypeLike:
		return newLikeEval(toEval(v.Arg), v.Value)
	case ast.NodeTypeIf:
		return newIfThenElseEval(toEval(v.If), toEval(v.Then), toEval(v.Else))
	case ast.NodeTypeIs:
		return newIsEval(toEval(v.Left), newLiteralEval(v.EntityType))
	case ast.NodeTypeIsIn:
		obj := toEval(v.Left)
		lhs := newIsEval(obj, newLiteralEval(v.EntityType))
		rhs := newInEval(obj, toEval(v.Entity))
		return newAndEval(lhs, rhs)
	case ast.NodeTypeExtensionCall:
		i, ok := extensions.ExtMap[v.Name]
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
	case ast.NodeValue:
		return newLiteralEval(v.Value)
	case ast.NodeTypeRecord:
		m := make(map[string]Evaler, len(v.Elements))
		for _, e := range v.Elements {
			m[string(e.Key)] = toEval(e.Value)
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
	case ast.NodeTypeIn:
		return newInEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeAnd:
		return newAndEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeEquals:
		return newEqualEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeGreaterThan:
		return newLongGreaterThanEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeGreaterThanOrEqual:
		return newLongGreaterThanOrEqualEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeLessThan:
		return newLongLessThanEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeLessThanOrEqual:
		return newLongLessThanOrEqualEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeSub:
		return newSubtractEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeAdd:
		return newAddEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeContains:
		return newContainsEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeContainsAll:
		return newContainsAllEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeContainsAny:
		return newContainsAnyEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeMult:
		return newMultiplyEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeNotEquals:
		return newNotEqualEval(toEval(v.Left), toEval(v.Right))
	case ast.NodeTypeOr:
		return newOrNode(toEval(v.Left), toEval(v.Right))
	default:
		panic(fmt.Sprintf("unknown node type %T", v))
	}
}

func scopeToNode(in ast.IsScopeNode) ast.Node {
	switch t := in.(type) {
	case ast.ScopeTypeAll:
		return ast.True()
	case ast.ScopeTypeEq:
		return ast.NewNode(t.Variable).Equals(ast.EntityUID(t.Entity))
	case ast.ScopeTypeIn:
		return ast.NewNode(t.Variable).In(ast.EntityUID(t.Entity))
	case ast.ScopeTypeInSet:
		set := make([]types.Value, len(t.Entities))
		for i, e := range t.Entities {
			set[i] = e
		}
		return ast.NewNode(t.Variable).In(ast.Set(set))
	case ast.ScopeTypeIs:
		return ast.NewNode(t.Variable).Is(t.Type)

	case ast.ScopeTypeIsIn:
		return ast.NewNode(t.Variable).IsIn(t.Type, ast.EntityUID(t.Entity))
	default:
		panic(fmt.Sprintf("unknown scope type %T", t))
	}
}
