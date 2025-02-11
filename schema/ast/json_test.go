package ast

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestParsesExampleSchema(t *testing.T) {
	exampleSchema, err := os.ReadFile("testdata/example_schema.json")
	if err != nil {
		t.Fatalf("Error reading example schema: %v", err)
	}

	var schema JsonSchema
	err = json.Unmarshal([]byte(exampleSchema), &schema)
	if err != nil {
		t.Fatalf("Error parsing schema: %v", err)
	}

	out, err := json.MarshalIndent(&schema, "", "    ")
	if err != nil {
		t.Fatalf("Error marshalling schema: %v", err)
	}
	if ok, err := jsonEq(exampleSchema, out); err != nil || !ok {
		t.Errorf("Schema does not match original:\n%s\n=========================================\n%s", exampleSchema, string(out))
	}
}

func jsonEq(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}
