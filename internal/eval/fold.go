package eval

import (
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/extensions"
	"github.com/cedar-policy/cedar-go/types"
)

func foldPolicy(p *ast.Policy) *ast.Policy {
	if len(p.Conditions) == 0 {
		return p
	}
	p2 := *p
	p2.Annotations = slices.Clone(p.Annotations)
	if p.Conditions != nil { // preserve nility for test purposes
		p2.Conditions = make([]ast.ConditionType, len(p.Conditions))
		for i, c := range p.Conditions {
			p2.Conditions[i] = ast.ConditionType{Condition: c.Condition, Body: fold(c.Body)}
		}
	}
	return &p2
}

// NOTE: nodes is modified in place, so be sure to send unique copy in
func tryFold(nodes []ast.IsNode,
	mkEval func(values []types.Value) Evaler,
	mkNode func(nodes []ast.IsNode) ast.IsNode,
) ast.IsNode {
	var values []types.Value
	ok := true
	for i, n := range nodes {
		n = fold(n)
		nodes[i] = n
		if !ok {
			continue
		}
		if v, vok := n.(ast.NodeValue); vok {
			values = append(values, v.Value)
			continue
		}
		ok = false
	}
	if ok {
		eval := mkEval(values)
		v, err := eval.Eval(nil)
		if err == nil {
			return ast.NodeValue{Value: v}
		}
	}
	return mkNode(nodes)
}

func tryFoldBinary(v ast.BinaryNode, mkEval func(a, b Evaler) Evaler, wrap func(b ast.BinaryNode) ast.IsNode) ast.IsNode {
	return tryFold([]ast.IsNode{v.Left, v.Right},
		func(values []types.Value) Evaler {
			return mkEval(newLiteralEval(values[0]), newLiteralEval(values[1]))
		},
		func(nodes []ast.IsNode) ast.IsNode {
			return wrap(ast.BinaryNode{Left: nodes[0], Right: nodes[1]})
		},
	)
}
func tryFoldUnary(v ast.UnaryNode, mkEval func(a Evaler) Evaler, wrap func(b ast.UnaryNode) ast.IsNode) ast.IsNode {
	return tryFold([]ast.IsNode{v.Arg},
		func(values []types.Value) Evaler { return mkEval(newLiteralEval(values[0])) },
		func(nodes []ast.IsNode) ast.IsNode { return wrap(ast.UnaryNode{Arg: nodes[0]}) },
	)
}

