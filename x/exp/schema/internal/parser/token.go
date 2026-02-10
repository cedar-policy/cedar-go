package parser

import (
	"fmt"
	"unicode/utf8"

	"github.com/cedar-policy/cedar-go/internal/rust"
)

type tokenType int

const (
	tokenEOF tokenType = iota
	tokenIdent
	tokenString
	tokenAt
	tokenLBrace
	tokenRBrace
	tokenLBracket
	tokenRBracket
	tokenLAngle
	tokenRAngle
	tokenLParen
	tokenRParen
	tokenComma
	tokenSemicolon
	tokenColon
	tokenDoubleColon
	tokenQuestion
	tokenEquals
)

type position struct {
	Filename string
	Line     int
	Column   int
	Offset   int
}

func (p position) String() string {
	name := p.Filename
	if name == "" {
		name = "<input>"
	}
	return fmt.Sprintf("%s:%d:%d", name, p.Line, p.Column)
}

type token struct {
	Type tokenType
	Pos  position
	Text string
}

type lexer struct {
	src      []byte
	pos      int
	line     int
	col      int
	filename string
}

func newLexer(filename string, src []byte) *lexer {
	return &lexer{
		src:      src,
		line:     1,
		col:      1,
		filename: filename,
	}
}

func (l *lexer) position() position {
	return position{
		Filename: l.filename,
		Line:     l.line,
		Column:   l.col,
		Offset:   l.pos,
	}
}

func (l *lexer) errorf(pos position, format string, args ...any) error {
	return fmt.Errorf("%s: %s", pos, fmt.Sprintf(format, args...))
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.src) {
		return -1
	}
	r, _ := utf8.DecodeRune(l.src[l.pos:])
	return r
}

func (l *lexer) advance() rune {
	if l.pos >= len(l.src) {
		return -1
	}
	r, size := utf8.DecodeRune(l.src[l.pos:])
	l.pos += size
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func (l *lexer) skipWhitespaceAndComments() error {
	for l.pos < len(l.src) {
		r := l.peek()
		if r == ' ' || r == '\t' || r == '\r' || r == '\n' {
			l.advance()
			continue
		}
		if r == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '/' {
			l.advance()
			l.advance()
			for l.pos < len(l.src) && l.peek() != '\n' {
				l.advance()
			}
			continue
		}
		if r == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '*' {
			pos := l.position()
			l.advance()
			l.advance()
			for {
				if l.pos >= len(l.src) {
					return l.errorf(pos, "unterminated block comment")
				}
				if l.peek() == '*' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '/' {
					l.advance()
					l.advance()
					break
				}
				l.advance()
			}
			continue
		}
		break
	}
	return nil
}

func isIdentStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func isIdentContinue(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}

func (l *lexer) scanIdent() string {
	start := l.pos
	l.advance()
	for l.pos < len(l.src) && isIdentContinue(l.peek()) {
		l.advance()
	}
	return string(l.src[start:l.pos])
}

func (l *lexer) scanString() (string, error) {
	pos := l.position()
	l.advance() // skip opening quote
	start := l.pos
	for l.pos < len(l.src) {
		r := l.peek()
		if r == '"' {
			raw := l.src[start:l.pos]
			l.advance() // skip closing quote
			unescaped, _, err := rust.Unquote(raw, false)
			if err != nil {
				return "", l.errorf(pos, "invalid string escape: %v", err)
			}
			return unescaped, nil
		}
		if r == '\n' {
			return "", l.errorf(pos, "unterminated string literal")
		}
		if r == '\\' {
			l.advance()
			if l.pos >= len(l.src) {
				return "", l.errorf(pos, "unterminated string literal")
			}
		}
		l.advance()
	}
	return "", l.errorf(pos, "unterminated string literal")
}

func (l *lexer) next() (token, error) {
	if err := l.skipWhitespaceAndComments(); err != nil {
		return token{}, err
	}

	pos := l.position()

	if l.pos >= len(l.src) {
		return token{Type: tokenEOF, Pos: pos}, nil
	}

	r := l.peek()

	if isIdentStart(r) {
		text := l.scanIdent()
		return token{Type: tokenIdent, Pos: pos, Text: text}, nil
	}

	if r == '"' {
		text, err := l.scanString()
		if err != nil {
			return token{}, err
		}
		return token{Type: tokenString, Pos: pos, Text: text}, nil
	}

	l.advance()
	switch r {
	case '@':
		return token{Type: tokenAt, Pos: pos, Text: "@"}, nil
	case '{':
		return token{Type: tokenLBrace, Pos: pos, Text: "{"}, nil
	case '}':
		return token{Type: tokenRBrace, Pos: pos, Text: "}"}, nil
	case '[':
		return token{Type: tokenLBracket, Pos: pos, Text: "["}, nil
	case ']':
		return token{Type: tokenRBracket, Pos: pos, Text: "]"}, nil
	case '<':
		return token{Type: tokenLAngle, Pos: pos, Text: "<"}, nil
	case '>':
		return token{Type: tokenRAngle, Pos: pos, Text: ">"}, nil
	case '(':
		return token{Type: tokenLParen, Pos: pos, Text: "("}, nil
	case ')':
		return token{Type: tokenRParen, Pos: pos, Text: ")"}, nil
	case ',':
		return token{Type: tokenComma, Pos: pos, Text: ","}, nil
	case ';':
		return token{Type: tokenSemicolon, Pos: pos, Text: ";"}, nil
	case '?':
		return token{Type: tokenQuestion, Pos: pos, Text: "?"}, nil
	case '=':
		return token{Type: tokenEquals, Pos: pos, Text: "="}, nil
	case ':':
		if l.peek() == ':' {
			l.advance()
			return token{Type: tokenDoubleColon, Pos: pos, Text: "::"}, nil
		}
		return token{Type: tokenColon, Pos: pos, Text: ":"}, nil
	default:
		return token{}, l.errorf(pos, "unexpected character %q", r)
	}
}
