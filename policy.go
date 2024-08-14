package cedar

import (
	"bytes"

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
		eval:        eval.Compile((internalast.Policy)(jsonPolicy)),
		ast:         (*internalast.Policy)(&jsonPolicy),
	}
	return nil
}

func (p *Policy) MarshalCedar(buf *bytes.Buffer) {
	cedarPolicy := &parser.Policy{Policy: *p.ast}
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
		eval:        eval.Compile(cedarPolicy.Policy),
		ast:         &cedarPolicy.Policy,
	}
	return nil
}

func NewPolicyFromAST(astIn *ast.Policy) *Policy {
	pp := (*internalast.Policy)(astIn)
	return &Policy{
		Position:    Position{},
		Annotations: newAnnotationsFromSlice(astIn.Annotations),
		Effect:      Effect(astIn.Effect),
		eval:        eval.Compile(*pp),
		ast:         pp,
	}
}
