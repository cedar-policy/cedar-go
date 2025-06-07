package templates

import (
	"bytes"
	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/types"

	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/internal/json"
	"github.com/cedar-policy/cedar-go/internal/parser"
	internalast "github.com/cedar-policy/cedar-go/x/exp/ast"
)

// A Policy is the parsed form of a single Cedar language policy statement.
type Policy struct {
	eval eval.BoolEvaler // determines if a policy matches a request.
	ast  *internalast.Policy
}

func newPolicy(astIn *internalast.Policy) *Policy {
	return &Policy{eval: eval.Compile(astIn), ast: astIn}
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

	*p = *newPolicy((*internalast.Policy)(&jsonPolicy))
	return nil
}

// MarshalCedar encodes a single Policy statement in the human-readable format specified by the [Cedar documentation].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/syntax-grammar.html
func (p *Policy) MarshalCedar() []byte {
	cedarPolicy := (*parser.Policy)(p.ast)

	var buf bytes.Buffer
	cedarPolicy.MarshalCedar(&buf)

	return buf.Bytes()
}

// UnmarshalCedar parses and compiles a single Policy statement in the human-readable format specified by the [Cedar documentation].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/syntax-grammar.html
func (p *Policy) UnmarshalCedar(b []byte) error {
	var cedarPolicy parser.Policy
	if err := cedarPolicy.UnmarshalCedar(b); err != nil {
		return err
	}
	*p = *newPolicy((*internalast.Policy)(&cedarPolicy))
	return nil
}

// NewPolicyFromAST lets you create a new policy statement from a programmatically created AST.
// Do not modify the *ast.Policy after passing it into NewPolicyFromAST.
func NewPolicyFromAST(astIn *ast.Policy) *Policy {
	p := newPolicy((*internalast.Policy)(astIn))
	return p
}

// Annotations retrieves the annotations associated with this policy.
func (p *Policy) Annotations() cedar.Annotations {
	res := make(cedar.Annotations, len(p.ast.Annotations))
	for _, e := range p.ast.Annotations {
		res[e.Key] = e.Value
	}
	return res
}

// Effect retrieves the effect of this policy.
func (p *Policy) Effect() cedar.Effect {
	return cedar.Effect(p.ast.Effect)
}

// Position retrieves the position of this policy.
func (p *Policy) Position() cedar.Position {
	return cedar.Position(p.ast.Position)
}

// SetFilename sets the filename of this policy.
func (p *Policy) SetFilename(fileName string) {
	p.ast.Position.Filename = fileName
}

// AST retrieves the AST of this policy.  Do not modify the AST, as the
// compiled policy will no longer be in sync with the AST.
func (p *Policy) AST() *ast.Policy {
	return (*ast.Policy)(p.ast)
}

// We use parser.Policy as the underlying type for Template because
// a templates.Policy or cedar.Policy would be compiled, however, a Template
// is not compilable until it is linked with slot values to create a concrete policy.

// Template represents a Cedar policy template that can be linked with slot values
// to create concrete policies. It's a wrapper around the internal parser.Policy type.
type Template parser.Policy

func newTemplate(astIn *internalast.Policy) *Template {
	t := (*Template)(astIn)
	return t
}

// MarshalCedar serializes the Template into its Cedar language representation.
// Returns the serialized template as a byte slice.
func (p *Template) MarshalCedar() []byte {
	cedarPolicy := (*parser.Policy)(p)

	var buf bytes.Buffer
	cedarPolicy.MarshalCedar(&buf)

	return buf.Bytes()
}

func (p *Template) UnmarshalCedar(b []byte) error {
	var cedarPolicy parser.Policy
	if err := cedarPolicy.UnmarshalCedar(b); err != nil {
		return err
	}

	*p = *newTemplate((*internalast.Policy)(&cedarPolicy))

	return nil
}

// MarshalJSON encodes a single Policy statement in the JSON format specified by the [Cedar documentation].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/json-format.html
func (p *Template) MarshalJSON() ([]byte, error) {
	policyAST := (*internalast.Policy)(p)
	jsonPolicy := (*json.Policy)(policyAST)

	return jsonPolicy.MarshalJSON()
}

// UnmarshalJSON parses and compiles a single Policy statement in the JSON format specified by the [Cedar documentation].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/json-format.html
func (p *Template) UnmarshalJSON(b []byte) error {
	var jsonPolicy json.Policy
	if err := jsonPolicy.UnmarshalJSON(b); err != nil {
		return err
	}

	*p = *newTemplate((*internalast.Policy)(&jsonPolicy))

	return nil
}

// SetFilename sets the filename of this template.
// This is useful for error reporting and debugging purposes.
func (p *Template) SetFilename(fileName string) {
	p.Position.Filename = fileName
}

func (p *Template) Slots() []types.SlotID {
	policyAST := (*internalast.Policy)(p)
	return policyAST.Slots()
}

// AST retrieves the AST of this Template. Do not modify the AST.
func (p *Template) AST() *ast.Policy {
	policyAST := (*internalast.Policy)(p)
	return (*ast.Policy)(policyAST)
}
