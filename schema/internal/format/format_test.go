package format

import (
	"os"
	"testing"
)

func TestFormatExamples(t *testing.T) {
	example, err := os.ReadFile("testdata/format_test.cedarschema")
	if err != nil {
		t.Fatalf("open testfile: %v", err)
	}

	got, err := Source(example, &Options{Indent: "    "})
	if err != nil {
		t.Fatalf("formatting error: %v", err)
	}
	if string(got) != string(example) {
		t.Errorf("Parsed schema does not match original:\n%s\n=========================================\n%s\n=========================================", example, string(got))
	}
}
