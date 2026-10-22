package types

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/go-json-experiment/json"
)

// An EntityGetter is an interface for retrieving an Entity by EntityUID.
type EntityGetter interface {
	Get(uid EntityUID) (Entity, bool)
}

var _ EntityGetter = EntityMap{}

// An EntityMap is a collection of all the entities that are needed to evaluate
// authorization requests.  The key is an EntityUID which uniquely identifies
// the Entity (it must be the same as the UID within the Entity itself.)
type EntityMap map[EntityUID]Entity

func (e EntityMap) MarshalJSON() ([]byte, error) {
	s := slices.Collect(maps.Values(e))
	slices.SortFunc(s, func(a, b Entity) int {
		return strings.Compare(a.UID.String(), b.UID.String())
	})
	return json.Marshal(s)
}

// UnmarshalJSON parses an EntityMap from a JSON array of entities. As in Rust
// Cedar, two entities with the same UID are an error.
//
// Unlike Rust Cedar, null is accepted as an empty EntityMap, because cedar-go
// v1.2.0 through v1.8.0 encoded an empty EntityMap as null.
func (e *EntityMap) UnmarshalJSON(b []byte) error {
	var s []Entity
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	var res = make(EntityMap, len(s))
	for _, ent := range s {
		if _, ok := res[ent.UID]; ok {
			return fmt.Errorf("%w: duplicate entity %s", errJSONInvalidEntity, ent.UID)
		}
		res[ent.UID] = ent
	}
	*e = res
	return nil
}

func (e EntityMap) Clone() EntityMap {
	return maps.Clone(e)
}

func (e EntityMap) Get(uid EntityUID) (Entity, bool) {
	ent, ok := e[uid]
	return ent, ok
}
