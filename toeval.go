package cedar

import (
	"fmt"
	"strings"

	"github.com/cedar-policy/cedar-go/x/exp/parser"
)

func toEval(n any) evaler {
	switch v := n.(type) {
	case parser.Policy:
		res := toEval(v.Principal)
		res = newAndEval(res, toEval(v.Action))
		res = newAndEval(res, toEval(v.Resource))
		for _, c := range v.Conditions {
			res = newAndEval(res, toEval(c))
		}
		return res
	case parser.Principal:
		var res evaler
		switch v.Type {
		case parser.MatchAny:
			res = newLiteralEval(Boolean(true))
		case parser.MatchEquals:
			res = newEqualEval(newVariableEval(variableNamePrincipal), toEval(v.Entity))
		case parser.MatchIn:
			res = newInEval(newVariableEval(variableNamePrincipal), toEval(v.Entity))
		case parser.MatchIs:
			res = newIsEval(newVariableEval(variableNamePrincipal), toEval(v.Path))
		case parser.MatchIsIn:
			lhs := newIsEval(newVariableEval(variableNamePrincipal), toEval(v.Path))
			rhs := newInEval(newVariableEval(variableNamePrincipal), toEval(v.Entity))
			res = newAndEval(lhs, rhs)
		}
		return res
	case parser.Action:
		var res evaler
		switch v.Type {
		case parser.MatchAny:
			res = newLiteralEval(Boolean(true))
		case parser.MatchEquals:
			res = newEqualEval(newVariableEval(variableNameAction), toEval(v.Entities[0]))
		case parser.MatchIn:
			res = newInEval(newVariableEval(variableNameAction), toEval(v.Entities[0]))
		case parser.MatchInList:
			var vals []evaler
			for _, e := range v.Entities {
				vals = append(vals, toEval(e))
			}
			sl := newSetLiteralEval(vals)
			res = newInEval(newVariableEval(variableNameAction), sl)
		}
		return res
	case parser.Resource:
		var res evaler
		switch v.Type {
		case parser.MatchAny:
			res = newLiteralEval(Boolean(true))
		case parser.MatchEquals:
			res = newEqualEval(newVariableEval(variableNameResource), toEval(v.Entity))
		case parser.MatchIn:
			res = newInEval(newVariableEval(variableNameResource), toEval(v.Entity))
		case parser.MatchIs:
			res = newIsEval(newVariableEval(variableNameResource), toEval(v.Path))
		case parser.MatchIsIn:
			lhs := newIsEval(newVariableEval(variableNameResource), toEval(v.Path))
			rhs := newInEval(newVariableEval(variableNameResource), toEval(v.Entity))
			res = newAndEval(lhs, rhs)
		}
		return res
	case parser.Entity:
		return newLiteralEval(entityValueFromSlice(v.Path))
	case parser.Path:
		return newLiteralEval(pathFromSlice(v.Path))
	case parser.Condition:
		var res evaler
		switch v.Type {
		case parser.ConditionWhen:
			res = toEval(v.Expression)
		case parser.ConditionUnless:
			res = newNotEval(toEval(v.Expression))
		}
		return res
	case parser.Expression:
		var res evaler
		switch v.Type {
		case parser.ExpressionOr:
			res = toEval(v.Or)
		case parser.ExpressionIf:
			res = toEval(*v.If)
		}
		return res
	case parser.If:
		return newIfThenElseEval(toEval(v.If), toEval(v.Then), toEval(v.Else))
	case parser.Or:
		res := toEval(v.Ands[len(v.Ands)-1])
		for i := len(v.Ands) - 2; i >= 0; i-- {
			res = newOrNode(toEval(v.Ands[i]), res)
		}
		return res
	case parser.And:
		res := toEval(v.Relations[len(v.Relations)-1])
		for i := len(v.Relations) - 2; i >= 0; i-- {
			res = newAndEval(toEval(v.Relations[i]), res)
		}
		return res
	case parser.Relation:
		lhs := toEval(v.Add)
		switch v.Type {
		case parser.RelationNone:
			return lhs
		case parser.RelationRelOp:
			rhs := toEval(v.RelOpRhs)
			switch v.RelOp {
			case parser.RelOpLt:
				return newLongLessThanEval(lhs, rhs)
			case parser.RelOpLe:
				return newLongLessThanOrEqualEval(lhs, rhs)
			case parser.RelOpGe:
				return newLongGreaterThanOrEqualEval(lhs, rhs)
			case parser.RelOpGt:
				return newLongGreaterThanEval(lhs, rhs)
			case parser.RelOpNe:
				return newNotEqualEval(lhs, rhs)
			case parser.RelOpEq:
				return newEqualEval(lhs, rhs)
			case parser.RelOpIn:
				return newInEval(lhs, rhs)
			default:
				panic("missing RelOp case")
			}
		case parser.RelationHasIdent, parser.RelationHasLiteral:
			return newHasEval(lhs, v.Str)
		case parser.RelationLike:
			return newLikeEval(lhs, v.Pat)
		case parser.RelationIs:
			return newIsEval(lhs, toEval(v.Path))
		case parser.RelationIsIn:
			lhs2 := newIsEval(lhs, toEval(v.Path))
			rhs2 := newInEval(lhs, toEval(v.Entity))
			return newAndEval(lhs2, rhs2)
		default:
			panic("missing RelationType case")
		}
	case parser.Add:
		res := toEval(v.Mults[len(v.Mults)-1])
		for i := len(v.AddOps) - 1; i >= 0; i-- {
			switch v.AddOps[i] {
			case parser.AddOpAdd:
				res = newAddEval(toEval(v.Mults[i]), res)
			case parser.AddOpSub:
				res = newSubtractEval(toEval(v.Mults[i]), res)
			default:
				panic("unknown AddOp")
			}
		}
		return res
	case parser.Mult:
		res := toEval(v.Unaries[len(v.Unaries)-1])
		for i := len(v.Unaries) - 2; i >= 0; i-- {
			res = newMultiplyEval(toEval(v.Unaries[i]), res)
		}
		return res

	case parser.Unary:
		res := toEval(v.Member)
		for i := len(v.Ops) - 1; i >= 0; i-- {
			switch v.Ops[i] {
			case parser.UnaryOpMinus:
				res = newNegateEval(res)
			case parser.UnaryOpNot:
				res = newNotEval(res)
			}
		}
		return res
	case parser.Member:
		res := toEval(v.Primary)
		for _, a := range v.Accesses {
			res = toAccess(a, res)
		}
		return res
	case parser.Primary:
		switch v.Type {
		case parser.PrimaryLiteral:
			return toEval(v.Literal)
		case parser.PrimaryVar:
			return toEval(v.Var)
		case parser.PrimaryEntity:
			return toEval(v.Entity)
		case parser.PrimaryExtFun:
			return toEval(v.ExtFun)
		case parser.PrimaryExpr:
			return toEval(v.Expression)
		case parser.PrimaryExprList:
			var nodes []evaler
			for _, e := range v.Expressions {
				nodes = append(nodes, toEval(e))
			}
			return newSetLiteralEval(nodes)
		case parser.PrimaryRecInits:
			nodes := map[string]evaler{}
			for _, r := range v.RecInits {
				nodes[r.Key] = toEval(r.Value)
			}
			return newRecordLiteralEval(nodes)
		default:
			panic("missing PrimaryType case")
		}
	case parser.Literal:
		switch v.Type {
		case parser.LiteralBool:
			return newLiteralEval(Boolean(v.Bool))
		case parser.LiteralInt:
			return newLiteralEval(Long(v.Long))
		case parser.LiteralString:
			return newLiteralEval(String(v.Str))
		default:
			panic("missing LiteralType case")
		}
	case parser.Var:
		switch v.Type {
		case parser.VarPrincipal:
			return newVariableEval(variableNamePrincipal)
		case parser.VarAction:
			return newVariableEval(variableNameAction)
		case parser.VarResource:
			return newVariableEval(variableNameResource)
		case parser.VarContext:
			return newVariableEval(variableNameContext)
		default:
			panic("missing VarType case")
		}
	case parser.ExtFun:
		funName := strings.Join(v.Path, "::")
		switch funName {
		case "decimal":
			if len(v.Expressions) != 1 {
				return newErrorEval(fmt.Errorf("%w: %s takes 1 parameter", errArity, funName))
			}
			return newDecimalLiteralEval(toEval(v.Expressions[0]))
		case "ip":
			if len(v.Expressions) != 1 {
				return newErrorEval(fmt.Errorf("%w: %s takes 1 parameter", errArity, funName))
			}
			return newIPLiteralEval(toEval(v.Expressions[0]))
		default:
			return newErrorEval(fmt.Errorf("%w: %s", errUnknownExtensionFunction, funName))
		}

	default:
		panic(fmt.Sprintf("unknown node type %T", v))
	}
}

