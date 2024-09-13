package types

import "fmt"

// A Request is the Principal, Action, Resource, and Context portion of an
// authorization request.
type Request struct {
	Principal EntityUID `json:"principal"`
	Action    EntityUID `json:"action"`
	Resource  EntityUID `json:"resource"`
	Context   Record    `json:"context"`
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

func (a *Decision) UnmarshalJSON(b []byte) error {
	*a = string(b) == `"allow"`
	return nil
}

// A Diagnostic details the errors and reasons for an authorization decision.
type Diagnostic struct {
	Reasons []DiagnosticReason `json:"reasons,omitempty"`
	Errors  []DiagnosticError  `json:"errors,omitempty"`
}

// A DiagnosticReason details the PolicyID within a PolicySet and the Position within the text document, if applicable.
type DiagnosticReason struct {
	PolicyID PolicyID `json:"policy"`
	Position Position `json:"position"`
}

// An DiagnosticError details the PolicyID within a PolicySet, the Position within the text document if applicable, and
// the resulting error message.
type DiagnosticError struct {
	PolicyID PolicyID `json:"policy"`
	Position Position `json:"position"`
	Message  string   `json:"message"`
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
