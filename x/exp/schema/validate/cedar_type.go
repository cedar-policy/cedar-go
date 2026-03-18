package validate

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// cedarType is the sum type representing Cedar types for the type checker.
//
//sumtype:decl
type cedarType interface {
	isCedarType()
}

type typeNever struct{}                  // bottom type, subtype of all
type typeTrue struct{}                   // singleton bool true
type typeFalse struct{}                  // singleton bool false
type typeBool struct{}                   // Bool primitive
type typeLong struct{}                   // Long primitive
type typeString struct{}                 // String primitive
type typeSet struct{ element cedarType } // Set with element type
type typeRecord struct {
	attrs  map[types.String]attributeType
	source *entityAttrSource // non-nil when record came from entity attribute access
}

// entityAttrSource tracks the origin of a record type from entity attribute access.
type entityAttrSource struct {
	lub  entityLUB
	attr types.String // the entity attribute name (e.g., "r" in resource.r["key"])
}                                             // Record with attribute types
type typeEntity struct{ lub entityLUB }       // Entity with LUB of types
type typeExtension struct{ name types.Ident } // Extension type (ipaddr, decimal, etc.)

func (typeNever) isCedarType()     { _ = 0 }
func (typeTrue) isCedarType()      { _ = 0 }
func (typeFalse) isCedarType()     { _ = 0 }
func (typeBool) isCedarType()      { _ = 0 }
func (typeLong) isCedarType()      { _ = 0 }
func (typeString) isCedarType()    { _ = 0 }
func (typeSet) isCedarType()       { _ = 0 }
func (typeRecord) isCedarType()    { _ = 0 }
func (typeEntity) isCedarType()    { _ = 0 }
func (typeExtension) isCedarType() { _ = 0 }

// typeIncompatErr creates a type incompatibility error with types sorted to match Rust's order.
func typeIncompatErr(a, b cedarType) *typeIncompatError {
	nameA := cedarTypeName(a)
	nameB := cedarTypeName(b)
	if compareCedarType(a, b) > 0 {
		nameA, nameB = nameB, nameA
	}
	return &typeIncompatError{msg: fmt.Sprintf("the types %s and %s are not compatible", nameA, nameB)}
}

// typeIncompatErrMulti reports type incompatibility for 3+ types (e.g., set elements).
// Matches Rust's format: "the types A, B, and C are not compatible"
func typeIncompatErrMulti(types []cedarType) *typeIncompatError {
	// Sort and dedupe by full structural type ordering (matches Rust's
	// BTreeSet<Type> behavior, not just display text).
	sorted := make([]cedarType, len(types))
	copy(sorted, types)
	slices.SortFunc(sorted, compareCedarType)

	var deduped []cedarType
	for i, t := range sorted {
		if i == 0 || compareCedarType(sorted[i-1], t) != 0 {
			deduped = append(deduped, t)
		}
	}

	names := make([]string, len(deduped))
	for i, t := range deduped {
		names[i] = cedarTypeName(t)
	}
	if len(names) == 2 {
		return &typeIncompatError{msg: fmt.Sprintf("the types %s and %s are not compatible", names[0], names[1])}
	}
	// "the types A, B, and C are not compatible"
	last := names[len(names)-1]
	rest := names[:len(names)-1]
	var sb strings.Builder
	sb.WriteString("the types ")
	for i, n := range rest {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(n)
	}
	sb.WriteString(", and ")
	sb.WriteString(last)
	sb.WriteString(" are not compatible")
	return &typeIncompatError{msg: sb.String()}
}

