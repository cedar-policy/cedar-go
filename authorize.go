package cedar

import (
	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/types"
)

type Request = types.Request
type Decision = types.Decision
type Diagnostic = types.Diagnostic
type DiagnosticReason = types.DiagnosticReason
type DiagnosticError = types.DiagnosticError

const (
	Allow = types.Allow
	Deny  = types.Deny
)

// IsAuthorized uses the combination of the PolicySet and Entities to determine
// if the given Request to determine Decision and Diagnostic.
func (p PolicySet) IsAuthorized(entityMap Entities, req Request) (Decision, Diagnostic) {
	c := eval.InitEnv(&eval.Env{
		Entities:  entityMap,
		Principal: req.Principal,
		Action:    req.Action,
		Resource:  req.Resource,
		Context:   req.Context,
	})
	var diag Diagnostic
	var forbids []DiagnosticReason
	var permits []DiagnosticReason
	// Don't try to short circuit this.
	// - Even though single forbid means forbid
	// - All policy should be run to collect errors
	// - For permit, all permits must be run to collect annotations
	// - For forbid, forbids must be run to collect annotations
	for id, po := range p.policies {
		result, err := po.eval.Eval(c)
		if err != nil {
			diag.Errors = append(diag.Errors, DiagnosticError{PolicyID: id, Position: po.Position(), Message: err.Error()})
			continue
		}
		if !result {
			continue
		}
		if po.Effect() == Forbid {
			forbids = append(forbids, DiagnosticReason{PolicyID: id, Position: po.Position()})
		} else {
			permits = append(permits, DiagnosticReason{PolicyID: id, Position: po.Position()})
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
