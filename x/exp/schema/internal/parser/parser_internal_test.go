package parser

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestLexerBasicTokens(t *testing.T) {
	src := `@{}<>[](),;:?=::`
	l := newLexer("", []byte(src))
	expected := []tokenType{
		tokenAt, tokenLBrace, tokenRBrace,
		tokenLAngle, tokenRAngle, tokenLBracket, tokenRBracket,
		tokenLParen, tokenRParen, tokenComma, tokenSemicolon,
		tokenColon, tokenQuestion, tokenEquals, tokenDoubleColon, tokenEOF,
	}
	for _, tt := range expected {
		tok, err := l.next()
		testutil.OK(t, err)
		testutil.Equals(t, tok.Type, tt)
	}
}

func TestLexerStringEscapes(t *testing.T) {
	src := `"hello\nworld"`
	l := newLexer("", []byte(src))
	tok, err := l.next()
	testutil.OK(t, err)
	testutil.Equals(t, tok.Type, tokenString)
	testutil.Equals(t, tok.Text, "hello\nworld")
}

func TestLexerUnterminatedString(t *testing.T) {
	src := `"hello`
	l := newLexer("", []byte(src))
	_, err := l.next()
	testutil.Error(t, err)
}

func TestLexerUnterminatedStringNewline(t *testing.T) {
	src := "\"hello\nworld\""
	l := newLexer("", []byte(src))
	_, err := l.next()
	testutil.Error(t, err)
}

func TestLexerUnterminatedStringBackslash(t *testing.T) {
	src := `"hello\`
	l := newLexer("", []byte(src))
	_, err := l.next()
	testutil.Error(t, err)
}

func TestLexerUnexpectedChar(t *testing.T) {
	src := `$`
	l := newLexer("", []byte(src))
	_, err := l.next()
	testutil.Error(t, err)
}

func TestLexerLineComment(t *testing.T) {
	src := "// comment\nfoo"
	l := newLexer("", []byte(src))
	tok, err := l.next()
	testutil.OK(t, err)
	testutil.Equals(t, tok.Type, tokenIdent)
	testutil.Equals(t, tok.Text, "foo")
}

func TestLexerBlockComment(t *testing.T) {
	src := "/* block */foo"
	l := newLexer("", []byte(src))
	tok, err := l.next()
	testutil.OK(t, err)
	testutil.Equals(t, tok.Type, tokenIdent)
	testutil.Equals(t, tok.Text, "foo")
}

func TestLexerUnterminatedBlockComment(t *testing.T) {
	src := "/* unterminated"
	l := newLexer("", []byte(src))
	_, err := l.next()
	testutil.Error(t, err)
}

func TestLexerPosition(t *testing.T) {
	src := "foo\nbar"
	l := newLexer("test.cedar", []byte(src))
	tok, err := l.next()
	testutil.OK(t, err)
	testutil.Equals(t, tok.Pos.Line, 1)
	testutil.Equals(t, tok.Pos.Column, 1)
	testutil.Equals(t, tok.Pos.Filename, "test.cedar")

	tok, err = l.next()
	testutil.OK(t, err)
	testutil.Equals(t, tok.Pos.Line, 2)
	testutil.Equals(t, tok.Pos.Column, 1)
}

func TestLexerEOF(t *testing.T) {
	l := newLexer("", []byte(""))
	tok, err := l.next()
	testutil.OK(t, err)
	testutil.Equals(t, tok.Type, tokenEOF)
}

func TestPositionString(t *testing.T) {
	p := position{Line: 1, Column: 5}
	testutil.Equals(t, p.String(), "<input>:1:5")

	p.Filename = "test.cedarschema"
	testutil.Equals(t, p.String(), "test.cedarschema:1:5")
}

func TestTokenName(t *testing.T) {
	tests := []struct {
		tt   tokenType
		want string
	}{
		{tokenEOF, "EOF"},
		{tokenIdent, "identifier"},
		{tokenString, "string"},
		{tokenAt, "'@'"},
		{tokenLBrace, "'{'"},
		{tokenRBrace, "'}'"},
		{tokenLBracket, "'['"},
		{tokenRBracket, "']'"},
		{tokenLAngle, "'<'"},
		{tokenRAngle, "'>'"},
		{tokenLParen, "'('"},
		{tokenRParen, "')'"},
		{tokenComma, "','"},
		{tokenSemicolon, "';'"},
		{tokenColon, "':'"},
		{tokenDoubleColon, "'::'"},
		{tokenQuestion, "'?'"},
		{tokenEquals, "'='"},
		{tokenType(999), "unknown"},
	}
	for _, tt := range tests {
		testutil.Equals(t, tokenName(tt.tt), tt.want)
	}
}

func TestTokenDesc(t *testing.T) {
	testutil.Equals(t, tokenDesc(token{Type: tokenEOF}), "EOF")
	testutil.Equals(t, tokenDesc(token{Type: tokenIdent, Text: "foo"}), `identifier "foo"`)
	testutil.Equals(t, tokenDesc(token{Type: tokenString, Text: "bar"}), `string "bar"`)
	testutil.Equals(t, tokenDesc(token{Type: tokenLBrace, Text: "{"}), `"{"`)
}

func TestIsValidIdent(t *testing.T) {
	testutil.Equals(t, isValidIdent("foo"), true)
	testutil.Equals(t, isValidIdent("_bar"), true)
	testutil.Equals(t, isValidIdent("a1"), true)
	testutil.Equals(t, isValidIdent(""), false)
	testutil.Equals(t, isValidIdent("1abc"), false)
	testutil.Equals(t, isValidIdent("foo bar"), false)
}

func TestLexerBadStringEscape(t *testing.T) {
	src := `"\q"`
	l := newLexer("", []byte(src))
	_, err := l.next()
	testutil.Error(t, err)
}

func TestLexerWhitespace(t *testing.T) {
	src := "  \t\r\n  foo"
	l := newLexer("", []byte(src))
	tok, err := l.next()
	testutil.OK(t, err)
	testutil.Equals(t, tok.Type, tokenIdent)
	testutil.Equals(t, tok.Text, "foo")
}

func TestQuoteCedar(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`hello`, `"hello"`},
		{`he"lo`, `"he\"lo"`},
		{`he\lo`, `"he\\lo"`},
		{"he\nlo", `"he\nlo"`},
		{"he\rlo", `"he\rlo"`},
		{"he\tlo", `"he\tlo"`},
		{"he\x00lo", `"he\0lo"`},
		{"he\vlo", `"he\u{b}lo"`},
		{"he\u0080lo", `"he\u{80}lo"`},
		{"he\U0001F600lo", `"he\u{1f600}lo"`},
	}
	for _, tt := range tests {
		testutil.Equals(t, quoteCedar(tt.input), tt.want)
	}
}

func TestLexerPeekAtEOF(t *testing.T) {
	l := newLexer("", []byte(""))
	testutil.Equals(t, l.peek(), rune(-1))
}

func TestLexerAdvanceAtEOF(t *testing.T) {
	l := newLexer("", []byte(""))
	testutil.Equals(t, l.advance(), rune(-1))
}
