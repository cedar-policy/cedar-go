package ast_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestConvertJsonToHumanRoundtrip(t *testing.T) {
	// Read the example JSON schema from embedded filesystem
	exampleJSON, err := fs.ReadFile(ast.Testdata, "testdata/convert/test_want.json")
	if err != nil {
		t.Fatalf("Error reading example JSON schema: %v", err)
	}

	// Parse the JSON schema
	var jsonSchema ast.JSONSchema
	if err := json.Unmarshal(exampleJSON, &jsonSchema); err != nil {
		t.Fatalf("Error parsing JSON schema: %v", err)
	}

	// Convert to human-readable format and back to JSON
	humanSchema := ast.ConvertJSON2Human(jsonSchema)
	jsonSchema2, err := ast.ConvertHuman2JSON(humanSchema)
	if err != nil {
		t.Fatalf("Error dumping schema: %v", err)
	}

	// Compare the JSON schemas
	json1, err := json.MarshalIndent(jsonSchema, "", "    ")
	testutil.OK(t, err)

	json2, err := json.MarshalIndent(jsonSchema2, "", "    ")
	testutil.OK(t, err)

	diff := cmp.Diff(string(json1), string(json2))
	testutil.FatalIf(t, diff != "", "mismatch -want +got:\n%v", diff)
}

func TestConvertJsonToHumanEmpty(t *testing.T) {
	// Test with an empty JSON schema
	emptySchema := ast.JSONSchema{}
	humanSchema := ast.ConvertJSON2Human(emptySchema)

	// Format the human-readable schema
	var got bytes.Buffer
	if err := ast.Format(humanSchema, &got); err != nil {
		t.Fatalf("Error formatting schema: %v", err)
	}

	// Should be empty
	if len(got.Bytes()) != 0 {
		t.Errorf("Expected empty output, got: %q", got.String())
	}
}

func TestConvertJsonToHumanInvalidType(t *testing.T) {
	// Test with an invalid JSON type
	invalidSchema := ast.JSONSchema{
		"": {
			EntityTypes: map[string]*ast.JSONEntity{
				"Test": {
					Shape: &ast.JSONType{
						Type: "InvalidType",
					},
				},
			},
		},
	}

	var panicMsg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicMsg = fmt.Sprint(r)
			}
		}()
		ast.ConvertJSON2Human(invalidSchema)
	}()

	if panicMsg == "" {
		t.Fatal("expected panic, got none")
	}

	expected := "unknown JSON type: InvalidType"
	if !strings.Contains(panicMsg, expected) {
		t.Errorf("expected panic message to contain %q, got %q", expected, panicMsg)
	}
}

func TestConvertHuman2JSON_NestedNamespace(t *testing.T) {
	namePath := &ast.Path{Parts: []*ast.Ident{{Value: "hi"}}}
	innerNamespace := &ast.Namespace{
		Name: namePath,
	}
	outerNamespace := &ast.Namespace{
		Name: namePath,
		Decls: []ast.Declaration{
			innerNamespace,
		},
	}
	schema := &ast.Schema{
		Decls: []ast.Declaration{
			outerNamespace,
		},
	}

	_, err := ast.ConvertHuman2JSON(schema)
	if err == nil {
		t.Errorf("should have failed")
	}
}
