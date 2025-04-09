package parser

import (
	"fmt"
	"io/fs"
	"reflect"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/schema/token"
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
    "name\0\n\r\t\"\'_\u{1}\u{001f}": id,
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
		"STRING :3:5 \"name\x00\n\r\t\"'_\x01\x00\x1f\"",
		"COLON :3:37 :",
		"IDENT :3:39 id",
		"COMMA :3:41 ,",
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

func TestLexerOk(t *testing.T) {
	files, err := fs.ReadDir(Testdata, "testdata/lex")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}
	for _, file := range files {
		t.Run(file.Name(), func(t *testing.T) {
			data, err := fs.ReadFile(Testdata, "testdata/lex/"+file.Name())
			if err != nil {
				t.Fatalf("Failed to read test data: %v", err)
			}
			l := NewLexer("<test>", data)
			l.All()
			if len(l.Errors) > 0 {
				t.Errorf("Errors: %v", l.Errors)
			}
		})
	}
}

func TestLexerNoPanic(t *testing.T) {
	cases := []string{
		// Basic tokens
		"{}[]<>?=,;::",
		// Keywords
		"action context entity type namespace principal resource tags in applies appliesTo",
		// Identifiers with various characters
		"abc ABC _123 A_b_C",
		// Strings with escape sequences
		`"hello\"world"`,
		`"escape sequences: \\ \' \? \a \b \f \n \r \t \v"`,
		// Comments
		"// this is a comment\n",
		"// comment with special chars: !@#$%^&*()\n",
		// Whitespace handling
		"  \t  \n\r\n",
		// Weird identifiers
		"____azxkljcqmoqiwerjqflkjazxklmzlkmdrfoiwqerjlakdsfsazljfdi",
		// Edge cases
		"\r\n",           // Carriage return + newline
		"\"unterminated", // Unterminated string
		"\\",             // Single backslash
		"@",              // Invalid character
		"\"\\z\"",        // Invalid escape sequence
		"\"\\\"",         // Unterminated escaped string
		`"simple string"`,
		`"string with \"escaped\" quotes"`,
		`"string with escape sequences \\ \' \? \a \b \f \n \r \t \v"`,
		`"string with
newline"`, // Invalid - should error
		`"unterminated`,       // Invalid - should error
		`"invalid escape \z"`, // Invalid escape sequence
	}

	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			// Recover from any panics to ensure test continues
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Lexer panicked on input: %q\nPanic: %v", input, r)
				}
			}()

			l := NewLexer("<test>", []byte(input))
			for {
				_, tok, _, _ := l.lex()
				if tok == token.EOF { // EOF
					break
				}
			}
		})
	}
}

func FuzzLexer(f *testing.F) {
	// Add some initial seeds
	seeds := []string{
		"namespace Demo {}",
		"\"hello\\\"world\"",
		"// comment\n",
		"\r\n",
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(_ *testing.T, data []byte) {
		l := NewLexer("<fuzz>", data)
		for {
			_, tok, _, _ := l.lex()
			if tok == token.EOF {
				break
			}
		}
	})
}
