package exptypes

import (
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// Entity wraps types.Entity to provide schema-guided JSON unmarshaling.
type Entity types.Entity

type rawEntity struct {
	UID     json.RawMessage            `json:"uid"`
	Attrs   map[string]json.RawMessage `json:"attrs"`
	Parents json.RawMessage            `json:"parents"`
	Tags    map[string]json.RawMessage `json:"tags"`
}

// UnmarshalJSONWithSchema unmarshals a single entity from JSON, using the
// resolved schema to guide attribute and tag parsing.
func (e *Entity) UnmarshalJSONWithSchema(data []byte, schema *resolved.Schema) error {
	var r rawEntity
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("unmarshaling entity: %w", err)
	}

	// Parse UID (always EntityUID, implicit form accepted)
	var uid types.EntityUID
	if err := json.Unmarshal(r.UID, &uid); err != nil {
		return fmt.Errorf("unmarshaling entity uid: %w", err)
	}

	// Parse parents (always []EntityUID, implicit form accepted)
	var parentUIDs []types.EntityUID
	if err := json.Unmarshal(r.Parents, &parentUIDs); err != nil {
		return fmt.Errorf("entity %s, parents: %w", uid, err)
	}
	parents := types.NewEntityUIDSet(parentUIDs...)

	// Look up entity type in schema
	if schemaEntity, ok := schema.Entities[uid.Type]; ok {
		return e.unmarshalWithShape(uid, parents, r, schemaEntity)
	}

	// Action and enum entities have no attributes or tags
	if _, ok := schema.Actions[uid]; ok {
		*e = Entity(types.Entity{
			UID:        uid,
			Parents:    parents,
			Attributes: types.NewRecord(nil),
			Tags:       types.NewRecord(nil),
		})
		return nil
	}
	if _, ok := schema.Enums[uid.Type]; ok {
		*e = Entity(types.Entity{
			UID:        uid,
			Parents:    parents,
			Attributes: types.NewRecord(nil),
			Tags:       types.NewRecord(nil),
		})
		return nil
	}

	return fmt.Errorf("entity type %q not found in schema", uid.Type)
}

func (e *Entity) unmarshalWithShape(
	uid types.EntityUID,
	parents types.EntityUIDSet,
	r rawEntity,
	schemaEntity resolved.Entity,
) error {
	attrs, err := unmarshalEntityAttrs(uid, r.Attrs, schemaEntity.Shape)
	if err != nil {
		return err
	}

	tags, err := unmarshalEntityTags(uid, r.Tags, schemaEntity.Tags)
	if err != nil {
		return err
	}

	*e = Entity(types.Entity{
		UID:        uid,
		Parents:    parents,
		Attributes: attrs,
		Tags:       tags,
	})
	return nil
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

