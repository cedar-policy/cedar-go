// Package cedar provides an implementation of the Cedar language authorizer.
package cedar

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/types"
	"golang.org/x/exp/maps"
)

// A PolicySet is a slice of policies.
type PolicySet []Policy

// A Policy is the parsed form of a single Cedar language policy statement. It
// includes the following elements, a Position, Annotations, and an Effect.
type Policy struct {
	Position    Position    // location within the policy text document
	Annotations Annotations // annotations found for this policy
	Effect      Effect      // the effect of this policy
	eval        evaler      // determines if a policy matches a request.
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

// An Effect specifies the intent of the policy, to either permit or forbid any
// request that matches the scope and conditions specified in the policy.
type Effect bool

// Each Policy has a Permit or Forbid effect that is determined during parsing.
const (
	Permit = Effect(true)
	Forbid = Effect(false)
)

func (a Effect) String() string {
	if a {
		return "permit"
	}
	return "forbid"
}
func (a Effect) MarshalJSON() ([]byte, error) { return []byte(`"` + a.String() + `"`), nil }

func (a *Effect) UnmarshalJSON(b []byte) error {
	*a = string(b) == `"permit"`
	return nil
}

// NewPolicySet will create a PolicySet from the given text document with the
// given file name used in Position data.  If there is an error parsing the
// document, it will be returned.
func NewPolicySet(fileName string, document []byte) (PolicySet, error) {
	var res ast.PolicySet
	if err := res.UnmarshalCedar(document); err != nil {
		return nil, fmt.Errorf("parser error: %w", err)
	}
	var policies PolicySet
	for _, p := range res {
		ann := Annotations(p.TmpGetAnnotations())
		policies = append(policies, Policy{
			Position: Position{
				Filename: fileName,
				Offset:   p.Position.Offset,
				Line:     p.Position.Line,
				Column:   p.Position.Column,
			},
			Annotations: ann,
			Effect:      Effect(p.TmpGetEffect()),
			eval:        ast.Compile(p.Policy),
		})
	}
	return policies, nil
}

type Entities = ast.Entities
type Entity = ast.Entity

func entitiesFromSlice(s []Entity) Entities {
	var res = Entities{}
	for _, e := range s {
		res[e.UID] = e
	}
	return res
}

func entitiesToSlice(e Entities) []Entity {
	s := maps.Values(e)
	slices.SortFunc(s, func(a, b Entity) int {
		return strings.Compare(a.UID.String(), b.UID.String())
	})
	return s
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

type evalContext = ast.EvalContext

type evaler = ast.Evaler

// IsAuthorized uses the combination of the PolicySet and Entities to determine
// if the given Request to determine Decision and Diagnostic.
func (p PolicySet) IsAuthorized(entities Entities, req Request) (Decision, Diagnostic) {
	c := &evalContext{
		Entities:  entities,
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
