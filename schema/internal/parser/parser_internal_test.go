package parser

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/schema/ast"
	"github.com/cedar-policy/cedar-go/types"
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

// readToken error tests: craft inputs that trigger lexer errors at specific parse points.
// The '$' character is invalid and causes a lexer error.

func TestParseSchemaReadTokenAfterNamespace(t *testing.T) {
	// parseSchema line 115: readToken after "namespace" keyword
	_, err := ParseSchema("", []byte("namespace $"))
	testutil.Error(t, err)
}

func TestParseNamespaceReadTokenAnnotations(t *testing.T) {
	// parseNamespace line 155: readToken error in inner annotations
	_, err := ParseSchema("", []byte("namespace Foo { @$"))
	testutil.Error(t, err)
}

func TestParseNamespaceReadTokenCloseBrace(t *testing.T) {
	// parseNamespace line 163: readToken consuming '}' — error at end
	// Trigger by having valid content then a lex error after '}'
	_, err := ParseSchema("", []byte("namespace Foo { entity Bar; }$"))
	testutil.Error(t, err)
}

func TestParseDeclReadTokenEntity(t *testing.T) {
	// parseDecl line 179: readToken after "entity" keyword
	_, err := ParseSchema("", []byte("entity $"))
	testutil.Error(t, err)
}

func TestParseDeclReadTokenAction(t *testing.T) {
	// parseDecl line 184: readToken after "action" keyword
	_, err := ParseSchema("", []byte("action $"))
	testutil.Error(t, err)
}

func TestParseDeclReadTokenType(t *testing.T) {
	// parseDecl line 189: readToken after "type" keyword
	_, err := ParseSchema("", []byte("type $"))
	testutil.Error(t, err)
}

func TestParseEntityReadTokenEnum(t *testing.T) {
	// parseEntity line 206: readToken after "enum" keyword
	_, err := ParseSchema("", []byte("entity Foo enum $"))
	testutil.Error(t, err)
}

func TestParseEntityReadTokenIn(t *testing.T) {
	// parseEntity line 215: readToken after "in" keyword
	_, err := ParseSchema("", []byte("entity Foo in $"))
	testutil.Error(t, err)
}

func TestParseEntityReadTokenEquals(t *testing.T) {
	// parseEntity line 228: readToken after "="
	_, err := ParseSchema("", []byte("entity Foo = $"))
	testutil.Error(t, err)
}

func TestParseEntityShapeError(t *testing.T) {
	// parseEntity line 232-233: parseRecordType error after "="
	_, err := ParseSchema("", []byte("entity Foo = { $ }"))
	testutil.Error(t, err)
}

func TestParseEntityReadTokenTags(t *testing.T) {
	// parseEntity line 247: readToken after "tags" keyword
	_, err := ParseSchema("", []byte("entity Foo tags $"))
	testutil.Error(t, err)
}

func TestParseEntityTagsTypeErrorLex(t *testing.T) {
	// parseEntity line 247: readToken error in tags
	_, err := ParseSchema("", []byte("entity Foo tags $"))
	testutil.Error(t, err)
}

func TestParseEnumEntityReadTokenAfterString(t *testing.T) {
	// parseEnumEntity line 285: readToken after string in enum
	_, err := ParseSchema("", []byte(`entity Foo enum ["a"$`))
	testutil.Error(t, err)
}

func TestParseEnumEntityReadTokenAfterComma(t *testing.T) {
	// parseEnumEntity line 289: readToken after comma in enum
	_, err := ParseSchema("", []byte(`entity Foo enum ["a",$`))
	testutil.Error(t, err)
}

func TestParseEnumEntityReadTokenAfterBracket(t *testing.T) {
	// parseEnumEntity line 296: readToken consuming ']'
	_, err := ParseSchema("", []byte(`entity Foo enum ["a"]$`))
	testutil.Error(t, err)
}

func TestParseEnumEntityMissingSemicolon(t *testing.T) {
	// parseEnumEntity line 299: expect semicolon
	_, err := ParseSchema("", []byte(`entity Foo enum ["a"]}`))
	testutil.Error(t, err)
}

