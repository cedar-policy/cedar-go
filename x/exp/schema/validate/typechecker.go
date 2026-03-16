package validate

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/internal/rust"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

// scopedError wraps an error with env-specific context for cross-env dedup.
// Matches Rust's EntityLUB-based structural identity on ValidationError.
type scopedError struct {
	err    error
	lubKey string
}

func (e *scopedError) Error() string { return e.err.Error() }
func (e *scopedError) Unwrap() error { return e.err }

// exprRootVar returns the root variable name of an expression, or "".
func exprRootVar(n ast.IsNode) types.String {
	switch v := n.(type) {
	case ast.NodeTypeVariable:
		return v.Name
	case ast.NodeTypeAccess:
		return exprRootVar(v.Arg)
	case ast.NodeTypeAdd, ast.NodeTypeAnd, ast.NodeTypeContains, ast.NodeTypeContainsAll,
		ast.NodeTypeContainsAny, ast.NodeTypeEquals, ast.NodeTypeExtensionCall,
		ast.NodeTypeGetTag, ast.NodeTypeGreaterThan, ast.NodeTypeGreaterThanOrEqual,
		ast.NodeTypeHas, ast.NodeTypeHasTag, ast.NodeTypeIfThenElse, ast.NodeTypeIn,
		ast.NodeTypeIs, ast.NodeTypeIsEmpty, ast.NodeTypeIsIn, ast.NodeTypeLessThan,
		ast.NodeTypeLessThanOrEqual, ast.NodeTypeLike, ast.NodeTypeMult, ast.NodeTypeNegate,
		ast.NodeTypeNot, ast.NodeTypeNotEquals, ast.NodeTypeOr, ast.NodeTypeRecord,
		ast.NodeTypeSet, ast.NodeTypeSub, ast.NodeValue:
	}
	return ""
}

// typeOfExpr infers the type of an expression given a request environment, schema, and capabilities.
// Returns the inferred type, updated capabilities (from `has` guards), and any type error.
func (v *Validator) typeOfExpr(env *requestEnv, expr ast.IsNode, caps capabilitySet) (cedarType, capabilitySet, error) {
	switch n := expr.(type) {
	case ast.NodeValue:
		ty, err := v.typeOfValue(n.Value)
		return ty, caps, err

	case ast.NodeTypeVariable:
		return typeOfVariable(env, n.Name), caps, nil

	case ast.NodeTypeAnd:
		return v.typeOfAnd(env, n, caps)

	case ast.NodeTypeOr:
		return v.typeOfOr(env, n, caps)

	case ast.NodeTypeNot:
		return v.typeOfNot(env, n, caps)

	case ast.NodeTypeIfThenElse:
		return v.typeOfIfThenElse(env, n, caps)

	case ast.NodeTypeEquals:
		return v.typeOfEquality(env, n.Left, n.Right, false, caps)

	case ast.NodeTypeNotEquals:
		return v.typeOfEquality(env, n.Left, n.Right, true, caps)

	case ast.NodeTypeLessThan:
		return v.typeOfComparison(env, n.Left, n.Right, caps, expectComparable, expectComparable)

	case ast.NodeTypeLessThanOrEqual:
		return v.typeOfComparison(env, n.Left, n.Right, caps, expectComparable, expectComparable)

	case ast.NodeTypeGreaterThan:
		return v.typeOfComparison(env, n.Left, n.Right, caps, expectComparable, expectComparable)

	case ast.NodeTypeGreaterThanOrEqual:
		return v.typeOfComparison(env, n.Left, n.Right, caps, expectComparable, expectComparable)

	case ast.NodeTypeAdd:
		return v.typeOfArith(env, n.Left, n.Right, caps)

	case ast.NodeTypeSub:
		return v.typeOfArith(env, n.Left, n.Right, caps)

	case ast.NodeTypeMult:
		return v.typeOfArith(env, n.Left, n.Right, caps)

	case ast.NodeTypeNegate:
		return v.typeOfNegate(env, n, caps)

	case ast.NodeTypeIn:
		return v.typeOfIn(env, n, caps)

	case ast.NodeTypeContains:
		return v.typeOfContains(env, n, caps)

	case ast.NodeTypeContainsAll:
		return v.typeOfContainsAllAny(env, n.Left, n.Right, caps)

	case ast.NodeTypeContainsAny:
		return v.typeOfContainsAllAny(env, n.Left, n.Right, caps)

	case ast.NodeTypeIsEmpty:
		return v.typeOfIsEmpty(env, n, caps)

	case ast.NodeTypeLike:
		return v.typeOfLike(env, n, caps)

	case ast.NodeTypeIs:
		return v.typeOfIs(env, n, caps)

	case ast.NodeTypeIsIn:
		return v.typeOfIsIn(env, n, caps)

	case ast.NodeTypeHas:
		return v.typeOfHas(env, n, caps)

	case ast.NodeTypeAccess:
		return v.typeOfAccess(env, n, caps)

	case ast.NodeTypeHasTag:
		return v.typeOfHasTag(env, n, caps)

	case ast.NodeTypeGetTag:
		return v.typeOfGetTag(env, n, caps)

	case ast.NodeTypeRecord:
		return v.typeOfRecord(env, n, caps)

	case ast.NodeTypeSet:
		return v.typeOfSet(env, n, caps)

	case ast.NodeTypeExtensionCall:
	}
	return v.typeOfExtensionCall(env, expr.(ast.NodeTypeExtensionCall), caps)
}

func (v *Validator) typeOfValue(val types.Value) (cedarType, error) {
	switch val := val.(type) {
	case types.Boolean:
		if val {
			return typeTrue{}, nil
		}
		return typeFalse{}, nil
	case types.Long:
		return typeLong{}, nil
	case types.String:
		return typeString{}, nil
	case types.EntityUID:
	}
	return v.typeOfEntityUID(val.(types.EntityUID))
}

func (v *Validator) typeOfEntityUID(uid types.EntityUID) (cedarType, error) {
	et := uid.Type
	if v.isKnownEntityType(et) {
		return typeEntity{lub: singleEntityLUB(et)}, nil
	}
	if isActionEntity(et) {
		if _, ok := v.schema.Actions[uid]; ok {
			return typeEntity{lub: singleEntityLUB(et)}, nil
		}
		for aUID := range v.schema.Actions {
			if aUID.Type == et {
				return nil, fmt.Errorf("unrecognized action `%s`", uid)
			}
		}
	}
	return nil, fmt.Errorf("unrecognized entity type `%s`", et)
}

func typeOfVariable(env *requestEnv, name types.String) cedarType {
	switch name {
	case "principal":
		return typeEntity{lub: singleEntityLUB(env.principalType)}
	case "action":
		return typeEntity{lub: singleEntityLUB(env.actionUID.Type)}
	case "resource":
		return typeEntity{lub: singleEntityLUB(env.resourceType)}
	}
	return env.contextType // "context"
}

