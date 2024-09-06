package cedar

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/types"
)

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
	Reasons []Reason `json:"reasons,omitempty"`
	Errors  []Error  `json:"errors,omitempty"`
}

// An Error details the Policy index within a PolicySet, the Position within the
// text document, and the resulting error message.
type Error struct {
	PolicyID PolicyID `json:"policy"`
	Position Position `json:"position"`
	Message  string   `json:"message"`
}

func (e Error) String() string {
	return fmt.Sprintf("while evaluating policy `%v`: %v", e.PolicyID, e.Message)
}

// A Reason details the Policy index within a PolicySet, and the Position within
// the text document.
type Reason struct {
	PolicyID PolicyID `json:"policy"`
	Position Position `json:"position"`
}

// A Request is the Principal, Action, Resource, and Context portion of an
// authorization request.
type Request struct {
	Principal types.EntityUID `json:"principal"`
	Action    types.EntityUID `json:"action"`
	Resource  types.EntityUID `json:"resource"`
	Context   types.Record    `json:"context"`
}

// IsAuthorized uses the combination of the PolicySet and Entities to determine
// if the given Request to determine Decision and Diagnostic.
func (p PolicySet) IsAuthorized(entityMap types.Entities, req Request) (Decision, Diagnostic) {
	c := eval.PrepContext(&eval.Context{
		Entities:  entityMap,
		Principal: req.Principal,
		Action:    req.Action,
		Resource:  req.Resource,
		Context:   req.Context,
	})
	var diag Diagnostic
	var forbids []Reason
	var permits []Reason
	// Don't try to short circuit this.
	// - Even though single forbid means forbid
	// - All policy should be run to collect errors
	// - For permit, all permits must be run to collect annotations
	// - For forbid, forbids must be run to collect annotations
	for id, po := range p.policies {
		result, err := po.eval.Eval(c)
		if err != nil {
			diag.Errors = append(diag.Errors, Error{PolicyID: id, Position: po.Position(), Message: err.Error()})
			continue
		}
		if !result {
			continue
		}
		if po.Effect() == Forbid {
			forbids = append(forbids, Reason{PolicyID: id, Position: po.Position()})
		} else {
			permits = append(permits, Reason{PolicyID: id, Position: po.Position()})
		}
	}
	if len(forbids) > 0 {
		diag.Reasons = forbids
		return Deny, diag
	}
	if len(permits) > 0 {
		diag.Reasons = permits
		return Allow, diag
	}
	return Deny, diag
}