func TestParseActionReadTokenIn(t *testing.T) {
	// parseAction line 325: readToken after "in"
	_, err := ParseSchema("", []byte("action view in $"))
	testutil.Error(t, err)
}

func TestParseActionReadTokenAppliesTo(t *testing.T) {
	// parseAction line 337: readToken after "appliesTo"
	_, err := ParseSchema("", []byte("action view appliesTo $"))
	testutil.Error(t, err)
}

func TestParseActionReadTokenAttributes(t *testing.T) {
	// parseAction line 349: readToken after "attributes" keyword
	_, err := ParseSchema("", []byte("action view attributes $"))
	testutil.Error(t, err)
}

func TestParseActionAttributesMissingLBrace(t *testing.T) {
	// parseAction line 352: expect tokenLBrace after "attributes"
	_, err := ParseSchema("", []byte("action view attributes foo"))
	testutil.Error(t, err)
}

func TestParseActionAttributesMissingRBrace(t *testing.T) {
	// parseAction line 355: expect tokenRBrace after "attributes {"
	_, err := ParseSchema("", []byte("action view attributes { foo"))
	testutil.Error(t, err)
}

func TestParseActionMissingSemicolon(t *testing.T) {
	// parseAction line 360: expect semicolon
	_, err := ParseSchema("", []byte("action view}"))
	testutil.Error(t, err)
}

func TestParseTypeDeclReadTokenAfterName(t *testing.T) {
	// parseTypeDecl line 385: readToken after type name
	_, err := ParseSchema("", []byte("type Foo$"))
	testutil.Error(t, err)
}

func TestParseTypeDeclTypeError(t *testing.T) {
	// parseTypeDecl line 391-393: parseType error
	_, err := ParseSchema("", []byte("type Foo = $"))
	testutil.Error(t, err)
}

func TestParseTypeDeclMissingSemicolon2(t *testing.T) {
	// parseTypeDecl line 395: expect semicolon after type
	_, err := ParseSchema("", []byte("type Foo = Long}"))
	testutil.Error(t, err)
}

func TestParseAnnotationsReadTokenAfterAt(t *testing.T) {
	// parseAnnotations line 412: readToken after '@'
	_, err := ParseSchema("", []byte("@$"))
	testutil.Error(t, err)
}

func TestParseAnnotationsReadTokenAfterName(t *testing.T) {
	// parseAnnotations line 419: readToken after annotation name
	_, err := ParseSchema("", []byte("@doc$"))
	testutil.Error(t, err)
}

func TestParseAnnotationsReadTokenAfterLParen(t *testing.T) {
	// parseAnnotations line 425: readToken after '('
	_, err := ParseSchema("", []byte("@doc($"))
	testutil.Error(t, err)
}

func TestParseAnnotationsReadTokenAfterString(t *testing.T) {
	// parseAnnotations line 433: readToken after annotation string value
	_, err := ParseSchema("", []byte(`@doc("x"$`))
	testutil.Error(t, err)
}

func TestParseAnnotationsMissingRParen(t *testing.T) {
	// parseAnnotations line 436: expect ')' after annotation value
	_, err := ParseSchema("", []byte(`@doc("x"}`))
	testutil.Error(t, err)
}

func TestParsePathReadTokenAfterIdent(t *testing.T) {
	// parsePath line 458: readToken after first ident
	_, err := ParseSchema("", []byte("entity Foo in Bar$"))
	testutil.Error(t, err)
}

func TestParsePathReadTokenAfterDoubleColon(t *testing.T) {
	// parsePath line 462: readToken after '::'
	_, err := ParseSchema("", []byte("entity Foo in Bar::$"))
	testutil.Error(t, err)
}

func TestParsePathReadTokenAfterSecondIdent(t *testing.T) {
	// parsePath line 469: readToken after second ident
	_, err := ParseSchema("", []byte("entity Foo in Bar::Baz$"))
	testutil.Error(t, err)
}

func TestParsePathForRefReadTokenAfterIdent(t *testing.T) {
	// parsePathForRef line 483: readToken after first ident
	_, err := ParseSchema("", []byte("action view in foo$"))
	testutil.Error(t, err)
}

