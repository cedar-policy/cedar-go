package resolved

import (
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

// Annotations is a resolved annotation map.
type Annotations ast.Annotations

// IsType is the sealed sum type for resolved Cedar schema types.
// Unlike ast.IsType, there is no TypeRef (common types are inlined)
// and EntityTypeRef is replaced with EntityType.
//
//sumtype:decl
type IsType interface {
	isType()
}

// StringType represents the Cedar String type.
type StringType struct{}

func (StringType) isType() { _ = 0 }

// LongType represents the Cedar Long type.
type LongType struct{}

func (LongType) isType() { _ = 0 }

// BoolType represents the Cedar Bool type.
type BoolType struct{}

func (BoolType) isType() { _ = 0 }

// ExtensionType represents a Cedar extension type.
type ExtensionType types.Ident

func (ExtensionType) isType() { _ = 0 }

// SetType represents the Cedar Set type.
type SetType struct {
	Element IsType
}

func (SetType) isType() { _ = 0 }

// Attribute describes a single attribute in a resolved record type.
type Attribute struct {
	Type        IsType
	Optional    bool
	Annotations Annotations
}

// RecordType maps attribute names to their resolved types.
type RecordType map[types.String]Attribute

func (RecordType) isType() { _ = 0 }

// EntityType represents a reference to an entity type in a resolved schema.
type EntityType types.EntityType

func (EntityType) isType() { _ = 0 }