// fold takes in an ast.Node and finds does as much constant folding as is possible given no PARC data.
func fold(n ast.IsNode) ast.IsNode {
	switch v := n.(type) {
	case ast.NodeTypeAccess:
		return tryFold(
			[]ast.IsNode{v.Arg},
			func(values []types.Value) Evaler {
				if _, ok := values[0].(types.EntityUID); ok {
					return newErrorEval(fmt.Errorf("fold.Access.EntityUID"))
				}
				return newAttributeAccessEval(newLiteralEval(values[0]), v.Value)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{Arg: nodes[0], Value: v.Value}}
			},
		)
	case ast.NodeTypeHas:
		return tryFold(
			[]ast.IsNode{v.Arg},
			func(values []types.Value) Evaler {
				if _, ok := values[0].(types.EntityUID); ok {
					return newErrorEval(fmt.Errorf("fold.Has.EntityUID"))
				}
				return newHasEval(newLiteralEval(values[0]), v.Value)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeHas{StrOpNode: ast.StrOpNode{Arg: nodes[0], Value: v.Value}}
			},
		)
	case ast.NodeTypeLike:
		return tryFold(
			[]ast.IsNode{v.Arg},
			func(values []types.Value) Evaler {
				return newLikeEval(newLiteralEval(values[0]), v.Value)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeLike{Arg: nodes[0], Value: v.Value}
			},
		)
	case ast.NodeTypeIfThenElse:
		return tryFold(
			[]ast.IsNode{v.If, v.Then, v.Else},
			func(values []types.Value) Evaler {
				return newIfThenElseEval(newLiteralEval(values[0]), newLiteralEval(values[1]), newLiteralEval(values[2]))
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeIfThenElse{If: nodes[0], Then: nodes[1], Else: nodes[2]}
			},
		)
	case ast.NodeTypeIs:
		return tryFold(
			[]ast.IsNode{v.Left},
			func(values []types.Value) Evaler {
				return newIsEval(newLiteralEval(values[0]), v.EntityType)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeIs{Left: nodes[0], EntityType: v.EntityType}
			},
		)
	case ast.NodeTypeIsIn:
		return tryFold(
			[]ast.IsNode{v.Left, v.Entity},
			func(values []types.Value) Evaler {
				return newErrorEval(fmt.Errorf("fold.IsIn.EntityUID"))
				// return newIsInEval(newLiteralEval(values[0]), v.EntityType, newLiteralEval(values[1]))
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeIsIn{NodeTypeIs: ast.NodeTypeIs{Left: nodes[0], EntityType: v.EntityType}, Entity: nodes[1]}
			},
		)

	case ast.NodeTypeExtensionCall:
		nodes := make([]ast.IsNode, len(v.Args))
		copy(nodes, v.Args)
		return tryFold(nodes,
			func(values []types.Value) Evaler {
				if i, ok := extensions.ExtMap[v.Name]; ok {
					if i.Args != len(values) {
						return newErrorEval(fmt.Errorf("%w: %s takes %d parameter(s)", errArity, v.Name, i.Args))
					}
					switch {
					case v.Name == "ip":
						return newIPLiteralEval(newLiteralEval(values[0]))
					case v.Name == "decimal":
						return newDecimalLiteralEval(newLiteralEval(values[0]))

					case v.Name == "lessThan":
						return newDecimalLessThanEval(newLiteralEval(values[0]), newLiteralEval(values[1]))
					case v.Name == "lessThanOrEqual":
						return newDecimalLessThanOrEqualEval(newLiteralEval(values[0]), newLiteralEval(values[1]))
					case v.Name == "greaterThan":
						return newDecimalGreaterThanEval(newLiteralEval(values[0]), newLiteralEval(values[1]))
					case v.Name == "greaterThanOrEqual":
						return newDecimalGreaterThanOrEqualEval(newLiteralEval(values[0]), newLiteralEval(values[1]))

					case v.Name == "isIpv4":
						return newIPTestEval(newLiteralEval(values[0]), ipTestIPv4)
					case v.Name == "isIpv6":
						return newIPTestEval(newLiteralEval(values[0]), ipTestIPv6)
					case v.Name == "isLoopback":
						return newIPTestEval(newLiteralEval(values[0]), ipTestLoopback)
					case v.Name == "isMulticast":
						return newIPTestEval(newLiteralEval(values[0]), ipTestMulticast)
					case v.Name == "isInRange":
						return newIPIsInRangeEval(newLiteralEval(values[0]), newLiteralEval(values[1]))
					}
				}
				return newErrorEval(fmt.Errorf("%w: %s", errUnknownExtensionFunction, v.Name))
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeExtensionCall{Name: v.Name, Args: nodes}
			},
		)
	case ast.NodeValue:
		return n
	case ast.NodeTypeRecord:
		elements := make([]ast.IsNode, len(v.Elements))
		for i, pair := range v.Elements {
			elements[i] = pair.Value
		}
		return tryFold(elements,
			func(values []types.Value) Evaler {
				m := make(map[types.String]Evaler, len(values))
				for i, val := range values {
					m[types.String(v.Elements[i].Key)] = newLiteralEval(val)
				}
				return newRecordLiteralEval(m)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				el := make([]ast.RecordElementNode, len(nodes))
				for i, val := range nodes {
					el[i] = ast.RecordElementNode{Key: v.Elements[i].Key, Value: val}
				}
				return ast.NodeTypeRecord{Elements: el}
			},
		)
	case ast.NodeTypeSet:
		elements := make([]ast.IsNode, len(v.Elements))
		copy(elements, v.Elements)
		return tryFold(elements,
			func(values []types.Value) Evaler {
				el := make([]Evaler, len(values))
				for i, v := range values {
					el[i] = newLiteralEval(v)
				}
				return newSetLiteralEval(el)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeSet{Elements: nodes}
			},
		)
	case ast.NodeTypeNegate:
		return tryFoldUnary(v.UnaryNode, newNegateEval, func(b ast.UnaryNode) ast.IsNode { return ast.NodeTypeNegate{UnaryNode: b} })
	case ast.NodeTypeNot:
		return tryFoldUnary(v.UnaryNode, newNotEval, func(b ast.UnaryNode) ast.IsNode { return ast.NodeTypeNot{UnaryNode: b} })
	case ast.NodeTypeVariable:
		return n
	case ast.NodeTypeIn:
		// return tryFoldBinary(v.BinaryNode, newInEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeIn{BinaryNode: b} })
		return tryFold(
			[]ast.IsNode{v.Left, v.Right},
			func(values []types.Value) Evaler {
				return newErrorEval(fmt.Errorf("fold.In.EntityUID"))
				// return newInEval(newLiteralEval(values[0]), newLiteralEval(values[1]))
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeIn{BinaryNode: ast.BinaryNode{Left: nodes[0], Right: nodes[1]}}
			},
		)
	case ast.NodeTypeAnd:
		return tryFoldBinary(v.BinaryNode, newAndEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeAnd{BinaryNode: b} })
	case ast.NodeTypeOr:
		return tryFoldBinary(v.BinaryNode, newOrEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeOr{BinaryNode: b} })
	case ast.NodeTypeEquals:
		return tryFoldBinary(v.BinaryNode, newEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeEquals{BinaryNode: b} })
	case ast.NodeTypeNotEquals:
		return tryFoldBinary(v.BinaryNode, newNotEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeNotEquals{BinaryNode: b} })
	case ast.NodeTypeGreaterThan:
		return tryFoldBinary(v.BinaryNode, newLongGreaterThanEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeGreaterThan{BinaryNode: b} })
	case ast.NodeTypeGreaterThanOrEqual:
		return tryFoldBinary(v.BinaryNode, newLongGreaterThanOrEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeGreaterThanOrEqual{BinaryNode: b} })
	case ast.NodeTypeLessThan:
		return tryFoldBinary(v.BinaryNode, newLongLessThanEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeLessThan{BinaryNode: b} })
	case ast.NodeTypeLessThanOrEqual:
		return tryFoldBinary(v.BinaryNode, newLongLessThanOrEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeLessThanOrEqual{BinaryNode: b} })
	case ast.NodeTypeSub:
		return tryFoldBinary(v.BinaryNode, newSubtractEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeSub{BinaryNode: b} })
	case ast.NodeTypeAdd:
		return tryFoldBinary(v.BinaryNode, newAddEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeAdd{BinaryNode: b} })
	case ast.NodeTypeMult:
		return tryFoldBinary(v.BinaryNode, newMultiplyEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeMult{BinaryNode: b} })
	case ast.NodeTypeContains:
		return tryFoldBinary(v.BinaryNode, newContainsEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeContains{BinaryNode: b} })
	case ast.NodeTypeContainsAll:
		return tryFoldBinary(v.BinaryNode, newContainsAllEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeContainsAll{BinaryNode: b} })
	case ast.NodeTypeContainsAny:
		return tryFoldBinary(v.BinaryNode, newContainsAnyEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeContainsAny{BinaryNode: b} })
	default:
		panic(fmt.Sprintf("unknown node type %T", v))
	}
}