func TestParsePathForRefReadTokenAfterDoubleColon(t *testing.T) {
	// parsePathForRef line 487: readToken after '::'
	_, err := ParseSchema("", []byte("action view in [foo::$]"))
	testutil.Error(t, err)
}

func TestParsePathForRefReadTokenAfterString(t *testing.T) {
	// parsePathForRef line 492: readToken after string in Path::STR
	_, err := ParseSchema("", []byte(`action view in [Foo::"bar"$]`))
	testutil.Error(t, err)
}

func TestParsePathForRefReadTokenAfterSecondIdent(t *testing.T) {
	// parsePathForRef line 501: readToken after ident in Path::Ident
	_, err := ParseSchema("", []byte("action view in [Foo::Bar$]"))
	testutil.Error(t, err)
}

func TestParseIdentsReadTokenAfterFirst(t *testing.T) {
	// parseIdents line 515: readToken after first ident
	_, err := ParseSchema("", []byte("entity Foo$"))
	testutil.Error(t, err)
}

func TestParseIdentsReadTokenAfterComma(t *testing.T) {
	// parseIdents line 519: readToken after comma
	_, err := ParseSchema("", []byte("entity Foo,$"))
	testutil.Error(t, err)
}

func TestParseIdentsReadTokenAfterSecondIdent(t *testing.T) {
	// parseIdents line 526: readToken after second ident
	_, err := ParseSchema("", []byte("entity Foo, Bar$"))
	testutil.Error(t, err)
}

func TestParseNamesReadTokenAfterComma(t *testing.T) {
	// parseNames line 541: readToken after comma
	_, err := ParseSchema("", []byte("action foo,$"))
	testutil.Error(t, err)
}

func TestParseNameReadTokenAfterIdent(t *testing.T) {
	// parseName line 557: readToken after ident name
	_, err := ParseSchema("", []byte(`action foo in [bar$]`))
	testutil.Error(t, err)
}

func TestParseNameReadTokenAfterString(t *testing.T) {
	// parseName line 563: readToken after string name
	_, err := ParseSchema("", []byte(`action "foo"$`))
	testutil.Error(t, err)
}

func TestParseEntityTypesReadTokenAfterLBracket(t *testing.T) {
	// parseEntityTypes line 575: readToken after '['
	_, err := ParseSchema("", []byte("entity Foo in [$"))
	testutil.Error(t, err)
}

func TestParseEntityTypesReadTokenAfterComma(t *testing.T) {
	// parseEntityTypes line 586: readToken after comma
	_, err := ParseSchema("", []byte("entity Foo in [Bar,$"))
	testutil.Error(t, err)
}

func TestParseEntityTypesReadTokenAfterRBracket(t *testing.T) {
	// parseEntityTypes line 593: readToken consuming ']'
	_, err := ParseSchema("", []byte("entity Foo in [Bar]$"))
	testutil.Error(t, err)
}

func TestParseActionParentsReadTokenAfterLBracket(t *testing.T) {
	// parseActionParents line 605: readToken after '['
	_, err := ParseSchema("", []byte("action view in [$"))
	testutil.Error(t, err)
}

func TestParseActionParentsReadTokenAfterComma(t *testing.T) {
	// parseActionParents line 616: readToken after comma in action parent list
	_, err := ParseSchema("", []byte("action view in [foo,$"))
	testutil.Error(t, err)
}

func TestParseActionParentsQualNameError(t *testing.T) {
	// parseActionParents line 626: parseQualName error (single parent)
	_, err := ParseSchema("", []byte("action view in 42"))
	testutil.Error(t, err)
}

func TestParseQualNameReadTokenAfterString(t *testing.T) {
	// parseQualName line 636: readToken after string literal
	_, err := ParseSchema("", []byte(`action view in ["foo"$]`))
	testutil.Error(t, err)
}

func TestParseAppliesToEOF(t *testing.T) {
	// parseAppliesTo line 659: EOF inside appliesTo
	_, err := ParseSchema("", []byte("action view appliesTo { principal: User"))
	testutil.Error(t, err)
}