// compareCedarType mirrors cedar-policy-core 4.9.1 derived Ord on validator::Type:
// Never < True < False < Bool < Long < String < Set < Record < Entity < Extension.
func compareCedarType(a, b cedarType) int {
	ra, rb := cedarTypeRank(a), cedarTypeRank(b)
	if ra != rb {
		if ra < rb {
			return -1
		}
		return 1
	}
	if ra <= 5 {
		return 0
	}
	if ra == 6 {
		return compareCedarType(a.(typeSet).element, b.(typeSet).element)
	}
	if ra == 7 {
		return compareRecordAttrs(a.(typeRecord).attrs, b.(typeRecord).attrs)
	}
	if ra == 8 {
		return compareEntityLUB(a.(typeEntity).lub, b.(typeEntity).lub)
	}
	return strings.Compare(string(a.(typeExtension).name), string(b.(typeExtension).name))
}

func cedarTypeRank(t cedarType) int {
	return cedarTypeRanks[reflect.TypeOf(t)]
}

var cedarTypeRanks = map[reflect.Type]int{
	reflect.TypeOf(typeNever{}):     0,
	reflect.TypeOf(typeTrue{}):      1,
	reflect.TypeOf(typeFalse{}):     2,
	reflect.TypeOf(typeBool{}):      3,
	reflect.TypeOf(typeLong{}):      4,
	reflect.TypeOf(typeString{}):    5,
	reflect.TypeOf(typeSet{}):       6,
	reflect.TypeOf(typeRecord{}):    7,
	reflect.TypeOf(typeEntity{}):    8,
	reflect.TypeOf(typeExtension{}): 9,
}

func compareEntityLUB(a, b entityLUB) int {
	n := min(len(a.elements), len(b.elements))
	for i := 0; i < n; i++ {
		if a.elements[i] < b.elements[i] {
			return -1
		}
		if a.elements[i] > b.elements[i] {
			return 1
		}
	}
	if len(a.elements) < len(b.elements) {
		return -1
	}
	if len(a.elements) > len(b.elements) {
		return 1
	}
	return 0
}

func compareRecordAttrs(a, b map[types.String]attributeType) int {
	ka := make([]string, 0, len(a))
	kb := make([]string, 0, len(b))
	for k := range a {
		ka = append(ka, string(k))
	}
	for k := range b {
		kb = append(kb, string(k))
	}
	slices.Sort(ka)
	slices.Sort(kb)

	n := min(len(ka), len(kb))
	for i := 0; i < n; i++ {
		if ka[i] < kb[i] {
			return -1
		}
		if ka[i] > kb[i] {
			return 1
		}
		at := a[types.String(ka[i])]
		bt := b[types.String(kb[i])]
		if c := compareCedarType(at.typ, bt.typ); c != 0 {
			return c
		}
		// Rust AttributeType field order is (attr_type, is_required).
		if at.required != bt.required {
			return requiredOrder[at.required] - requiredOrder[bt.required]
		}
	}
	if len(ka) < len(kb) {
		return -1
	}
	if len(ka) > len(kb) {
		return 1
	}
	return 0
}

// cedarTypeName returns the Rust Cedar display name for a type.
func cedarTypeName(t cedarType) string {
	switch tv := t.(type) {
	case typeNever:
		return "__cedar::internal::Never"
	case typeTrue:
		return "__cedar::internal::True"
	case typeFalse:
		return "__cedar::internal::False"
	case typeBool:
		return "Bool"
	case typeLong:
		return "Long"
	case typeString:
		return "String"
	case typeSet:
		return "Set<" + cedarTypeName(tv.element) + ">"
	case typeRecord:
		return cedarRecordTypeName(tv)
	case typeEntity:
		return cedarEntityTypeName(tv.lub)
	case typeExtension:
	}
	return string(t.(typeExtension).name)
}

func cedarEntityTypeName(lub entityLUB) string {
	if len(lub.elements) == 1 {
		return string(lub.elements[0])
	}
	// Multiple entity types - format as __cedar::internal::Union<A, B>
	names := make([]string, len(lub.elements))
	for i, et := range lub.elements {
		names[i] = string(et)
	}
	return "__cedar::internal::Union<" + strings.Join(names, ", ") + ">"
}

