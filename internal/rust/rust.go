package rust

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func nextRune(b []byte, i int) (rune, int, error) {
	ch, size := utf8.DecodeRune(b[i:])
	if ch == utf8.RuneError {
		return ch, i, fmt.Errorf("bad unicode rune")
	}
	return ch, i + size, nil
}

func parseHexEscape(b []byte, i int) (rune, int, error) {
	var ch rune
	var err error
	ch, i, err = nextRune(b, i)
	if err != nil {
		return 0, i, err
	}
	if !IsHexadecimal(ch) {
		return 0, i, fmt.Errorf("bad hex escape sequence")
	}
	res := digitVal(ch)
	ch, i, err = nextRune(b, i)
	if err != nil {
		return 0, i, err
	}
	if !IsHexadecimal(ch) {
		return 0, i, fmt.Errorf("bad hex escape sequence")
	}
	res = 16*res + digitVal(ch)
	if res > 127 {
		return 0, i, fmt.Errorf("bad hex escape sequence")
	}
	return rune(res), i, nil
}

func parseUnicodeEscape(b []byte, i int) (rune, int, error) {
	var ch rune
	var err error

	ch, i, err = nextRune(b, i)
	if err != nil {
		return 0, i, err
	}
	if ch != '{' {
		return 0, i, fmt.Errorf("bad unicode escape sequence")
	}

	digits := 0
	res := 0
	for {
		ch, i, err = nextRune(b, i)
		if err != nil {
			return 0, i, err
		}
		if ch == '}' {
			break
		}
		if !IsHexadecimal(ch) {
			return 0, i, fmt.Errorf("bad unicode escape sequence")
		}
		res = 16*res + digitVal(ch)
		digits++
	}

	if digits == 0 || digits > 6 || !utf8.ValidRune(rune(res)) {
		return 0, i, fmt.Errorf("bad unicode escape sequence")
	}

	return rune(res), i, nil
}

func unquote(s string) (string, error) {
	s = strings.TrimPrefix(s, "\"")
	s = strings.TrimSuffix(s, "\"")
	res, _, err := Unquote([]byte(s), false)
	return res, err
}

func Unquote(b []byte, star bool) (string, []byte, error) {
	var sb strings.Builder
	var ch rune
	var err error
	i := 0
	for i < len(b) {
		ch, i, err = nextRune(b, i)
		if err != nil {
			return "", nil, err
		}
		if star && ch == '*' {
			i--
			return sb.String(), b[i:], nil
		}
		if ch != '\\' {
			sb.WriteRune(ch)
			continue
		}
		ch, i, err = nextRune(b, i)
		if err != nil {
			return "", nil, err
		}
		switch ch {
		case 'n':
			sb.WriteRune('\n')
		case 'r':
			sb.WriteRune('\r')
		case 't':
			sb.WriteRune('\t')
		case '\\':
			sb.WriteRune('\\')
		case '0':
			sb.WriteRune('\x00')
		case '\'':
			sb.WriteRune('\'')
		case '"':
			sb.WriteRune('"')
		case 'x':
			ch, i, err = parseHexEscape(b, i)
			if err != nil {
				return "", nil, err
			}
			sb.WriteRune(ch)
		case 'u':
			ch, i, err = parseUnicodeEscape(b, i)
			if err != nil {
				return "", nil, err
			}
			sb.WriteRune(ch)
		case '*':
			if !star {
				return "", nil, fmt.Errorf("bad char escape")
			}
			sb.WriteRune('*')
		default:
			return "", nil, fmt.Errorf("bad char escape")
		}
	}
	return sb.String(), b[i:], nil
}

func IsHexadecimal(ch rune) bool {
	return IsDecimal(ch) || ('a' <= lower(ch) && lower(ch) <= 'f')
}

func lower(ch rune) rune     { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter
func IsDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= lower(ch) && lower(ch) <= 'f':
		return int(lower(ch) - 'a' + 10)
	}
	return 16 // larger than any legal digit val
}

// EscapeStringDebug escapes a string using Rust's debug-style escaping.
func EscapeStringDebug(s string) string {
	return escapeStringMode(s, true)
}

// EscapeStringDisplay escapes a string using Rust's display-style escaping.
// This keeps printable unicode (including combining marks) intact while
// escaping control and non-printable runes.
func EscapeStringDisplay(s string) string {
	return escapeStringMode(s, false)
}

// EscapeString is kept as a compatibility alias for debug-style escaping.
func EscapeString(s string) string {
	return EscapeStringDebug(s)
}

func escapeStringMode(s string, debug bool) string {
	var b []byte
	for _, r := range s {
		switch r {
		case 0:
			b = append(b, `\0`...)
		case '\t':
			b = append(b, `\t`...)
		case '\n':
			b = append(b, `\n`...)
		case '\r':
			b = append(b, `\r`...)
		case '\\':
			b = append(b, `\\`...)
		case '\'':
			b = append(b, `\'`...)
		case '"':
			b = append(b, `\"`...)
		default:
			if shouldEscapeByMode(r, debug) {
				b = append(b, fmt.Sprintf(`\u{%x}`, r)...)
			} else {
				b = append(b, []byte(string(r))...)
			}
		}
	}
	return string(b)
}

// ShouldEscape returns true if a rune should be escaped as \u{xx} in
// debug-style escaping. This includes combining marks.
func ShouldEscape(r rune) bool {
	return shouldEscapeByMode(r, true)
}

func shouldEscapeByMode(r rune, debug bool) bool {
	if r < 0x20 || r == 0x7f {
		return true
	}
	if r >= 0x80 && r <= 0x9f {
		return true
	}
	// Rust debug escaping escapes nonspacing/enclosing marks; display escaping
	// keeps them when otherwise printable.
	if debug && (unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r)) {
		return true
	}
	if !unicode.IsPrint(r) {
		return true
	}
	return false
}
