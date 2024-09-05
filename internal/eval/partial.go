package eval

import (
	"errors"
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/types"
)

// partialPolicy itself cannot error, it can only return the error that would happen
func partialPolicy(ctx *Context, p *ast.Policy) (policy *ast.Policy, keep bool, policyErr error) {
	p2 := *p
	if p2.Principal, keep = partialPrincipalScope(ctx, ctx.Principal, p2.Principal); !keep {
		return nil, false, nil
	}
	if p2.Action, keep = partialActionScope(ctx, ctx.Action, p2.Action); !keep {
		return nil, false, nil
	}
	if p2.Resource, keep = partialResourceScope(ctx, ctx.Resource, p2.Resource); !keep {
		return nil, false, nil
	}
	p2.Conditions = nil
	for _, c := range p.Conditions {
		body, err := partial(ctx, c.Body)
		if err != nil {
			return nil, false, err
		}
		if v, ok := body.(ast.NodeValue); ok {
			if b, bok := v.Value.(types.Boolean); bok {
				if bool(b) != bool(c.Condition) {
					return nil, false, nil
				}
				continue
			}
			return nil, false, fmt.Errorf("%w: condition expected bool", ErrType)
		}
		p2.Conditions = append(p2.Conditions, ast.ConditionType{Condition: c.Condition, Body: body})
	}
	p2.Annotations = slices.Clone(p.Annotations)
	return &p2, true, nil
}

func partialPrincipalScope(ctx *Context, ent types.Value, scope ast.IsPrincipalScopeNode) (ast.IsPrincipalScopeNode, bool) {
	evaled, result := partialScopeEval(ctx, ent, scope)
	switch {
	case evaled && !result:
		return nil, false
	case evaled && result:
		return ast.ScopeTypeAll{}, true
	default:
		return scope, true
	}
}

func partialActionScope(ctx *Context, ent types.Value, scope ast.IsActionScopeNode) (ast.IsActionScopeNode, bool) {
	evaled, result := partialScopeEval(ctx, ent, scope)
	switch {
	case evaled && !result:
		return nil, false
	case evaled && result:
		return ast.ScopeTypeAll{}, true
	default:
		return scope, true
	}
}

func partialResourceScope(ctx *Context, ent types.Value, scope ast.IsResourceScopeNode) (ast.IsResourceScopeNode, bool) {
	evaled, result := partialScopeEval(ctx, ent, scope)
	switch {
	case evaled && !result:
		return nil, false
	case evaled && result:
		return ast.ScopeTypeAll{}, true
	default:
		return scope, true
	}
}

func partialScopeEval(ctx *Context, ent types.Value, in ast.IsScopeNode) (evaled bool, result bool) {
	e, ok := ent.(types.EntityUID)
	if !ok {
		return false, false
	}
	if e.Type == variableEntityType {
		return false, false
	}
	switch t := in.(type) {
	case ast.ScopeTypeAll:
		return true, true
	case ast.ScopeTypeEq:
		return true, e == t.Entity
	case ast.ScopeTypeIn:
		return true, entityInOne(ctx, e, t.Entity)
	case ast.ScopeTypeInSet:
		set := make(map[types.EntityUID]struct{}, len(t.Entities))
		for _, e := range t.Entities {
			set[e] = struct{}{}
		}
		return true, entityInSet(ctx, e, set)
	case ast.ScopeTypeIs:
		return true, e.Type == t.Type
	case ast.ScopeTypeIsIn:
		return true, e.Type == t.Type && entityInOne(ctx, e, t.Entity)
	default:
		panic(fmt.Sprintf("unknown scope type %T", t))
	}
}

var errVariable = fmt.Errorf("variable")

// NOTE: nodes is modified in place, so be sure to send unique copy in
func tryPartial(ctx *Context, nodes []ast.IsNode,
	mkEval func(values []types.Value) Evaler,
	mkNode func(nodes []ast.IsNode) ast.IsNode,
) (ast.IsNode, error) {
	var values []types.Value
	ok := true
	for i, n := range nodes {
		n, err := partial(ctx, n)
		if errors.Is(err, errVariable) {
			ok = false
			continue
		}
		if err != nil {
			return nil, err
		}
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
		v, err := eval.Eval(ctx)
		if err != nil {
			return nil, err
		}
		if v == nil { // unknown (old)
			return mkNode(nodes), nil
		}
		if ent, ok := v.(types.EntityUID); ok && ent.Type == variableEntityType { // unknown (new)
			return mkNode(nodes), errVariable
		}
		return ast.NodeValue{Value: v}, nil
	}
	return mkNode(nodes), nil
}

