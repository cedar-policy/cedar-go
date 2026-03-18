package validate

import (
	"errors"
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// typeIncompatError represents a type incompatibility error that should NOT get
// the "for policy `policyN`, " prefix. Rust emits these as standalone messages.
type typeIncompatError struct{ msg string }

func (e *typeIncompatError) Error() string { return e.msg }

// Policy validates a policy against the schema, performing scope validation
// and expression type checking.
func (v *Validator) Policy(policyID string, policy *ast.Policy) error {
	var errs []error

	// RBAC scope validation
	principalTypes, err := v.validatePrincipalScope(policy.Principal)
	if err != nil {
		errs = append(errs, flattenErrors(err)...)
	}

	// Validate action scope
	actionUIDs, err := v.validateAndGetActionUIDs(policy.Action)
	if err != nil {
		errs = append(errs, flattenErrors(err)...)
	}

	resourceTypes, err := v.validateResourceScope(policy.Resource)
	if err != nil {
		errs = append(errs, flattenErrors(err)...)
	}

	// Check action application
	if err := v.validateActionApplication(principalTypes, resourceTypes, actionUIDs); err != nil {
		errs = append(errs, err)
	}

	// Expression type checking
	allEnvs := v.generateRequestEnvs()
	envs := v.filterEnvsForPolicy(allEnvs, principalTypes, resourceTypes, actionUIDs)
	actionScopeEmptySet := false
	if sc, ok := policy.Action.(ast.ScopeTypeInSet); ok && len(sc.Entities) == 0 {
		actionScopeEmptySet = true
	}

	// `action in []` has mode-dependent behavior in Rust:
	// permissive: conditions are unreachable (skip condition typechecking)
	// strict: still typecheck conditions (plus emit empty-set-literal error)
	if actionScopeEmptySet {
		if v.strict {
			envs = v.filterEnvsForPolicy(allEnvs, principalTypes, resourceTypes, nil)
		} else {
			envs = nil
		}
	}

	// Rust reports empty action set literals in strict mode even when no request
	// environment is applicable for the scoped policy.
	if v.strict && actionScopeEmptySet {
		errs = append(errs, fmt.Errorf("empty set literals are forbidden in policies"))
	}

	if len(envs) > 0 {
		if len(policy.Conditions) > 0 {
			if err := v.typecheckConditions(envs, policy.Conditions); err != nil {
				errs = append(errs, flattenErrors(err)...)
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}

	// Prefix each error with "for policy `policyN`, " except type incompatibility errors
	var result []error
	for _, e := range errs {
		var tie *typeIncompatError
		if errors.As(e, &tie) || policyID == "" {
			result = append(result, e)
		} else {
			result = append(result, fmt.Errorf("for policy `%s`, %s", policyID, e.Error()))
		}
	}
	return errors.Join(result...)
}

// flattenErrors recursively unwraps joined errors into a flat list.
func flattenErrors(err error) []error {
	if ue, ok := err.(interface{ Unwrap() []error }); ok {
		var result []error
		for _, e := range ue.Unwrap() {
			result = append(result, flattenErrors(e)...)
		}
		return result
	}
	return []error{err}
}

// validatePrincipalScope validates the principal scope and returns the entity types it constrains to.
func (v *Validator) validatePrincipalScope(scope ast.IsPrincipalScopeNode) ([]types.EntityType, error) {
	switch sc := scope.(type) {
	case ast.ScopeTypeAll:
		return nil, nil
	case ast.ScopeTypeEq:
		entityTypes, err := v.validateScopeEntity(sc.Entity)
		if err != nil {
			return []types.EntityType{}, err
		}
		return entityTypes, nil
	case ast.ScopeTypeIn:
		_, err := v.validateScopeEntity(sc.Entity)
		if err != nil {
			return []types.EntityType{}, err
		}
		return v.getEntityTypesIn(sc.Entity.Type), nil
	case ast.ScopeTypeIs:
		entityTypes, err := v.validateScopeType(sc.Type)
		if err != nil {
			return []types.EntityType{}, err
		}
		return entityTypes, nil
	case ast.ScopeTypeIsIn:
	}
	isIn := scope.(ast.ScopeTypeIsIn)
	entityTypes, err := v.validateScopeType(isIn.Type)
	if err != nil {
		return []types.EntityType{}, err
	}
	if _, err := v.validateScopeEntity(isIn.Entity); err != nil {
		return []types.EntityType{}, err
	}
	typesIn := v.getEntityTypesIn(isIn.Entity.Type)
	if slices.Contains(typesIn, isIn.Type) {
		return entityTypes, nil
	}
	return []types.EntityType{}, nil
}

// validateAndGetActionUIDs validates that actions in the scope exist in the schema.
func (v *Validator) validateAndGetActionUIDs(scope ast.IsActionScopeNode) ([]types.EntityUID, error) {
	var errs []error
	var actionUIDs []types.EntityUID

	switch sc := scope.(type) {
	case ast.ScopeTypeAll:
		return nil, nil
	case ast.ScopeTypeEq:
		if _, ok := v.schema.Actions[sc.Entity]; !ok {
			errs = append(errs, fmt.Errorf("unrecognized action `%s`", sc.Entity))
		}
		actionUIDs = []types.EntityUID{sc.Entity}
	case ast.ScopeTypeIn:
		if _, ok := v.schema.Actions[sc.Entity]; !ok {
			errs = append(errs, fmt.Errorf("unrecognized action `%s`", sc.Entity))
		}
		actionUIDs = v.getActionsInSet([]types.EntityUID{sc.Entity})
	case ast.ScopeTypeInSet:
		for _, uid := range sc.Entities {
			if _, ok := v.schema.Actions[uid]; !ok {
				errs = append(errs, fmt.Errorf("unrecognized action `%s`", uid))
			}
		}
		actionUIDs = v.getActionsInSet(sc.Entities)
	}

	return actionUIDs, errors.Join(errs...)
}

// validateResourceScope validates the resource scope and returns the entity types it constrains to.
func (v *Validator) validateResourceScope(scope ast.IsResourceScopeNode) ([]types.EntityType, error) {
	switch sc := scope.(type) {
	case ast.ScopeTypeAll:
		return nil, nil
	case ast.ScopeTypeEq:
		entityTypes, err := v.validateScopeEntity(sc.Entity)
		if err != nil {
			return []types.EntityType{}, err
		}
		return entityTypes, nil
	case ast.ScopeTypeIn:
		_, err := v.validateScopeEntity(sc.Entity)
		if err != nil {
			return []types.EntityType{}, err
		}
		return v.getEntityTypesIn(sc.Entity.Type), nil
	case ast.ScopeTypeIs:
		entityTypes, err := v.validateScopeType(sc.Type)
		if err != nil {
			return []types.EntityType{}, err
		}
		return entityTypes, nil
	case ast.ScopeTypeIsIn:
	}
	isIn := scope.(ast.ScopeTypeIsIn)
	entityTypes, err := v.validateScopeType(isIn.Type)
	if err != nil {
		return []types.EntityType{}, err
	}
	if _, err := v.validateScopeEntity(isIn.Entity); err != nil {
		return []types.EntityType{}, err
	}
	typesIn := v.getEntityTypesIn(isIn.Entity.Type)
	if slices.Contains(typesIn, isIn.Type) {
		return entityTypes, nil
	}
	return []types.EntityType{}, nil
}

func (v *Validator) validateScopeEntity(uid types.EntityUID) ([]types.EntityType, error) {
	et := uid.Type
	if v.isKnownEntityType(et) {
		return []types.EntityType{et}, nil
	}
	if isActionEntity(et) {
		if _, ok := v.schema.Actions[uid]; ok {
			return []types.EntityType{et}, nil
		}
	}
	return nil, fmt.Errorf("unrecognized entity type `%s`", et)
}

func (v *Validator) validateScopeType(et types.EntityType) ([]types.EntityType, error) {
	if v.isKnownEntityType(et) {
		return []types.EntityType{et}, nil
	}
	return nil, fmt.Errorf("unrecognized entity type `%s`", et)
}

// validateActionApplication checks that at least one action's AppliesTo intersects
// the policy's principal AND resource constraints.
func (v *Validator) validateActionApplication(principalTypes, resourceTypes []types.EntityType, actionUIDs []types.EntityUID) error {
	if principalTypes == nil && resourceTypes == nil && actionUIDs == nil {
		return nil
	}

	var actions []resolved.Action
	hasUnknownAction := false
	if actionUIDs == nil {
		for _, a := range v.schema.Actions {
			actions = append(actions, a)
		}
	} else {
		for _, uid := range actionUIDs {
			if a, ok := v.schema.Actions[uid]; ok {
				actions = append(actions, a)
			} else {
				hasUnknownAction = true
			}
		}
	}

	if hasUnknownAction {
		return fmt.Errorf("unable to find an applicable action given the policy scope constraints")
	}

	for _, action := range actions {
		if action.AppliesTo == nil {
			continue
		}
		principalMatch := principalTypes == nil
		if !principalMatch {
			for _, pt := range principalTypes {
				if slices.Contains(action.AppliesTo.Principals, pt) {
					principalMatch = true
					break
				}
			}
		}
		resourceMatch := resourceTypes == nil
		if !resourceMatch {
			for _, rt := range resourceTypes {
				if slices.Contains(action.AppliesTo.Resources, rt) {
					resourceMatch = true
					break
				}
			}
		}
		if principalMatch && resourceMatch {
			return nil
		}
	}

	return fmt.Errorf("unable to find an applicable action given the policy scope constraints")
}

func (v *Validator) getActionsInSet(uids []types.EntityUID) []types.EntityUID {
	result := make([]types.EntityUID, 0, len(uids))
	for _, uid := range uids {
		result = append(result, uid)
		for aUID := range v.schema.Actions {
			if aUID == uid {
				continue
			}
			if v.isActionDescendant(aUID, uid) {
				result = append(result, aUID)
			}
		}
	}
	return result
}

func (v *Validator) isActionDescendant(actionUID, ancestorUID types.EntityUID) bool {
	action := v.schema.Actions[actionUID]
	for parent := range action.Entity.Parents.All() {
		if parent == ancestorUID {
			return true
		}
		if v.isActionDescendant(parent, ancestorUID) {
			return true
		}
	}
	return false
}

func (v *Validator) getEntityTypesIn(target types.EntityType) []types.EntityType {
	result := []types.EntityType{target}
	for name, entity := range v.schema.Entities {
		if slices.Contains(entity.ParentTypes, target) {
			result = append(result, name)
		}
	}
	changed := true
	for changed {
		changed = false
		for name, entity := range v.schema.Entities {
			if slices.Contains(result, name) {
				continue
			}
			for _, parent := range entity.ParentTypes {
				if slices.Contains(result, parent) {
					result = append(result, name)
					changed = true
					break
				}
			}
		}
	}
	return result
}

func (v *Validator) typecheckConditions(envs []requestEnv, conditions []ast.ConditionType) error {
	var allErrs []error
	for _, cond := range conditions {
		// Rust deduplicates ValidationError values across request environments
		// using HashSet semantics (identity includes source location). We
		// approximate identity using the leaf's unwrap-path plus rendered message:
		// this preserves same-text diagnostics from distinct AST sites while
		// collapsing repeats of the same site across environments.
		merged := map[string]error{}
		var mergedOrder []string
		for _, env := range envs {
			caps := newCapabilitySet()
			t, _, err := v.typeOfExpr(&env, cond.Body, caps)
			var envErrors []keyedError
			if err != nil {
				envErrors = append(envErrors, flattenErrorsWithPath(err, "e")...)
			}
			if t != nil && !isBoolType(t) {
				te := fmt.Errorf("unexpected type: expected Bool but saw %s", cedarTypeName(t))
				envErrors = append(envErrors, keyedError{key: "t", err: te})
			}
			for _, ke := range envErrors {
				key := ke.key + "\x00" + ke.err.Error()
				if _, ok := merged[key]; !ok {
					merged[key] = ke.err
					mergedOrder = append(mergedOrder, key)
				}
			}
		}
		for _, key := range mergedOrder {
			allErrs = append(allErrs, merged[key])
		}
	}
	return errors.Join(allErrs...)
}

type keyedError struct {
	key string
	err error
}

func flattenErrorsWithPath(err error, path string) []keyedError {
	if ue, ok := err.(interface{ Unwrap() []error }); ok {
		var out []keyedError
		for i, child := range ue.Unwrap() {
			if child == nil {
				continue
			}
			childPath := fmt.Sprintf("%s.%d", path, i)
			out = append(out, flattenErrorsWithPath(child, childPath)...)
		}
		return out
	}
	if mk, ok := mergeIdentityKey(err); ok {
		return []keyedError{{key: path + "\x1f" + mk, err: err}}
	}
	return []keyedError{{key: path, err: err}}
}
