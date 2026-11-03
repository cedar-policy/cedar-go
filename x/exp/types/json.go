package exptypes

import (
	"github.com/go-json-experiment/json"

	"github.com/cedar-policy/cedar-go/internal/scanner"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
	"github.com/cedar-policy/cedar-go/x/exp/schema/validate"
)

type EntityMap types.EntityMap

func (em *EntityMap) UnmarshalJSONWithSchema(b []byte, schema *resolved.Schema) error {
	var uncoerced types.EntityMap
	if err := json.Unmarshal(b, &uncoerced); err != nil {
		return err
	}

	coerced := make(types.EntityMap, len(uncoerced))
	for uid, entity := range uncoerced {
		coerced[uid] = coerceEntity(entity, schema)
	}

	v := validate.New(schema, validate.WithStrict())
	if err := v.Entities(coerced); err != nil {
		return err
	}

	*em = EntityMap(coerced)
	return nil
}

type Entity types.Entity

func (e *Entity) UnmarshalJSONWithSchema(b []byte, schema *resolved.Schema) error {
	var uncoerced types.Entity
	if err := json.Unmarshal(b, &uncoerced); err != nil {
		return err
	}

	coerced := coerceEntity(uncoerced, schema)

	v := validate.New(schema, validate.WithStrict())
	if err := v.Entity(coerced); err != nil {
		return err
	}

	*e = Entity(coerced)
	return nil
}

func coerceEntity(uncoerced types.Entity, schema *resolved.Schema) types.Entity {
	if schemaEntity, ok := schema.Entities[uncoerced.UID.Type]; ok {
		return types.Entity{
			UID:        uncoerced.UID,
			Parents:    uncoerced.Parents,
			Attributes: coerceRecord(uncoerced.Attributes, schemaEntity.Shape).(types.Record),
			Tags:       coerceTagValues(uncoerced.Tags, schemaEntity.Tags),
		}
	}
	return uncoerced
}

// coerceRecord coerces each known attribute in a Record according to its schema
// type. Attributes absent from the schema pass through unchanged.
func coerceRecord(v types.Value, typ resolved.RecordType) types.Value {
	rec, ok := v.(types.Record)
	if !ok {
		return v
	}
	m := rec.Map()
	if m == nil {
		return v
	}
	changed := false
	for name, attr := range typ {
		val, ok := m[name]
		if !ok {
			continue
		}
		coerced := coerceValue(val, attr.Type)
		if !coerced.Equal(val) {
			m[name] = coerced
			changed = true
		}
	}
	if !changed {
		return v
	}
	return types.NewRecord(m)
}

// coerceTagValues coerces each tag value according to the tag type.
func coerceTagValues(tags types.Record, tagType resolved.IsType) types.Record {
	if tagType == nil {
		return tags
	}
	m := tags.Map()
	if m == nil {
		return tags
	}
	changed := false
	for k, v := range m {
		coerced := coerceValue(v, tagType)
		if !coerced.Equal(v) {
			m[k] = coerced
			changed = true
		}
	}
	if !changed {
		return tags
	}
	return types.NewRecord(m)
}

// coerceValue attempts to convert a parsed value to match the expected schema
// type. This handles implicit forms that the unguided parser cannot
// disambiguate: {"type":"X","id":"Y"} as EntityUID, and bare strings as
// extension types. If coercion fails, the original value is returned unchanged.
func coerceValue(v types.Value, typ resolved.IsType) (val types.Value) {
	switch typ := typ.(type) {
	case resolved.EntityType:
		val = coerceEntityUID(v)
	case resolved.ExtensionType:
		val = coerceExtension(v, typ)
	case resolved.SetType:
		val = coerceSet(v, typ)
	case resolved.RecordType:
		val = coerceRecord(v, typ)
	case resolved.StringType, resolved.LongType, resolved.BoolType:
		// No coercion needed
		val = v
	}
	return
}

// coerceEntityUID attempts to convert a Record into an EntityUID, accepting
// either the explicit form {"__entity": {"type": ..., "id": ...}}, which
// unguided parsing leaves as a Record when it has other keys, or the implicit
// form {"type": ..., "id": ...}. As in Rust Cedar's schema-based parsing, extra
// keys are ignored and the explicit form takes precedence. Returns v unchanged
// otherwise, including when the type isn't a valid Cedar name.
func coerceEntityUID(v types.Value) types.Value {
	rec, ok := v.(types.Record)
	if !ok {
		return v
	}
	if entity, ok := rec.Get("__entity"); ok {
		if uid, ok := coerceValueToEntityUID(entity); ok && scanner.IsPath(string(uid.Type)) {
			return uid
		}
	}
	if uid, ok := coerceValueToEntityUID(rec); ok && scanner.IsPath(string(uid.Type)) {
		return uid
	}
	return v
}

// This is a copy of the unexported function of the same name in the types
// package, so keep the two in sync.
func coerceValueToEntityUID(v types.Value) (types.EntityUID, bool) {
	rec, ok := v.(types.Record)
	if !ok {
		return types.EntityUID{}, false
	}
	typ, ok := rec.Get("type")
	if !ok {
		return types.EntityUID{}, false
	}
	id, ok := rec.Get("id")
	if !ok {
		return types.EntityUID{}, false
	}
	typStr, ok := typ.(types.String)
	if !ok {
		return types.EntityUID{}, false
	}
	idStr, ok := id.(types.String)
	if !ok {
		return types.EntityUID{}, false
	}
	return types.NewEntityUID(types.EntityType(typStr), idStr), true
}

// coerceExtension attempts to parse a String value as the given extension type.
// Returns v unchanged if v is not a String or parsing fails.
func coerceExtension(v types.Value, typ resolved.ExtensionType) types.Value {
	s, ok := v.(types.String)
	if !ok {
		return v
	}
	switch typ {
	case "ipaddr":
		if parsed, err := types.ParseIPAddr(string(s)); err == nil {
			return parsed
		}
	case "decimal":
		if parsed, err := types.ParseDecimal(string(s)); err == nil {
			return parsed
		}
	case "datetime":
		if parsed, err := types.ParseDatetime(string(s)); err == nil {
			return parsed
		}
	case "duration":
		if parsed, err := types.ParseDuration(string(s)); err == nil {
			return parsed
		}
	}
	return v
}

// coerceSet coerces each element in a Set according to the element type.
func coerceSet(v types.Value, typ resolved.SetType) types.Value {
	set, ok := v.(types.Set)
	if !ok {
		return v
	}
	var elems []types.Value
	changed := false
	for elem := range set.All() {
		coerced := coerceValue(elem, typ.Element)
		if !coerced.Equal(elem) {
			changed = true
		}
		elems = append(elems, coerced)
	}
	if !changed {
		return v
	}
	return types.NewSet(elems...)
}
