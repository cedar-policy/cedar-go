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

func TestEscapeString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"empty", "", ""},
		{"plain", "hello", "hello"},
		{"null", "\x00", `\0`},
		{"tab", "\t", `\t`},
		{"newline", "\n", `\n`},
		{"cr", "\r", `\r`},
		{"backslash", `\`, `\\`},
		{"single_quote", "'", `\'`},
		{"double_quote", `"`, `\"`},
		{"mixed", "a\tb\nc", `a\tb\nc`},
		{"ascii_control", "\x01", `\u{1}`},
		{"del", "\x7f", `\u{7f}`},
		{"c1_control", string(rune(0x80)), `\u{80}`},
		{"c1_control_end", string(rune(0x9f)), `\u{9f}`},
		{"no_break_space", string(rune(0xa0)), `\u{a0}`}, // 0xa0 is not printable in Rust's tables
		{"combining_mark_first", "\u0300", `\u{300}`},    // Mn: escaped as first char (grapheme extend)
		{"combining_mark_cont", "a\u0300", "a\u0300"},    // Mn: NOT escaped in continuation position
		{"enclosing_mark_first", "\u20DD", `\u{20dd}`},   // Me: escaped as first char (grapheme extend)
		{"spacing_mark", "\u0903", "\u0903"},             // Mc: NOT grapheme extend, printable → pass through
		{"normal_unicode", "\u00e9", "\u00e9"},           // é is printable
		{"emoji", "\U0001F600", "\U0001F600"},            // emoji is printable
		{"non_printable_high", "\uFFFE", `\u{fffe}`},     // non-printable
		{"soft_hyphen", string(rune(0xad)), `\u{ad}`},    // U+00AD SOFT HYPHEN: not printable in Rust
		{"katakana_voiced_first", "\uFF9E", `\u{ff9e}`},  // Lm but Grapheme_Extend: escaped as first char
		{"katakana_voiced_cont", "a\uFF9E", "a\uFF9E"},   // Lm but Grapheme_Extend: NOT escaped in continuation
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := EscapeString(tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}

func TestEscapeCharAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"combining_mark_always", "a\u0300", `a\u{300}`},   // Mn: always escaped with ESCAPE_ALL
		{"katakana_voiced_always", "a\uFF9E", `a\u{ff9e}`}, // Grapheme_Extend: always escaped
		{"spacing_mark_not_escaped", "a\u0903", "a\u0903"}, // Mc only: NOT grapheme extend → pass through
		{"plain", "hello", "hello"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := EscapeCharAll(tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}

func TestIsPrintable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   rune
		out  bool
	}{
		{"ascii_control_start", 0x01, false},
		{"ascii_control_end", 0x1f, false},
		{"space", ' ', true},
		{"tilde", '~', true},
		{"del", 0x7f, false},
		{"c1_start", 0x80, false},
		{"c1_end", 0x9f, false},
		{"no_break_space", 0xa0, false}, // not printable in Rust's tables
		{"soft_hyphen", 0xad, false},    // Rust singleton: not printable
		{"letter", 'A', true},
		{"combining_mark", 0x0300, true}, // printable (grapheme extend handled separately)
		{"printable_unicode", 0x00e9, true},
		{"non_printable_fffe", 0xfffe, false},
		{"bmp_singleton_hit", 0x0378, false},            // BMP singleton: not printable
		{"bmp_normal_printable", 0x0100, true},          // BMP normal range: printable
		{"smp_printable", 0x10000, true},                // SMP: printable
		{"smp_singleton_hit", 0x1000C, false},           // SMP singleton: not printable
		{"smp_normal_printable", 0x10100, true},         // SMP normal range: printable
		{"supplementary_printable", 0x20000, true},      // supplementary plane: printable
		{"supplementary_non_printable", 0x2a6e0, false}, // hardcoded supplementary range
		{"supplementary_non_printable_2", 0x2b81e, false},
		{"supplementary_non_printable_3", 0x2ceae, false},
		{"supplementary_non_printable_4", 0x2ebe1, false},
		{"supplementary_non_printable_5", 0x2ee5e, false},
		{"supplementary_non_printable_6", 0x2fa1e, false},
		{"supplementary_non_printable_7", 0x3134b, false},
		{"supplementary_non_printable_8", 0x3347a, false},
		{"supplementary_non_printable_9", 0xe01f0, false},
		{"supplementary_printable_after", 0xe0100, true}, // just inside last printable range
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := isPrintable(tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}

func TestIsGraphemeExtended(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   rune
		out  bool
	}{
		{"ascii_letter", 'A', false},
		{"ascii_control", 0x01, false},
		{"combining_grave", 0x0300, true},      // Mn
		{"enclosing_circle", 0x20DD, true},     // Me
		{"devanagari_visarga", 0x0903, false},  // Mc only, NOT Grapheme_Extend
		{"devanagari_vowel_aa", 0x093E, false}, // Mc only, NOT in Other_Grapheme_Extend
		{"zwj", 0x200D, false},                 // Cf, NOT in Grapheme_Extend
		{"katakana_voiced", 0xFF9E, true},      // Lm but Other_Grapheme_Extend
		{"katakana_semi_voiced", 0xFF9F, true}, // Lm but Other_Grapheme_Extend
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := isGraphemeExtended(tt.in)
			testutil.Equals(t, out, tt.out)
		})
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
