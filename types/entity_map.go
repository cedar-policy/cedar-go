package types

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
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

func (e EntityMap) Get(uid EntityUID) (Entity, bool) {
	ent, ok := e[uid]
	return ent, ok
}

// ToDotStr writee entities into a DOT graph.
//
// This function only returns an error on a failing write to w, so it is infallible if the Writer implementation cannot fail.
func (entities EntityMap) ToDotStr(w io.Writer) error {
	// write prelude
	if _, err := fmt.Fprintln(w, "strict digraph {\n\tordering=\"out\"\n\tnode[shape=box]"); err != nil {
		return err
	}

	// write clusters (subgraphs)
	entitiesByType := entities.getEntitiesByEntityType()

	for et, entities := range entitiesByType {
		if _, err := fmt.Fprintf(w, "\tsubgraph \"cluster_%s\" {\n\t\tlabel=%s\n", et, toDotID(string(et))); err != nil {
			return err
		}
		for _, entity := range entities {
			if _, err := fmt.Fprintf(w, "\t\t%s [label=%s]\n", toDotID(entity.UID.String()), toDotID(entity.UID.ID.String())); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, "\t}"); err != nil {
			return err
		}
	}

	// adding edges
	for _, entity := range entities {
		for ancestor := range entity.Parents.All() {
			if _, err := fmt.Fprintf(w, "\t%s -> %s\n", toDotID(entity.UID.String()), toDotID(ancestor.String())); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintln(w, "}"); err != nil {
		return err
	}
	return nil
}

func toDotID(v string) string {
	// From DOT language reference:
	// An ID is one of the following:
	// Any string of alphabetic ([a-zA-Z\200-\377]) characters, underscores ('_') or digits([0-9]), not beginning with a digit;
	// a numeral [-]?(.[0-9]⁺ | [0-9]⁺(.[0-9]*)? );
	// any double-quoted string ("...") possibly containing escaped quotes (\");
	// an HTML string (<...>).
	// The best option to convert a `Name` or an `EntityUid` is to use double-quoted string.
	// The `strconv.Quote` function should be sufficient for our purpose.
	return strconv.Quote(v)
}

func (entities EntityMap) getEntitiesByEntityType() map[EntityType][]Entity {
	entitiesByType := map[EntityType][]Entity{}
	for _, entity := range entities {
		euid := entity.UID
		entityType := euid.Type
		if entities, ok := entitiesByType[entityType]; ok {
			entitiesByType[entityType] = append(entities, entity)
		} else {
			entitiesByType[entityType] = []Entity{entity}
		}
	}
	return entitiesByType
}
