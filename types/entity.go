package types

import "encoding/json"

// An Entity defines the parents and attributes for an EntityUID.
type Entity struct {
	UID        EntityUID    `json:"uid"`
	Parents    EntityUIDSet `json:"parents"`
	Attributes Record       `json:"attrs"`
}

// MarshalJSON serializes Entity as a JSON object, using the implicit form of EntityUID encoding to match the Rust
// SDK's behavior.
func (e Entity) MarshalJSON() ([]byte, error) {
	parents := make([]ImplicitlyMarshaledEntityUID, 0, e.Parents.Len())
	e.Parents.Iterate(func(p EntityUID) bool {
		parents = append(parents, ImplicitlyMarshaledEntityUID(p))
		return true
	})

	m := struct {
		UID        ImplicitlyMarshaledEntityUID   `json:"uid"`
		Parents    []ImplicitlyMarshaledEntityUID `json:"parents"`
		Attributes Record                         `json:"attrs"`
	}{
		ImplicitlyMarshaledEntityUID(e.UID),
		parents,
		e.Attributes,
	}
	return json.Marshal(m)
}
