package templates

import (
	"github.com/cedar-policy/cedar-go"
	"iter"

	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/types"
)

// PolicyIterator is an interface which abstracts an iterable set of policies.
type PolicyIterator interface {
	// All returns an iterator over all the policies in the set
	All() iter.Seq2[cedar.PolicyID, *Policy]
}

// Authorize uses the combination of the PolicySet and Entities to determine
// if the given Request to determine Decision and Diagnostic.
func Authorize(policies PolicyIterator, entities types.EntityGetter, req cedar.Request) (cedar.Decision, cedar.Diagnostic) {
	if entities == nil {
		var zero types.EntityMap
		entities = zero
	}
	env := eval.Env{
		Entities:  entities,
		Principal: req.Principal,
		Action:    req.Action,
		Resource:  req.Resource,
		Context:   req.Context,
	}
	var diag cedar.Diagnostic
	var forbids []cedar.DiagnosticReason
	var permits []cedar.DiagnosticReason
	// Don't try to short circuit this.
	// - Even though single forbid means forbid
	// - All policy should be run to collect errors
	// - For permit, all permits must be run to collect annotations
	// - For forbid, forbids must be run to collect annotations
	for id, po := range policies.All() {
		result, err := po.eval.Eval(env)
		if err != nil {
			diag.Errors = append(diag.Errors, cedar.DiagnosticError{PolicyID: id, Position: po.Position(), Message: err.Error()})
			continue
		}
		if !result {
			continue
		}
		if po.Effect() == cedar.Forbid {
			forbids = append(forbids, cedar.DiagnosticReason{PolicyID: id, Position: po.Position()})
		} else {
			permits = append(permits, cedar.DiagnosticReason{PolicyID: id, Position: po.Position()})
		}
	}
	if len(forbids) > 0 {
		diag.Reasons = forbids
		return cedar.Deny, diag
	}
	if len(permits) > 0 {
		diag.Reasons = permits
		return cedar.Allow, diag
	}

	return cedar.Deny, diag
}
