package cedar

import (
	"bytes"
	"encoding/json"

	"github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/schema/parser"
)

// Schema is a description of entities and actions that are allowed for a PolicySet. They can be used to validate policies
// and entity definitions and also provide documentation.
//
// Schemas can be represented in either JSON (*JSON functions) or Human-readable formats (*Cedar functions) just like policies.
// Marshalling and unmarshalling between the formats is allowed, except from a JSON schema into a human-readable one.
type Schema struct {
	filename    string
	jsonSchema  ast.JsonSchema
	humanSchema *ast.Schema
}

// UnmarshalCedar parses and stores the human-readable schema from src and returns an error if the schema is invalid.
//
// Any errors returned will have file positions matching filename.
func (old *Schema) UnmarshalCedar(src []byte) (err error) {
	var new Schema
	new.humanSchema, err = parser.ParseFile(old.filename, src, nil)
	if err != nil {
		return err
	}
	*old = new
	return nil
}

// MarshalCedar serializes the schema into the human readable format. This function can only be called on schemas
// that are initialized with UnmarshalCedar, not with UnmarshalJSON.
func (s *Schema) MarshalCedar() ([]byte, error) {
	if s.jsonSchema != nil {
		s.humanSchema = ast.ConvertJSON2Human(s.jsonSchema)
	}
	if s.humanSchema == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	err := ast.Format(s.humanSchema, &buf)
	return buf.Bytes(), err
}

// UnmarshalJSON deserializes the JSON schema from src or returns an error if the JSON is not valid schema JSON.
func (old *Schema) UnmarshalJSON(src []byte) error {
	var new Schema
	err := json.Unmarshal(src, &new.jsonSchema)
	if err != nil {
		return err
	}
	new.filename = old.filename
	*old = new
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