func cedarRecordTypeName(r typeRecord) string {
	if len(r.attrs) == 0 {
		return "{}"
	}
	keys := make([]string, 0, len(r.attrs))
	for k := range r.attrs {
		keys = append(keys, string(k))
	}
	slices.Sort(keys)
	var sb strings.Builder
	sb.WriteRune('{')
	for _, k := range keys {
		at := r.attrs[types.String(k)]
		sb.WriteString(k)
		sb.WriteString(optionalAttrSuffix[at.required])
		sb.WriteString(": ")
		sb.WriteString(cedarTypeName(at.typ))
		sb.WriteRune(',')
	}
	sb.WriteRune('}')
	return sb.String()
}

var requiredOrder = map[bool]int{
	false: 0,
	true:  1,
}

var optionalAttrSuffix = map[bool]string{
	true:  "",
	false: "?",
}

type attributeType struct {
	typ      cedarType
	required bool
}

// entityLUB represents the least upper bound (LUB) of a set of entity types.
// In type theory, the LUB is the most specific type that is a supertype of all
// given types. For entities, this is the union of the entity type names: e.g.
// the LUB of User and Admin is {User, Admin}. Elements are stored sorted for
// deterministic equality comparison.
type entityLUB struct {
	elements []types.EntityType // sorted, unique
}

// singleEntityLUB is an optimized constructor for the common single-element case,
// avoiding the clone/sort/compact overhead of newEntityLUB.
func singleEntityLUB(et types.EntityType) entityLUB {
	return entityLUB{elements: []types.EntityType{et}}
}

// isDisjoint returns true if the two entity LUBs have no entity types in common.
func (a entityLUB) isDisjoint(b entityLUB) bool {
	// Both LUBs are sorted, so we can check for intersection efficiently
	i, j := 0, 0
	for i < len(a.elements) && j < len(b.elements) {
		if a.elements[i] == b.elements[j] {
			return false // found a common element
		}
		if a.elements[i] < b.elements[j] {
			i++
		} else {
			j++
		}
	}
	return true // no common elements found
}

// isSubtype returns true if a is a subtype of b.
// Currently only called from extension function argument type checking,
// which only uses typeString and typeExtension argument types.
func (v *Validator) isSubtype(a, b cedarType) bool {
	switch b.(type) {
	case typeString:
		_, ok := a.(typeString)
		return ok
	case typeNever, typeTrue, typeFalse, typeBool, typeLong, typeSet, typeRecord, typeEntity, typeExtension:
	}
	av, ok := a.(typeExtension)
	bv := b.(typeExtension)
	return ok && av.name == bv.name
}

// leastUpperBound computes the LUB of two types.
func (v *Validator) leastUpperBound(a, b cedarType) (cedarType, error) {
	if _, ok := a.(typeNever); ok {
		return b, nil
	}
	if _, ok := b.(typeNever); ok {
		return a, nil
	}

	switch av := a.(type) {
	case typeTrue:
		switch b.(type) {
		case typeTrue:
			return typeTrue{}, nil
		case typeFalse, typeBool:
			return typeBool{}, nil
		case typeNever, typeLong, typeString, typeSet, typeRecord, typeEntity, typeExtension:
		}
	case typeFalse:
		switch b.(type) {
		case typeFalse:
			return typeFalse{}, nil
		case typeTrue, typeBool:
			return typeBool{}, nil
		case typeNever, typeLong, typeString, typeSet, typeRecord, typeEntity, typeExtension:
		}
	case typeBool:
		switch b.(type) {
		case typeTrue, typeFalse, typeBool:
			return typeBool{}, nil
		case typeNever, typeLong, typeString, typeSet, typeRecord, typeEntity, typeExtension:
		}
	case typeLong:
		if _, ok := b.(typeLong); ok {
			return typeLong{}, nil
		}
	case typeString:
		if _, ok := b.(typeString); ok {
			return typeString{}, nil
		}
	case typeSet:
		if bv, ok := b.(typeSet); ok {
			elem, err := v.leastUpperBound(av.element, bv.element)
			if err != nil {
				return nil, err
			}
			return typeSet{element: elem}, nil
		}
	case typeRecord:
		if bv, ok := b.(typeRecord); ok {
			return v.lubRecord(av, bv)
		}
	case typeEntity:
		if bv, ok := b.(typeEntity); ok {
			return typeEntity{lub: unionLUB(av.lub, bv.lub)}, nil
		}
	case typeExtension:
		if bv, ok := b.(typeExtension); ok && av.name == bv.name {
			return av, nil
		}
	case typeNever:
		// Already handled above; unreachable.
	}

	return nil, fmt.Errorf("incompatible types for least upper bound")
}