func unexpectedTypeErr(expected string, actual cedarType) error {
	return fmt.Errorf("unexpected type: expected %s but saw %s", expected, cedarTypeName(actual))
}

func (v *Validator) typeOfAnd(env *requestEnv, n ast.NodeTypeAnd, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, lCaps, err := v.typeOfExpr(env, n.Left, caps)
	if err != nil {
		var errs []error
		if lt != nil && !isBoolType(lt) {
			errs = append(errs, err, unexpectedTypeErr("Bool", lt))
		} else {
			errs = append(errs, err)
		}
		// Short-circuit when left recovery type is typeFalse (matches Rust behavior:
		// even when left has errors, if its recovery type is False, skip typechecking right)
		if _, ok := lt.(typeFalse); ok {
			if refErr := v.validateEntityRefs(n.Right); refErr != nil {
				collectErrors(&errs, refErr)
			}
			return typeFalse{}, caps, errors.Join(errs...)
		}
		rt, _, rightErr := v.typeOfExpr(env, n.Right, caps)
		if rightErr != nil {
			collectErrors(&errs, rightErr)
		}
		// Propagate typeFalse recovery type when right is false
		if _, ok := rt.(typeFalse); ok {
			return typeFalse{}, caps, errors.Join(errs...)
		}
		return typeBool{}, caps, errors.Join(errs...)
	}
	if !isBoolType(lt) {
		return nil, caps, unexpectedTypeErr("Bool", lt)
	}

	if _, ok := lt.(typeFalse); ok {
		if err := v.validateEntityRefs(n.Right); err != nil {
			return nil, caps, err
		}
		return typeFalse{}, caps, nil
	}

	rt, rCaps, err := v.typeOfExpr(env, n.Right, caps.merge(lCaps))
	if err != nil {
		if _, ok := rt.(typeFalse); ok {
			return typeFalse{}, caps, err
		}
		return typeBool{}, caps, err
	}
	if !isBoolType(rt) {
		return nil, caps, unexpectedTypeErr("Bool", rt)
	}

	if _, ok := lt.(typeTrue); ok {
		return rt, rCaps, nil
	}
	if _, ok := rt.(typeFalse); ok {
		return typeFalse{}, rCaps, nil
	}

	return typeBool{}, rCaps, nil
}

