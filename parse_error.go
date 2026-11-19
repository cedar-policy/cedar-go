package cedar

import (
	"errors"

	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/internal/scanner"
)

// A ParseError is returned when Cedar policy text cannot be parsed. It reports where in the input the problem was
// found so callers can surface it programmatically, for example to highlight the offending text in an editor.
//
// Use [errors.As] to retrieve a ParseError from the error returned by [Policy.UnmarshalCedar],
// [PolicyList.UnmarshalCedar], [NewPolicySetFromBytes], [NewPolicyListFromBytes] or [Decoder.Decode].
// Position is the location the parser was at when it detected the problem, which for syntax errors is the token it
// was examining at that point.
type ParseError struct {
	// Position is the location of the problem in the input. Filename is populated by [NewPolicySetFromBytes] and
	// [NewPolicyListFromBytes]; Line and Column start at 1 and Offset starts at 0.
	Position Position

	// Message describes the problem without any location information.
	Message string

	err error
}

// Error implements the error interface.
func (e *ParseError) Error() string { return e.err.Error() }

// Unwrap returns the underlying error.
func (e *ParseError) Unwrap() error { return e.err }

// wrapParseError converts errors produced by the internal scanner and parser into a *ParseError. Any other error is
// returned unchanged.
func wrapParseError(err error) error {
	var scanErr *scanner.Error
	if errors.As(err, &scanErr) {
		return &ParseError{
			Position: Position{Offset: scanErr.Pos.Offset, Line: scanErr.Pos.Line, Column: scanErr.Pos.Column},
			Message:  scanErr.Message,
			err:      err,
		}
	}
	var parseErr *parser.Error
	if errors.As(err, &parseErr) {
		return &ParseError{
			Position: Position{Offset: parseErr.Pos.Offset, Line: parseErr.Pos.Line, Column: parseErr.Pos.Column},
			Message:  parseErr.Err.Error(),
			err:      err,
		}
	}
	return err
}