func (v *Validator) lubRecord(a, b typeRecord) (cedarType, error) {
	// Strict mode: records with different key sets cannot be combined (no width subtyping)
	if v.strict {
		if len(a.attrs) != len(b.attrs) {
			return nil, fmt.Errorf("record types have different attributes in strict mode")
		}
		for k := range a.attrs {
			if _, ok := b.attrs[k]; !ok {
				return nil, fmt.Errorf("record types have different attributes in strict mode")
			}
		}
	}

	attrs := make(map[types.String]attributeType)
	// Attributes in both
	for k, aAttr := range a.attrs {
		if bAttr, ok := b.attrs[k]; ok {
			lub, err := v.leastUpperBound(aAttr.typ, bAttr.typ)
			if err != nil {
				if v.strict {
					return nil, err
				}
				// Permissive mode: drop attributes with incompatible types
				continue
			}
			attrs[k] = attributeType{
				typ:      lub,
				required: aAttr.required && bAttr.required,
			}
		} else {
			attrs[k] = attributeType{typ: aAttr.typ, required: false}
		}
	}
	for k, bAttr := range b.attrs {
		if _, ok := a.attrs[k]; !ok {
			attrs[k] = attributeType{typ: bAttr.typ, required: false}
		}
	}
	return typeRecord{attrs: attrs}, nil
}

func unionLUB(a, b entityLUB) entityLUB {
	combined := append(slices.Clone(a.elements), b.elements...)
	slices.Sort(combined)
	combined = slices.Compact(combined)
	return entityLUB{elements: combined}
}

// schemaTypeToCedarType converts a resolved schema type to a cedarType.
func schemaTypeToCedarType(t resolved.IsType) cedarType {
	switch tv := t.(type) {
	case resolved.StringType:
		return typeString{}
	case resolved.LongType:
		return typeLong{}
	case resolved.BoolType:
		return typeBool{}
	case resolved.ExtensionType:
		return typeExtension{name: types.Ident(tv)}
	case resolved.SetType:
		return typeSet{element: schemaTypeToCedarType(tv.Element)}
	case resolved.RecordType:
		return schemaRecordToCedarType(tv)
	case resolved.EntityType:
	}
	return typeEntity{lub: singleEntityLUB(types.EntityType(t.(resolved.EntityType)))}
}

func schemaRecordToCedarType(rec resolved.RecordType) typeRecord {
	attrs := make(map[types.String]attributeType, len(rec))
	for name, attr := range rec {
		attrs[name] = attributeType{
			typ:      schemaTypeToCedarType(attr.Type),
			required: !attr.Optional,
		}
	}
	return typeRecord{attrs: attrs}
}

// lookupAttributeType looks up an attribute on a type using schema information.
// Called only when ty is already known to be a record or entity type.
func (v *Validator) lookupAttributeType(ty cedarType, attr types.String) *attributeType {
	switch tv := ty.(type) {
	case typeRecord:
		if a, ok := tv.attrs[attr]; ok {
			return &a
		}
		return nil
	case typeNever, typeTrue, typeFalse, typeBool, typeLong, typeString, typeSet, typeEntity, typeExtension:
	}
	return v.lookupEntityAttr(ty.(typeEntity).lub, attr)
}