func tryPartialBinary(ctx *Context, v ast.BinaryNode, mkEval func(a, b Evaler) Evaler, wrap func(b ast.BinaryNode) ast.IsNode) (ast.IsNode, error) {
	return tryPartial(ctx, []ast.IsNode{v.Left, v.Right},
		func(values []types.Value) Evaler { return mkEval(newLiteralEval(values[0]), newLiteralEval(values[1])) },
		func(nodes []ast.IsNode) ast.IsNode { return wrap(ast.BinaryNode{Left: nodes[0], Right: nodes[1]}) },
	)
}
func tryPartialUnary(ctx *Context, v ast.UnaryNode, mkEval func(a Evaler) Evaler, wrap func(b ast.UnaryNode) ast.IsNode) (ast.IsNode, error) {
	return tryPartial(ctx, []ast.IsNode{v.Arg},
		func(values []types.Value) Evaler { return mkEval(newLiteralEval(values[0])) },
		func(nodes []ast.IsNode) ast.IsNode { return wrap(ast.UnaryNode{Arg: nodes[0]}) },
	)
}

// partial takes in an ast.Node and finds does as much as is possible given the context
func partial(ctx *Context, n ast.IsNode) (ast.IsNode, error) {
	switch v := n.(type) {
	case ast.NodeTypeAccess:
		return tryPartial(ctx,
			[]ast.IsNode{v.Arg},
			func(values []types.Value) Evaler {
				return newAttributeAccessEval(newLiteralEval(values[0]), v.Value)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{Arg: nodes[0], Value: v.Value}}
			},
		)
	case ast.NodeTypeHas:
		return tryPartial(ctx,
			[]ast.IsNode{v.Arg},
			func(values []types.Value) Evaler {
				return newHasEval(newLiteralEval(values[0]), v.Value)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeHas{StrOpNode: ast.StrOpNode{Arg: nodes[0], Value: v.Value}}
			},
		)
	case ast.NodeTypeLike:
		return tryPartial(ctx,
			[]ast.IsNode{v.Arg},
			func(values []types.Value) Evaler {
				return newLikeEval(newLiteralEval(values[0]), v.Value)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeLike{Arg: nodes[0], Value: v.Value}
			},
		)
	case ast.NodeTypeIfThenElse:
		return partialIfThenElse(ctx, v)
	case ast.NodeTypeIs:
		return tryPartial(ctx,
			[]ast.IsNode{v.Left},
			func(values []types.Value) Evaler {
				return newIsEval(newLiteralEval(values[0]), v.EntityType)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeIs{Left: nodes[0], EntityType: v.EntityType}
			},
		)
	case ast.NodeTypeIsIn:
		return tryPartial(ctx,
			[]ast.IsNode{v.Left, v.Entity},
			func(values []types.Value) Evaler {
				return newIsInEval(newLiteralEval(values[0]), v.EntityType, newLiteralEval(values[1]))
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeIsIn{NodeTypeIs: ast.NodeTypeIs{Left: nodes[0], EntityType: v.EntityType}, Entity: nodes[1]}
			},
		)

	case ast.NodeTypeExtensionCall:
		nodes := make([]ast.IsNode, len(v.Args))
		copy(nodes, v.Args)
		return tryPartial(ctx, nodes,
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
		return n, nil
	case ast.NodeTypeRecord:
		elements := make([]ast.IsNode, len(v.Elements))
		for i, pair := range v.Elements {
			elements[i] = pair.Value
		}
		return tryPartial(ctx, elements,
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
		return tryPartial(ctx, elements,
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
		return tryPartialUnary(ctx, v.UnaryNode, newNegateEval, func(b ast.UnaryNode) ast.IsNode { return ast.NodeTypeNegate{UnaryNode: b} })
	case ast.NodeTypeNot:
		return tryPartialUnary(ctx, v.UnaryNode, newNotEval, func(b ast.UnaryNode) ast.IsNode { return ast.NodeTypeNot{UnaryNode: b} })
	case ast.NodeTypeVariable:
		return tryPartial(ctx,
			[]ast.IsNode{},
			func(_ []types.Value) Evaler {
				return newVariableEval(v.Name)
			},
			func(_ []ast.IsNode) ast.IsNode {
				return ast.NodeTypeVariable{Name: v.Name}
			},
		)
	case ast.NodeTypeIn:
		return tryPartialBinary(ctx, v.BinaryNode, newInEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeIn{BinaryNode: b} })
	case ast.NodeTypeAnd:
		return partialAnd(ctx, v)
	case ast.NodeTypeOr:
		return partialOr(ctx, v)
	case ast.NodeTypeEquals:
		return tryPartialBinary(ctx, v.BinaryNode, newEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeEquals{BinaryNode: b} })
	case ast.NodeTypeNotEquals:
		return tryPartialBinary(ctx, v.BinaryNode, newNotEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeNotEquals{BinaryNode: b} })
	case ast.NodeTypeGreaterThan:
		return tryPartialBinary(ctx, v.BinaryNode, newLongGreaterThanEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeGreaterThan{BinaryNode: b} })
	case ast.NodeTypeGreaterThanOrEqual:
		return tryPartialBinary(ctx, v.BinaryNode, newLongGreaterThanOrEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeGreaterThanOrEqual{BinaryNode: b} })
	case ast.NodeTypeLessThan:
		return tryPartialBinary(ctx, v.BinaryNode, newLongLessThanEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeLessThan{BinaryNode: b} })
	case ast.NodeTypeLessThanOrEqual:
		return tryPartialBinary(ctx, v.BinaryNode, newLongLessThanOrEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeLessThanOrEqual{BinaryNode: b} })
	case ast.NodeTypeSub:
		return tryPartialBinary(ctx, v.BinaryNode, newSubtractEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeSub{BinaryNode: b} })
	case ast.NodeTypeAdd:
		return tryPartialBinary(ctx, v.BinaryNode, newAddEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeAdd{BinaryNode: b} })
	case ast.NodeTypeMult:
		return tryPartialBinary(ctx, v.BinaryNode, newMultiplyEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeMult{BinaryNode: b} })
	case ast.NodeTypeContains:
		return tryPartialBinary(ctx, v.BinaryNode, newContainsEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeContains{BinaryNode: b} })
	case ast.NodeTypeContainsAll:
		return tryPartialBinary(ctx, v.BinaryNode, newContainsAllEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeContainsAll{BinaryNode: b} })
	case ast.NodeTypeContainsAny:
		return tryPartialBinary(ctx, v.BinaryNode, newContainsAnyEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeContainsAny{BinaryNode: b} })
	default:
		panic(fmt.Sprintf("unknown node type %T", v))
	}
}

