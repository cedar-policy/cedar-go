package eval

import (
	"errors"
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/internal/mapset"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

const variableEntityType = "__cedar::variable"

func Variable(v types.String) types.Value {
	return types.NewEntityUID(variableEntityType, v)
}

const ignoreEntityType = "__cedar::ignore"

func Ignore() types.Value {
	return types.NewEntityUID(ignoreEntityType, "")
}

func IsVariable(v types.Value) bool {
	if ent, ok := v.(types.EntityUID); ok && ent.Type == variableEntityType {
		return true
	}
	return false
}

func ToVariable(ent types.EntityUID) (types.String, bool) {
	if ent.Type == variableEntityType {
		return ent.ID, true
	}
	return "", false
}

func IsIgnore(v types.Value) bool {
	if ent, ok := v.(types.EntityUID); ok && ent.Type == ignoreEntityType {
		return true
	}
	return false
}

// PartialPolicyToNode returns node compiled from a partially evaluated version of the policy
// and a boolean indicating if the policy should be kept.
// (Policies that are determined to evaluate to false are not kept.)
func PartialPolicyToNode(env Env, p *ast.Policy) (node ast.Node, keep bool) {
	pp, keep := PartialPolicy(env, p)
	if !keep {
		return ast.False(), keep
	}
	return policyToNode(pp), keep
}

// PartialPolicy returns a partially evaluated version of the policy and a boolean indicating if the policy should be kept.
// (Policies that are determined to evaluate to false are not kept.)
func PartialPolicy(env Env, p *ast.Policy) (policy *ast.Policy, keep bool) {
	p2 := *p
	if p2.Principal, keep = partialPrincipalScope(env, env.Principal, p2.Principal); !keep {
		return nil, false
	}
	if p2.Action, keep = partialActionScope(env, env.Action, p2.Action); !keep {
		return nil, false
	}
	if p2.Resource, keep = partialResourceScope(env, env.Resource, p2.Resource); !keep {
		return nil, false
	}
	p2.Annotations = slices.Clone(p.Annotations)
	p2.Conditions = nil
	for _, c := range p.Conditions {
		body, err := partial(env, c.Body)
		if errors.Is(err, errVariable) {
			p2.Conditions = append(p2.Conditions, c)
			continue
		} else if errors.Is(err, errIgnore) {
			if types.Effect(p.Effect) == types.Permit {
				continue
			}
			return nil, false
		} else if err != nil {
			p2.Conditions = append(p2.Conditions, ast.ConditionType{Condition: c.Condition, Body: extError(err)})
			return &p2, true
		} else if v, ok := body.(ast.NodeValue); ok {
			if b, bok := v.Value.(types.Boolean); bok {
				if bool(b) != bool(c.Condition) {
					return nil, false
				}
				continue
			}
			err := fmt.Errorf("%w: condition expected bool", ErrType)
			p2.Conditions = append(p2.Conditions, ast.ConditionType{Condition: c.Condition, Body: extError(err)})
			return &p2, true
		}
		p2.Conditions = append(p2.Conditions, ast.ConditionType{Condition: c.Condition, Body: body})
	}
	return &p2, true
}

func partialPrincipalScope(env Env, ent types.Value, scope ast.IsPrincipalScopeNode) (ast.IsPrincipalScopeNode, bool) {
	evaled, result := partialScopeEval(env, ent, scope)
	switch {
	case evaled && !result:
		return nil, false
	case evaled && result:
		return ast.ScopeTypeAll{}, true
	default:
		return scope, true
	}
}

func partialActionScope(env Env, ent types.Value, scope ast.IsActionScopeNode) (ast.IsActionScopeNode, bool) {
	evaled, result := partialScopeEval(env, ent, scope)
	switch {
	case evaled && !result:
		return nil, false
	case evaled && result:
		return ast.ScopeTypeAll{}, true
	default:
		return scope, true
	}
}

func partialResourceScope(env Env, ent types.Value, scope ast.IsResourceScopeNode) (ast.IsResourceScopeNode, bool) {
	evaled, result := partialScopeEval(env, ent, scope)
	switch {
	case evaled && !result:
		return nil, false
	case evaled && result:
		return ast.ScopeTypeAll{}, true
	default:
		return scope, true
	}
}

func partialScopeEval(env Env, ent types.Value, in ast.IsScopeNode) (evaled bool, result bool) {
	if IsVariable(ent) {
		return false, false
	} else if IsIgnore(ent) {
		return true, true
	}
	e, ok := ent.(types.EntityUID)
	if !ok {
		return false, false
	}
	switch t := in.(type) {
	case ast.ScopeTypeAll:
		return true, true
	case ast.ScopeTypeEq:
		return true, e == t.Entity
	case ast.ScopeTypeIn:
		return true, entityInOne(env, e, t.Entity)
	case ast.ScopeTypeInSet:
		set := mapset.Immutable(t.Entities...)
		return true, entityInSet(env, e, set)
	case ast.ScopeTypeIs:
		return true, e.Type == t.Type
	case ast.ScopeTypeIsIn:
		return true, e.Type == t.Type && entityInOne(env, e, t.Entity)
	default:
		panic(fmt.Sprintf("unknown scope type %T", t))
	}
}

var errVariable = fmt.Errorf("variable")
var errIgnore = fmt.Errorf("ignore")

// NOTE: nodes is modified in place, so be sure to send unique copy in
func tryPartial(env Env, nodes []ast.IsNode,
	mkEval func(values []types.Value) Evaler,
	mkNode func(nodes []ast.IsNode) ast.IsNode,
) (ast.IsNode, error) {
	var values []types.Value
	ok := true
	for i, n := range nodes {
		n, err := partial(env, n)
		if errors.Is(err, errVariable) {
			ok = false
			continue
		} else if err != nil {
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
		v, err := eval.Eval(env)
		if err != nil {
			return nil, err
		}
		if IsVariable(v) {
			return mkNode(nodes), errVariable
		} else if IsIgnore(v) {
			return nil, errIgnore
		}
		return ast.NodeValue{Value: v}, nil
	}
	return mkNode(nodes), nil
}

func tryPartialBinary(env Env, v ast.BinaryNode, mkEval func(a, b Evaler) Evaler, wrap func(b ast.BinaryNode) ast.IsNode) (ast.IsNode, error) {
	return tryPartial(env, []ast.IsNode{v.Left, v.Right},
		func(values []types.Value) Evaler { return mkEval(newLiteralEval(values[0]), newLiteralEval(values[1])) },
		func(nodes []ast.IsNode) ast.IsNode { return wrap(ast.BinaryNode{Left: nodes[0], Right: nodes[1]}) },
	)
}
func tryPartialUnary(env Env, v ast.UnaryNode, mkEval func(a Evaler) Evaler, wrap func(b ast.UnaryNode) ast.IsNode) (ast.IsNode, error) {
	return tryPartial(env, []ast.IsNode{v.Arg},
		func(values []types.Value) Evaler { return mkEval(newLiteralEval(values[0])) },
		func(nodes []ast.IsNode) ast.IsNode { return wrap(ast.UnaryNode{Arg: nodes[0]}) },
	)
}

// partial takes in an ast.Node and finds does as much as is possible given the context
func partial(env Env, n ast.IsNode) (ast.IsNode, error) {
	switch v := n.(type) {
	case ast.NodeTypeAccess:
		return tryPartial(env,
			[]ast.IsNode{v.Arg},
			func(values []types.Value) Evaler {
				return newAttributeAccessEval(newLiteralEval(values[0]), v.Value)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{Arg: nodes[0], Value: v.Value}}
			},
		)
	case ast.NodeTypeHas:
		return tryPartial(env,
			[]ast.IsNode{v.Arg},
			func(values []types.Value) Evaler {
				return newPartialHasEval(newLiteralEval(values[0]), v.Value)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeHas{StrOpNode: ast.StrOpNode{Arg: nodes[0], Value: v.Value}}
			},
		)
	case ast.NodeTypeGetTag:
		return tryPartial(env,
			[]ast.IsNode{v.Left, v.Right},
			func(values []types.Value) Evaler {
				return newGetTagEval(newLiteralEval(values[0]), newLiteralEval(values[1]))
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeGetTag{BinaryNode: ast.BinaryNode{Left: nodes[0], Right: nodes[1]}}
			},
		)
	case ast.NodeTypeHasTag:
		return tryPartial(env,
			[]ast.IsNode{v.Left, v.Right},
			func(values []types.Value) Evaler {
				return newHasTagEval(newLiteralEval(values[0]), newLiteralEval(values[1]))
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeHasTag{BinaryNode: ast.BinaryNode{Left: nodes[0], Right: nodes[1]}}
			},
		)
	case ast.NodeTypeLike:
		return tryPartial(env,
			[]ast.IsNode{v.Arg},
			func(values []types.Value) Evaler {
				return newLikeEval(newLiteralEval(values[0]), v.Value)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeLike{Arg: nodes[0], Value: v.Value}
			},
		)
	case ast.NodeTypeIfThenElse:
		return partialIfThenElse(env, v)
	case ast.NodeTypeIs:
		return tryPartial(env,
			[]ast.IsNode{v.Left},
			func(values []types.Value) Evaler {
				return newIsEval(newLiteralEval(values[0]), v.EntityType)
			},
			func(nodes []ast.IsNode) ast.IsNode {
				return ast.NodeTypeIs{Left: nodes[0], EntityType: v.EntityType}
			},
		)
	case ast.NodeTypeIsIn:
		return tryPartial(env,
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
		return tryPartial(env, nodes,
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
		return tryPartial(env, elements,
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
		return tryPartial(env, elements,
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
		return tryPartialUnary(env, v.UnaryNode, newNegateEval, func(b ast.UnaryNode) ast.IsNode { return ast.NodeTypeNegate{UnaryNode: b} })
	case ast.NodeTypeNot:
		return tryPartialUnary(env, v.UnaryNode, newNotEval, func(b ast.UnaryNode) ast.IsNode { return ast.NodeTypeNot{UnaryNode: b} })
	case ast.NodeTypeVariable:
		return tryPartial(env,
			[]ast.IsNode{},
			func(_ []types.Value) Evaler {
				return newVariableEval(v.Name)
			},
			func(_ []ast.IsNode) ast.IsNode {
				return ast.NodeTypeVariable{Name: v.Name}
			},
		)
	case ast.NodeTypeIn:
		return tryPartialBinary(env, v.BinaryNode, newInEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeIn{BinaryNode: b} })
	case ast.NodeTypeAnd:
		return partialAnd(env, v)
	case ast.NodeTypeOr:
		return partialOr(env, v)
	case ast.NodeTypeEquals:
		return tryPartialBinary(env, v.BinaryNode, newEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeEquals{BinaryNode: b} })
	case ast.NodeTypeNotEquals:
		return tryPartialBinary(env, v.BinaryNode, newNotEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeNotEquals{BinaryNode: b} })
	case ast.NodeTypeGreaterThan:
		return tryPartialBinary(env, v.BinaryNode, newComparableValueGreaterThanEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeGreaterThan{BinaryNode: b} })
	case ast.NodeTypeGreaterThanOrEqual:
		return tryPartialBinary(env, v.BinaryNode, newComparableValueGreaterThanOrEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeGreaterThanOrEqual{BinaryNode: b} })
	case ast.NodeTypeLessThan:
		return tryPartialBinary(env, v.BinaryNode, newComparableValueLessThanEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeLessThan{BinaryNode: b} })
	case ast.NodeTypeLessThanOrEqual:
		return tryPartialBinary(env, v.BinaryNode, newComparableValueLessThanOrEqualEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeLessThanOrEqual{BinaryNode: b} })
	case ast.NodeTypeSub:
		return tryPartialBinary(env, v.BinaryNode, newSubtractEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeSub{BinaryNode: b} })
	case ast.NodeTypeAdd:
		return tryPartialBinary(env, v.BinaryNode, newAddEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeAdd{BinaryNode: b} })
	case ast.NodeTypeMult:
		return tryPartialBinary(env, v.BinaryNode, newMultiplyEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeMult{BinaryNode: b} })
	case ast.NodeTypeContains:
		return tryPartialBinary(env, v.BinaryNode, newContainsEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeContains{BinaryNode: b} })
	case ast.NodeTypeContainsAll:
		return tryPartialBinary(env, v.BinaryNode, newContainsAllEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeContainsAll{BinaryNode: b} })
	case ast.NodeTypeContainsAny:
		return tryPartialBinary(env, v.BinaryNode, newContainsAnyEval, func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeContainsAny{BinaryNode: b} })
	case ast.NodeTypeIsEmpty:
		return tryPartialUnary(env, v.UnaryNode, newIsEmptyEval, func(b ast.UnaryNode) ast.IsNode { return ast.NodeTypeIsEmpty{UnaryNode: b} })
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

