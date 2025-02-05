// Package schema is able to load, parse, format, and convert Cedar schemas in both JSON and human
// readable formats.
//
// The Cedar schema format like the policy format has a human-readable syntax, which can be parsed
// with Parse, and a JSON format, which can be loaded with standard JSON marshallers using the ast.Json* types.
//
// The package considers the JSON format the canonical format, but the AST for the human-readable format is still available
// in the schema/ast package to allow users for introspection and transformation.
//
// Once the schema is loaded, it can be converted to JSON from the human readable using the
// ast.Convert function, or formatted out in a pretty-printed format using ast.Format. To dump the JSON
// format use a standard JSON marshaller on the ast.Json* structs.
package schema

import (
	"io"
	"strings"

	"github.com/cedar-policy/cedar-go/schema/ast"
	"github.com/cedar-policy/cedar-go/schema/internal/convert"
	"github.com/cedar-policy/cedar-go/schema/internal/format"
	"github.com/cedar-policy/cedar-go/schema/internal/parser"
)

// ParseOptions allow customization of the parsing process.
type ParseOptions struct {
	// Intentionally empty for future use
}

// ParseFile parses a human-readable Cedar schema and returns the AST along with any errors.
//
// You can pass optional parameters to change how the parsing is done. If nil, the default options
// are used.
//
// If there are errors, the parser will still attempt to continue and return any errors it finds.
// All errors are of type []token.Error and contain position information.
func ParseFile(filename string, src io.Reader, options *ParseOptions) (*ast.Schema, error) {
	if options == nil {
		options = &ParseOptions{}
	}
	contents, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return parser.ParseFile(filename, contents, nil)
}

// Parse parses the human-readable schema in src and returns the schema it represents if valid.
func Parse(src string) (*ast.Schema, error) {
	return ParseFile("<input>", strings.NewReader(src), nil)
}

type FormatOptions = format.Options

// Format will pretty-print the AST node to out using the options provided.
func Format(node ast.Node, out io.Writer, opts *FormatOptions) error {
	return format.Node(node, out, opts)
}

// Convert will convert the AST schema to a JSON schema. The conversion process loses all comments and formatting info.
func Convert(schema *ast.Schema) (ast.JsonSchema, error) {
	return convert.Schema(schema)
}
