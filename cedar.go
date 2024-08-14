// Package cedar provides an implementation of the Cedar language authorizer.
package cedar

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/entities"
	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/internal/json"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/types"
)

// A Policy is the parsed form of a single Cedar language policy statement.
type Policy struct {
	Position    Position    // location within the policy text document
	Annotations Annotations // annotations found for this policy
	Effect      Effect      // the effect of this policy
	eval        evaler      // determines if a policy matches a request.
	ast         ast.Policy
}

// A Position describes an arbitrary source position including the file, line, and column location.
type Position struct {
	Filename string // filename, if any
	Offset   int    // byte offset, starting at 0
	Line     int    // line number, starting at 1
	Column   int    // column number, starting at 1 (character count per line)
}

// An Annotations is a map of key, value pairs found in the policy. Annotations
// have no impact on policy evaluation.
type Annotations map[string]string

// TODO: Is this where we should deal with duplicate keys?
func newAnnotationsFromSlice(annotations []ast.AnnotationType) Annotations {
	res := make(map[string]string, len(annotations))
	for _, e := range annotations {
		res[string(e.Key)] = string(e.Value)
	}
	return res
}

// An Effect specifies the intent of the policy, to either permit or forbid any
// request that matches the scope and conditions specified in the policy.
type Effect ast.Effect

// Each Policy has a Permit or Forbid effect that is determined during parsing.
const (
	Permit = Effect(true)
	Forbid = Effect(false)
)

// MarshalJSON encodes a single Policy statement in the JSON format specified by the [Cedar documentation].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/json-format.html
func (p *Policy) MarshalJSON() ([]byte, error) {
	jsonPolicy := &json.Policy{Policy: p.ast}
	return jsonPolicy.MarshalJSON()
}

// UnmarshalJSON parses and compiles a single Policy statement in the JSON format specified by the [Cedar documentation].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/json-format.html
func (p *Policy) UnmarshalJSON(b []byte) error {
	var jsonPolicy json.Policy
	if err := jsonPolicy.UnmarshalJSON(b); err != nil {
		return err
	}
	*p = Policy{
		Position:    Position{},
		Annotations: newAnnotationsFromSlice(jsonPolicy.Annotations),
		Effect:      Effect(jsonPolicy.Effect),
		eval:        eval.Compile(jsonPolicy.Policy),
		ast:         jsonPolicy.Policy,
	}
	return nil
}

// A PolicySet is a slice of policies.
type PolicySet []Policy

// NewPolicySet will create a PolicySet from the given text document with the
// given file name used in Position data.  If there is an error parsing the
// document, it will be returned.
func NewPolicySet(fileName string, document []byte) (PolicySet, error) {
	var res parser.PolicySet
	if err := res.UnmarshalCedar(document); err != nil {
		return nil, fmt.Errorf("parser error: %w", err)
	}
	var policies PolicySet
	for _, p := range res {
		policies = append(policies, Policy{
			Position: Position{
				Filename: fileName,
				Offset:   p.Position.Offset,
				Line:     p.Position.Line,
				Column:   p.Position.Column,
			},
			Annotations: newAnnotationsFromSlice(p.Policy.Annotations),
			Effect:      Effect(p.Policy.Effect),
			eval:        eval.Compile(p.Policy.Policy),
			ast:         p.Policy.Policy,
		})
	}
	return policies, nil
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
	Reasons []Reason `json:"reasons,omitempty"`
	Errors  []Error  `json:"errors,omitempty"`
}

// An Error details the Policy index within a PolicySet, the Position within the
// text document, and the resulting error message.
type Error struct {
	Policy   int      `json:"policy"`
	Position Position `json:"position"`
	Message  string   `json:"message"`
}

func (e Error) String() string {
	return fmt.Sprintf("while evaluating policy `policy%d`: %v", e.Policy, e.Message)
}

// A Reason details the Policy index within a PolicySet, and the Position within
// the text document.
type Reason struct {
	Policy   int      `json:"policy"`
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
	for n, po := range p {
		v, err := po.eval.Eval(c)
		if err != nil {
			diag.Errors = append(diag.Errors, Error{Policy: n, Position: po.Position, Message: err.Error()})
			continue
		}
		vb, err := types.ValueToBool(v)
		if err != nil {
			// should never happen, maybe remove this case
			diag.Errors = append(diag.Errors, Error{Policy: n, Position: po.Position, Message: err.Error()})
			continue
		}
		if !vb {
			continue
		}
		if po.Effect == Forbid {
			forbidReasons = append(forbidReasons, Reason{Policy: n, Position: po.Position})
			gotForbid = true
		} else {
			permitReasons = append(permitReasons, Reason{Policy: n, Position: po.Position})
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
