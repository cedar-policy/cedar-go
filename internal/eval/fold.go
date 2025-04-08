package eval

import (
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

// foldPolicy takes in a given policy and attempts as much constant folding as possible.
// It is not given any environment (entities, parc) so any references of those sort will
// stop the folding.  The kinds of things that this will fold:
//
//   - 1+1 -> 2
//   - [1,2,3].contains(2) -> true
//   - a nodes set [1,2,3] -> a value node of the set [1,2,3]
//   - an extension call to Decimal("42") -> a value node of the Decimal("42")
//
// Expressions that will cause errors are not folded, for example, `Decimal("hello")` will
// remain as an extension call and not be folded into a Decimal value.
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
	allFolded := true
	for i, n := range nodes {
		n = fold(n)
		nodes[i] = n
		if !allFolded {
			continue
		}
		if v, vok := n.(ast.NodeValue); vok {
			values = append(values, v.Value)
			continue
		}
		allFolded = false
	}
	if allFolded {
		eval := mkEval(values)
		v, err := eval.Eval(Env{Entities: types.EntityMap{}})
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
//
//nolint:revive
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
	case ast.NodeTypeGetTag:
		return tryFold(
			[]ast.IsNode{v.Left, v.Right},
			func(_ []types.Value) Evaler {
				return newErrorEval(fmt.Errorf("fold.GetTag.EntityUID"))
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeGetTag{BinaryNode: ast.BinaryNode{Left: nodes[0], Right: nodes[1]}}
			},
		)
	case ast.NodeTypeHasTag:
		return tryFold(
			[]ast.IsNode{v.Left, v.Right},
			func(_ []types.Value) Evaler {
				return newErrorEval(fmt.Errorf("fold.HasTag.EntityUID"))
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeHasTag{BinaryNode: ast.BinaryNode{Left: nodes[0], Right: nodes[1]}}
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
			func(_ []types.Value) Evaler {
				return newErrorEval(fmt.Errorf("fold.IsIn.EntityUID"))
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
				args := make([]Evaler, len(values))
				for i, a := range values {
					args[i] = newLiteralEval(a)
				}
				return newExtensionEval(v.Name, args)
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
		return tryFold(
			[]ast.IsNode{v.Left, v.Right},
			func(_ []types.Value) Evaler {
				return newErrorEval(fmt.Errorf("fold.In.EntityUID"))
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
		return tryFoldBinary(v.BinaryNode, newComparableValueGreaterThanEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeGreaterThan{BinaryNode: b} })
	case ast.NodeTypeGreaterThanOrEqual:
		return tryFoldBinary(v.BinaryNode, newComparableValueGreaterThanOrEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeGreaterThanOrEqual{BinaryNode: b} })
	case ast.NodeTypeLessThan:
		return tryFoldBinary(v.BinaryNode, newComparableValueLessThanEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeLessThan{BinaryNode: b} })
	case ast.NodeTypeLessThanOrEqual:
		return tryFoldBinary(v.BinaryNode, newComparableValueLessThanOrEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeLessThanOrEqual{BinaryNode: b} })
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
	case ast.NodeTypeIsEmpty:
		return tryFoldUnary(v.UnaryNode, newIsEmptyEval, func(b ast.UnaryNode) ast.IsNode { return ast.NodeTypeIsEmpty{UnaryNode: b} })
	default:
		panic(fmt.Sprintf("unknown node type %T", v))
	}
}