func TestParseAppliesToReadTokenPrincipal(t *testing.T) {
	// parseAppliesTo line 667: readToken after "principal"
	_, err := ParseSchema("", []byte("action view appliesTo { principal$"))
	testutil.Error(t, err)
}

func TestParseAppliesToPrincipalColonError(t *testing.T) {
	// parseAppliesTo line 670: expect colon after "principal"
	_, err := ParseSchema("", []byte("action view appliesTo { principal User }"))
	testutil.Error(t, err)
}

func TestParseAppliesToPrincipalTypeError(t *testing.T) {
	// parseAppliesTo line 673: parseEntityTypes error after "principal:"
	_, err := ParseSchema("", []byte("action view appliesTo { principal: $ }"))
	testutil.Error(t, err)
}

func TestParseAppliesToReadTokenResource(t *testing.T) {
	// parseAppliesTo line 679: readToken after "resource"
	_, err := ParseSchema("", []byte("action view appliesTo { resource$"))
	testutil.Error(t, err)
}

func TestParseAppliesToResourceColonError(t *testing.T) {
	// parseAppliesTo line 682: expect colon after "resource"
	_, err := ParseSchema("", []byte("action view appliesTo { resource User }"))
	testutil.Error(t, err)
}

func TestParseAppliesToResourceTypeError(t *testing.T) {
	// parseAppliesTo line 685: parseEntityTypes error after "resource:"
	_, err := ParseSchema("", []byte("action view appliesTo { resource: $ }"))
	testutil.Error(t, err)
}

func TestParseAppliesToReadTokenContext(t *testing.T) {
	// parseAppliesTo line 691: readToken after "context"
	_, err := ParseSchema("", []byte("action view appliesTo { context$"))
	testutil.Error(t, err)
}

func TestParseAppliesToContextColonError(t *testing.T) {
	// parseAppliesTo line 694: expect colon after "context"
	_, err := ParseSchema("", []byte("action view appliesTo { context User }"))
	testutil.Error(t, err)
}

func TestParseAppliesToContextTypeError(t *testing.T) {
	// parseAppliesTo line 697: parseType error after "context:"
	_, err := ParseSchema("", []byte("action view appliesTo { context: $ }"))
	testutil.Error(t, err)
}

func TestParseAppliesToReadTokenAfterComma(t *testing.T) {
	// parseAppliesTo line 706: readToken after comma in appliesTo
	_, err := ParseSchema("", []byte("action view appliesTo { principal: User,$"))
	testutil.Error(t, err)
}

func TestParseTypeRecordError(t *testing.T) {
	// parseType line 717-719: parseRecordType error
	_, err := ParseSchema("", []byte("entity Foo { x: { $ } };"))
	testutil.Error(t, err)
}

func TestParseTypeSetReadTokenAfterSet(t *testing.T) {
	// parseType line 725: readToken after "Set"
	_, err := ParseSchema("", []byte("entity Foo { x: Set$"))
	testutil.Error(t, err)
}

func TestParseTypeSetMissingLAngle(t *testing.T) {
	// parseType line 728: expect '<' after "Set"
	_, err := ParseSchema("", []byte("entity Foo { x: Set(Long) };"))
	testutil.Error(t, err)
}

func TestParseTypeSetElemError(t *testing.T) {
	// parseType line 731-733: parseType error inside Set<>
	_, err := ParseSchema("", []byte("entity Foo { x: Set<$> };"))
	testutil.Error(t, err)
}

func TestParseTypeSetMissingRAngle(t *testing.T) {
	// parseType line 735: expect '>' after Set element
	_, err := ParseSchema("", []byte("entity Foo { x: Set<Long; };"))
	testutil.Error(t, err)
}

func TestParseTypePathError(t *testing.T) {
	// parseType line 741-743: parsePath error (not ident, not Set, not '{')
	_, err := ParseSchema("", []byte("entity Foo { x: 42 };"))
	testutil.Error(t, err)
}

