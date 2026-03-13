package exptypes

import (
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// EntityMap wraps types.EntityMap to provide schema-guided JSON unmarshaling.
type EntityMap types.EntityMap

type rawEntity struct {
	UID     json.RawMessage            `json:"uid"`
	Attrs   map[string]json.RawMessage `json:"attrs"`
	Parents json.RawMessage            `json:"parents"`
	Tags    map[string]json.RawMessage `json:"tags"`
}

// UnmarshalJSONWithSchema unmarshals a JSON array of entities, using the
// resolved schema to guide attribute and tag parsing.
func (e *EntityMap) UnmarshalJSONWithSchema(data []byte, schema *resolved.Schema) error {
	var raw []rawEntity
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshaling entity array: %w", err)
	}

	result := make(types.EntityMap, len(raw))
	for _, r := range raw {
		entity, err := unmarshalEntity(r, schema)
		if err != nil {
			return err
		}

		// Check for duplicate UIDs
		if existing, ok := result[entity.UID]; ok {
			if !existing.Equal(entity) {
				return fmt.Errorf("duplicate entity %s with different content", entity.UID)
			}
			continue // identical duplicate, skip
		}
		result[entity.UID] = entity
	}

	*e = EntityMap(result)
	return nil
}

func unmarshalEntity(r rawEntity, schema *resolved.Schema) (types.Entity, error) {
	// Parse UID (always EntityUID, implicit form accepted)
	var uid types.EntityUID
	if err := json.Unmarshal(r.UID, &uid); err != nil {
		return types.Entity{}, fmt.Errorf("unmarshaling entity uid: %w", err)
	}

	// Parse parents (always []EntityUID, implicit form accepted)
	var parentUIDs []types.EntityUID
	if err := json.Unmarshal(r.Parents, &parentUIDs); err != nil {
		return types.Entity{}, fmt.Errorf("entity %s, parents: %w", uid, err)
	}
	parents := types.NewEntityUIDSet(parentUIDs...)

	// Look up entity type in schema
	if schemaEntity, ok := schema.Entities[uid.Type]; ok {
		return unmarshalSchemaEntity(uid, parents, r, schemaEntity)
	}

	// Action entities: parse without schema guidance
	if _, ok := schema.Actions[uid]; ok {
		return unmarshalActionEntity(uid, parents, r)
	}

	// Enum entities: accept with no attrs/tags
	if _, ok := schema.Enums[uid.Type]; ok {
		return types.Entity{
			UID:        uid,
			Parents:    parents,
			Attributes: types.NewRecord(nil),
			Tags:       types.NewRecord(nil),
		}, nil
	}

	return types.Entity{}, fmt.Errorf("entity type %q not found in schema", uid.Type)
}

func unmarshalSchemaEntity(
	uid types.EntityUID,
	parents types.EntityUIDSet,
	r rawEntity,
	schemaEntity resolved.Entity,
) (types.Entity, error) {
	attrs, err := unmarshalEntityAttrs(uid, r.Attrs, schemaEntity.Shape)
	if err != nil {
		return types.Entity{}, err
	}

	tags, err := unmarshalEntityTags(uid, r.Tags, schemaEntity.Tags)
	if err != nil {
		return types.Entity{}, err
	}

	return types.Entity{
		UID:        uid,
		Parents:    parents,
		Attributes: attrs,
		Tags:       tags,
	}, nil
}

func unmarshalEntityAttrs(
	uid types.EntityUID,
	rawAttrs map[string]json.RawMessage,
	shape resolved.RecordType,
) (types.Record, error) {
	// Reject unknown attributes (closed record)
	for k := range rawAttrs {
		if _, ok := shape[types.String(k)]; !ok {
			return types.Record{}, fmt.Errorf("entity %s: unexpected attribute %q", uid, k)
		}
	}

	m := make(types.RecordMap, len(rawAttrs))
	for name, attr := range shape {
		rawVal, ok := rawAttrs[string(name)]
		if !ok {
			if !attr.Optional {
				return types.Record{}, fmt.Errorf("entity %s: missing required attribute %q", uid, name)
			}
			continue
		}
		v, err := unmarshalValueJSON(rawVal, attr.Type)
		if err != nil {
			return types.Record{}, fmt.Errorf("entity %s, attribute %q: %w", uid, name, err)
		}
		m[name] = v
	}
	return types.NewRecord(m), nil
}

func unmarshalEntityTags(
	uid types.EntityUID,
	rawTags map[string]json.RawMessage,
	tagType resolved.IsType,
) (types.Record, error) {
	if tagType == nil {
		if len(rawTags) > 0 {
			return types.Record{}, fmt.Errorf("entity %s: tags not allowed by schema", uid)
		}
		return types.NewRecord(nil), nil
	}

	m := make(types.RecordMap, len(rawTags))
	for k, rawVal := range rawTags {
		v, err := unmarshalValueJSON(rawVal, tagType)
		if err != nil {
			return types.Record{}, fmt.Errorf("entity %s, tag %q: %w", uid, k, err)
		}
		m[types.String(k)] = v
	}
	return types.NewRecord(m), nil
}

// unmarshalActionEntity parses an action entity without schema guidance.
// After Commit 2 (removing implicit EntityUID from Value unmarshaling),
// action entity attributes containing EntityUIDs will need the explicit
// {"__entity":{...}} form.
func unmarshalActionEntity(
	uid types.EntityUID,
	parents types.EntityUIDSet,
	r rawEntity,
) (types.Entity, error) {
	attrs, err := unmarshalRawRecord(uid, "attrs", r.Attrs)
	if err != nil {
		return types.Entity{}, err
	}

	tags, err := unmarshalRawRecord(uid, "tags", r.Tags)
	if err != nil {
		return types.Entity{}, err
	}

	return types.Entity{
		UID:        uid,
		Parents:    parents,
		Attributes: attrs,
		Tags:       tags,
	}, nil
}

func unmarshalRawRecord(uid types.EntityUID, field string, raw map[string]json.RawMessage) (types.Record, error) {
	if len(raw) == 0 {
		return types.NewRecord(nil), nil
	}
	// Re-marshal and use standard Record unmarshaling (no schema guidance)
	rawBytes, err := json.Marshal(raw)
	if err != nil {
		return types.Record{}, fmt.Errorf("entity %s, %s: %w", uid, field, err)
	}
	var rec types.Record
	if err := json.Unmarshal(rawBytes, &rec); err != nil {
		return types.Record{}, fmt.Errorf("entity %s, %s: %w", uid, field, err)
	}
	return rec, nil
}