func partialIfThenElse(env Env, v ast.NodeTypeIfThenElse) (ast.IsNode, error) {
	ifNode, ifErr := partial(env, v.If)
	switch {
	case errors.Is(ifErr, errVariable):
	case ifErr != nil:
		return nil, ifErr
	case isNonBoolValue(ifNode):
		return nil, fmt.Errorf("%w: ifThenElse expected bool", ErrType)
	case isTrue(ifNode):
		return partial(env, v.Then)
	case isFalse(ifNode):
		return partial(env, v.Else)
	}
	thenNode, thenErr := partial(env, v.Then)
	if errors.Is(thenErr, errIgnore) {
		return nil, thenErr
	} else if thenErr != nil && !errors.Is(thenErr, errVariable) {
		thenNode = extError(thenErr)
	}
	elseNode, elseErr := partial(env, v.Else)
	if errors.Is(elseErr, errIgnore) {
		return nil, elseErr
	} else if elseErr != nil && !errors.Is(elseErr, errVariable) {
		elseNode = extError(elseErr)
	}
	return ast.NodeTypeIfThenElse{If: ifNode, Then: thenNode, Else: elseNode}, nil
}

func partialAnd(env Env, v ast.NodeTypeAnd) (ast.IsNode, error) {
	left, leftErr := partial(env, v.Left)
	switch {
	case errors.Is(leftErr, errVariable):
	case leftErr != nil:
		return nil, leftErr
	case isNonBoolValue(left):
		return nil, fmt.Errorf("%w: and expected bool", ErrType)
	case isFalse(left):
		return ast.NodeValue{Value: types.False}, nil
	case isTrue(left):
		return tryPartialBinary(env,
			ast.BinaryNode{Left: ast.NodeValue{Value: types.True}, Right: v.Right},
			newAndEval,
			func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeAnd{BinaryNode: b} },
		)
	}
	right, rightErr := partial(env, v.Right)
	if errors.Is(rightErr, errIgnore) {
		return nil, rightErr
	} else if rightErr != nil && !errors.Is(rightErr, errVariable) {
		right = extError(rightErr)
	}
	return ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{Left: left, Right: right}}, nil
}

