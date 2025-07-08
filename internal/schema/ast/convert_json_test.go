package ast_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestConvertJsonToHumanRoundtrip(t *testing.T) {
	// Read the example JSON schema from embedded filesystem
	exampleJson, err := fs.ReadFile(ast.Testdata, "testdata/convert/test_want.json")
	if err != nil {
		t.Fatalf("Error reading example JSON schema: %v", err)
	}

	// Parse the JSON schema
	var jsonSchema ast.JsonSchema
	if err := json.Unmarshal(exampleJson, &jsonSchema); err != nil {
		t.Fatalf("Error parsing JSON schema: %v", err)
	}

	// Convert to human-readable format and back to JSON
	humanSchema := ast.ConvertJSON2Human(jsonSchema)
	jsonSchema2 := ast.ConvertHuman2Json(humanSchema)

	// Compare the JSON schemas
	json1, err := json.MarshalIndent(jsonSchema, "", "    ")
	testutil.OK(t, err)

	json2, err := json.MarshalIndent(jsonSchema2, "", "    ")
	testutil.OK(t, err)

	testutil.Equals(t, string(json1), string(json2))
}

func TestConvertJsonToHumanEmpty(t *testing.T) {
	// Test with an empty JSON schema
	emptySchema := ast.JsonSchema{}
	humanSchema := ast.ConvertJSON2Human(emptySchema)

	// Format the human-readable schema
	var got bytes.Buffer
	if err := ast.Format(humanSchema, &got); err != nil {
		t.Fatalf("Error formatting schema: %v", err)
	}

	// Should be empty
	if len(got.Bytes()) != 0 {
		t.Errorf("Expected empty output, got: %q", string(got.Bytes()))
	}
}

func TestConvertJsonToHumanInvalidType(t *testing.T) {
	// Test with an invalid JSON type
	invalidSchema := ast.JsonSchema{
		"": {
			EntityTypes: map[string]*ast.JsonEntity{
				"Test": {
					Shape: &ast.JsonType{
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
