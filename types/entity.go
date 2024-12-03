package types

import (
	"encoding/json"
	"slices"
	"strings"
)

// An Entity defines the parents and attributes for an EntityUID.
type Entity struct {
	UID        EntityUID    `json:"uid"`
	Parents    EntityUIDSet `json:"parents"`
	Attributes Record       `json:"attrs"`
	Tags       Record       `json:"tags,omitempty"`
}

// MarshalJSON serializes Entity as a JSON object, using the implicit form of EntityUID encoding to match the Rust
// SDK's behavior.
func (e Entity) MarshalJSON() ([]byte, error) {
	parents := make([]ImplicitlyMarshaledEntityUID, 0, e.Parents.Len())
	e.Parents.Iterate(func(p EntityUID) bool {
		parents = append(parents, ImplicitlyMarshaledEntityUID(p))
		return true
	})
	slices.SortFunc(parents, func(a, b ImplicitlyMarshaledEntityUID) int {
		if cmp := strings.Compare(string(a.Type), string(b.Type)); cmp != 0 {
			return cmp
		}

		return strings.Compare(string(a.ID), string(b.ID))
	})

	m := struct {
		UID        ImplicitlyMarshaledEntityUID   `json:"uid"`
		Parents    []ImplicitlyMarshaledEntityUID `json:"parents"`
		Attributes Record                         `json:"attrs"`
		Tags       Record                         `json:"tags"`
	}{
		ImplicitlyMarshaledEntityUID(e.UID),
		parents,
		e.Attributes,
		e.Tags,
	}
	return json.Marshal(m)
}
