// Package ast provides types for constructing Cedar schema ASTs programmatically.
package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

// Annotations maps annotation keys to their string values.
type Annotations map[types.Ident]types.String

// Entities maps entity type names to their definitions.
type Entities map[types.Ident]Entity

// Enums maps entity type names to their enum definitions.
type Enums map[types.Ident]Enum

// Actions maps action names to their definitions.
type Actions map[types.String]Action

// CommonTypes maps common type names to their definitions.
type CommonTypes map[types.Ident]CommonType

// Namespaces maps namespace paths to their definitions.
type Namespaces map[types.Namespace]Namespace

// Schema is the top-level Cedar schema AST.
// The Entities, Enums, Actions, and CommonTypes are for the top-level namespace.
type Schema struct {
	Entities    Entities
	Enums       Enums
	Actions     Actions
	CommonTypes CommonTypes
	Namespaces  Namespaces
}

// Namespace groups declarations under a namespace path.
type Namespace struct {
	Annotations Annotations
	Entities    Entities
	Enums       Enums
	Actions     Actions
	CommonTypes CommonTypes
}

// CommonType is a named type alias declaration.
type CommonType struct {
	Annotations Annotations
	Type        IsType
}

// Entity defines the shape and membership of an entity type.
type Entity struct {
	Annotations Annotations
	ParentTypes []EntityTypeRef
	Shape       RecordType
	Tags        IsType
}

// Enum defines an entity type whose valid values are a fixed set of strings.
type Enum struct {
	Annotations Annotations
	Values      []types.String
}

// Action defines what principals can do to resources.
// If AppliesTo is nil, the action never applies.
type Action struct {
	Annotations Annotations
	Parents     []ParentRef
	AppliesTo   *AppliesTo
}

// AppliesTo specifies the principal and resource types an action can apply to.
type AppliesTo struct {
	Principals []EntityTypeRef
	Resources  []EntityTypeRef
	Context    IsType
}

// ParentRef identifies an action parent by type and ID.
type ParentRef struct {
	Type EntityTypeRef
	ID   types.String
}

// ParentRefFromID creates a ParentRef with only an ID.
// Type is inferred as Action and namespaced during resolution.
func ParentRefFromID(id types.String) ParentRef {
	return ParentRef{
		ID: id,
	}
}

// NewParentRef creates a ParentRef with type and ID.
// Type will be namespaced during resolution.
func NewParentRef(typ EntityTypeRef, id types.String) ParentRef {
	return ParentRef{
		Type: typ,
		ID:   id,
	}
}