func (v *Validator) lookupEntityAttr(lub entityLUB, attr types.String) *attributeType {
	// LUB always has at least one element (from singleEntityLUB or unionLUB),
	// and entity types always exist in the schema (validated during scope checking).
	var result *attributeType
	for _, et := range lub.elements {
		entity := v.schema.Entities[et]
		schemaAttr, ok := entity.Shape[attr]
		if !ok {
			return nil
		}
		at := &attributeType{
			typ:      schemaTypeToCedarType(schemaAttr.Type),
			required: !schemaAttr.Optional,
		}
		if result == nil {
			result = at
		} else {
			lubType, err := v.leastUpperBound(result.typ, at.typ)
			if err != nil {
				return nil
			}
			result = &attributeType{
				typ:      lubType,
				required: result.required && at.required,
			}
		}
	}
	return result
}

// entityHasTags returns true if all entities in the LUB have tags defined.
func (v *Validator) entityHasTags(lub entityLUB) bool {
	for _, et := range lub.elements {
		entity := v.schema.Entities[et]
		if entity.Tags == nil {
			return false
		}
	}
	return true
}

// entityTagType returns the LUB of the tag types for all entities in the LUB.
// Returns the LUB type and an error if the tag types are incompatible.
func (v *Validator) entityTagType(lub entityLUB) (cedarType, error) {
	var result cedarType = typeNever{}
	for _, et := range lub.elements {
		entity := v.schema.Entities[et]
		if entity.Tags == nil {
			return typeNever{}, nil
		}
		tagType := schemaTypeToCedarType(entity.Tags)
		tagLUB, err := v.leastUpperBound(result, tagType)
		if err != nil {
			return typeNever{}, typeIncompatErr(result, tagType)
		}
		result = tagLUB
	}
	return result, nil
}

// checkStrictEntityLUB checks if two types have compatible entity types in strict mode.
// In strict mode, entity LUBs between unrelated entity types are disallowed.
func (v *Validator) checkStrictEntityLUB(a, b cedarType) error {
	if !v.strict {
		return nil
	}
	if _, ok := a.(typeNever); ok {
		return nil
	}
	ae, aOk := a.(typeEntity)
	be, bOk := b.(typeEntity)
	if !aOk || !bOk {
		return nil
	}
	if !entityLUBsRelated(ae.lub, be.lub) {
		return fmt.Errorf("entity types are incompatible in strict mode")
	}
	return nil
}

// entityLUBsRelated returns true if any entity type in LUB a is the same as
// any entity type in LUB b. Matching Rust Cedar, hierarchy relationships
// (ancestor/descendant) do NOT count — only exact type equality.
func entityLUBsRelated(a, b entityLUB) bool {
	for _, at := range a.elements {
		if slices.Contains(b.elements, at) {
			return true
		}
	}
	return false
}

// isEntityDescendant returns true if childType can be a descendant (member) of ancestorType.
// This means childType lists ancestorType (directly or transitively) in its ParentTypes.
func (v *Validator) isEntityDescendant(childType, ancestorType types.EntityType) bool {
	// Entity types always exist in the schema (validated during scope checking).
	entity := v.schema.Entities[childType]
	for _, parent := range entity.ParentTypes {
		if parent == ancestorType {
			return true
		}
		if v.isEntityDescendant(parent, ancestorType) {
			return true
		}
	}
	return false
}

// anyEntityDescendantOf returns true if any entity type in lhs can be a
// descendant (member) of any entity type in rhs, or if lhs and rhs share a
// common entity type (same type means "in" can be true for the same entity).
func (v *Validator) anyEntityDescendantOf(lhs, rhs entityLUB) bool {
	for _, lt := range lhs.elements {
		for _, rt := range rhs.elements {
			if lt == rt {
				return true
			}
			if v.isEntityDescendant(lt, rt) {
				return true
			}
		}
	}
	return false
}

// isActionEntity returns true if the entity type is an action type.
func isActionEntity(et types.EntityType) bool {
	s := string(et)
	return s == "Action" || strings.HasSuffix(s, "::Action")
}
