package ast_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/schema/parser"
	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestConvertHumanToJson(t *testing.T) {
	// Generate testdata/test_want.json by running:
	//  cedar translate-schema --direction cedar-to-json -s test.cedarschema | jq . > test_want.json
	// Note that as of cedar-policy-cli 4.4.1, the "required: true" attribute is omitted from the JSON output while
	// our JSON serialization always includes it. You'll have to add the expected "required: true" fields to
	// the emitted output.
	exampleHuman, err := fs.ReadFile(ast.Testdata, "testdata/convert/test.cedarschema")
	if err != nil {
		t.Fatalf("Error reading example schema: %v", err)
	}
	schema, err := parser.ParseFile("<test>", exampleHuman)
	if err != nil {
		t.Fatalf("Error parsing example schema: %v", err)
	}

	jsonSchema, err := ast.ConvertHuman2JSON(schema)
	if err != nil {
		t.Fatalf("Error in schema: %v", err)
	}
	var got bytes.Buffer
	enc := json.NewEncoder(&got)
	enc.SetIndent("", "    ")
	err = enc.Encode(jsonSchema)
	if err != nil {
		t.Fatalf("Error dumping JSON: %v", err)
	}

	want, err := fs.ReadFile(ast.Testdata, "testdata/convert/test_want.json")
	if err != nil {
		t.Fatalf("Error reading example JSON schema: %v", err)
	}
	var gotJ, wantJ interface{}
	testutil.OK(t, json.Unmarshal(want, &wantJ))
	testutil.OK(t, json.Unmarshal(got.Bytes(), &gotJ))
	diff := cmp.Diff(gotJ, wantJ)
	testutil.FatalIf(t, diff != "", "mismatch -want +got:\n%v", diff)
}

func TestConvertHumanToJson_NestedNamespace(t *testing.T) {
	namespace := &ast.Schema{
		Decls: []ast.Declaration{
			&ast.Namespace{
				Name: &ast.Path{Parts: []*ast.Ident{{Value: "hi"}}},
				Decls: []ast.Declaration{
					&ast.Namespace{
						Name: &ast.Path{Parts: []*ast.Ident{{Value: "hi"}}},
					},
				},
			},
		},
	}
	_, err := ast.ConvertHuman2JSON(namespace)
	if err == nil {
		t.Error("error should not be nil")
	}
	if !strings.Contains(err.Error(), "namespace") {
		t.Errorf("bad error %v", err)
	}
}
