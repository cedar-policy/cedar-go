package rust

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestParseUnicodeEscape(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   []byte
		out  rune
		outN int
		err  func(t testutil.TB, err error)
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
		err  func(t testutil.TB, err error)
	}{
		{"happy", `"test"`, `test`, testutil.OK},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := unquote(tt.in)
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
				got, rem, err := Unquote([]byte(tt.input), false)
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
				got, rem, err := Unquote([]byte(tt.input), true)
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
