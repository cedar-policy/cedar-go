// Needs to be in a dedicated test package to avoid circular dependency with format
package parser_test

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/schema/parser"
	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestParseSimple(t *testing.T) {
	tests := []string{
		// Empty namespace
		`namespace Demo {
}
`,
		// Simple namespace with single entity
		`namespace Demo {
  entity User in UserGroup = {
    name: Demo::id,
    "department": UserGroup,
  };
}
`,
		// Anonymous namespace references
		`@annotation("entity")
// Entity example
entity User;
@annotation("type")
// Type example
type Id = String;
@annotation("action")
// Action example
action run;
namespace NS {
  // empty
} // footer
`,
	}

	for _, test := range tests {
		schema, err := parser.ParseFile("<test>", []byte(test))
		if err != nil {
			t.Fatalf("Error parsing schema: %v", err)
		}
		var got bytes.Buffer
		err = ast.Format(schema, &got) // tab format to match Go
		testutil.OK(t, err)
		diff := cmp.Diff(got.String(), test)
		testutil.FatalIf(t, diff != "", "mismatch -want +got:\n%v", diff)
	}
}

func TestParserHasErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "missing closing bracket",
			input: `namespace PhotoFlash {`,
			want:  `<test>:1:23: expected }, got EOF`,
		},
		{
			name:  "missing entity name",
			input: `namespace PhotoFlash { entity { "department": String }; }`,
			want:  `<test>:1:31: expected identifier, got {`,
		},
	}

	for _, test := range tests {
		_, err := parser.ParseFile("<test>", []byte(test.input))
		if err == nil {
			t.Fatalf("Expected error parsing schema, got none")
		}
		if err.Error() != test.want {
			t.Errorf("Expected error %q, got %q", test.want, err.Error())
		}
	}
}

func TestRealFiles(t *testing.T) {
	files, err := fs.ReadDir(parser.Testdata, "testdata/cases")
	if err != nil {
		t.Fatalf("Error reading testdata: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".cedarschema") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			input, err := fs.ReadFile(parser.Testdata, "testdata/cases/"+file.Name())
			if err != nil {
				t.Fatalf("Error reading example schema: %v", err)
			}
			schema, err := parser.ParseFile("<test>", input)
			if err != nil {
				t.Fatalf("Error parsing schema: %v", err)
			}

			var gotBytes bytes.Buffer
			err = ast.Format(schema, &gotBytes)
			testutil.OK(t, err)

			got := strings.TrimSpace(gotBytes.String())
			testutil.OK(t, err)
			diff := cmp.Diff(got, strings.TrimSpace(string(input)))
			testutil.FatalIf(t, diff != "", "mismatch -want +got:\n%v", diff)
		})
	}
}
