package parser

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/cedar-policy/cedar-go/internal/rust"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

//go:generate moq -pkg parser -fmt goimports -out tokenize_mocks_test.go . reader

// This type alias is for test purposes only.
type reader = io.Reader

type TokenType int

const (
	TokenEOF = TokenType(iota)
	TokenIdent
	TokenInt
	TokenReservedKeyword
	TokenString
	TokenOperator
	TokenUnknown
)

type Token struct {
	Type TokenType
	Pos  Position
	Text string
}

// N.B. "is" is included here for compatibility with the Rust implementation. The Cedar specification does not list
// "is" as a reserved keyword
var reservedKeywords = []string{"true", "false", "if", "then", "else", "in", "like", "has", "is"}

func (t Token) isEOF() bool {
	return t.Type == TokenEOF
}

func (t Token) isIdent() bool {
	return t.Type == TokenIdent
}

func (t Token) isInt() bool {
	return t.Type == TokenInt
}

func (t Token) isReservedKeyword() bool {
	return t.Type == TokenReservedKeyword
}

func (t Token) isString() bool {
	return t.Type == TokenString
}

func (t Token) stringValue() (string, error) {
	s := t.Text
	s = strings.TrimPrefix(s, "\"")
	s = strings.TrimSuffix(s, "\"")
	b := []byte(s)
	res, _, err := rust.Unquote(b, false)
	return res, err
}

func (t Token) intValue() (int64, error) {
	return strconv.ParseInt(t.Text, 10, 64)
}

func Tokenize(src []byte) ([]Token, error) {
	var res []Token
	var s scanner
	s.Init(bytes.NewBuffer(src))
	for tok := s.nextToken(); s.err == nil && tok.Type != TokenEOF; tok = s.nextToken() {
		res = append(res, tok)
	}
	if s.err != nil {
		return nil, s.err
	}
	res = append(res, Token{Type: TokenEOF, Pos: s.position})
	return res, nil
}

type Position ast.Position

func (pos Position) String() string {
	return fmt.Sprintf("<input>:%d:%d", pos.Line, pos.Column)
}

const (
	specialRuneEOF = rune(-(iota + 1))
	specialRuneBOF
)

const bufLen = 1024 // at least utf8.UTFMax

// A scanner implements reading of Unicode characters and tokens from an io.Reader.
type scanner struct {
	// Input
	src io.Reader

	// Source buffer
	srcBuf [bufLen + 1]byte // +1 for sentinel for common case of s.next()
	srcPos int              // reading position (srcBuf index)
	srcEnd int              // source end (srcBuf index)

	// Source position
	srcBufOffset int // byte offset of srcBuf[0] in source
	line         int // line count
	column       int // character count
	lastLineLen  int // length of last line in characters (for correct column reporting)
	lastCharLen  int // length of last character in bytes

	// Token text buffer
	// Typically, token text is stored completely in srcBuf, but in general
	// the token text's head may be buffered in tokBuf while the token text's
	// tail is stored in srcBuf.
	tokBuf bytes.Buffer // token text head that is not in srcBuf anymore
	tokPos int          // token text tail position (srcBuf index); valid if >= 0
	tokEnd int          // token text tail end (srcBuf index)

	// One character look-ahead
	ch rune // character before current srcPos

	// Last error encountered by nextToken.
	err error

	// Start position of most recently scanned token; set by nextToken.
	// Calling Init or Next invalidates the position (Line == 0).
	// If an error is reported (via Error) and position is invalid,
	// the scanner is not inside a token. Call Pos to obtain an error
	// position in that case, or to obtain the position immediately
	// after the most recently scanned token.
	position Position
}

// Init initializes a Scanner with a new source and returns s.
func (s *scanner) Init(src io.Reader) *scanner {
	s.src = src

	// initialize source buffer
	// (the first call to next() will fill it by calling src.Read)
	s.srcBuf[0] = utf8.RuneSelf // sentinel
	s.srcPos = 0
	s.srcEnd = 0

	// initialize source Position
	s.srcBufOffset = 0
	s.line = 1
	s.column = 0
	s.lastLineLen = 0
	s.lastCharLen = 0

	// initialize token text buffer
	// (required for first call to next()).
	s.tokPos = -1

	// initialize one character look-ahead
	s.ch = specialRuneBOF // no char read yet, not EOF

	// initialize public fields
	s.position.Line = 0 // invalidate token Position

	return s
}

