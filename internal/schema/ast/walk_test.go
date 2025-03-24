package ast_test

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/schema/parser"
)

func TestAstScope(t *testing.T) {
	// For each node, we verify that all of its children nodes are entirely within the Start and End of the functions.
	// This is a quick sanity test that we didn't implement any of our start or end positions incorrectly (or not set them!)
	tests := []string{
		"testdata/format/test.cedarschema",
		"testdata/format/emptynamespace.cedarschema",
		"testdata/format/nocomments.cedarschema",
		"testdata/walk/example.cedarschema",
		"testdata/walk/emptyfile.cedarschema",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			src, err := fs.ReadFile(ast.Testdata, test)
			if err != nil {
				t.Fatalf("Error reading test schema: %v", err)
			}
			schema, err := parser.ParseFile("<test>", []byte(src))
			if err != nil {
				t.Fatalf("Error parsing example schema: %v", err)
			}

			chain := []ast.Node{}
			Walk(schema, func(node ast.Node) bool {
				if node == nil {
					panic("node should not be nil")
				}
				if len(chain) > 0 {
					parent := chain[len(chain)-1]
					err := assertWithin(node, parent)
					if err != nil {
						chainstrs := make([]string, 0, len(chain)+1)
						for _, n := range chain {
							chainstrs = append(chainstrs, nodeid(n))
						}
						chainstrs = append(chainstrs, nodeid(node))

						t.Errorf("%s: %v", strings.Join(chainstrs, " -> "), err)
						return false
					}
				}
				chain = append(chain, node)
				return true
			}, func(ast.Node) bool {
				chain = chain[:len(chain)-1]
				return true
			})
		})
	}
}

func nodeid(n ast.Node) string {
	return fmt.Sprintf("%T(%d/%d:%d-%d/%d:%d)", n, n.Pos().Offset, n.Pos().Line, n.Pos().Column, n.End().Offset, n.End().Line, n.End().Column)
}

func assertWithin(node ast.Node, parent ast.Node) error {
	if node.Pos().Offset == 0 && node.End().Offset == 0 {
		return nil
	}

	if node.Pos().Offset == node.End().Offset {
		return fmt.Errorf("zero length node")
	}

	if node.Pos().Line == 0 {
		return fmt.Errorf("missing start position")
	}
	if node.Pos().Offset < parent.Pos().Offset {
		return fmt.Errorf("node start < parent start (%d < %d)", node.Pos().Offset, parent.Pos().Offset)
	}

	if node.End().Line == 0 {
		return fmt.Errorf("missing end position")
	}
	if node.End().Offset > parent.End().Offset {
		return fmt.Errorf("node end > parent end (%d > %d)", node.End().Offset, parent.End().Offset)
	}
	return nil
}

type visitor struct {
	stop bool
}

func Walk(n ast.Node, open, exit func(ast.Node) bool) {
	var v visitor
	v.walk(n, open, exit)
}

func (vis *visitor) walk(n ast.Node, open, exit func(ast.Node) bool) {
	if vis.stop {
		return
	}
	if n == nil || !open(n) {
		vis.stop = true
		return
	}
	defer func() {
		if n != nil {
			exit(n)
		}
	}()

	switch v := n.(type) {
	case *ast.Schema:
		for _, decl := range v.Decls {
			vis.walk(decl, open, exit)
		}
		vis.walk(v.Remaining, open, exit)
	case *ast.Namespace:
		vis.walk(v.Before, open, exit)
		vis.walk(v.Name, open, exit)
		if v.Inline != nil {
			vis.walk(v.Inline, open, exit)
		}
		for _, decl := range v.Decls {
			vis.walk(decl, open, exit)
		}
		vis.walk(v.Remaining, open, exit)
		if v.Footer != nil {
			vis.walk(v.Footer, open, exit)
		}
	case ast.CommentBlock:
		for _, c := range v {
			vis.walk(c, open, exit)
		}
	case *ast.CommonTypeDecl:
		vis.walk(v.Name, open, exit)
		vis.walk(v.Value, open, exit)
	case *ast.Entity:
		for _, name := range v.Names {
			vis.walk(name, open, exit)
		}
		for _, in := range v.In {
			vis.walk(in, open, exit)
		}
		if v.Shape != nil {
			vis.walk(v.Shape, open, exit)
		}
		if v.Tags != nil {
			vis.walk(v.Tags, open, exit)
		}
	case *ast.Action:
		for _, name := range v.Names {
			vis.walk(name, open, exit)
		}
		for _, in := range v.In {
			vis.walk(in, open, exit)
		}
		if v.AppliesTo != nil {
			vis.walk(v.AppliesTo, open, exit)
		}
	case *ast.AppliesTo:
		for _, p := range v.Principal {
			vis.walk(p, open, exit)
		}
		for _, r := range v.Resource {
			vis.walk(r, open, exit)
		}
		if v.Context != nil {
			vis.walk(v.Context, open, exit)
		}
	case *ast.RecordType:
		for _, attr := range v.Attributes {
			vis.walk(attr, open, exit)
		}
	case *ast.SetType:
		vis.walk(v.Element, open, exit)
	case *ast.Path:
		for _, part := range v.Parts {
			vis.walk(part, open, exit)
		}
	case *ast.Attribute:
		vis.walk(v.Key, open, exit)
		vis.walk(v.Type, open, exit)
	case *ast.Ref:
		for _, n := range v.Namespace {
			vis.walk(n, open, exit)
		}
		vis.walk(v.Name, open, exit)
	}
}
