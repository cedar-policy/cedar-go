package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

// IsType is the sealed sum type for all Cedar schema types.
//
//sumtype:decl
type IsType interface {
	isType()
}

// StringType represents the Cedar String type.
type StringType struct{}

func (StringType) isType() { _ = 0 }

// String returns a StringType.
func String() StringType { return StringType{} }

// LongType represents the Cedar Long type.
type LongType struct{}

func (LongType) isType() { _ = 0 }

// Long returns a LongType.
func Long() LongType { return LongType{} }

// BoolType represents the Cedar Bool type.
type BoolType struct{}

func (BoolType) isType() { _ = 0 }

// Bool returns a BoolType.
func Bool() BoolType { return BoolType{} }

// ExtensionType represents a Cedar extension type (e.g. ipaddr, decimal).
type ExtensionType types.Ident

func (ExtensionType) isType() { _ = 0 }

// IPAddr returns an ExtensionType for ipaddr.
func IPAddr() ExtensionType { return ExtensionType("ipaddr") }

// Decimal returns an ExtensionType for decimal.
func Decimal() ExtensionType { return ExtensionType("decimal") }

// Datetime returns an ExtensionType for datetime.
func Datetime() ExtensionType { return ExtensionType("datetime") }

// Duration returns an ExtensionType for duration.
func Duration() ExtensionType { return ExtensionType("duration") }

// SetType represents the Cedar Set type.
type SetType struct {
	Element IsType
}

func (SetType) isType() { _ = 0 }

// Set returns a SetType with the given element type.
func Set(element IsType) SetType {
	return SetType{Element: element}
}

// Attribute describes a single attribute in a record type.
type Attribute struct {
	Type        IsType
	Optional    bool
	Annotations Annotations
}

// RecordType maps attribute names to their types and optionality.
type RecordType map[types.String]Attribute

func (RecordType) isType() { _ = 0 }

// EntityTypeRef is a reference to an entity type in the schema.
type EntityTypeRef types.EntityType

func (EntityTypeRef) isType() { _ = 0 }

// EntityType returns an EntityTypeRef for the given entity type name.
func EntityType(name types.EntityType) EntityTypeRef {
	return EntityTypeRef(name)
}

// TypeRef is a reference to a common type or entity type by path, not yet resolved.
type TypeRef types.Path

func (TypeRef) isType() { _ = 0 }

// Type returns a TypeRef for the given path.
func Type(name types.Path) TypeRef {
	return TypeRef(name)
}
