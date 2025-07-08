package ast_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
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

	jsonSchema := ast.ConvertHuman2Json(schema)
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
