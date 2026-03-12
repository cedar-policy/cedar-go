package validate

import (
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// Validator validates Cedar policies, entities, and requests against a resolved schema.
type Validator struct {
	schema *resolved.Schema
	strict bool
}

// Option configures a Validator.
type Option func(*Validator)

// WithStrict returns an Option that enables strict validation mode (default).
func WithStrict() Option { return func(v *Validator) { v.strict = true } }

// WithPermissive returns an Option that enables permissive validation mode.
func WithPermissive() Option { return func(v *Validator) { v.strict = false } }

// New creates a Validator for the given schema. By default, strict mode is enabled.
func New(s *resolved.Schema, opts ...Option) *Validator {
	v := &Validator{schema: s, strict: true}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// isKnownEntityType returns true if the entity type exists in the schema
// as either a regular entity type or an enum type.
func (v *Validator) isKnownEntityType(et types.EntityType) bool {
	_, inEntities := v.schema.Entities[et]
	_, inEnums := v.schema.Enums[et]
	return inEntities || inEnums
}
