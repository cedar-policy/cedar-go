package parser

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/schema/token"
)

func TestLexer(t *testing.T) {
	tests := []string{
		"namespace Demo {}",
	}

	for _, test := range tests {
		lex := NewLexer("test", []byte(test))
		lex.All()
		next := lex.NextToken() // make sure we can call this as many times as we want and it will always return EOF
		if next.Type != token.EOF {
			t.Errorf("Expected EOF, got %v", next)
		}
		if len(lex.Errors) > 0 {
			t.Errorf("Errors: %v", lex.Errors)
		}
	}
}

func TestLexerExample(t *testing.T) {
	src := `namespace Demo {
  entity User {
    name: id,
  };
  // Comment
  type id = String;
}`
	lex := NewLexer("", []byte(src))
	tokens := lex.All()
	if len(lex.Errors) > 0 {
		t.Errorf("Errors: %v", lex.Errors)
	}

	want := []string{
		"NAMESPACE :1:1 namespace",
		"IDENT :1:11 Demo",
		"LEFTBRACE :1:16 {",
		"ENTITY :2:3 entity",
		"IDENT :2:10 User",
		"LEFTBRACE :2:15 {",
		"IDENT :3:5 name",
		"COLON :3:9 :",
		"IDENT :3:11 id",
		"COMMA :3:13 ,",
		"RIGHTBRACE :4:3 }",
		"SEMICOLON :4:4 ;",
		"COMMENT :5:3 // Comment",
		"TYPE :6:3 type",
		"IDENT :6:8 id",
		"EQUALS :6:11 =",
		"IDENT :6:13 String",
		"SEMICOLON :6:19 ;",
		"RIGHTBRACE :7:1 }",
	}
	var got []string
	for _, tok := range tokens {
		got = append(got, fmt.Sprintf("%s %s %s", tok.Type.String(), fmtpos(tok.Pos), tok.String()))
	}

	if !reflect.DeepEqual(got, want) {
		t.Logf("want: %v", strings.Join(want, "\n"))
		t.Logf(" got: %v", strings.Join(got, "\n"))
		t.Fail()
	}
}

func fmtpos(pos token.Position) string {
	return fmt.Sprintf("%s:%d:%d", pos.Filename, pos.Line, pos.Column)
}
