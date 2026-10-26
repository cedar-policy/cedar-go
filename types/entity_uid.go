package types

import (
	"errors"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/go-json-experiment/json"

	"github.com/cedar-policy/cedar-go/internal/mapset"
	"github.com/cedar-policy/cedar-go/internal/names"
	"github.com/cedar-policy/cedar-go/internal/rust"
	"github.com/cedar-policy/cedar-go/internal/scanner"
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
func (e EntityUID) String() string {
	return string(e.Type) + "::\"" + rust.EscapeString(string(e.ID)) + "\""
}

// MarshalCedar produces a valid MarshalCedar language representation of the EntityUID, e.g. `Type::"id"`.
func (e EntityUID) MarshalCedar() []byte {
	return []byte(e.String())
}

var errInvalidUID = errors.New("invalid EntityUID")

// UnmarshalCedar parses a Cedar language representation of an EntityUID.
func (e *EntityUID) UnmarshalCedar(data []byte) error {
	// TODO: we use the scanner here but the code below it duplicates parser
	// code, eventually we'd like to use a single function.
	tokens, err := scanner.Tokenize(data)
	if err != nil || !tokens[0].IsIdent() {
		return errInvalidUID
	}
	path := []string{tokens[0].Text}
	for i := 1; i+1 < len(tokens); i += 2 {
		if tokens[i].Text != "::" {
			return errInvalidUID
		}
		switch t := tokens[i+1]; {
		case t.IsIdent():
			path = append(path, t.Text)
		case t.IsString() && tokens[i+2].IsEOF():
			id, err := t.StringValue()
			if err != nil {
				return errInvalidUID
			}
			*e = NewEntityUID(EntityType(strings.Join(path, "::")), String(id))
			return nil
		default:
			return errInvalidUID
		}
	}
	return errInvalidUID
}

// UnmarshalJSON decodes an EntityUID from its explicit form,
// {"__entity": {"type": ..., "id": ...}}, or its implicit form,
// {"type": ..., "id": ...}.
func (e *EntityUID) UnmarshalJSON(b []byte) error {
	var res entityValueJSON
	if err := json.Unmarshal(b, &res); err != nil {
		return err
	}
	var typ, id *string
	if res.Entity != nil {
		if res.Entity.Type == nil || res.Entity.ID == nil {
			return errJSONEntityNotFound
		}
		typ, id = res.Entity.Type, res.Entity.ID
	} else if res.Type != nil && res.ID != nil { // require both Type and ID to parse "implicit" JSON
		typ, id = res.Type, res.ID
	} else {
		return errJSONEntityNotFound
	}
	if !names.IsName(*typ) {
		return fmt.Errorf("%w: %q", errJSONInvalidType, *typ)
	}
	e.Type = EntityType(*typ)
	e.ID = String(*id)
	return nil
}

// MarshalJSON marshals the EntityUID into JSON using the explicit form.
func (e EntityUID) MarshalJSON() ([]byte, error) {
	typ, id := string(e.Type), string(e.ID)
	return json.Marshal(entityValueJSON{
		Entity: &extEntity{Type: &typ, ID: &id},
	})
}

func (e EntityUID) MarshalBinary() ([]byte, error) {
	return e.MarshalCedar(), nil
}

func (e *EntityUID) UnmarshalBinary(data []byte) error {
	return e.UnmarshalCedar(data)
}

func (e EntityUID) hash() uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(e.Type))
	_, _ = h.Write([]byte(e.ID))
	return h.Sum64()
}

// ImplicitlyMarshaledEntityUID exists to allow the marshaling of the EntityUID
// into JSON using the implicit form. Users can opt in to this form if they
// know that this EntityUID will be serialized to a place where its type will
// be unambiguous.
//
// ImplicitlyMarshaledEntityUID can be unmarshaled from either the implicit or
// explicit form.
type ImplicitlyMarshaledEntityUID EntityUID

func (i *ImplicitlyMarshaledEntityUID) UnmarshalJSON(b []byte) error {
	return (*EntityUID)(i).UnmarshalJSON(b)
}

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
