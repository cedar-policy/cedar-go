package cedar

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/schema/parser"
)

// Schema is a description of entities and actions that are allowed for a PolicySet. They can be used to validate policies
// and entity definitions and also provide documentation.
//
// Schemas can be represented in either JSON (*JSON functions) or Human-readable formats (*Cedar functions) just like policies.
// Marshalling and unmarshalling between the formats is allowed, except from a JSON schema into a human-readable one.
type Schema struct {
	jsonSchema  ast.JsonSchema
	humanSchema *ast.Schema
}

// NewSchema creates an empty schema.
func NewSchema() *Schema {
	return &Schema{}
}

// UnmarshalCedar parses and stores the human-readable schema from src and returns an error if the schema is invalid.
//
// Any errors returned will have file positions matching filename.
func (s *Schema) UnmarshalCedar(filename string, src []byte) (err error) {
	s.humanSchema, err = parser.ParseFile(filename, src, nil)
	return
}

// MarshalCedar serializes the schema into the human readable format. This function can only be called on schemas
// that are initialized with UnmarshalCedar, not with UnmarshalJSON.
func (s *Schema) MarshalCedar() ([]byte, error) {
	if s.jsonSchema != nil {
		return nil, fmt.Errorf("cannot call MarshalCedar after UnmarshalJSON")
	}
	if s.humanSchema == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	err := ast.Format(s.humanSchema, &buf)
	return buf.Bytes(), err
}

// UnmarshalJSON deserializes the JSON schema from src or returns an error if the JSON is not valid schema JSON.
func (s *Schema) UnmarshalJSON(src []byte) error {
	return json.Unmarshal(src, &s.jsonSchema)
}

// MarshalJSON serializes the schema into the JSON format.
//
// If the schema was loaded from UnmarshalCedar, it will convert the human-readable format into the JSON format.
// An error is returned if the schema is invalid.
func (s *Schema) MarshalJSON() (out []byte, err error) {
	if s.humanSchema != nil {
		// Error should not be possible since s.humanSchema comes from our parser.
		// If it happens, we return empty JSON.
		s.jsonSchema, _ = ast.Convert(s.humanSchema)
	}
	if s.jsonSchema == nil {
		return nil, nil
	}
	return json.Marshal(s.jsonSchema)
}