func partialOr(env Env, v ast.NodeTypeOr) (ast.IsNode, error) {
	left, leftErr := partial(env, v.Left)
	switch {
	case errors.Is(leftErr, errVariable):
	case leftErr != nil:
		return nil, leftErr
	case isNonBoolValue(left):
		return nil, fmt.Errorf("%w: or expected bool", ErrType)
	case isTrue(left):
		return ast.NodeValue{Value: types.True}, nil
	case isFalse(left):
		return tryPartialBinary(env,
			ast.BinaryNode{Left: ast.NodeValue{Value: types.False}, Right: v.Right},
			newOrEval,
			func(b ast.BinaryNode) ast.IsNode { return ast.NodeTypeOr{BinaryNode: b} },
		)
	}
	right, rightErr := partial(env, v.Right)
	if errors.Is(rightErr, errIgnore) {
		return nil, rightErr
	} else if rightErr != nil && !errors.Is(rightErr, errVariable) {
		right = extError(rightErr)
	}
	return ast.NodeTypeOr{BinaryNode: ast.BinaryNode{Left: left, Right: right}}, nil
}

const partialErrorName = "__cedar::partialError"

func extError(err error) ast.NodeTypeExtensionCall {
	return ast.NodeTypeExtensionCall{Name: partialErrorName, Args: []ast.IsNode{ast.NodeValue{Value: types.String(err.Error())}}}
}