func isNonBoolValue(in ast.IsNode) bool {
	n, ok := in.(ast.NodeValue)
	if !ok {
		return false
	}
	_, ok = n.Value.(types.Boolean)
	return !ok
}

func isTrue(in ast.IsNode) bool {
	n, ok := in.(ast.NodeValue)
	if !ok {
		return false
	}
	v, ok := n.Value.(types.Boolean)
	if !ok {
		return false
	}
	return v == types.Boolean(true)
}

func isFalse(in ast.IsNode) bool {
	n, ok := in.(ast.NodeValue)
	if !ok {
		return false
	}
	v, ok := n.Value.(types.Boolean)
	if !ok {
		return false
	}
	return v == types.Boolean(false)
}

func partialIfThenElse(ctx *Context, v ast.NodeTypeIfThenElse) (ast.IsNode, error) {
	if_, ifErr := partial(ctx, v.If)
	switch {
	case ifErr != nil:
		return nil, ifErr
	case isNonBoolValue(if_):
		return nil, fmt.Errorf("%w: ifThenElse expected bool", ErrType)
	case isTrue(if_):
		return partial(ctx, v.Then)
	case isFalse(if_):
		return partial(ctx, v.Else)
	}
	then, thenErr := partial(ctx, v.Then)
	if thenErr != nil {
		then = ast.NodeTypeExtensionCall{Name: "error"}
	}
	else_, elseErr := partial(ctx, v.Else)
	if elseErr != nil {
		else_ = ast.NodeTypeExtensionCall{Name: "error"}
	}
	return ast.NodeTypeIfThenElse{If: if_, Then: then, Else: else_}, nil
}

func partialAnd(ctx *Context, v ast.NodeTypeAnd) (ast.IsNode, error) {
	left, leftErr := partial(ctx, v.Left)
	switch {
	case leftErr != nil:
		return nil, leftErr
	case isNonBoolValue(left):
		return nil, fmt.Errorf("%w: and expected bool", ErrType)
	case isFalse(left):
		return ast.NodeValue{Value: types.False}, nil
	case isTrue(left):
		return tryPartialBinary(ctx,
			ast.BinaryNode{Left: ast.NodeValue{Value: types.True}, Right: v.Right},
			newAndEval,
			func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeAnd{BinaryNode: b} },
		)
	}
	right, rightErr := partial(ctx, v.Right)
	if rightErr != nil {
		right = ast.NodeTypeExtensionCall{Name: "error"}
	}
	return ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{Left: left, Right: right}}, nil
}

func partialOr(ctx *Context, v ast.NodeTypeOr) (ast.IsNode, error) {
	left, leftErr := partial(ctx, v.Left)
	switch {
	case leftErr != nil:
		return nil, leftErr
	case isNonBoolValue(left):
		return nil, fmt.Errorf("%w: or expected bool", ErrType)
	case isTrue(left):
		return ast.NodeValue{Value: types.True}, nil
	case isFalse(left):
		return tryPartialBinary(ctx,
			ast.BinaryNode{Left: ast.NodeValue{Value: types.False}, Right: v.Right},
			newOrEval,
			func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeOr{BinaryNode: b} },
		)
	}
	right, rightErr := partial(ctx, v.Right)
	if rightErr != nil {
		right = ast.NodeTypeExtensionCall{Name: "error"}
	}
	return ast.NodeTypeOr{BinaryNode: ast.BinaryNode{Left: left, Right: right}}, nil
}
