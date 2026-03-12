package validate

import (
	"errors"
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// Entity validates a single entity against the schema.
func (v *Validator) Entity(entity types.Entity) error {
	et := entity.UID.Type

	// Check if it's an action entity
	if isActionEntity(et) {
		return v.validateActionEntity(entity)
	}

	// Look up in entities
	if schemaEntity, ok := v.schema.Entities[et]; ok {
		return v.validateEntity(entity, schemaEntity)
	}

	// Enum entities are accepted if the type exists
	if _, ok := v.schema.Enums[et]; ok {
		return nil
	}

	return newDeserError(fmt.Sprintf("entity type %q not found in schema", et))
}

// Entities validates all entities in the map against the schema.
// Returns a single flat error matching Rust Cedar's format:
// "entity does not conform to the schema" or "error during entity deserialization".
func (v *Validator) Entities(entities types.EntityMap) error {
	for _, entity := range entities {
		if err := v.Entity(entity); err != nil {
			var de *entityDeserError
			if errors.As(err, &de) {
				return fmt.Errorf("error during entity deserialization")
			}
			return fmt.Errorf("entity does not conform to the schema")
		}
	}
	return nil
}

func (v *Validator) validateActionEntity(entity types.Entity) error {
	action, ok := v.schema.Actions[entity.UID]
	if !ok {
		return fmt.Errorf("action %s not found in schema", entity.UID)
	}

	// Action entities should not have attributes or tags
	if entity.Attributes.Len() > 0 {
		return fmt.Errorf("action %s should not have attributes", entity.UID)
	}
	if entity.Tags.Len() > 0 {
		return fmt.Errorf("action %s should not have tags", entity.UID)
	}

	// Compute transitive closure of the action's parents from the schema
	closure := make(map[types.EntityUID]bool)
	var walk func(types.EntityUID)
	walk = func(uid types.EntityUID) {
		if closure[uid] {
			return
		}
		closure[uid] = true
		if a, ok := v.schema.Actions[uid]; ok {
			for p := range a.Entity.Parents.All() {
				walk(p)
			}
		}
	}
	for p := range action.Entity.Parents.All() {
		walk(p)
	}

	// Verify entity parents match the transitive closure
	for parent := range entity.Parents.All() {
		if !closure[parent] {
			return fmt.Errorf("action %s has unexpected parent %s", entity.UID, parent)
		}
	}
	for parent := range closure {
		if !entity.Parents.Contains(parent) {
			return fmt.Errorf("action %s missing expected parent %s", entity.UID, parent)
		}
	}

	return nil
}

func (v *Validator) validateEntity(entity types.Entity, schemaEntity resolved.Entity) error {
	// Validate parents
	for parent := range entity.Parents.All() {
		if !slices.Contains(schemaEntity.ParentTypes, parent.Type) {
			return fmt.Errorf("invalid parent type %q for entity type %q", parent.Type, entity.UID.Type)
		}
	}

	// Validate attributes
	if err := checkRecord(entity.Attributes, schemaEntity.Shape); err != nil {
		return err
	}

	// Validate tags
	if schemaEntity.Tags == nil {
		if entity.Tags.Len() > 0 {
			return newDeserError(fmt.Sprintf("entity type %q does not allow tags", entity.UID.Type))
		}
	} else {
		for tagVal := range entity.Tags.Values() {
			if err := checkValue(tagVal, schemaEntity.Tags); err != nil {
				return err
			}
		}
	}

	return nil
}