// PartialError returns a node that represents a partial error.
func PartialError(err error) ast.IsNode {
	return ast.NodeTypeExtensionCall{Name: partialErrorName, Args: []ast.IsNode{ast.NodeValue{Value: types.String(err.Error())}}}
}

// IsPartialError returns true if the node is a partial error.
func IsPartialError(n ast.IsNode) bool {
	ec, ok := n.(ast.NodeTypeExtensionCall)
	if !ok {
		return false
	}
	return ec.Name == partialErrorName
}

// partialHasEval
type partialHasEval struct {
	object    Evaler
	attribute types.String
}

func newPartialHasEval(record Evaler, attribute types.String) *partialHasEval {
	return &partialHasEval{object: record, attribute: attribute}
}

func (n *partialHasEval) Eval(env Env) (types.Value, error) {
	v, err := n.object.Eval(env)
	if err != nil {
		return zeroValue(), err
	}
	var record types.Record
	switch vv := v.(type) {
	case types.EntityUID:
		if rec, ok := env.Entities.Get(vv); ok {
			record = rec.Attributes
		}
	case types.Record:
		record = vv
	default:
		return zeroValue(), fmt.Errorf("%w: expected one of [record, (entity of type `any_entity_type`)], got %v", ErrType, TypeName(v))
	}
	v, ok := record.Get(n.attribute)
	if IsIgnore(v) {
		return nil, errIgnore
	}
	return types.Boolean(ok), nil
}

// partialErrorEval
type partialErrorEval struct {
	arg Evaler
}

func newPartialErrorEval(err Evaler) *partialErrorEval {
	return &partialErrorEval{
		arg: err,
	}
}

func (n *partialErrorEval) Eval(env Env) (types.Value, error) {
	v, err := evalString(n.arg, env)
	if err != nil {
		return nil, err
	}
	return zeroValue(), errors.New(string(v))
}
