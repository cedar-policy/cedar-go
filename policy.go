package cedar

import (
	"bytes"
	"fmt"

	"github.com/cedar-policy/cedar-go/ast"
	internalast "github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/internal/json"
	"github.com/cedar-policy/cedar-go/internal/parser"
)

// A Policy is the parsed form of a single Cedar language policy statement.
type Policy struct {
	Position    Position    // location within the policy text document
	Annotations Annotations // annotations found for this policy
	Effect      Effect      // the effect of this policy
	eval        evaler      // determines if a policy matches a request.
	ast         *internalast.Policy
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
func newAnnotationsFromSlice(annotations []internalast.AnnotationType) Annotations {
	res := make(map[string]string, len(annotations))
	for _, e := range annotations {
		res[string(e.Key)] = string(e.Value)
	}
	return res
}

// An Effect specifies the intent of the policy, to either permit or forbid any
// request that matches the scope and conditions specified in the policy.
type Effect internalast.Effect

// Each Policy has a Permit or Forbid effect that is determined during parsing.
const (
	Permit = Effect(true)
	Forbid = Effect(false)
)

// MarshalJSON encodes a single Policy statement in the JSON format specified by the [Cedar documentation].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/json-format.html
func (p *Policy) MarshalJSON() ([]byte, error) {
	jsonPolicy := (*json.Policy)(p.ast)
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
		eval:        eval.Compile((*internalast.Policy)(&jsonPolicy)),
		ast:         (*internalast.Policy)(&jsonPolicy),
	}
	return nil
}

func (p *Policy) MarshalCedar(buf *bytes.Buffer) {
	cedarPolicy := (*parser.Policy)(p.ast)
	cedarPolicy.MarshalCedar(buf)
}

func (p *Policy) UnmarshalCedar(b []byte) error {
	var cedarPolicy parser.Policy
	if err := cedarPolicy.UnmarshalCedar(b); err != nil {
		return err
	}

	*p = Policy{
		Position:    Position{},
		Annotations: newAnnotationsFromSlice(cedarPolicy.Annotations),
		Effect:      Effect(cedarPolicy.Effect),
		eval:        eval.Compile((*internalast.Policy)(&cedarPolicy)),
		ast:         (*internalast.Policy)(&cedarPolicy),
	}
	return nil
}

func NewPolicyFromAST(astIn *ast.Policy) *Policy {
	pp := (*internalast.Policy)(astIn)
	return &Policy{
		Position:    Position{},
		Annotations: newAnnotationsFromSlice(astIn.Annotations),
		Effect:      Effect(astIn.Effect),
		eval:        eval.Compile(pp),
		ast:         pp,
	}
}

// PolicySlice represents a set of un-named Policy's. Cedar documents, unlike the JSON format, don't have a means of
// naming individual policies.
type PolicySlice []*Policy

// UnmarshalCedar parses a concatenation of un-named Cedar policy statements. Names can be assigned to these policies
// when adding them to a PolicySet.
func (p *PolicySlice) UnmarshalCedar(b []byte) error {
	var res parser.PolicySlice
	if err := res.UnmarshalCedar(b); err != nil {
		return fmt.Errorf("parser error: %w", err)
	}
	policySlice := make([]*Policy, 0, len(res))
	for _, p := range res {
		policySlice = append(policySlice, &Policy{
			Position: Position{
				Offset: p.Position.Offset,
				Line:   p.Position.Line,
				Column: p.Position.Column,
			},
			Annotations: newAnnotationsFromSlice(p.Annotations),
			Effect:      Effect(p.Effect),
			eval:        eval.Compile((*internalast.Policy)(p)),
			ast:         (*internalast.Policy)(p),
		})
	}
	*p = policySlice
	return nil
}

// MarshalCedar emits a concatenated Cedar representation of a PolicySlice
func (p PolicySlice) MarshalCedar(buf *bytes.Buffer) {
	for i, policy := range p {
		policy.MarshalCedar(buf)

		if i < len(p)-1 {
			buf.WriteString("\n\n")
		}
	}
}
