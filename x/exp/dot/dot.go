package dot

import (
	"fmt"
	"io"
	"iter"
	"strconv"

	"github.com/cedar-policy/cedar-go/types"
)

// Write takes an entity iterator and writes a DOT graph representing entities relationship.
//
// This function only returns an error on a failing write to w, so it is infallible if the Writer implementation cannot fail.
func Write(w io.Writer, entities iter.Seq[types.Entity]) error {
	// write prelude
	if _, err := fmt.Fprintln(w, "strict digraph {\n\tordering=\"out\"\n\tnode[shape=box]"); err != nil {
		return err
	}

	// write clusters (subgraphs)
	entitiesByType := getEntitiesByEntityType(entities)

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
	for entity := range entities {
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

func getEntitiesByEntityType(entities iter.Seq[types.Entity]) map[types.EntityType][]types.Entity {
	entitiesByType := map[types.EntityType][]types.Entity{}
	for entity := range entities {
		euid := entity.UID
		entityType := euid.Type
		if entities, ok := entitiesByType[entityType]; ok {
			entitiesByType[entityType] = append(entities, entity)
		} else {
			entitiesByType[entityType] = []types.Entity{entity}
		}
	}
	return entitiesByType
}
