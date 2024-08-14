package rust

import (
	"fmt"
	"strings"
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
