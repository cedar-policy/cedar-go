package types

import (
	"encoding/json"
	"hash/fnv"
	"strconv"
)

// Path is a series of idents separated by ::
type Path string

// EntityType is the type portion of an EntityUID
type EntityType Path

// An EntityUID is the identifier for a principal, action, or resource.
type EntityUID struct {
	Type EntityType
	ID   String
}

// NewEntityUID returns an EntityUID given an EntityType and identifier
func NewEntityUID(typ EntityType, id String) EntityUID {
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
func (v EntityUID) String() string { return string(v.Type) + "::" + strconv.Quote(string(v.ID)) }

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
		v.Type = EntityType(res.Entity.Type)
		v.ID = String(res.Entity.ID)
		return nil
	} else if res.Type != nil && res.ID != nil { // require both Type and ID to parse "implicit" JSON
		v.Type = EntityType(*res.Type)
		v.ID = String(*res.ID)
		return nil
	}
	return errJSONEntityNotFound
}

// ExplicitMarshalJSON marshals the EntityUID into JSON using the implicit form.
func (v EntityUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(entityValueJSON{
		Type: (*string)(&v.Type),
		ID:   (*string)(&v.ID),
	})
}

// ExplicitMarshalJSON marshals the EntityUID into JSON using the explicit form.
func (v EntityUID) ExplicitMarshalJSON() ([]byte, error) {
	return json.Marshal(entityValueJSON{
		Entity: &extEntity{
			Type: string(v.Type),
			ID:   string(v.ID),
		},
	})
}

func (v EntityUID) hash() uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(v.Type))
	_, _ = h.Write([]byte(v.ID))
	return h.Sum64()
}
