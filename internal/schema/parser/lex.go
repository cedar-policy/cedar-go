package parser

// We use re2go to automatically generate a correct lexer.
// You can install re2go here: https://re2c.org/manual/manual_go.html
// Unless you are changing cedarschema.re, you should not need to rerun this re2go
// and regenerate cedarschema.re.
//go:generate re2go cedarschema.re -o cedarschema.go -i

import (
	"errors"

	"github.com/cedar-policy/cedar-go/internal/schema/token"
)

var (
	ErrUnrecognizedToken   = errors.New("unrecognized token")
	ErrInvalidString       = errors.New("invalid string")
	ErrUnterminatedString  = errors.New("unterminated string")
	ErrUnterminatedComment = errors.New("unterminated multiline comment")
)

type TokenType int

type Token struct {
	Pos  token.Position
	Type token.Type
	Lit  string
}

func (t Token) String() string {
	if t.Lit != "" {
		return t.Lit
	} else {
		return t.Type.String()
	}
}

type Lexer struct {
	input     []byte
	cursor    int // internal use by lexer
	token     int // marks the start of the currently scanned token
	prevToken Token

	lineStart int            // byte offset from start of last line
	pos       token.Position // marks position of the scanner

	Errors token.Errors
}

func (l *Lexer) error(pos token.Position, err error) {
	l.Errors = append(l.Errors, token.Error{Pos: pos, Err: err})
}

// NewLexer creates a new lexer for the given input.
//
// All tokens returned from NextToken will have the filename set to the given filename.
// If the input byte array is not null-terminated, a NULL character will automatically be added to the end.
// If the input contains null characters, the lexer will stop at the first one.
func NewLexer(filename string, input []byte) *Lexer {
	if len(input) == 0 || input[len(input)-1] != '\x00' {
		// termination char, faster copying than branching every time in the lexer
		input = append(input, '\x00')
	}
	return &Lexer{input: input, pos: token.Position{Filename: filename, Line: 1}}
}

// All will scan all tokens from input until it sees the EOF (NULL) token.
func (l *Lexer) All() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

func (l *Lexer) literal() string { return string(l.input[l.token:l.cursor]) }

// NextToken returns a single token from the input.
//
// If the returned token is EOF, then NextToken will always return EOF on subsequent calls.
func (l *Lexer) NextToken() (tok Token) {
	pos, typ, lit, err := l.lex()
	if err != nil {
		l.error(pos, err)
	}

	tok.Pos = pos
	tok.Lit = lit
	tok.Type = typ
	if tok.Type != token.COMMENT {
		l.prevToken = tok
	}
	return
}
