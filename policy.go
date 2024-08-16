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
	eval evaler // determines if a policy matches a request.
	ast  *internalast.Policy
}

func newPolicy(astIn *internalast.Policy) Policy {
	return Policy{eval: eval.Compile(astIn), ast: astIn}
}

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

	*p = newPolicy((*internalast.Policy)(&jsonPolicy))
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

	*p = newPolicy((*internalast.Policy)(&cedarPolicy))
	return nil
}

func NewPolicyFromAST(astIn *ast.Policy) *Policy {
	p := newPolicy((*internalast.Policy)(astIn))
	return &p
}

// An Annotations is a map of key, value pairs found in the policy. Annotations
// have no impact on policy evaluation.
type Annotations map[string]string

func (p Policy) Annotations() Annotations {
	// TODO: Where should we deal with duplicate keys?
	res := make(map[string]string, len(p.ast.Annotations))
	for _, e := range p.ast.Annotations {
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

func (p Policy) Effect() Effect {
	return Effect(p.ast.Effect)
}

// A Position describes an arbitrary source position including the file, line, and column location.
type Position internalast.Position

func (p Policy) Position() Position {
	return Position(p.ast.Position)
}

func (p *Policy) SetSourceFile(path string) {
	p.ast.Position.FileName = path
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
		newPolicy := newPolicy((*internalast.Policy)(p))
		policySlice = append(policySlice, &newPolicy)
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
