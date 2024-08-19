package types

import (
	"encoding/json"
	"strconv"
)

// An EntityUID is the identifier for a principal, action, or resource.
type EntityUID struct {
	Type EntityType
	ID   string
}

func NewEntityUID(typ EntityType, id string) EntityUID {
	return EntityUID{
		Type: typ,
		ID:   id,
	}
}

// IsZero returns true if the EntityUID has an empty Type and ID.
func (a EntityUID) IsZero() bool {
	return a.Type == "" && a.ID == ""
}

func (a EntityUID) Equal(bi Value) bool {
	b, ok := bi.(EntityUID)
	return ok && a == b
}

// String produces a string representation of the EntityUID, e.g. `Type::"id"`.
func (v EntityUID) String() string { return string(v.Type.String() + "::" + strconv.Quote(v.ID)) }

// MarshalCedar produces a valid MarshalCedar language representation of the EntityUID, e.g. `Type::"id"`.
func (v EntityUID) MarshalCedar() []byte {
	return []byte(v.String())
}

func (v *EntityUID) UnmarshalJSON(b []byte) error {
	// TODO: review after adding support for schemas
	var res entityValueJSON
	if err := json.Unmarshal(b, &res); err != nil {
		return err
	}
	if res.Entity != nil {
		v.Type = res.Entity.Type
		v.ID = res.Entity.ID
		return nil
	} else if res.Type != nil && res.ID != nil { // require both Type and ID to parse "implicit" JSON
		v.Type = *res.Type
		v.ID = *res.ID
		return nil
	}
	return errJSONEntityNotFound
}

// ExplicitMarshalJSON marshals the EntityUID into JSON using the implicit form.
func (v EntityUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(entityValueJSON{
		Type: &v.Type,
		ID:   &v.ID,
	})
}

// ExplicitMarshalJSON marshals the EntityUID into JSON using the explicit form.
func (v EntityUID) ExplicitMarshalJSON() ([]byte, error) {
	return json.Marshal(entityValueJSON{
		Entity: &extEntity{
			Type: v.Type,
			ID:   v.ID,
		},
	})
}
func (v EntityUID) deepClone() Value { return v }