func TestParseRecordTypeAnnotationsError(t *testing.T) {
	// parseRecordType line 758-760: parseAnnotations error inside record
	_, err := ParseSchema("", []byte("entity Foo { @$ name: Long };"))
	testutil.Error(t, err)
}

func TestParseRecordTypeNameError(t *testing.T) {
	// parseRecordType line 762-764: parseName error for attr name
	_, err := ParseSchema("", []byte("entity Foo { 42: Long };"))
	testutil.Error(t, err)
}

func TestParseRecordTypeReadTokenAfterQuestion(t *testing.T) {
	// parseRecordType line 769: readToken after '?'
	_, err := ParseSchema("", []byte("entity Foo { x?$"))
	testutil.Error(t, err)
}

func TestParseRecordTypeTypeError(t *testing.T) {
	// parseRecordType line 776-778: parseType error
	_, err := ParseSchema("", []byte("entity Foo { x: $ };"))
	testutil.Error(t, err)
}

func TestParseRecordTypeReadTokenAfterComma(t *testing.T) {
	// parseRecordType line 786: readToken after comma
	_, err := ParseSchema("", []byte("entity Foo { x: Long,$"))
	testutil.Error(t, err)
}

func TestParseNamespacePathError(t *testing.T) {
	// parseNamespace line 142: parsePath gets non-ident token
	_, err := ParseSchema("", []byte(`namespace "bad" {}`))
	testutil.Error(t, err)
}

func TestParseDeclNotIdentInNamespace(t *testing.T) {
	// parseDecl line 174: non-ident token where declaration expected
	_, err := ParseSchema("", []byte("namespace Foo { ; }"))
	testutil.Error(t, err)
}

func TestParseEntityTagsTypeError(t *testing.T) {
	// parseEntity line 251: parseType error after "tags"
	_, err := ParseSchema("", []byte("entity Foo tags ;"))
	testutil.Error(t, err)
}

func TestParseEnumEntityNotString(t *testing.T) {
	// parseEnumEntity line 281: ident where string expected
	_, err := ParseSchema("", []byte("entity Foo enum [Bar];"))
	testutil.Error(t, err)
}

func TestParseTypeDeclParseTypeError(t *testing.T) {
	// parseTypeDecl line 392 (and parseType line 742): parseType error
	_, err := ParseSchema("", []byte("type Foo = ;"))
	testutil.Error(t, err)
}

func TestParseAnnotationsNotString(t *testing.T) {
	// parseAnnotations line 428: non-string after '(' in annotation value
	_, err := ParseSchema("", []byte("@doc(;) entity Foo;"))
	testutil.Error(t, err)
}

func TestParsePathNotIdent(t *testing.T) {
	// parsePath line 454: non-ident token at start of path
	_, err := ParseSchema("", []byte(`namespace Foo { entity Bar in "bad"; }`))
	testutil.Error(t, err)
}

func TestParsePathNotIdentAfterColon(t *testing.T) {
	// parsePath line 465: non-ident after "::" in path
	_, err := ParseSchema("", []byte(`namespace Foo { entity Bar in Baz::"bad"; }`))
	testutil.Error(t, err)
}

func TestParsePathForRefNotIdent(t *testing.T) {
	// parsePathForRef line 479: non-ident/non-string at start
	// parseQualName checks tokenString first, so we need a non-string non-ident valid token
	_, err := ParseSchema("", []byte("action view in [;];"))
	testutil.Error(t, err)
}

func TestParsePathForRefNotIdentAfterColon(t *testing.T) {
	// parsePathForRef line 497: not ident or string after "::" in ref
	_, err := ParseSchema("", []byte("action view in [Foo::;];"))
	testutil.Error(t, err)
}

func TestParseIdentsNotIdentAfterComma(t *testing.T) {
	// parseIdents line 522: non-ident after ',' in ident list
	_, err := ParseSchema("", []byte(`entity Foo, "bar" {};`))
	testutil.Error(t, err)
}

func TestParseNamesParseNameError(t *testing.T) {
	// parseNames line 545: parseName error after comma
	_, err := ParseSchema("", []byte("action foo, ;"))
	testutil.Error(t, err)
}

