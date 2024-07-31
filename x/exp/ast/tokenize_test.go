package ast

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/cedar-policy/cedar-go/testutil"
)

func TestTokenize(t *testing.T) {
	t.Parallel()
	input := `
These are some identifiers
0 1 1234
-1 9223372036854775807 -9223372036854775808
"" "string" "\"\'\n\r\t\\\0" "\x123" "\u{0}\u{10fFfF}"
"*" "\*" "*\**"
@.,;(){}[]+-*
:::
!!=<<=>>=
||&&
// single line comment
/*
multiline comment
// embedded comment does nothing
*/
'/%|&=`
	want := []Token{
		{Type: TokenIdent, Text: "These", Pos: Position{Offset: 1, Line: 2, Column: 1}},
		{Type: TokenIdent, Text: "are", Pos: Position{Offset: 7, Line: 2, Column: 7}},
		{Type: TokenIdent, Text: "some", Pos: Position{Offset: 11, Line: 2, Column: 11}},
		{Type: TokenIdent, Text: "identifiers", Pos: Position{Offset: 16, Line: 2, Column: 16}},

		{Type: TokenInt, Text: "0", Pos: Position{Offset: 28, Line: 3, Column: 1}},
		{Type: TokenInt, Text: "1", Pos: Position{Offset: 30, Line: 3, Column: 3}},
		{Type: TokenInt, Text: "1234", Pos: Position{Offset: 32, Line: 3, Column: 5}},

		{Type: TokenOperator, Text: "-", Pos: Position{Offset: 37, Line: 4, Column: 1}},
		{Type: TokenInt, Text: "1", Pos: Position{Offset: 38, Line: 4, Column: 2}},
		{Type: TokenInt, Text: "9223372036854775807", Pos: Position{Offset: 40, Line: 4, Column: 4}},
		{Type: TokenOperator, Text: "-", Pos: Position{Offset: 60, Line: 4, Column: 24}},
		{Type: TokenInt, Text: "9223372036854775808", Pos: Position{Offset: 61, Line: 4, Column: 25}},

		{Type: TokenString, Text: `""`, Pos: Position{Offset: 81, Line: 5, Column: 1}},
		{Type: TokenString, Text: `"string"`, Pos: Position{Offset: 84, Line: 5, Column: 4}},
		{Type: TokenString, Text: `"\"\'\n\r\t\\\0"`, Pos: Position{Offset: 93, Line: 5, Column: 13}},
		{Type: TokenString, Text: `"\x123"`, Pos: Position{Offset: 110, Line: 5, Column: 30}},
		{Type: TokenString, Text: `"\u{0}\u{10fFfF}"`, Pos: Position{Offset: 118, Line: 5, Column: 38}},

		{Type: TokenString, Text: `"*"`, Pos: Position{Offset: 136, Line: 6, Column: 1}},
		{Type: TokenString, Text: `"\*"`, Pos: Position{Offset: 140, Line: 6, Column: 5}},
		{Type: TokenString, Text: `"*\**"`, Pos: Position{Offset: 145, Line: 6, Column: 10}},

		{Type: TokenOperator, Text: "@", Pos: Position{Offset: 152, Line: 7, Column: 1}},
		{Type: TokenOperator, Text: ".", Pos: Position{Offset: 153, Line: 7, Column: 2}},
		{Type: TokenOperator, Text: ",", Pos: Position{Offset: 154, Line: 7, Column: 3}},
		{Type: TokenOperator, Text: ";", Pos: Position{Offset: 155, Line: 7, Column: 4}},
		{Type: TokenOperator, Text: "(", Pos: Position{Offset: 156, Line: 7, Column: 5}},
		{Type: TokenOperator, Text: ")", Pos: Position{Offset: 157, Line: 7, Column: 6}},
		{Type: TokenOperator, Text: "{", Pos: Position{Offset: 158, Line: 7, Column: 7}},
		{Type: TokenOperator, Text: "}", Pos: Position{Offset: 159, Line: 7, Column: 8}},
		{Type: TokenOperator, Text: "[", Pos: Position{Offset: 160, Line: 7, Column: 9}},
		{Type: TokenOperator, Text: "]", Pos: Position{Offset: 161, Line: 7, Column: 10}},
		{Type: TokenOperator, Text: "+", Pos: Position{Offset: 162, Line: 7, Column: 11}},
		{Type: TokenOperator, Text: "-", Pos: Position{Offset: 163, Line: 7, Column: 12}},
		{Type: TokenOperator, Text: "*", Pos: Position{Offset: 164, Line: 7, Column: 13}},

		{Type: TokenOperator, Text: "::", Pos: Position{Offset: 166, Line: 8, Column: 1}},
		{Type: TokenOperator, Text: ":", Pos: Position{Offset: 168, Line: 8, Column: 3}},

		{Type: TokenOperator, Text: "!", Pos: Position{Offset: 170, Line: 9, Column: 1}},
		{Type: TokenOperator, Text: "!=", Pos: Position{Offset: 171, Line: 9, Column: 2}},
		{Type: TokenOperator, Text: "<", Pos: Position{Offset: 173, Line: 9, Column: 4}},
		{Type: TokenOperator, Text: "<=", Pos: Position{Offset: 174, Line: 9, Column: 5}},
		{Type: TokenOperator, Text: ">", Pos: Position{Offset: 176, Line: 9, Column: 7}},
		{Type: TokenOperator, Text: ">=", Pos: Position{Offset: 177, Line: 9, Column: 8}},

		{Type: TokenOperator, Text: "||", Pos: Position{Offset: 180, Line: 10, Column: 1}},
		{Type: TokenOperator, Text: "&&", Pos: Position{Offset: 182, Line: 10, Column: 3}},

		{Type: TokenUnknown, Text: "'", Pos: Position{Offset: 265, Line: 16, Column: 1}},
		{Type: TokenUnknown, Text: "/", Pos: Position{Offset: 266, Line: 16, Column: 2}},
		{Type: TokenUnknown, Text: "%", Pos: Position{Offset: 267, Line: 16, Column: 3}},
		{Type: TokenUnknown, Text: "|", Pos: Position{Offset: 268, Line: 16, Column: 4}},
		{Type: TokenUnknown, Text: "&", Pos: Position{Offset: 269, Line: 16, Column: 5}},
		{Type: TokenUnknown, Text: "=", Pos: Position{Offset: 270, Line: 16, Column: 6}},

		{Type: TokenEOF, Text: "", Pos: Position{Offset: 271, Line: 16, Column: 7}},
	}
	got, err := Tokenize([]byte(input))
	testutil.OK(t, err)
	testutil.Equals(t, got, want)
}

func TestTokenizeErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input      string
		wantErrStr string
		wantErrPos Position
	}{
		{"okay\x00not okay", "invalid character NUL", Position{Line: 1, Column: 1}},
		{`okay /*
        stuff
        `, "comment not terminated", Position{Line: 1, Column: 6}},
		{`okay "
        " foo bar`, "literal not terminated", Position{Line: 1, Column: 6}},
		{`"okay" "\a"`, "invalid char escape", Position{Line: 1, Column: 8}},
		{`"okay" "\b"`, "invalid char escape", Position{Line: 1, Column: 8}},
		{`"okay" "\f"`, "invalid char escape", Position{Line: 1, Column: 8}},
		{`"okay" "\v"`, "invalid char escape", Position{Line: 1, Column: 8}},
		{`"okay" "\1"`, "invalid char escape", Position{Line: 1, Column: 8}},
		{`"okay" "\x"`, "invalid char escape", Position{Line: 1, Column: 8}},
		{`"okay" "\x1"`, "invalid char escape", Position{Line: 1, Column: 8}},
		{`"okay" "\ubadf"`, "invalid char escape", Position{Line: 1, Column: 8}},
		{`"okay" "\U0000badf"`, "invalid char escape", Position{Line: 1, Column: 8}},
		{`"okay" "\u{}"`, "invalid char escape", Position{Line: 1, Column: 8}},
		{`"okay" "\u{0000000}"`, "invalid char escape", Position{Line: 1, Column: 8}},
		{`"okay" "\u{z"`, "invalid char escape", Position{Line: 1, Column: 8}},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()
			got, gotErr := Tokenize([]byte(tt.input))
			wantErrStr := fmt.Sprintf("%v: %s", tt.wantErrPos, tt.wantErrStr)
			testutil.Error(t, gotErr)
			testutil.Equals(t, gotErr.Error(), wantErrStr)
			testutil.Equals(t, got, nil)
		})
	}
}

func TestIntTokenValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		wantOk  bool
		want    int64
		wantErr string
	}{
		{"0", true, 0, ""},
		{"9223372036854775807", true, 9223372036854775807, ""},
		{"9223372036854775808", false, 0, `strconv.ParseInt: parsing "9223372036854775808": value out of range`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := Tokenize([]byte(tt.input))
			testutil.OK(t, err)
			testutil.Equals(t, len(got), 2)
			testutil.Equals(t, got[0].Type, TokenInt)
			gotInt, err := got[0].intValue()
			if err != nil {
				testutil.Equals(t, tt.wantOk, false)
				testutil.Equals(t, err.Error(), tt.wantErr)
			} else {
				testutil.Equals(t, tt.wantOk, true)
				testutil.Equals(t, gotInt, tt.want)
			}
		})
	}
}

func TestStringTokenValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		wantOk  bool
		want    string
		wantErr string
	}{
		{`""`, true, "", ""},
		{`"hello"`, true, "hello", ""},
		{`"a\n\r\t\\\0b"`, true, "a\n\r\t\\\x00b", ""},
		{`"a\"b"`, true, "a\"b", ""},
		{`"a\'b"`, true, "a'b", ""},

		{`"a\x00b"`, true, "a\x00b", ""},
		{`"a\x7fb"`, true, "a\x7fb", ""},
		{`"a\x80b"`, false, "", "bad hex escape sequence"},

		{`"a\u{A}b"`, true, "a\u000ab", ""},
		{`"a\u{aB}b"`, true, "a\u00abb", ""},
		{`"a\u{AbC}b"`, true, "a\u0abcb", ""},
		{`"a\u{aBcD}b"`, true, "a\uabcdb", ""},
		{`"a\u{AbCdE}b"`, true, "a\U000abcdeb", ""},
		{`"a\u{10cDeF}b"`, true, "a\U0010cdefb", ""},
		{`"a\u{ffffff}b"`, false, "", "bad unicode escape sequence"},
		{`"a\u{d7ff}b"`, true, "a\ud7ffb", ""},
		{`"a\u{d800}b"`, false, "", "bad unicode escape sequence"},
		{`"a\u{dfff}b"`, false, "", "bad unicode escape sequence"},
		{`"a\u{e000}b"`, true, "a\ue000b", ""},
		{`"a\u{10ffff}b"`, true, "a\U0010ffffb", ""},
		{`"a\u{110000}b"`, false, "", "bad unicode escape sequence"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := Tokenize([]byte(tt.input))
			testutil.OK(t, err)
			testutil.Equals(t, len(got), 2)
			testutil.Equals(t, got[0].Type, TokenString)
			gotStr, err := got[0].stringValue()
			if err != nil {
				testutil.Equals(t, tt.wantOk, false)
				testutil.Equals(t, err.Error(), tt.wantErr)
			} else {
				testutil.Equals(t, tt.wantOk, true)
				testutil.Equals(t, gotStr, tt.want)
			}
		})
	}
}

