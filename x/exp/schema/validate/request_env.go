package validate

import (
	"slices"

	"github.com/cedar-policy/cedar-go/types"
)

// requestEnv represents the type environment for type checking a policy condition.
type requestEnv struct {
	principalType types.EntityType
	actionUID     types.EntityUID
	resourceType  types.EntityType
	contextType   typeRecord
}

// generateRequestEnvs builds request environments from the schema for all action/principal/resource combos.
func (v *Validator) generateRequestEnvs() []requestEnv {
	var envs []requestEnv
	for uid, action := range v.schema.Actions {
		if action.AppliesTo == nil {
			continue
		}
		ctx := schemaRecordToCedarType(action.AppliesTo.Context)
		for _, pt := range action.AppliesTo.Principals {
			for _, rt := range action.AppliesTo.Resources {
				envs = append(envs, requestEnv{
					principalType: pt,
					actionUID:     uid,
					resourceType:  rt,
					contextType:   ctx,
				})
			}
		}
	}
	return envs
}

// filterEnvsForPolicy filters request environments to only those that match the policy's scope constraints.
// actionUIDs already includes descendants for `in` scopes, so only direct matching is needed.
func (v *Validator) filterEnvsForPolicy(envs []requestEnv, principalTypes, resourceTypes []types.EntityType, actionUIDs []types.EntityUID) []requestEnv {
	var filtered []requestEnv
	for _, env := range envs {
		if !matchesEntityTypeConstraint(env.principalType, principalTypes) {
			continue
		}
		if !matchesEntityTypeConstraint(env.resourceType, resourceTypes) {
			continue
		}
		if !matchesActionConstraint(env.actionUID, actionUIDs) {
			continue
		}
		filtered = append(filtered, env)
	}
	return filtered
}

func matchesEntityTypeConstraint(et types.EntityType, constraints []types.EntityType) bool {
	if constraints == nil {
		return true // ScopeTypeAll — no constraint
	}
	return slices.Contains(constraints, et)
}

func matchesActionConstraint(actionUID types.EntityUID, constraints []types.EntityUID) bool {
	if len(constraints) == 0 {
		return true
	}
	return slices.Contains(constraints, actionUID)
}
