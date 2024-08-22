package types

import (
	"encoding/json"
	"slices"
	"strings"

	"golang.org/x/exp/maps"
)

// An Entities is a collection of all the Entities that are needed to evaluate
// authorization requests.  The key is an EntityUID which uniquely identifies
// the Entity (it must be the same as the UID within the Entity itself.)
type Entities map[EntityUID]*Entity

// An Entity defines the parents and attributes for an EntityUID.
type Entity struct {
	UID        EntityUID   `json:"uid"`
	Parents    []EntityUID `json:"parents,omitempty"`
	Attributes Record      `json:"attrs"`
}

func (e Entities) MarshalJSON() ([]byte, error) {
	s := e.toSlice()
	return json.Marshal(s)
}

func (e *Entities) UnmarshalJSON(b []byte) error {
	var s []*Entity
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*e = entitiesFromSlice(s)
	return nil
}

func entitiesFromSlice(s []*Entity) Entities {
	var res = Entities{}
	for _, e := range s {
		res[e.UID] = e
	}
	return res
}

func (e Entities) toSlice() []*Entity {
	s := maps.Values(e)
	slices.SortFunc(s, func(a, b *Entity) int {
		return strings.Compare(a.UID.String(), b.UID.String())
	})
	return s
}

func (e Entities) Clone() Entities {
	return maps.Clone(e)
}