func TestParseNameDefault(t *testing.T) {
	// parseName line 567-568: default case (neither ident nor string)
	_, err := ParseSchema("", []byte("action ;"))
	testutil.Error(t, err)
}

func TestParseEntityTypesPathError(t *testing.T) {
	// parseEntityTypes line 581: parsePath error inside bracketed list
	_, err := ParseSchema("", []byte(`entity Foo in ["bad"];`))
	testutil.Error(t, err)
}

func TestParseAppliesToNotIdentCheck(t *testing.T) {
	// parseAppliesTo line 662: non-ident token where principal/resource/context expected
	_, err := ParseSchema("", []byte(`action view appliesTo { "bad" };`))
	testutil.Error(t, err)
}

func TestParseAppliesToPrincipalTypeError2(t *testing.T) {
	// parseAppliesTo line 674: parseEntityTypes error for principal
	_, err := ParseSchema("", []byte(`action view appliesTo { principal: "bad" };`))
	testutil.Error(t, err)
}

func TestParseAppliesToResourceTypeError2(t *testing.T) {
	// parseAppliesTo line 686: parseEntityTypes error for resource
	_, err := ParseSchema("", []byte(`action view appliesTo { resource: "bad" };`))
	testutil.Error(t, err)
}

func TestParseAppliesToContextTypeError2(t *testing.T) {
	// parseAppliesTo line 698: parseType error for context
	_, err := ParseSchema("", []byte("action view appliesTo { context: ; };"))
	testutil.Error(t, err)
}

func TestParseTypeSetInnerError(t *testing.T) {
	// parseType line 732: parseType error inside Set<>
	_, err := ParseSchema("", []byte("entity Foo { x: Set<;> };"))
	testutil.Error(t, err)
}

func TestParseRecordTypeEOF(t *testing.T) {
	// parseRecordType line 755: EOF inside record
	_, err := ParseSchema("", []byte("entity Foo {"))
	testutil.Error(t, err)
}

func TestParseRecordTypeNameError2(t *testing.T) {
	// parseRecordType line 763: parseName error for attr (non-ident, non-string)
	_, err := ParseSchema("", []byte("entity Foo { ;: Long };"))
	testutil.Error(t, err)
}

func TestParseNamespaceMissingBrace(t *testing.T) {
	// parseNamespace line 145: expect '{' error after path
	_, err := ParseSchema("", []byte("namespace Foo entity Bar;"))
	testutil.Error(t, err)
}

func TestParseTypeDeclNotIdent(t *testing.T) {
	// parseTypeDecl line 378: non-ident where type name expected
	_, err := ParseSchema("", []byte(`type "bad" = Long;`))
	testutil.Error(t, err)
}

func TestParseNameReadTokenAfterIdentInAction(t *testing.T) {
	// parseName line 557: readToken error after valid ident in action name position
	_, err := ParseSchema("", []byte("action foo$"))
	testutil.Error(t, err)
}

func TestMarshalMultipleDecls(t *testing.T) {
	// marshal.go lines 64, 105: !*first newline separator between declarations
	// Also lines 333, 342: sort comparison with 2+ keys
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS1": ast.Namespace{},
			"NS2": ast.Namespace{
				CommonTypes: ast.CommonTypes{
					"A": ast.CommonType{Type: ast.StringType{}},
					"B": ast.CommonType{Type: ast.LongType{}},
				},
				Enums: ast.Enums{
					"NS2::Color": ast.Enum{Values: []types.String{"red"}},
					"NS2::Size":  ast.Enum{Values: []types.String{"small"}},
				},
			},
		},
	}
	result := MarshalSchema(s)
	testutil.Equals(t, len(result) > 0, true)
	// Round-trip to verify
	_, err := ParseSchema("", result)
	testutil.OK(t, err)
}

func TestLexerPeekAtEOF(t *testing.T) {
	l := newLexer("", []byte(""))
	testutil.Equals(t, l.peek(), rune(-1))
}

func TestLexerAdvanceAtEOF(t *testing.T) {
	l := newLexer("", []byte(""))
	testutil.Equals(t, l.advance(), rune(-1))
}