func TestParseUnicodeEscape(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   []byte
		out  rune
		outN int
		err  func(t testing.TB, err error)
	}{
		{"happy", []byte{'{', '4', '2', '}'}, 0x42, 4, testutil.OK},
		{"badRune", []byte{'{', 0x80, 0x81}, 0, 1, testutil.Error},
		{"notHex", []byte{'{', 'g'}, 0, 2, testutil.Error},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, n, err := parseUnicodeEscape(tt.in, 0)
			testutil.Equals(t, out, tt.out)
			testutil.Equals(t, n, tt.outN)
			tt.err(t, err)
		})
	}
}

func TestUnquote(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		out  string
		err  func(t testing.TB, err error)
	}{
		{"happy", `"test"`, `test`, testutil.OK},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := Unquote(tt.in)
			testutil.Equals(t, out, tt.out)
			tt.err(t, err)
		})
	}
}

func TestRustUnquote(t *testing.T) {
	t.Parallel()
	// star == false
	{
		tests := []struct {
			input   string
			wantOk  bool
			want    string
			wantErr string
		}{
			{``, true, "", ""},
			{`hello`, true, "hello", ""},
			{`a\n\r\t\\\0b`, true, "a\n\r\t\\\x00b", ""},
			{`a\"b`, true, "a\"b", ""},
			{`a\'b`, true, "a'b", ""},

			{`a\x00b`, true, "a\x00b", ""},
			{`a\x7fb`, true, "a\x7fb", ""},
			{`a\x80b`, false, "", "bad hex escape sequence"},

			{string([]byte{0x80, 0x81}), false, "", "bad unicode rune"},
			{`a\u`, false, "", "bad unicode rune"},
			{`a\uz`, false, "", "bad unicode escape sequence"},
			{`a\u{}b`, false, "", "bad unicode escape sequence"},
			{`a\u{A}b`, true, "a\u000ab", ""},
			{`a\u{aB}b`, true, "a\u00abb", ""},
			{`a\u{AbC}b`, true, "a\u0abcb", ""},
			{`a\u{aBcD}b`, true, "a\uabcdb", ""},
			{`a\u{AbCdE}b`, true, "a\U000abcdeb", ""},
			{`a\u{10cDeF}b`, true, "a\U0010cdefb", ""},
			{`a\u{ffffff}b`, false, "", "bad unicode escape sequence"},
			{`a\u{0000000}b`, false, "", "bad unicode escape sequence"},
			{`a\u{d7ff}b`, true, "a\ud7ffb", ""},
			{`a\u{d800}b`, false, "", "bad unicode escape sequence"},
			{`a\u{dfff}b`, false, "", "bad unicode escape sequence"},
			{`a\u{e000}b`, true, "a\ue000b", ""},
			{`a\u{10ffff}b`, true, "a\U0010ffffb", ""},
			{`a\u{110000}b`, false, "", "bad unicode escape sequence"},

			{`\`, false, "", "bad unicode rune"},
			{`\a`, false, "", "bad char escape"},
			{`\*`, false, "", "bad char escape"},
			{`\x`, false, "", "bad unicode rune"},
			{`\xz`, false, "", "bad hex escape sequence"},
			{`\xa`, false, "", "bad unicode rune"},
			{`\xaz`, false, "", "bad hex escape sequence"},
			{`\{`, false, "", "bad char escape"},
			{`\{z`, false, "", "bad char escape"},
			{`\{0`, false, "", "bad char escape"},
			{`\{0z`, false, "", "bad char escape"},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.input, func(t *testing.T) {
				t.Parallel()
				got, rem, err := rustUnquote([]byte(tt.input), false)
				if err != nil {
					testutil.Equals(t, tt.wantOk, false)
					testutil.Equals(t, err.Error(), tt.wantErr)
					testutil.Equals(t, got, tt.want)
				} else {
					testutil.Equals(t, tt.wantOk, true)
					testutil.Equals(t, got, tt.want)
					testutil.Equals(t, rem, []byte(""))
				}
			})
		}
	}

	// star == true
	{
		tests := []struct {
			input   string
			wantOk  bool
			want    string
			wantRem string
			wantErr string
		}{
			{``, true, "", "", ""},
			{`hello`, true, "hello", "", ""},
			{`a\n\r\t\\\0b`, true, "a\n\r\t\\\x00b", "", ""},
			{`a\"b`, true, "a\"b", "", ""},
			{`a\'b`, true, "a'b", "", ""},

			{`a\x00b`, true, "a\x00b", "", ""},
			{`a\x7fb`, true, "a\x7fb", "", ""},
			{`a\x80b`, false, "", "", "bad hex escape sequence"},

			{`a\u`, false, "", "", "bad unicode rune"},
			{`a\uz`, false, "", "", "bad unicode escape sequence"},
			{`a\u{}b`, false, "", "", "bad unicode escape sequence"},
			{`a\u{A}b`, true, "a\u000ab", "", ""},
			{`a\u{aB}b`, true, "a\u00abb", "", ""},
			{`a\u{AbC}b`, true, "a\u0abcb", "", ""},
			{`a\u{aBcD}b`, true, "a\uabcdb", "", ""},
			{`a\u{AbCdE}b`, true, "a\U000abcdeb", "", ""},
			{`a\u{10cDeF}b`, true, "a\U0010cdefb", "", ""},
			{`a\u{ffffff}b`, false, "", "", "bad unicode escape sequence"},
			{`a\u{0000000}b`, false, "", "", "bad unicode escape sequence"},
			{`a\u{d7ff}b`, true, "a\ud7ffb", "", ""},
			{`a\u{d800}b`, false, "", "", "bad unicode escape sequence"},
			{`a\u{dfff}b`, false, "", "", "bad unicode escape sequence"},
			{`a\u{e000}b`, true, "a\ue000b", "", ""},
			{`a\u{10ffff}b`, true, "a\U0010ffffb", "", ""},
			{`a\u{110000}b`, false, "", "", "bad unicode escape sequence"},

			{`*`, true, "", "*", ""},
			{`*hello*how*are*you`, true, "", "*hello*how*are*you", ""},
			{`hello*how*are*you`, true, "hello", "*how*are*you", ""},
			{`\**`, true, "*", "*", ""},

			{`\`, false, "", "", "bad unicode rune"},
			{`\a`, false, "", "", "bad char escape"},
			{`\x`, false, "", "", "bad unicode rune"},
			{`\xz`, false, "", "", "bad hex escape sequence"},
			{`\xa`, false, "", "", "bad unicode rune"},
			{`\xaz`, false, "", "", "bad hex escape sequence"},
			{`\{`, false, "", "", "bad char escape"},
			{`\{z`, false, "", "", "bad char escape"},
			{`\{0`, false, "", "", "bad char escape"},
			{`\{0z`, false, "", "", "bad char escape"},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.input, func(t *testing.T) {
				t.Parallel()
				got, rem, err := rustUnquote([]byte(tt.input), true)
				if err != nil {
					testutil.Equals(t, tt.wantOk, false)
					testutil.Equals(t, err.Error(), tt.wantErr)
					testutil.Equals(t, got, tt.want)
				} else {
					testutil.Equals(t, tt.wantOk, true)
					testutil.Equals(t, got, tt.want)
					testutil.Equals(t, string(rem), tt.wantRem)
				}
			})
		}
	}
}

func TestFakeRustQuote(t *testing.T) {
	t.Parallel()
	out := FakeRustQuote("hello")
	testutil.Equals(t, out, `"hello"`)
}

func TestPatternFromStringLiteral(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		wantOk  bool
		want    []PatternComponent
		wantErr string
	}{
		{`""`, true, nil, ""},
		{`"a"`, true, []PatternComponent{{false, "a"}}, ""},
		{`"*"`, true, []PatternComponent{{true, ""}}, ""},
		{`"*a"`, true, []PatternComponent{{true, "a"}}, ""},
		{`"a*"`, true, []PatternComponent{{false, "a"}, {true, ""}}, ""},
		{`"**"`, true, []PatternComponent{{true, ""}}, ""},
		{`"**a"`, true, []PatternComponent{{true, "a"}}, ""},
		{`"a**"`, true, []PatternComponent{{false, "a"}, {true, ""}}, ""},
		{`"*a*"`, true, []PatternComponent{{true, "a"}, {true, ""}}, ""},
		{`"**a**"`, true, []PatternComponent{{true, "a"}, {true, ""}}, ""},
		{`"abra*ca"`, true, []PatternComponent{
			{false, "abra"}, {true, "ca"},
		}, ""},
		{`"abra**ca"`, true, []PatternComponent{
			{false, "abra"}, {true, "ca"},
		}, ""},
		{`"*abra*ca"`, true, []PatternComponent{
			{true, "abra"}, {true, "ca"},
		}, ""},
		{`"abra*ca*"`, true, []PatternComponent{
			{false, "abra"}, {true, "ca"}, {true, ""},
		}, ""},
		{`"*abra*ca*"`, true, []PatternComponent{
			{true, "abra"}, {true, "ca"}, {true, ""},
		}, ""},
		{`"*abra*ca*dabra"`, true, []PatternComponent{
			{true, "abra"}, {true, "ca"}, {true, "dabra"},
		}, ""},
		{`"*abra*c\**da\*ra"`, true, []PatternComponent{
			{true, "abra"}, {true, "c*"}, {true, "da*ra"},
		}, ""},
		{`"\u"`, false, nil, "bad unicode rune"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := NewPattern(tt.input)
			if err != nil {
				testutil.Equals(t, tt.wantOk, false)
				testutil.Equals(t, err.Error(), tt.wantErr)
			} else {
				testutil.Equals(t, tt.wantOk, true)
				testutil.Equals(t, got.Comps, tt.want)
				testutil.Equals(t, got.String(), tt.input)
			}
		})
	}
}

func TestScanner(t *testing.T) {
	t.Parallel()
	t.Run("SrcError", func(t *testing.T) {
		t.Parallel()
		wantErr := fmt.Errorf("wantErr")
		r := &readerMock{
			ReadFunc: func(_ []byte) (int, error) {
				return 0, wantErr
			},
		}
		var s scanner
		s.Init(r)
		out := s.next()
		testutil.Equals(t, out, specialRuneEOF)
	})

	t.Run("MidEmojiEOF", func(t *testing.T) {
		t.Parallel()
		var s scanner
		var eof bool
		str := []byte(string(`🐐`))
		r := &readerMock{
			ReadFunc: func(p []byte) (int, error) {
				if eof {
					return 0, io.EOF
				}
				p[0] = str[0]
				eof = true
				return 1, nil
			},
		}
		s.Init(r)
		out := s.next()
		testutil.Equals(t, out, utf8.RuneError)
		out = s.next()
		testutil.Equals(t, out, specialRuneEOF)
	})

	t.Run("NotAsciiEmoji", func(t *testing.T) {
		t.Parallel()
		var s scanner
		s.Init(strings.NewReader(`🐐`))
		out := s.next()
		testutil.Equals(t, out, '🐐')
	})

	t.Run("InvalidUTF8", func(t *testing.T) {
		t.Parallel()
		var s scanner
		s.Init(strings.NewReader(string([]byte{0x80, 0x81})))
		out := s.next()
		testutil.Equals(t, out, utf8.RuneError)
	})

	t.Run("tokenTextNone", func(t *testing.T) {
		t.Parallel()
		var s scanner
		s.Init(strings.NewReader(""))
		out := s.tokenText()
		testutil.Equals(t, out, "")
	})
}

func TestDigitVal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   rune
		out  int
	}{
		{"happy", '0', 0},
		{"hex", 'f', 15},
		{"sad", 'g', 16},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := digitVal(tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}
