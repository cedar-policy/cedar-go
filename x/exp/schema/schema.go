// Package schema provides schema parsing, serialization, and resolution.
package schema

import (
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema/internal/json"
	"github.com/cedar-policy/cedar-go/x/exp/schema/internal/parser"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// Schema provides parsing and marshaling for Cedar schemas.
type Schema struct {
	filename string
	schema   *ast.Schema
}

// NewSchemaFromAST creates a Schema from an AST.
func NewSchemaFromAST(in *ast.Schema) *Schema {
	return &Schema{schema: in}
}

// SetFilename sets the filename for error reporting.
func (s *Schema) SetFilename(filename string) {
	s.filename = filename
}

// MarshalJSON encodes the Schema in the JSON format.
func (s *Schema) MarshalJSON() ([]byte, error) {
	jsonSchema := (*json.Schema)(s.schema)
	return jsonSchema.MarshalJSON()
}

// UnmarshalJSON parses a Schema in the JSON format.
func (s *Schema) UnmarshalJSON(b []byte) error {
	var jsonSchema json.Schema
	if err := jsonSchema.UnmarshalJSON(b); err != nil {
		return err
	}
	s.schema = (*ast.Schema)(&jsonSchema)
	return nil
}

// MarshalCedar encodes the Schema in the human-readable format.
func (s *Schema) MarshalCedar() ([]byte, error) {
	return parser.MarshalSchema(s.schema), nil
}

// UnmarshalCedar parses a Schema in the human-readable format.
func (s *Schema) UnmarshalCedar(b []byte) error {
	schema, err := parser.ParseSchema(s.filename, b)
	if err != nil {
		return err
	}
	s.schema = schema
	return nil
}

// AST returns the underlying AST.
func (s *Schema) AST() *ast.Schema {
	return s.schema
}

// Resolve returns a resolved.Schema with type references resolved and declarations indexed.
func (s *Schema) Resolve() (*resolved.Schema, error) {
	return resolved.Resolve(s.schema)
}
