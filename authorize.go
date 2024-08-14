package cedar

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/entities"
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

type evalContext = eval.Context

type evaler = eval.Evaler

// IsAuthorized uses the combination of the PolicySet and Entities to determine
// if the given Request to determine Decision and Diagnostic.
func (p PolicySet) IsAuthorized(entityMap entities.Entities, req Request) (Decision, Diagnostic) {
	c := &evalContext{
		Entities:  entityMap,
		Principal: req.Principal,
		Action:    req.Action,
		Resource:  req.Resource,
		Context:   req.Context,
	}
	var diag Diagnostic
	var gotForbid bool
	var forbidReasons []Reason
	var gotPermit bool
	var permitReasons []Reason
	// Don't try to short circuit this.
	// - Even though single forbid means forbid
	// - All policy should be run to collect errors
	// - For permit, all permits must be run to collect annotations
	// - For forbid, forbids must be run to collect annotations
	for id, po := range p.policies {
		v, err := po.eval.Eval(c)
		if err != nil {
			diag.Errors = append(diag.Errors, Error{PolicyID: id, Position: po.Position, Message: err.Error()})
			continue
		}
		vb, err := types.ValueToBool(v)
		if err != nil {
			// should never happen, maybe remove this case
			diag.Errors = append(diag.Errors, Error{PolicyID: id, Position: po.Position, Message: err.Error()})
			continue
		}
		if !vb {
			continue
		}
		if po.Effect == Forbid {
			forbidReasons = append(forbidReasons, Reason{PolicyID: id, Position: po.Position})
			gotForbid = true
		} else {
			permitReasons = append(permitReasons, Reason{PolicyID: id, Position: po.Position})
			gotPermit = true
		}
	}
	if gotForbid {
		diag.Reasons = forbidReasons
	} else if gotPermit {
		diag.Reasons = permitReasons
	}
	return Decision(gotPermit && !gotForbid), diag
}
