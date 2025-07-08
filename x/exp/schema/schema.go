package schema

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
// Marshalling and unmarshalling between the formats is allowed.
type Schema struct {
	filename    string
	jsonSchema  ast.JsonSchema
	humanSchema *ast.Schema
}

// UnmarshalCedar parses and stores the human-readable schema from src and returns an error if the schema is invalid.
//
// Any errors returned will have file positions matching filename.
func (old *Schema) UnmarshalCedar(src []byte) (err error) {
	var s Schema
	s.humanSchema, err = parser.ParseFile(old.filename, src)
	if err != nil {
		return err
	}
	if old.filename != "" {
		s.filename = old.filename
	}
	*old = s
	return nil
}

// MarshalCedar serializes the schema into the human readable format.
func (s *Schema) MarshalCedar() ([]byte, error) {
	if s.jsonSchema != nil {
		s.humanSchema = ast.ConvertJSON2Human(s.jsonSchema)
	}
	if s.humanSchema == nil {
		return nil, fmt.Errorf("schema is empty")
	}
	var buf bytes.Buffer
	err := ast.Format(s.humanSchema, &buf)
	return buf.Bytes(), err
}

// UnmarshalJSON deserializes the JSON schema from src or returns an error if the JSON is not valid schema JSON.
func (old *Schema) UnmarshalJSON(src []byte) error {
	var s Schema
	err := json.Unmarshal(src, &s.jsonSchema)
	if err != nil {
		return err
	}
	s.filename = old.filename
	*old = s
	return nil
}

// MarshalJSON serializes the schema into the JSON format.
//
// If the schema was loaded from UnmarshalCedar, it will convert the human-readable format into the JSON format.
// An error is returned if the schema is invalid.
func (s *Schema) MarshalJSON() (out []byte, err error) {
	if s.humanSchema != nil {
		// Error should not be possible since s.humanSchema comes from our parser.
		// If it happens, we return empty JSON.
		s.jsonSchema = ast.ConvertHuman2Json(s.humanSchema)
	}
	if s.jsonSchema == nil {
		return nil, nil
	}
	return json.Marshal(s.jsonSchema)
}

// SetFilename sets the filename for the schema in the returned error messagers from Unmarshal*.
func (s *Schema) SetFilename(filename string) {
	s.filename = filename
}
