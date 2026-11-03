package types

import (
	"fmt"
	"slices"
	"strings"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

var errJSONInvalidEntity = fmt.Errorf("invalid entity")

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

// entityJSON is the JSON form of an Entity. Its members are pointers so that
// missing and null members can be detected. Tags is kept as raw JSON because
// it's optional but, unlike a missing member, may not be null.
type entityJSON struct {
	UID        *EntityUID     `json:"uid"`
	Parents    *EntityUIDSet  `json:"parents"`
	Attributes *Record        `json:"attrs"`
	Tags       jsontext.Value `json:"tags"`
}

// UnmarshalJSON parses an Entity from JSON. As in Rust Cedar, "uid", "parents",
// and "attrs" are required, "tags" is optional, none of them may be null, and
// unknown members are ignored.
func (e *Entity) UnmarshalJSON(b []byte) error {
	var ej entityJSON
	if err := json.Unmarshal(b, &ej); err != nil {
		return err
	}
	switch {
	case ej.UID == nil:
		return fmt.Errorf(`%w: "uid" is missing or null`, errJSONInvalidEntity)
	case ej.Parents == nil:
		return fmt.Errorf(`%w: "parents" is missing or null`, errJSONInvalidEntity)
	case ej.Attributes == nil:
		return fmt.Errorf(`%w: "attrs" is missing or null`, errJSONInvalidEntity)
	case ej.Tags.Kind() == jsontext.KindNull:
		return fmt.Errorf(`%w: "tags" is null`, errJSONInvalidEntity)
	}

	res := Entity{UID: *ej.UID, Parents: *ej.Parents, Attributes: *ej.Attributes}
	if ej.Tags != nil {
		if err := json.Unmarshal(ej.Tags, &res.Tags); err != nil {
			return err
		}
	}
	*e = res
	return nil
}

func (e Entity) Equal(other Entity) bool {
	return e.UID.Equal(other.UID) &&
		e.Parents.Equal(other.Parents) &&
		e.Attributes.Equal(other.Attributes) &&
		e.Tags.Equal(other.Tags)
}