// next reads and returns the next Unicode character. It is designed such
// that only a minimal amount of work needs to be done in the common ASCII
// case (one test to check for both ASCII and end-of-buffer, and one test
// to check for newlines).
func (s *scanner) next() rune {
	ch, width := rune(s.srcBuf[s.srcPos]), 1

	if ch >= utf8.RuneSelf {
		// uncommon case: not ASCII or not enough bytes
		for s.srcPos+utf8.UTFMax > s.srcEnd && !utf8.FullRune(s.srcBuf[s.srcPos:s.srcEnd]) {
			// not enough bytes: read some more, but first
			// save away token text if any
			if s.tokPos >= 0 {
				s.tokBuf.Write(s.srcBuf[s.tokPos:s.srcPos])
				s.tokPos = 0
				// s.tokEnd is set by nextToken()
			}
			// move unread bytes to beginning of buffer
			copy(s.srcBuf[0:], s.srcBuf[s.srcPos:s.srcEnd])
			s.srcBufOffset += s.srcPos
			// read more bytes
			// (an io.Reader must return io.EOF when it reaches
			// the end of what it is reading - simply returning
			// n == 0 will make this loop retry forever; but the
			// error is in the reader implementation in that case)
			i := s.srcEnd - s.srcPos
			n, err := s.src.Read(s.srcBuf[i:bufLen])
			s.srcPos = 0
			s.srcEnd = i + n
			s.srcBuf[s.srcEnd] = utf8.RuneSelf // sentinel
			if err != nil {
				if err != io.EOF {
					s.error(err.Error())
				}
				if s.srcEnd == 0 {
					if s.lastCharLen > 0 {
						// previous character was not EOF
						s.column++
					}
					s.lastCharLen = 0
					return specialRuneEOF
				}
				// If err == EOF, we won't be getting more
				// bytes; break to avoid infinite loop. If
				// err is something else, we don't know if
				// we can get more bytes; thus also break.
				break
			}
		}
		// at least one byte
		ch = rune(s.srcBuf[s.srcPos])
		if ch >= utf8.RuneSelf {
			// uncommon case: not ASCII
			ch, width = utf8.DecodeRune(s.srcBuf[s.srcPos:s.srcEnd])
			if ch == utf8.RuneError && width == 1 {
				// advance for correct error Position
				s.srcPos += width
				s.lastCharLen = width
				s.column++
				s.error("invalid UTF-8 encoding")
				return ch
			}
		}
	}

	// advance
	s.srcPos += width
	s.lastCharLen = width
	s.column++

	// special situations
	switch ch {
	case 0:
		// for compatibility with other tools
		s.error("invalid character NUL")
	case '\n':
		s.line++
		s.lastLineLen = s.column
		s.column = 0
	}

	return ch
}

func (s *scanner) error(msg string) {
	s.tokEnd = s.srcPos - s.lastCharLen // make sure token text is terminated
	s.err = fmt.Errorf("%v: %v", s.position, msg)
}

func isASCIILetter(ch rune) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
}

