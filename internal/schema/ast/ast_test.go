package ast

// Tests in this file are for reaching 100% coverage of the ast package. These code paths should normally not be executed,
// but are forcefully exercised here to ensure that they are implemented correctly.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/schema/token"
)

func TestIsNode(*testing.T) {
	// Test all isNode() implementations
	(&Schema{}).isNode()
	(&Namespace{}).isNode()
	(&CommonTypeDecl{}).isNode()
	(&RecordType{}).isNode()
	(&SetType{}).isNode()
	(&Path{}).isNode()
	(&Ident{}).isNode()
	(&Entity{}).isNode()
	(&Action{}).isNode()
	(&AppliesTo{}).isNode()
	(&Ref{}).isNode()
	(&Attribute{}).isNode()
	(&String{}).isNode()
	(CommentBlock{}).isNode()
	(&Comment{}).isNode()

	// No assertions needed since we just want coverage for these marker methods
}

func TestIsDecl(*testing.T) {
	// Test all isDecl() implementations
	(&Entity{}).isDecl()
	(&Action{}).isDecl()
	(&Namespace{}).isDecl()
	(&CommonTypeDecl{}).isDecl()
	(&CommentBlock{}).isDecl()

	// No assertions needed since we just want coverage for these marker methods
}

func TestIsType(*testing.T) {
	// Test all isType() implementations
	(&RecordType{}).isType()
	(&SetType{}).isType()
	(&Path{}).isType()

	// No assertions needed since we just want coverage for these marker methods
}

func TestIsName(*testing.T) {
	// Test all isName() implementations
	(&String{}).isName()
	(&Ident{}).isName()

	// No assertions needed since we just want coverage for these marker methods
}

func TestPathEmptyParts(t *testing.T) {
	p := &Path{Parts: nil}

	// Test Pos() with empty Parts
	pos := p.Pos()
	if pos != (token.Position{}) {
		t.Errorf("Expected empty Position for Pos(), got %v", pos)
	}

	// Test End() with empty Parts
	end := p.End()
	if end != (token.Position{}) {
		t.Errorf("Expected empty Position for End(), got %v", end)
	}
}

func TestSchemaEmpty(t *testing.T) {
	s := &Schema{}

	// Test Pos() with empty Schema
	pos := s.Pos()
	if pos != (token.Position{}) {
		t.Errorf("Expected empty Position for Pos(), got %v", pos)
	}

	// Test End() with empty Schema
	end := s.End()
	if end != (token.Position{}) {
		t.Errorf("Expected empty Position for End(), got %v", end)
	}
}

func Test_formatter_printInd_panic(t *testing.T) {
	p := &formatter{
		w:        &bytes.Buffer{},
		lastchar: 'x', // Not a newline
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string, got %T", r)
		}
		if !strings.Contains(msg, "lastchar must be newline") {
			t.Errorf("expected panic message about newline, got %q", msg)
		}
	}()

	p.printInd("test")
}

type unknownNode struct {
	Node // Embed Node interface to satisfy type checker
}

func Test_formatter_print_panic(t *testing.T) {
	p := &formatter{
		w: &bytes.Buffer{},
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !strings.Contains(msg, "unhandled node type") {
			t.Errorf("expected panic message about unhandled type, got %q", msg)
		}
	}()

	p.print(unknownNode{})
}

func Test_printBracketList_panic(t *testing.T) {
	p := &formatter{
		w: &bytes.Buffer{},
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !strings.Contains(msg, "list must not be empty") {
			t.Errorf("expected panic message about empty list, got %q", msg)
		}
	}()

	var emptyList []Node
	printBracketList(p, emptyList)
}

type unknownType struct {
	Type // Embed Type interface to satisfy it
}

func TestConvertType_Panic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		expected := "unknownType is not an AST type"
		if !strings.Contains(msg, expected) {
			t.Errorf("expected panic message to contain %q, got %q", expected, msg)
		}
	}()

	convertType(unknownType{})
}