func (v *Validator) typeOfOr(env *requestEnv, n ast.NodeTypeOr, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, lCaps, err := v.typeOfExpr(env, n.Left, caps)
	if err != nil {
		var errs []error
		if lt != nil && !isBoolType(lt) {
			errs = append(errs, err, unexpectedTypeErr("Bool", lt))
		} else {
			errs = append(errs, err)
		}
		if rt, _, rightErr := v.typeOfExpr(env, n.Right, caps); rightErr != nil {
			collectErrors(&errs, rightErr)
			if rt != nil && !isBoolType(rt) {
				errs = append(errs, unexpectedTypeErr("Bool", rt))
			}
		}
		return nil, caps, errors.Join(errs...)
	}
	if !isBoolType(lt) {
		return nil, caps, unexpectedTypeErr("Bool", lt)
	}

	if _, ok := lt.(typeTrue); ok {
		if err := v.validateEntityRefs(n.Right); err != nil {
			return nil, caps, err
		}
		return typeTrue{}, lCaps, nil
	}

	rt, rCaps, err := v.typeOfExpr(env, n.Right, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isBoolType(rt) {
		return nil, caps, unexpectedTypeErr("Bool", rt)
	}

	if _, ok := lt.(typeFalse); ok {
		return rt, rCaps, nil
	}
	if _, ok := rt.(typeTrue); ok {
		return typeTrue{}, rCaps, nil
	}
	if _, ok := rt.(typeFalse); ok {
		return lt, lCaps, nil
	}

	return typeBool{}, lCaps.intersect(rCaps), nil
}

func (v *Validator) typeOfNot(env *requestEnv, n ast.NodeTypeNot, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := v.typeOfExpr(env, n.Arg, caps)
	if err != nil {
		if t != nil && !isBoolType(t) {
			return nil, caps, errors.Join(err, unexpectedTypeErr("Bool", t))
		}
		return nil, caps, err
	}
	if !isBoolType(t) {
		return nil, caps, unexpectedTypeErr("Bool", t)
	}
	switch t.(type) {
	case typeTrue:
		return typeFalse{}, caps, nil
	case typeFalse:
		return typeTrue{}, caps, nil
	case typeBool, typeNever, typeLong, typeString, typeSet, typeRecord, typeEntity, typeExtension:
	}
	return typeBool{}, caps, nil
}

func (v *Validator) typeOfIfThenElse(env *requestEnv, n ast.NodeTypeIfThenElse, caps capabilitySet) (cedarType, capabilitySet, error) {
	condType, condCaps, condErr := v.typeOfExpr(env, n.If, caps)

	// If condition failed or has wrong type, still typecheck both branches for error collection
	if condErr != nil || (condType != nil && !isBoolType(condType)) {
		var errs []error
		if condErr != nil {
			collectErrors(&errs, condErr)
		}
		if condType != nil && !isBoolType(condType) {
			errs = append(errs, unexpectedTypeErr("Bool", condType))
		}
		if condType == nil || !isBoolType(condType) {
			// Condition has hard failure or wrong type — typecheck branches for errors
			thenType, _, thenErr := v.typeOfExpr(env, n.Then, caps)
			if thenErr != nil {
				collectErrors(&errs, thenErr)
			}
			elseType, _, elseErr := v.typeOfExpr(env, n.Else, caps)
			if elseErr != nil {
				collectErrors(&errs, elseErr)
			}
			// Return recovery type from branches if available (for parent type checks)
			var resultType cedarType
			if thenType != nil && elseType != nil {
				if lub, err := v.leastUpperBound(thenType, elseType); err == nil {
					resultType = lub
				}
			}
			return resultType, caps, errors.Join(errs...)
		}
		// condType is Bool with soft errors — fall through to evaluate branches properly
	}

	thenCaps := caps.merge(condCaps)

	// Constant folding: skip dead branches even when condition has soft errors.
	// This matches Rust where typeTrue/typeFalse conditions trigger dead-branch
	// elimination regardless of sub-expression errors.
	if _, ok := condType.(typeFalse); ok {
		if err := v.validateEntityRefs(n.Then); err != nil {
			if condErr != nil {
				return nil, caps, errors.Join(condErr, err)
			}
			return nil, caps, err
		}
		t, c, e := v.typeOfExpr(env, n.Else, caps)
		if condErr != nil {
			return t, c, errors.Join(condErr, e)
		}
		return t, c, e
	}
	if _, ok := condType.(typeTrue); ok {
		if err := v.validateEntityRefs(n.Else); err != nil {
			if condErr != nil {
				return nil, caps, errors.Join(condErr, err)
			}
			return nil, caps, err
		}
		t, c, e := v.typeOfExpr(env, n.Then, thenCaps)
		if condErr != nil {
			return t, c, errors.Join(condErr, e)
		}
		return t, c, e
	}

	// Evaluate both branches
	var errs []error
	if condErr != nil {
		collectErrors(&errs, condErr)
	}
	thenType, thenResultCaps, thenErr := v.typeOfExpr(env, n.Then, thenCaps)
	if thenErr != nil {
		collectErrors(&errs, thenErr)
	}
	elseType, elseResultCaps, elseErr := v.typeOfExpr(env, n.Else, caps)
	if elseErr != nil {
		collectErrors(&errs, elseErr)
	}
	if len(errs) > 0 {
		// Return recovery type from branches if available (for parent type checks)
		var resultType cedarType
		if thenType != nil && elseType != nil {
			if lub, lubErr := v.leastUpperBound(thenType, elseType); lubErr == nil {
				resultType = lub
			}
		}
		return resultType, caps, errors.Join(errs...)
	}

	if err := v.checkStrictEntityLUB(thenType, elseType); err != nil {
		return nil, caps, typeIncompatErr(thenType, elseType)
	}
	result, err := v.leastUpperBound(thenType, elseType)
	if err != nil {
		return nil, caps, typeIncompatErr(thenType, elseType)
	}
	return result, thenResultCaps.intersect(elseResultCaps), nil
}

func (v *Validator) typeOfEquality(env *requestEnv, left, right ast.IsNode, negated bool, caps capabilitySet) (cedarType, capabilitySet, error) {
	var errs []error
	lt, _, err := v.typeOfExpr(env, left, caps)
	if err != nil {
		if ue, ok := err.(interface{ Unwrap() []error }); ok {
			errs = append(errs, ue.Unwrap()...)
		} else {
			errs = append(errs, err)
		}
	}
	rt, _, err := v.typeOfExpr(env, right, caps)
	if err != nil {
		if ue, ok := err.(interface{ Unwrap() []error }); ok {
			errs = append(errs, ue.Unwrap()...)
		} else {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		if result, ok := evalLiteralEquality(left, right); ok {
			if negated {
				result = !result
			}
			if result {
				return typeTrue{}, caps, nil
			}
			return typeFalse{}, caps, nil
		}
		if lv, lok := left.(ast.NodeTypeVariable); lok {
			if rv, rok := right.(ast.NodeTypeVariable); rok && lv.Name == rv.Name {
				if negated {
					return typeFalse{}, caps, nil
				}
				return typeTrue{}, caps, nil
			}
		}
		if areTypesDisjoint(lt, rt) {
			if negated {
				return typeTrue{}, caps, nil
			}
			return typeFalse{}, caps, nil
		}
	}
	if v.strict && lt != nil && rt != nil && !areTypesDisjoint(lt, rt) {
		if _, err := v.leastUpperBound(lt, rt); err != nil {
			errs = append(errs, typeIncompatErr(lt, rt))
		}
	}
	if len(errs) > 0 {
		return typeBool{}, caps, errors.Join(errs...)
	}
	return typeBool{}, caps, nil
}

type typeExpectation func(cedarType) error

var expectComparable typeExpectation = func(t cedarType) error {
	if _, ok := t.(typeLong); ok {
		return nil
	}
	if ext, ok := t.(typeExtension); ok {
		if ext.name == "datetime" || ext.name == "duration" {
			return nil
		}
	}
	return unexpectedTypeErr("datetime, or duration, or Long", t)
}

func (v *Validator) typeOfComparison(env *requestEnv, left, right ast.IsNode, caps capabilitySet, expectLeft, expectRight typeExpectation) (cedarType, capabilitySet, error) {
	lt, _, leftErr := v.typeOfExpr(env, left, caps)
	rt, _, rightErr := v.typeOfExpr(env, right, caps)

	var leftExpectErr, rightExpectErr error
	if expectLeft != nil && lt != nil {
		leftExpectErr = expectLeft(lt)
	}
	if expectRight != nil && rt != nil {
		rightExpectErr = expectRight(rt)
	}

	var errs []error
	if leftErr != nil {
		if ue, ok := leftErr.(interface{ Unwrap() []error }); ok {
			errs = append(errs, ue.Unwrap()...)
		} else {
			errs = append(errs, leftErr)
		}
	}
	if leftExpectErr != nil {
		errs = append(errs, leftExpectErr)
	}
	if rightErr != nil {
		if ue, ok := rightErr.(interface{ Unwrap() []error }); ok {
			errs = append(errs, ue.Unwrap()...)
		} else {
			errs = append(errs, rightErr)
		}
	}
	if rightExpectErr != nil {
		errs = append(errs, rightExpectErr)
	}

	if len(errs) > 0 {
		return typeBool{}, caps, errors.Join(errs...)
	}
	return typeBool{}, caps, nil
}

func (v *Validator) typeOfArith(env *requestEnv, left, right ast.IsNode, caps capabilitySet) (cedarType, capabilitySet, error) {
	var errs []error
	lt, _, leftErr := v.typeOfExpr(env, left, caps)
	rt, _, rightErr := v.typeOfExpr(env, right, caps)

	if leftErr != nil {
		collectErrors(&errs, leftErr)
	} else if _, ok := lt.(typeLong); !ok {
		errs = append(errs, unexpectedTypeErr("Long", lt))
	}
	if rightErr != nil {
		collectErrors(&errs, rightErr)
	} else if _, ok := rt.(typeLong); !ok {
		errs = append(errs, unexpectedTypeErr("Long", rt))
	}
	if len(errs) > 0 {
		return typeLong{}, caps, errors.Join(errs...)
	}
	return typeLong{}, caps, nil
}

func (v *Validator) typeOfNegate(env *requestEnv, n ast.NodeTypeNegate, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := v.typeOfExpr(env, n.Arg, caps)
	var typeErr error
	if err != nil {
		typeErr = err
	}
	if _, ok := t.(typeLong); !ok && t != nil {
		if typeErr == nil {
			typeErr = unexpectedTypeErr("Long", t)
		} else {
			typeErr = errors.Join(typeErr, unexpectedTypeErr("Long", t))
		}
	}
	return typeLong{}, caps, typeErr
}

func (v *Validator) typeOfIn(env *requestEnv, n ast.NodeTypeIn, caps capabilitySet) (cedarType, capabilitySet, error) {
	var errs []error
	lt, _, leftErr := v.typeOfExpr(env, n.Left, caps)
	rt, _, rightErr := v.typeOfExpr(env, n.Right, caps)

	if leftErr != nil {
		collectErrors(&errs, leftErr)
	} else if !isEntityType(lt) {
		errs = append(errs, unexpectedTypeErr("__cedar::internal::AnyEntity", lt))
	}
	if rightErr != nil {
		collectErrors(&errs, rightErr)
	} else if !isEntityOrSetOfEntity(rt) {
		errs = append(errs, unexpectedTypeErr("Set<__cedar::internal::AnyEntity>, or __cedar::internal::AnyEntity", rt))
	}
	if len(errs) > 0 {
		return typeBool{}, caps, errors.Join(errs...)
	}

	// Special case: when LHS resolves to a known action EUID (action variable or
	// action literal), and RHS resolves to action EUID(s), check the action
	// hierarchy to constant-fold to True/False (matches Rust behavior).
	if lhsEUID := v.exprToActionEUID(env, n.Left); lhsEUID != nil {
		if rhsEUIDs := v.exprToActionEUIDs(env, n.Right); rhsEUIDs != nil {
			// Filter to action entities only
			var rhsActions []types.EntityUID
			for _, uid := range rhsEUIDs {
				if _, isAction := v.schema.Actions[uid]; isAction {
					rhsActions = append(rhsActions, uid)
				}
			}
			if len(rhsActions) > 0 {
				if v.isActionInSet(*lhsEUID, rhsActions) {
					return typeTrue{}, caps, nil
				}
				return typeFalse{}, caps, nil
			}
			// No actions on RHS — action can't be `in` non-actions
			return typeFalse{}, caps, nil
		}
	}

	if le, ok := lt.(typeEntity); ok {
		var rhsLUB *entityLUB
		if re, ok := rt.(typeEntity); ok {
			rhsLUB = &re.lub
		} else if rs, ok := rt.(typeSet); ok {
			if re, ok := rs.element.(typeEntity); ok {
				rhsLUB = &re.lub
			}
		}
		if rhsLUB != nil && !v.anyEntityDescendantOf(le.lub, *rhsLUB) {
			return typeFalse{}, caps, nil
		}
	}

	return typeBool{}, caps, nil
}

func (v *Validator) typeOfContains(env *requestEnv, n ast.NodeTypeContains, caps capabilitySet) (cedarType, capabilitySet, error) {
	var errs []error
	lt, _, leftErr := v.typeOfExpr(env, n.Left, caps)
	rt, _, rightErr := v.typeOfExpr(env, n.Right, caps)

	if leftErr != nil {
		collectErrors(&errs, leftErr)
	}
	if rightErr != nil {
		collectErrors(&errs, rightErr)
	}
	if len(errs) > 0 {
		return typeBool{}, caps, errors.Join(errs...)
	}

	st, ok := lt.(typeSet)
	if !ok {
		return nil, caps, unexpectedTypeErr("Set<__cedar::internal::Any>", lt)
	}
	if _, isNever := st.element.(typeNever); !isNever && v.strict {
		if _, err := v.leastUpperBound(st.element, rt); err != nil {
			return nil, caps, typeIncompatErr(st.element, rt)
		}
		if err := v.checkStrictEntityLUB(st.element, rt); err != nil {
			return nil, caps, typeIncompatErr(st.element, rt)
		}
	}
	return typeBool{}, caps, nil
}

func (v *Validator) typeOfContainsAllAny(env *requestEnv, left, right ast.IsNode, caps capabilitySet) (cedarType, capabilitySet, error) {
	var errs []error
	lt, _, leftErr := v.typeOfExpr(env, left, caps)
	rt, _, rightErr := v.typeOfExpr(env, right, caps)

	if leftErr != nil {
		collectErrors(&errs, leftErr)
	}
	if rightErr != nil {
		collectErrors(&errs, rightErr)
	}

	// Do strict type check even when there are sub-expression errors,
	// as long as both sides have valid set types (matches Rust behavior)
	lSet, lOk := lt.(typeSet)
	rSet, rOk := rt.(typeSet)
	if len(errs) > 0 {
		if v.strict && lOk && rOk {
			if _, err := v.leastUpperBound(lSet.element, rSet.element); err != nil {
				errs = append(errs, typeIncompatErr(lt, rt))
			}
		}
		return typeBool{}, caps, errors.Join(errs...)
	}

	if !lOk {
		return nil, caps, unexpectedTypeErr("Set<__cedar::internal::Any>", lt)
	}
	if !rOk {
		return nil, caps, unexpectedTypeErr("Set<__cedar::internal::Any>", rt)
	}
	if v.strict {
		if _, err := v.leastUpperBound(lSet.element, rSet.element); err != nil {
			return nil, caps, typeIncompatErr(lt, rt)
		}
	}
	return typeBool{}, caps, nil
}

func (v *Validator) typeOfIsEmpty(env *requestEnv, n ast.NodeTypeIsEmpty, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := v.typeOfExpr(env, n.Arg, caps)
	if err != nil {
		if t == nil {
			return nil, caps, err
		}
		// Non-nil type with error: collect error, continue type-checking
		if _, ok := t.(typeSet); !ok {
			return nil, caps, errors.Join(err, unexpectedTypeErr("Set<__cedar::internal::Any>", t))
		}
		return typeBool{}, caps, err
	}
	if _, ok := t.(typeSet); !ok {
		return nil, caps, unexpectedTypeErr("Set<__cedar::internal::Any>", t)
	}
	return typeBool{}, caps, nil
}

func (v *Validator) typeOfLike(env *requestEnv, n ast.NodeTypeLike, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := v.typeOfExpr(env, n.Arg, caps)
	if err != nil {
		if t == nil {
			return nil, caps, err
		}
		if _, ok := t.(typeString); !ok {
			return nil, caps, errors.Join(err, unexpectedTypeErr("String", t))
		}
		return typeBool{}, caps, err
	}
	if _, ok := t.(typeString); !ok {
		return nil, caps, unexpectedTypeErr("String", t)
	}
	return typeBool{}, caps, nil
}

func (v *Validator) typeOfIs(env *requestEnv, n ast.NodeTypeIs, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := v.typeOfExpr(env, n.Left, caps)
	if err != nil {
		if t == nil {
			return nil, caps, err
		}
		if !isEntityType(t) {
			return nil, caps, errors.Join(err, unexpectedTypeErr("__cedar::internal::AnyEntity", t))
		}
		return typeBool{}, caps, err
	}
	if !isEntityType(t) {
		return nil, caps, unexpectedTypeErr("__cedar::internal::AnyEntity", t)
	}

	if et, ok := t.(typeEntity); ok {
		if !slices.Contains(et.lub.elements, n.EntityType) {
			return typeFalse{}, caps, nil
		}
		if len(et.lub.elements) == 1 && et.lub.elements[0] == n.EntityType {
			return typeTrue{}, caps, nil
		}
	}

	return typeBool{}, caps, nil
}

func (v *Validator) typeOfIsIn(env *requestEnv, n ast.NodeTypeIsIn, caps capabilitySet) (cedarType, capabilitySet, error) {
	var errs []error

	lt, _, leftErr := v.typeOfExpr(env, n.Left, caps)

	if lt != nil && !isEntityType(lt) {
		errs = append(errs, unexpectedTypeErr("__cedar::internal::AnyEntity", lt))
	} else if leftErr != nil {
		errs = append(errs, leftErr)
	}

	if leftErr != nil && lt != nil && !isEntityType(lt) {
		errs = append(errs, leftErr)
	}

	if lt != nil && !isEntityType(lt) {
		errs = append(errs, unexpectedTypeErr("__cedar::internal::AnyEntity", lt))
	} else if leftErr != nil {
		errs = append(errs, leftErr)
	}

	rt, _, rightErr := v.typeOfExpr(env, n.Entity, caps)
	if rt != nil && !isEntityOrSetOfEntity(rt) {
		errs = append(errs, unexpectedTypeErr("Set<__cedar::internal::AnyEntity>, or __cedar::internal::AnyEntity", rt))
	} else if rightErr != nil {
		errs = append(errs, rightErr)
	}

	if len(errs) > 0 {
		return typeBool{}, caps, errors.Join(errs...)
	}
	return typeBool{}, caps, nil
}

func (v *Validator) typeOfHas(env *requestEnv, n ast.NodeTypeHas, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := v.typeOfExpr(env, n.Arg, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityOrRecordType(t) {
		return nil, caps, unexpectedTypeErr("__cedar::internal::AnyEntity, or __cedar::internal::OpenRecord{}", t)
	}

	resultType := v.hasResultType(t, n.Value)

	if _, isBool := resultType.(typeBool); isBool {
		if varName := exprVarName(n.Arg); varName != "" {
			if caps.has(capability{varName: varName, attr: n.Value}) {
				resultType = typeTrue{}
			}
		}
	}

	newCaps := caps
	if varName := exprVarName(n.Arg); varName != "" {
		newCaps = caps.add(capability{varName: varName, attr: n.Value})
	}

	return resultType, newCaps, nil
}

func (v *Validator) hasResultType(t cedarType, attr types.String) cedarType {
	if tv, ok := t.(typeRecord); ok {
		a, ok := tv.attrs[attr]
		if !ok {
			return typeFalse{}
		}
		if a.required {
			return typeTrue{}
		}
		return typeBool{}
	}
	return v.hasResultTypeEntity(t.(typeEntity).lub, attr)
}

func (v *Validator) hasResultTypeEntity(lub entityLUB, attr types.String) cedarType {
	anyHas := false
	for _, et := range lub.elements {
		entity, ok := v.schema.Entities[et]
		if !ok {
			continue
		}
		if _, ok := entity.Shape[attr]; ok {
			anyHas = true
		}
	}
	if !anyHas {
		return typeFalse{}
	}
	return typeBool{}
}

func (v *Validator) typeOfAccess(env *requestEnv, n ast.NodeTypeAccess, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, subErr := v.typeOfExpr(env, n.Arg, caps)
	if subErr != nil && t == nil {
		// Sub-expression failed with no recovery type — can't continue
		return nil, caps, subErr
	}
	if !isEntityOrRecordType(t) {
		if subErr != nil {
			return nil, caps, errors.Join(subErr, unexpectedTypeErr("__cedar::internal::AnyEntity, or __cedar::internal::OpenRecord{}", t))
		}
		return nil, caps, unexpectedTypeErr("__cedar::internal::AnyEntity, or __cedar::internal::OpenRecord{}", t)
	}

	var errs []error
	if subErr != nil {
		collectErrors(&errs, subErr)
	}

	attrType := v.lookupAttributeType(t, n.Value)
	if attrType == nil {
		// Format error message based on type
		errs = append(errs, v.attrNotFoundError(env, t, n.Value))
		return nil, caps, errors.Join(errs...)
	}

	// Check if the attribute is optional and requires a `has` guard
	if !attrType.required {
		varName := exprVarName(n.Arg)
		if varName == "" || !caps.has(capability{varName: varName, attr: n.Value}) {
			errs = append(errs, v.unsafeOptionalAccessError(env, t, n.Value, exprVarName(n.Arg)))
		}
	}

	// Tag the result with entity source when accessing entity attributes that return records
	result := attrType.typ
	if et, ok := t.(typeEntity); ok {
		if rec, ok := result.(typeRecord); ok {
			rec.source = &entityAttrSource{lub: et.lub, attr: n.Value}
			result = rec
		}
	}

	if len(errs) > 0 {
		return result, caps, errors.Join(errs...)
	}
	return result, caps, nil
}

// attrNotFoundError formats the "attribute not found" error matching Rust Cedar's format.
// Called only when t is known to be a record or entity type.
func (v *Validator) attrNotFoundError(env *requestEnv, t cedarType, attr types.String) error {
	if rec, ok := t.(typeRecord); ok {
		if rec.source != nil {
			// Record came from entity attribute access: format as entity attribute with path
			// Use dot notation for valid identifiers, bracket notation otherwise
			fullPath := formatEntityAttrPath(rec.source.attr, attr)
			if len(rec.source.lub.elements) == 1 {
				return fmt.Errorf("attribute `%s` on entity type `%s` not found", fullPath, rec.source.lub.elements[0])
			}
			names := make([]string, len(rec.source.lub.elements))
			for i, et := range rec.source.lub.elements {
				names[i] = fmt.Sprintf("`%s`", et)
			}
			return fmt.Errorf("attribute `%s` on entity types %s not found", fullPath, joinComma(names))
		}
		// Context record access
		return fmt.Errorf("attribute `%s` in context for %s not found", attr, env.actionUID)
	}
	// Must be entity type (caller guarantees record or entity)
	te := t.(typeEntity)
	if len(te.lub.elements) == 1 {
		return fmt.Errorf("attribute `%s` on entity type `%s` not found", attr, te.lub.elements[0])
	}
	names := make([]string, len(te.lub.elements))
	for i, et := range te.lub.elements {
		names[i] = fmt.Sprintf("`%s`", et)
	}
	return fmt.Errorf("attribute `%s` on entity types %s not found", attr, joinComma(names))
}

// unsafeOptionalAccessError formats the "unsafe optional access" error matching Rust Cedar's format.
func (v *Validator) unsafeOptionalAccessError(env *requestEnv, t cedarType, attr types.String, varName types.String) error {
	if _, ok := t.(typeRecord); ok {
		// Context attribute - include path and action
		fullPath := string(attr)
		if varName != "" && varName != "context" {
			// nested path like context.session.token
			fullPath = string(varName)[len("context."):] + "." + string(attr)
		}
		return fmt.Errorf("unable to guarantee safety of access to optional attribute `%s` in context for %s", fullPath, env.actionUID)
	}
	// Must be entity type (caller guarantees record or entity via isEntityOrRecordType check)
	te := t.(typeEntity)
	if len(te.lub.elements) == 1 {
		return fmt.Errorf("unable to guarantee safety of access to optional attribute `%s` on entity type `%s`", attr, te.lub.elements[0])
	}
	names := make([]string, len(te.lub.elements))
	for i, et := range te.lub.elements {
		names[i] = fmt.Sprintf("`%s`", et)
	}
	return fmt.Errorf("unable to guarantee safety of access to optional attribute `%s` on entity types %s", attr, joinComma(names))
}

// formatEntityAttrPath formats a nested entity attribute path.
// Uses dot notation for valid Cedar identifiers, bracket notation otherwise.
func formatEntityAttrPath(parent, child types.String) string {
	var sb strings.Builder
	ps := string(parent)
	if isValidCedarIdent(ps) {
		sb.WriteString(ps)
	} else {
		fmt.Fprintf(&sb, `["%s"]`, rust.EscapeString(ps))
	}
	cs := string(child)
	if isValidCedarIdent(cs) {
		sb.WriteByte('.')
		sb.WriteString(cs)
	} else {
		fmt.Fprintf(&sb, `["%s"]`, rust.EscapeString(cs))
	}
	return sb.String()
}

// isValidCedarIdent returns true if s is a valid Cedar identifier (alphanumeric + underscore, not starting with digit).
func isValidCedarIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if i == 0 && !isIdentStart(r) {
			return false
		} else if i > 0 && !isIdentContinue(r) {
			return false
		}
	}
	return true
}

func isIdentStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func isIdentContinue(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}

func joinComma(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}

func (v *Validator) typeOfHasTag(env *requestEnv, n ast.NodeTypeHasTag, caps capabilitySet) (cedarType, capabilitySet, error) {
	var errs []error
	lt, _, leftErr := v.typeOfExpr(env, n.Left, caps)
	rt, _, rightErr := v.typeOfExpr(env, n.Right, caps)

	if leftErr != nil {
		collectErrors(&errs, leftErr)
	} else if !isEntityType(lt) {
		errs = append(errs, unexpectedTypeErr("__cedar::internal::AnyEntity", lt))
	}
	if rightErr != nil {
		collectErrors(&errs, rightErr)
	} else if _, ok := rt.(typeString); !ok {
		errs = append(errs, unexpectedTypeErr("String", rt))
	}
	if len(errs) > 0 {
		return typeBool{}, caps, errors.Join(errs...)
	}

	if et, ok := lt.(typeEntity); ok {
		if !v.entityHasTags(et.lub) {
			return typeFalse{}, caps, nil
		}
	}

	newCaps := caps
	if varName := exprVarName(n.Left); varName != "" {
		tagKey := tagCapabilityKey(n.Right)
		if tagKey != "" {
			newCaps = caps.add(capability{varName: varName, attr: types.String("__tag:" + tagKey)})
		}
	}

	return typeBool{}, newCaps, nil
}

func (v *Validator) typeOfGetTag(env *requestEnv, n ast.NodeTypeGetTag, caps capabilitySet) (cedarType, capabilitySet, error) {
	var errs []error

	lt, _, err := v.typeOfExpr(env, n.Left, caps)
	if err != nil {
		errs = append(errs, err)
	}
	et, ok := lt.(typeEntity)
	if lt == nil {
		// Return nil (no recovery type) when left is unknown. This prevents
		// downstream false positives (e.g., `is` check seeing typeString).
		return nil, caps, errors.Join(errs...)
	}
	if !ok {
		errs = append(errs, unexpectedTypeErr("__cedar::internal::AnyEntity", lt))
		return typeString{}, caps, errors.Join(errs...)
	}

	rt, _, err := v.typeOfExpr(env, n.Right, caps)
	if err != nil {
		errs = append(errs, err)
	}
	if rt != nil {
		if _, ok := rt.(typeString); !ok {
			errs = append(errs, unexpectedTypeErr("String", rt))
		}
	}

	// Check tag type compatibility first - if tag types are incompatible,
	// report that error instead of the unsafe access error
	var tagType cedarType = typeString{}
	if ok {
		tt, tagErr := v.entityTagType(et.lub)
		tagType = tt
		if tagErr != nil {
			errs = append(errs, tagErr)
			return tagType, caps, errors.Join(errs...)
		}
	}

	varName := exprVarName(n.Left)
	tagKey := tagCapabilityKey(n.Right)
	hasCapability := varName != "" && tagKey != "" && caps.has(capability{varName: varName, attr: types.String("__tag:" + tagKey)})

	if hasCapability {
		// Capability is only set by hasTag when entity supports tags
	} else {
		tagExpr := parser.MarshalExpr(rewriteConstITE(n.Right))

		var entityTypeMsg string
		if ok {
			if len(et.lub.elements) == 1 {
				entityTypeMsg = fmt.Sprintf(" on entity type `%s`", et.lub.elements[0])
			}
		}

		var lubKey string
		switch string(exprRootVar(n.Right)) {
		case "principal":
			lubKey = string(env.principalType)
		case "resource":
			lubKey = string(env.resourceType)
		case "action", "context":
			lubKey = env.actionUID.String()
		}
		errs = append(errs, &scopedError{
			err:    fmt.Errorf("unable to guarantee safety of access to tag `%s`%s", tagExpr, entityTypeMsg),
			lubKey: lubKey,
		})
	}
	if len(errs) > 0 {
		return nil, caps, errors.Join(errs...)
	}
	return tagType, caps, nil
}

// rewriteConstITE recursively rewrites if-then-else nodes with constant boolean
// conditions, replacing the dead branch with a copy of the taken branch. This
// matches Rust Cedar's typechecker behavior where the AST is rewritten during
// typechecking. The rewrite does not modify the original AST nodes.
func rewriteConstITE(n ast.IsNode) ast.IsNode {
	ite, ok := n.(ast.NodeTypeIfThenElse)
	if !ok {
		return n
	}
	cond := rewriteConstITE(ite.If)
	thenBranch := rewriteConstITE(ite.Then)
	elseBranch := rewriteConstITE(ite.Else)
	if nv, ok := cond.(ast.NodeValue); ok {
		if b, ok := nv.Value.(types.Boolean); ok {
			if bool(b) {
				return ast.NodeTypeIfThenElse{If: cond, Then: thenBranch, Else: thenBranch}
			}
			return ast.NodeTypeIfThenElse{If: cond, Then: elseBranch, Else: elseBranch}
		}
	}
	return ast.NodeTypeIfThenElse{If: cond, Then: thenBranch, Else: elseBranch}
}

func (v *Validator) typeOfRecord(env *requestEnv, n ast.NodeTypeRecord, caps capabilitySet) (cedarType, capabilitySet, error) {
	attrs := make(map[types.String]attributeType, len(n.Elements))
	var errs []error
	allTyped := true
	for _, elem := range n.Elements {
		elemType, _, err := v.typeOfExpr(env, elem.Value, caps)
		if err != nil {
			collectErrors(&errs, err)
		}
		if elemType != nil {
			attrs[elem.Key] = attributeType{typ: elemType, required: true}
		} else {
			allTyped = false
		}
	}
	if len(errs) > 0 {
		// Return recovery type only when all elements have types (soft errors like parse failures)
		if allTyped {
			return typeRecord{attrs: attrs}, caps, errors.Join(errs...)
		}
		return nil, caps, errors.Join(errs...)
	}
	return typeRecord{attrs: attrs}, caps, nil
}

func (v *Validator) typeOfSet(env *requestEnv, n ast.NodeTypeSet, caps capabilitySet) (cedarType, capabilitySet, error) {
	if v.strict && len(n.Elements) == 0 {
		// Return nil type (no recovery) to match Rust behavior where empty set
		// returns TypecheckFail with no type data. This prevents parents from
		// adding spurious type incompatibility errors.
		return nil, caps, fmt.Errorf("empty set literals are forbidden in policies")
	}
	// Phase 1: typecheck all elements, collecting nested errors.
	// Include recovery types (non-nil types with errors) for LUB computation.
	var errs []error
	elemTypes := make([]cedarType, 0, len(n.Elements))
	for _, elem := range n.Elements {
		et, _, err := v.typeOfExpr(env, elem, caps)
		if err != nil {
			collectErrors(&errs, err)
		}
		if et != nil {
			elemTypes = append(elemTypes, et)
		}
	}
	// Phase 2: compute LUB. When there are sub-expression errors, provide a
	// recovery type only if ALL elements had types. This matches Rust where
	// Option::collect() short-circuits to None if any element has no type,
	// preventing spurious type incompatibility errors in parent expressions.
	if len(errs) > 0 {
		if len(elemTypes) < len(n.Elements) {
			// Some elements had no recovery type — return nil (no recovery)
			return nil, caps, errors.Join(errs...)
		}
		var elemType cedarType = typeNever{}
		for _, et := range elemTypes {
			lub, err := v.leastUpperBound(elemType, et)
			if err != nil {
				if len(elemTypes) > 2 {
					errs = append(errs, typeIncompatErrMulti(elemTypes))
				} else {
					errs = append(errs, typeIncompatErr(elemType, et))
				}
				break
			}
			elemType = lub
		}
		return typeSet{element: elemType}, caps, errors.Join(errs...)
	}
	var elemType cedarType = typeNever{}
	for _, et := range elemTypes {
		if err := v.checkStrictEntityLUB(elemType, et); err != nil {
			if len(elemTypes) > 2 {
				return nil, caps, typeIncompatErrMulti(elemTypes)
			}
			return nil, caps, typeIncompatErr(elemType, et)
		}
		lub, err := v.leastUpperBound(elemType, et)
		if err != nil {
			if len(elemTypes) > 2 {
				return nil, caps, typeIncompatErrMulti(elemTypes)
			}
			return nil, caps, typeIncompatErr(elemType, et)
		}
		elemType = lub
	}
	return typeSet{element: elemType}, caps, nil
}

func (v *Validator) typeOfExtensionCall(env *requestEnv, n ast.NodeTypeExtensionCall, caps capabilitySet) (cedarType, capabilitySet, error) {
	sig := extFuncTypes[n.Name]

	var errs []error

	if len(n.Args) != len(sig.argTypes) {
		// Still typecheck all args to collect nested errors (matches Rust behavior)
		for _, arg := range n.Args {
			if _, _, err := v.typeOfExpr(env, arg, caps); err != nil {
				collectErrors(&errs, err)
			}
		}
		errs = append(errs, fmt.Errorf("wrong number of arguments in extension function application. Expected %d, got %d", len(sig.argTypes), len(n.Args)))
		// Non-literal check is independent of arg count (matches Rust behavior)
		if sig.isConstructor && v.strict {
			allLiteral := true
			for _, arg := range n.Args {
				if _, ok := arg.(ast.NodeValue); !ok {
					allLiteral = false
					break
				}
			}
			if !allLiteral {
				errs = append(errs, fmt.Errorf("extension constructors may not be called with non-literal expressions"))
			}
		}
		return nil, caps, errors.Join(errs...)
	}

	// Extension constructors: check literal requirement BEFORE type check
	// Rust reports "non-literal expressions" when the arg is not a literal at all,
	// but reports a type error when the arg IS a literal of the wrong type.
	if sig.isConstructor && v.strict && len(n.Args) == 1 {
		if _, ok := n.Args[0].(ast.NodeValue); !ok {
			// Still typecheck the arg to collect nested errors
			if _, _, err := v.typeOfExpr(env, n.Args[0], caps); err != nil {
				collectErrors(&errs, err)
			}
			errs = append(errs, fmt.Errorf("extension constructors may not be called with non-literal expressions"))
			return sig.returnType, caps, errors.Join(errs...)
		}
	}

	// Check argument types — continue on error to collect all nested errors
	for i, arg := range n.Args {
		argType, _, err := v.typeOfExpr(env, arg, caps)
		if err != nil {
			collectErrors(&errs, err)
			continue
		}
		if !v.isSubtype(argType, sig.argTypes[i]) {
			errs = append(errs, unexpectedTypeErr(cedarTypeName(sig.argTypes[i]), argType))
		}
	}

	// Extension constructors: validate string literal values (even if there were other errors)
	if sig.isConstructor && len(n.Args) == 1 {
		if nv, ok := n.Args[0].(ast.NodeValue); ok {
			if s, ok := nv.Value.(types.String); ok {
				if err := validateExtensionValue(n.Name, string(s)); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	if len(errs) > 0 {
		return sig.returnType, caps, errors.Join(errs...)
	}
	return sig.returnType, caps, nil
}

// collectErrors unwraps joined errors and appends them individually.
func collectErrors(errs *[]error, err error) {
	if ue, ok := err.(interface{ Unwrap() []error }); ok {
		*errs = append(*errs, ue.Unwrap()...)
	} else {
		*errs = append(*errs, err)
	}
}

func areTypesDisjoint(a, b cedarType) bool {
	ae, aOk := a.(typeEntity)
	be, bOk := b.(typeEntity)
	if !aOk || !bOk {
		return false
	}
	return ae.lub.isDisjoint(be.lub)
}

func validateExtensionValue(funcName types.Path, value string) error {
	switch funcName {
	case "ip":
		if _, err := types.ParseIPAddr(value); err != nil {
			return fmt.Errorf("error during extension function argument validation: Failed to parse as IP address: `\"%s\"`", rust.EscapeString(value))
		}
	case "decimal":
		if _, err := types.ParseDecimal(value); err != nil {
			return fmt.Errorf("error during extension function argument validation: Failed to parse as a decimal value: `\"%s\"`", rust.EscapeString(value))
		}
	case "datetime":
		if _, err := types.ParseDatetime(value); err != nil {
			return fmt.Errorf("error during extension function argument validation: Failed to parse as a datetime value: `\"%s\"`", rust.EscapeString(value))
		}
	case "duration":
		if _, err := types.ParseDuration(value); err != nil {
			return fmt.Errorf("error during extension function argument validation: Failed to parse as a duration value: `\"%s\"`", rust.EscapeString(value))
		}
	}
	return nil
}

func isBoolType(t cedarType) bool {
	switch t.(type) {
	case typeBool, typeTrue, typeFalse:
		return true
	case typeNever, typeLong, typeString, typeSet, typeRecord, typeEntity, typeExtension:
	}
	return false
}

func isEntityType(t cedarType) bool {
	switch t.(type) {
	case typeEntity:
		return true
	case typeNever, typeTrue, typeFalse, typeBool, typeLong, typeString, typeSet, typeRecord, typeExtension:
	}
	return false
}

func isEntityOrRecordType(t cedarType) bool {
	switch t.(type) {
	case typeEntity, typeRecord:
		return true
	case typeNever, typeTrue, typeFalse, typeBool, typeLong, typeString, typeSet, typeExtension:
	}
	return false
}

func isEntityOrSetOfEntity(t cedarType) bool {
	if isEntityType(t) {
		return true
	}
	if st, ok := t.(typeSet); ok {
		if _, isNever := st.element.(typeNever); isNever {
			return true
		}
		return isEntityType(st.element)
	}
	return false
}

func exprVarName(n ast.IsNode) types.String {
	if nd, ok := n.(ast.NodeTypeVariable); ok {
		return nd.Name
	}
	if nd, ok := n.(ast.NodeTypeAccess); ok {
		if parent := exprVarName(nd.Arg); parent != "" {
			return parent + "." + nd.Value
		}
	}
	return ""
}

func (v *Validator) validateEntityRefs(n ast.IsNode) error {
	switch nd := n.(type) {
	case ast.NodeValue:
		if uid, ok := nd.Value.(types.EntityUID); ok {
			if _, err := v.typeOfEntityUID(uid); err != nil {
				return err
			}
		}
	case ast.NodeTypeVariable:
	case ast.NodeTypeIfThenElse:
		return errors.Join(
			v.validateEntityRefs(nd.If),
			v.validateEntityRefs(nd.Then),
			v.validateEntityRefs(nd.Else),
		)
	case ast.NodeTypeExtensionCall:
		var errs []error
		for _, arg := range nd.Args {
			if err := v.validateEntityRefs(arg); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	case ast.NodeTypeRecord:
		var errs []error
		for _, elem := range nd.Elements {
			if err := v.validateEntityRefs(elem.Value); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	case ast.NodeTypeSet:
		var errs []error
		for _, elem := range nd.Elements {
			if err := v.validateEntityRefs(elem); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	case ast.NodeTypeAnd:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeOr:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeEquals:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeNotEquals:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeLessThan:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeLessThanOrEqual:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeGreaterThan:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeGreaterThanOrEqual:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeAdd:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeSub:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeMult:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeIn:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeContains:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeContainsAll:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeContainsAny:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeHasTag:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeGetTag:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeNegate:
		return v.validateEntityRefs(nd.Arg)
	case ast.NodeTypeNot:
		return v.validateEntityRefs(nd.Arg)
	case ast.NodeTypeIsEmpty:
		return v.validateEntityRefs(nd.Arg)
	case ast.NodeTypeHas:
		return v.validateEntityRefs(nd.Arg)
	case ast.NodeTypeAccess:
		return v.validateEntityRefs(nd.Arg)
	case ast.NodeTypeLike:
		return v.validateEntityRefs(nd.Arg)
	case ast.NodeTypeIs:
		return v.validateEntityRefs(nd.Left)
	case ast.NodeTypeIsIn:
		return v.validateEntityRefsPair(nd.Left, nd.Entity)
	}
	return nil
}

func (v *Validator) validateEntityRefsPair(a, b ast.IsNode) error {
	return errors.Join(v.validateEntityRefs(a), v.validateEntityRefs(b))
}

func evalLiteralEquality(left, right ast.IsNode) (bool, bool) {
	lv, lok := left.(ast.NodeValue)
	rv, rok := right.(ast.NodeValue)
	if !lok || !rok {
		return false, false
	}
	return lv.Value.Equal(rv.Value), true
}

// exprToActionEUID resolves an expression to an action EntityUID if possible.
// Returns the EUID for the `action` variable or an action entity literal.
func (v *Validator) exprToActionEUID(env *requestEnv, n ast.IsNode) *types.EntityUID {
	if nd, ok := n.(ast.NodeTypeVariable); ok && nd.Name == "action" {
		return &env.actionUID
	}
	if nd, ok := n.(ast.NodeValue); ok {
		if uid, ok := nd.Value.(types.EntityUID); ok {
			if _, isAction := v.schema.Actions[uid]; isAction {
				return &uid
			}
		}
	}
	return nil
}

// exprToActionEUIDs resolves an expression to action EntityUIDs.
// Handles single entities, action variable, and set literals.
func (v *Validator) exprToActionEUIDs(env *requestEnv, n ast.IsNode) []types.EntityUID {
	if uid := v.exprToActionEUID(env, n); uid != nil {
		return []types.EntityUID{*uid}
	}
	if s, ok := n.(ast.NodeTypeSet); ok {
		var uids []types.EntityUID
		for _, elem := range s.Elements {
			uid := v.exprToActionEUID(env, elem)
			if uid == nil {
				// Non-literal element — can't constant fold
				if nv, ok := elem.(ast.NodeValue); ok {
					if euid, ok := nv.Value.(types.EntityUID); ok {
						uids = append(uids, euid)
						continue
					}
				}
				return nil
			}
			uids = append(uids, *uid)
		}
		return uids
	}
	return nil
}

// isActionInSet checks if an action is `in` a set of actions using the schema hierarchy.
func (v *Validator) isActionInSet(action types.EntityUID, targets []types.EntityUID) bool {
	descendants := v.getActionsInSet(targets)
	return slices.Contains(descendants, action)
}

func tagCapabilityKey(n ast.IsNode) types.String {
	nv, ok := n.(ast.NodeValue)
	if !ok {
		return ""
	}
	s, ok := nv.Value.(types.String)
	if !ok {
		return ""
	}
	return s
}
