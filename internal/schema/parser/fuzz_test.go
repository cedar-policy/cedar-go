package parser

import (
	"io/fs"
	"testing"
)

// Make sure that we aren't panicing or looping forever for any byte sequences
// that are passed to the parser.
// Run with:
//
//	go test -fuzz=FuzzParseSchema -fuzztime=60s github.com/cedar-policy/cedar-go/schema/internal/parser
func FuzzParseSchema(f *testing.F) {
	f.Add([]byte("namespace Demo {}"))
	f.Add([]byte("namespace D A0!00000000000000\"0"))
	f.Add([]byte("namespace 0A0 action \"\" appliesTo 0 principal!0//0000"))
	f.Add([]byte("namespace Demo { action Test, Test2; entity Test { id: String } }"))
	f.Add(read("testdata/cases/example.cedarschema"))
	f.Fuzz(func(t *testing.T, data []byte) {
		schema, _ := ParseFile("<fuzz>", data)
		if schema == nil {
			t.Fatalf("Schema should never be nil")
		}
	})
}

func read(file string) []byte {
	contents, err := fs.ReadFile(Testdata, file)
	if err != nil {
		panic(err)
	}
	return contents
}
