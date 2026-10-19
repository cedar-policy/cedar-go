package types

import (
	"fmt"

	"github.com/go-json-experiment/json"
)

// A Request is the Principal, Action, Resource, and Context portion of an
// authorization request.
type Request struct {
	Principal EntityUID `json:"principal"`
	Action    EntityUID `json:"action"`
	Resource  EntityUID `json:"resource"`
	Context   Record    `json:"context"`
}

// UnmarshalJSON parses a Request from JSON using encoding/json/v2 semantics,
// even when the caller decodes with encoding/json (v1). Without this method,
// v1 would decode the struct with its own lenient rules (e.g. case-insensitive
// names, duplicate keys, etc). Rust Cedar rejects all of those. Decoding
// through an unexported alias type keeps v2's stricter behavior and is
// necessary to avoid recursing back into this method.
func (r *Request) UnmarshalJSON(b []byte) error {
	type request Request
	return json.Unmarshal(b, (*request)(r))
}

// Equal comapres two requests for sameness
func (r Request) Equal(other Request) bool {
	return (r.Principal == other.Principal &&
		r.Action == other.Action &&
		r.Resource == other.Resource &&
		r.Context.Equal(other.Context))
}

// A Decision is the result of the authorization.
type Decision bool

// Each authorization results in one of these Decisions.
const (
	Allow = Decision(true)
	Deny  = Decision(false)
)

func (a Decision) String() string {
	if a {
		return "allow"
	}
	return "deny"
}

func (a Decision) MarshalJSON() ([]byte, error) { return []byte(`"` + a.String() + `"`), nil }

// UnmarshalJSON decodes a Decision from "allow" or "deny". Anything else is an
// error, rather than Deny.
func (a *Decision) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case `"allow"`:
		*a = Allow
	case `"deny"`:
		*a = Deny
	default:
		return fmt.Errorf(`invalid decision %s: must be "allow" or "deny"`, b)
	}
	return nil
}

// A Diagnostic details the errors and reasons for an authorization decision.
type Diagnostic struct {
	Reasons []DiagnosticReason `json:"reasons,omitempty"`
	Errors  []DiagnosticError  `json:"errors,omitempty"`
}

// UnmarshalJSON parses a Diagnostic from JSON using encoding/json/v2 semantics,
// even when the caller decodes with encoding/json (v1). See
// Request.UnmarshalJSON for why.
func (d *Diagnostic) UnmarshalJSON(b []byte) error {
	type diagnostic Diagnostic
	return json.Unmarshal(b, (*diagnostic)(d))
}

// A DiagnosticReason details the PolicyID within a PolicySet and the Position within the text document, if applicable.
type DiagnosticReason struct {
	PolicyID PolicyID `json:"policy"`
	Position Position `json:"position"`
}

// UnmarshalJSON parses a DiagnosticReason from JSON using encoding/json/v2
// semantics, even when the caller decodes with encoding/json (v1). See
// Request.UnmarshalJSON for why.
func (d *DiagnosticReason) UnmarshalJSON(b []byte) error {
	type diagnosticReason DiagnosticReason
	return json.Unmarshal(b, (*diagnosticReason)(d))
}

// An DiagnosticError details the PolicyID within a PolicySet, the Position within the text document if applicable, and
// the resulting error message.
type DiagnosticError struct {
	PolicyID PolicyID `json:"policy"`
	Position Position `json:"position"`
	Message  string   `json:"message"`
}

// UnmarshalJSON parses a DiagnosticError from JSON using encoding/json/v2
// semantics, even when the caller decodes with encoding/json (v1). See
// Request.UnmarshalJSON for why.
func (e *DiagnosticError) UnmarshalJSON(b []byte) error {
	type diagnosticError DiagnosticError
	return json.Unmarshal(b, (*diagnosticError)(e))
}

func (e DiagnosticError) String() string {
	return fmt.Sprintf("while evaluating policy `%v`: %v", e.PolicyID, e.Message)
}

// PolicyID is a string identifier for the policy within the PolicySet
type PolicyID string

// A Position describes an arbitrary source position including the file, line, and column location.
type Position struct {
	// Filename is the optional name of the source file for the enclosing policy, "" if the source is unknown or not a named file
	Filename string `json:"filename"`

	// Offset is the byte offset, starting at 0
	Offset int `json:"offset"`

	// Line is the line number, starting at 1
	Line int `json:"line"`

	// Column is the column number, starting at 1 (character count per line)
	Column int `json:"column"`
}

// UnmarshalJSON parses a Position from JSON using encoding/json/v2 semantics,
// even when the caller decodes with encoding/json (v1). See
// Request.UnmarshalJSON for why.
func (p *Position) UnmarshalJSON(b []byte) error {
	type position Position
	return json.Unmarshal(b, (*position)(p))
}

// An Effect specifies the intent of the policy, to either permit or forbid any
// request that matches the scope and conditions specified in the policy.
type Effect bool

// Each Policy has a Permit or Forbid effect that is determined during parsing.
const (
	Permit = Effect(true)
	Forbid = Effect(false)
)

// An Annotations is a map of key, value pairs found in the policy. Annotations
// have no impact on policy evaluation.
type Annotations map[Ident]String
