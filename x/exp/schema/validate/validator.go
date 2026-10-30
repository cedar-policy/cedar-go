package validate

import (
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/internal/rust"
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

// invalidEnumEntityError reports an entity UID whose type is an enumerated
// entity type but whose ID is not one of the declared values.
type invalidEnumEntityError struct{ uid types.EntityUID }

func (e *invalidEnumEntityError) Error() string {
	return fmt.Sprintf("entity `%s` is of an enumerated entity type, but `\"%s\"` is not declared as a valid eid", e.uid, rust.EscapeString(string(e.uid.ID)))
}

// checkEnumEntity returns an error if uid is of an enumerated entity type
// but its ID is not one of the type's declared values.
func (v *Validator) checkEnumEntity(uid types.EntityUID) error {
	if enum, ok := v.schema.Enums[uid.Type]; ok && !slices.Contains(enum.Values, uid) {
		return &invalidEnumEntityError{uid: uid}
	}
	return nil
}
