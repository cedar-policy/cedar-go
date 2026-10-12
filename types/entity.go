package types

import (
	"encoding/json"
	jsonv2 "encoding/json/v2"
	"slices"
	"strings"

	"github.com/cedar-policy/cedar-go/internal"
)

// An Entity defines the parents and attributes for an EntityUID.
type Entity struct {
	UID        EntityUID    `json:"uid"`
	Parents    EntityUIDSet `json:"parents"`
	Attributes Record       `json:"attrs"`
	Tags       Record       `json:"tags"`
}

// MarshalJSON serializes Entity as a JSON object, using the implicit form of EntityUID encoding to match the Rust
// SDK's behavior.
func (e Entity) MarshalJSON() ([]byte, error) {
	parents := make([]ImplicitlyMarshaledEntityUID, 0, e.Parents.Len())
	for p := range e.Parents.All() {
		parents = append(parents, ImplicitlyMarshaledEntityUID(p))
	}
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

// UnmarshalJSON deserializes an Entity from JSON. Keys are case-sensitive, as in the Rust SDK.
func (e *Entity) UnmarshalJSON(b []byte) error {
	type entity Entity
	return jsonv2.Unmarshal(b, (*entity)(e), internal.JSONOptions)
}

func (e Entity) Equal(other Entity) bool {
	return e.UID.Equal(other.UID) &&
		e.Parents.Equal(other.Parents) &&
		e.Attributes.Equal(other.Attributes) &&
		e.Tags.Equal(other.Tags)
}
