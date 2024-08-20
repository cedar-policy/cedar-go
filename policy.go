package cedar

import (
	"bytes"

	"github.com/cedar-policy/cedar-go/ast"
	internalast "github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/internal/json"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/types"
)

// A Policy is the parsed form of a single Cedar language policy statement.
type Policy struct {
	eval eval.Evaler // determines if a policy matches a request.
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

func (p *Policy) MarshalCedar() []byte {
	cedarPolicy := (*parser.Policy)(p.ast)

	var buf bytes.Buffer
	cedarPolicy.MarshalCedar(&buf)

	return buf.Bytes()
}

func (p *Policy) UnmarshalCedar(b []byte) error {
	var cedarPolicy parser.Policy
	if err := cedarPolicy.UnmarshalCedar(b); err != nil {
		return err
	}

	*p = newPolicy((*internalast.Policy)(&cedarPolicy))
	return nil
}

func NewPolicyFromAST(astIn *ast.Policy) Policy {
	p := newPolicy((*internalast.Policy)(astIn))
	return p
}

// An Annotations is a map of key, value pairs found in the policy. Annotations
// have no impact on policy evaluation.
type Annotations map[types.Ident]types.String

func (p Policy) Annotations() Annotations {
	res := make(Annotations, len(p.ast.Annotations))
	for _, e := range p.ast.Annotations {
		res[e.Key] = e.Value
	}
	return res
}

// An Effect specifies the intent of the policy, to either permit or forbid any
// request that matches the scope and conditions specified in the policy.
type Effect bool

// Each Policy has a Permit or Forbid effect that is determined during parsing.
const (
	Permit = Effect(true)
	Forbid = Effect(false)
)

func (p Policy) Effect() Effect {
	return Effect(p.ast.Effect)
}

// A Position describes an arbitrary source position including the file, line, and column location.
type Position struct {
	Filename string // optional name of the source file for the enclosing policy, "" if the source is unknown or not a named file
	Offset   int    // byte offset, starting at 0
	Line     int    // line number, starting at 1
	Column   int    // column number, starting at 1 (character count per line)
}

func (p Policy) Position() Position {
	return Position(p.ast.Position)
}

func (p *Policy) SetFilename(fileName string) {
	p.ast.Position.Filename = fileName
}
