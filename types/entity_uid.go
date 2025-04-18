package types

import (
	"encoding/json"
	"hash/fnv"
	"strconv"

	"github.com/cedar-policy/cedar-go/internal/mapset"
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
func (e EntityUID) IsZero() bool {
	return e.Type == "" && e.ID == ""
}

func (e EntityUID) Equal(bi Value) bool {
	b, ok := bi.(EntityUID)
	return ok && e == b
}

// String produces a string representation of the EntityUID, e.g. `Type::"id"`.
func (e EntityUID) String() string { return string(e.Type) + "::" + strconv.Quote(string(e.ID)) }

// MarshalCedar produces a valid MarshalCedar language representation of the EntityUID, e.g. `Type::"id"`.
func (e EntityUID) MarshalCedar() []byte {
	return []byte(e.String())
}

func (e *EntityUID) UnmarshalJSON(b []byte) error {
	// TODO: review after adding support for schemas
	var res entityValueJSON
	if err := json.Unmarshal(b, &res); err != nil {
		return err
	}
	if res.Entity != nil {
		e.Type = EntityType(res.Entity.Type)
		e.ID = String(res.Entity.ID)
		return nil
	} else if res.Type != nil && res.ID != nil { // require both Type and ID to parse "implicit" JSON
		e.Type = EntityType(*res.Type)
		e.ID = String(*res.ID)
		return nil
	}
	return errJSONEntityNotFound
}

// MarshalJSON marshals the EntityUID into JSON using the explicit form.
func (e EntityUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(entityValueJSON{
		Entity: &extEntity{
			Type: string(e.Type),
			ID:   string(e.ID),
		},
	})
}

func (e EntityUID) hash() uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(e.Type))
	_, _ = h.Write([]byte(e.ID))
	return h.Sum64()
}

// ImplicitlyMarshaledEntityUID exists to allow the marshaling of the EntityUID into JSON using the implicit form. Users
// can opt in to this form if they know that this EntityUID will be serialized to a place where its type will be
// unambiguous.
type ImplicitlyMarshaledEntityUID EntityUID

func (i ImplicitlyMarshaledEntityUID) MarshalJSON() ([]byte, error) {
	s := struct {
		Type EntityType `json:"type"`
		ID   String     `json:"id"`
	}{i.Type, i.ID}
	return json.Marshal(s)
}

type EntityUIDSet = mapset.ImmutableMapSet[EntityUID]

// NewEntityUIDSet returns an immutable EntityUIDSet ready for use.
func NewEntityUIDSet(args ...EntityUID) EntityUIDSet {
	return mapset.Immutable[EntityUID](args...)
}