func isASCIINumber(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isIdentRune(ch rune, first bool) bool {
	return ch == '_' || isASCIILetter(ch) || isASCIINumber(ch) && !first
}

func (s *scanner) scanIdentifier() rune {
	// we know the zeroth rune is OK; start scanning at the next one
	ch := s.next()
	for isIdentRune(ch, false) {
		ch = s.next()
	}
	return ch
}

func (s *scanner) scanInteger(ch rune) rune {
	for rust.IsDecimal(ch) {
		ch = s.next()
	}
	return ch
}

func (s *scanner) scanHexDigits(ch rune, minDigits, maxDigits int) rune {
	n := 0
	for n < maxDigits && rust.IsHexadecimal(ch) {
		ch = s.next()
		n++
	}
	if n < minDigits || n > maxDigits {
		s.error("invalid char escape")
	}
	return ch
}

func (s *scanner) scanEscape() rune {
	ch := s.next() // read character after '/'
	switch ch {
	case 'n', 'r', 't', '\\', '0', '\'', '"', '*':
		// nothing to do
		ch = s.next()
	case 'x':
		ch = s.scanHexDigits(s.next(), 2, 2)
	case 'u':
		ch = s.next()
		if ch != '{' {
			s.error("invalid char escape")
			return ch
		}
		ch = s.scanHexDigits(s.next(), 1, 6)
		if ch != '}' {
			s.error("invalid char escape")
			return ch
		}
		ch = s.next()
	default:
		s.error("invalid char escape")
	}
	return ch
}

func (s *scanner) scanString() (n int) {
	ch := s.next() // read character after quote
	for ch != '"' {
		if ch == '\n' || ch < 0 {
			s.error("literal not terminated")
			return
		}
		if ch == '\\' {
			ch = s.scanEscape()
		} else {
			ch = s.next()
		}
		n++
	}
	return
}

func (s *scanner) scanComment(ch rune) rune {
	// ch == '/' || ch == '*'
	if ch == '/' {
		// line comment
		ch = s.next() // read character after "//"
		for ch != '\n' && ch >= 0 {
			ch = s.next()
		}
		return ch
	}

	// general comment
	ch = s.next() // read character after "/*"
	for {
		if ch < 0 {
			s.error("comment not terminated")
			break
		}
		ch0 := ch
		ch = s.next()
		if ch0 == '*' && ch == '/' {
			ch = s.next()
			break
		}
	}
	return ch
}

func (s *scanner) scanOperator(ch0, ch rune) (TokenType, rune) {
	switch ch0 {
	case '@', '.', ',', ';', '(', ')', '{', '}', '[', ']', '+', '-', '*':
	case ':':
		if ch == ':' {
			ch = s.next()
		}
	case '!', '<', '>':
		if ch == '=' {
			ch = s.next()
		}
	case '=':
		if ch != '=' {
			return TokenUnknown, ch
		}
		ch = s.next()
	case '|':
		if ch != '|' {
			return TokenUnknown, ch
		}
		ch = s.next()
	case '&':
		if ch != '&' {
			return TokenUnknown, ch
		}
		ch = s.next()
	default:
		return TokenUnknown, ch
	}
	return TokenOperator, ch
}

func isWhitespace(c rune) bool {
	switch c {
	case '\t', '\n', '\r', ' ':
		return true
	default:
		return false
	}
}

// nextToken reads the next token or Unicode character from source and returns
// it. It returns specialRuneEOF at the end of the source. It reports scanner
// errors (read and token errors) by calling s.Error, if not nil; otherwise it
// prints an error message to os.Stderr.
func (s *scanner) nextToken() Token {
	if s.ch == specialRuneBOF {
		s.ch = s.next()
	}

	ch := s.ch

	// reset token text Position
	s.tokPos = -1
	s.position.Line = 0

redo:
	// skip white space
	for isWhitespace(ch) {
		ch = s.next()
	}

	// start collecting token text
	s.tokBuf.Reset()
	s.tokPos = s.srcPos - s.lastCharLen

	// set token Position
	s.position.Offset = s.srcBufOffset + s.tokPos
	if s.column > 0 {
		// common case: last character was not a '\n'
		s.position.Line = s.line
		s.position.Column = s.column
	} else {
		// last character was a '\n'
		// (we cannot be at the beginning of the source
		// since we have called next() at least once)
		s.position.Line = s.line - 1
		s.position.Column = s.lastLineLen
	}

	// determine token value
	var tt TokenType
	switch {
	case ch == specialRuneEOF:
		tt = TokenEOF
	case isIdentRune(ch, true):
		ch = s.scanIdentifier()
		tt = TokenIdent
	case rust.IsDecimal(ch):
		ch = s.scanInteger(ch)
		tt = TokenInt
	case ch == '"':
		s.scanString()
		ch = s.next()
		tt = TokenString
	case ch == '/':
		ch0 := ch
		ch = s.next()
		if ch == '/' || ch == '*' {
			s.tokPos = -1 // don't collect token text
			ch = s.scanComment(ch)
			goto redo
		}
		tt, ch = s.scanOperator(ch0, ch)
	default:
		tt, ch = s.scanOperator(ch, s.next())
	}

	// end of token text
	s.tokEnd = s.srcPos - s.lastCharLen
	s.ch = ch

	// last minute check for reserved keywords
	text := s.tokenText()
	if tt == TokenIdent && slices.Contains(reservedKeywords, text) {
		tt = TokenReservedKeyword
	}

	return Token{
		Type: tt,
		Pos:  s.position,
		Text: text,
	}
}

// tokenText returns the string corresponding to the most recently scanned token.
// Valid after calling nextToken and in calls of Scanner.Error.
func (s *scanner) tokenText() string {
	if s.tokPos < 0 {
		// no token text
		return ""
	}

	if s.tokBuf.Len() == 0 {
		// common case: the entire token text is still in srcBuf
		return string(s.srcBuf[s.tokPos:s.tokEnd])
	}

	// part of the token text was saved in tokBuf: save the rest in
	// tokBuf as well and return its content
	s.tokBuf.Write(s.srcBuf[s.tokPos:s.tokEnd])
	s.tokPos = s.tokEnd // ensure idempotency of TokenText() call
	return s.tokBuf.String()
}
