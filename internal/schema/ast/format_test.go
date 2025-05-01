package ast_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/schema/parser"
)

// Source will pretty-print src in the returned byte slice. If src is malformed Cedar schema, an error will be returned.
func Source(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	tree, err := parser.ParseFile("<input>", src)
	if err != nil {
		return nil, err
	}

	err = ast.Format(tree, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestFormatExamples(t *testing.T) {
	tests := []struct {
		file string
	}{
		{file: "testdata/format/nocomments.cedarschema"},
		{file: "testdata/format/test.cedarschema"},
		{file: "testdata/format/emptynamespace.cedarschema"},
		{file: "testdata/walk/emptyfile.cedarschema"},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			example, err := fs.ReadFile(ast.Testdata, tt.file)
			if err != nil {
				t.Fatalf("open testfile %s: %v", tt.file, err)
			}

			got, err := Source(example)
			if err != nil {
				t.Fatalf("formatting error: %v", err)
			}
			if string(got) != string(example) {
				t.Errorf("Parsed schema does not match original:\n%s\n=========================================\n%s\n=========================================", example, string(got))
			}
		})
	}
}

func TestFormatEmpty(t *testing.T) {
	got, err := Source([]byte(""))
	if err != nil {
		t.Fatalf("formatting empty file failed: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty output, got: %q", string(got))
	}
}

type errorWriter struct{}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("intentional write error")
}

func TestFormat_WriterError(t *testing.T) {
	schema := &ast.Schema{
		Decls: []ast.Declaration{
			&ast.CommonTypeDecl{
				Name:  &ast.Ident{Value: "Test"},
				Value: &ast.Path{Parts: []*ast.Ident{{Value: "Test"}}},
			},
		},
	}

	err := ast.Format(schema, errorWriter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "intentional write error") {
		t.Errorf("expected error to contain 'intentional write error', got %v", err)
	}
}