func toAccess(v parser.Access, lhs evaler) evaler {
	switch v.Type {
	case parser.AccessField:
		return newAttributeAccessEval(lhs, v.Name)
	case parser.AccessCall:
		var ctor1 func(evaler, evaler) evaler
		var ctor0 func(evaler) evaler
		switch v.Name {
		case "contains":
			ctor1 = func(lhs, rhs evaler) evaler { return newContainsEval(lhs, rhs) }
		case "containsAll":
			ctor1 = func(lhs, rhs evaler) evaler { return newContainsAllEval(lhs, rhs) }
		case "containsAny":
			ctor1 = func(lhs, rhs evaler) evaler { return newContainsAnyEval(lhs, rhs) }
		case "lessThan":
			ctor1 = func(lhs, rhs evaler) evaler { return newDecimalLessThanEval(lhs, rhs) }
		case "lessThanOrEqual":
			ctor1 = func(lhs, rhs evaler) evaler { return newDecimalLessThanOrEqualEval(lhs, rhs) }
		case "greaterThan":
			ctor1 = func(lhs, rhs evaler) evaler { return newDecimalGreaterThanEval(lhs, rhs) }
		case "greaterThanOrEqual":
			ctor1 = func(lhs, rhs evaler) evaler { return newDecimalGreaterThanOrEqualEval(lhs, rhs) }
		case "isIpv4":
			ctor0 = func(lhs evaler) evaler { return newIPTestEval(lhs, ipTestIPv4) }
		case "isIpv6":
			ctor0 = func(lhs evaler) evaler { return newIPTestEval(lhs, ipTestIPv6) }
		case "isLoopback":
			ctor0 = func(lhs evaler) evaler { return newIPTestEval(lhs, ipTestLoopback) }
		case "isMulticast":
			ctor0 = func(lhs evaler) evaler { return newIPTestEval(lhs, ipTestMulticast) }
		case "isInRange":
			ctor1 = func(lhs, rhs evaler) evaler { return newIPIsInRangeEval(lhs, rhs) }
		default:
			return newErrorEval(fmt.Errorf("%w: %s", errUnknownMethod, v.Name))
		}
		if ctor0 != nil {
			if len(v.Expressions) != 0 {
				return newErrorEval(fmt.Errorf("%w `%s`: expected 1, got %d", errArity, v.Name, len(v.Expressions)+1))
			}
			return ctor0(lhs)
		}
		if len(v.Expressions) != 1 {
			return newErrorEval(fmt.Errorf("%w `%s`: expected 2, got %d", errArity, v.Name, len(v.Expressions)+1))
		}
		return ctor1(lhs, toEval(v.Expressions[0]))
	case parser.AccessIndex:
		return newAttributeAccessEval(lhs, v.Name)
	default:
		panic("missing AccessType case")
	}
}
