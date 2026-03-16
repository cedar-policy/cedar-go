package exptypes

import (
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// EntityMap wraps types.EntityMap to provide schema-guided JSON unmarshaling.
type EntityMap types.EntityMap

// UnmarshalJSONWithSchema unmarshals a JSON array of entities, using the
// resolved schema to guide attribute and tag parsing.
func (em *EntityMap) UnmarshalJSONWithSchema(data []byte, schema *resolved.Schema) error {
	var rawEntities []json.RawMessage
	if err := json.Unmarshal(data, &rawEntities); err != nil {
		return fmt.Errorf("unmarshaling entity array: %w", err)
	}

	result := make(types.EntityMap, len(rawEntities))
	for _, raw := range rawEntities {
		var entity Entity
		if err := entity.UnmarshalJSONWithSchema(raw, schema); err != nil {
			return err
		}

		te := types.Entity(entity)
		if existing, ok := result[te.UID]; ok {
			if !existing.Equal(te) {
				return fmt.Errorf("duplicate entity %s with different content", te.UID)
			}
			continue // identical duplicate, skip
		}
		result[te.UID] = te
	}

	*em = EntityMap(result)
	return nil
}
