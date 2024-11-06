package types

import (
	"encoding/json"
	"slices"
	"strings"

	"golang.org/x/exp/maps"
)

// EntityLoader defines the interface for loading entities from an entity store.
type EntityLoader interface {
	Load(EntityUID) (Entity, bool)
}

// An EntityMap is a collection of all the entities that are needed to evaluate
// authorization requests.  The key is an EntityUID which uniquely identifies
// the Entity (it must be the same as the UID within the Entity itself.)
type EntityMap map[EntityUID]Entity

func (e EntityMap) Load(k EntityUID) (Entity, bool) {
	v, ok := e[k]
	return v, ok
}

func (e EntityMap) MarshalJSON() ([]byte, error) {
	s := maps.Values(e)
	slices.SortFunc(s, func(a, b Entity) int {
		return strings.Compare(a.UID.String(), b.UID.String())
	})
	return json.Marshal(s)
}

func (e *EntityMap) UnmarshalJSON(b []byte) error {
	var s []Entity
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	var res = EntityMap{}
	for _, e := range s {
		res[e.UID] = e
	}
	*e = res
	return nil
}

func (e EntityMap) Clone() EntityMap {
	return maps.Clone(e)
}
