// Package ast provides functions for programmatically constructing a Cedar policy AST.
//
// Programmatically generated policies are germinated by calling one of the following top-level functions:
//   - [Permit]
//   - [Forbid]
//   - [Annotation]
package ast

import (
	"bytes"

	"github.com/cedar-policy/cedar-go/internal/json"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

// Policy represents a single Cedar policy statement
type Policy ast.Policy

func wrapPolicy(p *ast.Policy) *Policy {
	return (*Policy)(p)
}

func (p *Policy) unwrap() *ast.Policy {
	return (*ast.Policy)(p)
}

// Permit creates a new Permit policy.
func Permit() *Policy {
	return wrapPolicy(ast.Permit())
}

// Forbid creates a new Forbid policy.
func Forbid() *Policy {
	return wrapPolicy(ast.Forbid())
}

// When adds a conditional clause.
func (p *Policy) When(node Node) *Policy {
	return wrapPolicy(p.unwrap().When(node.Node))
}

// Unless adds a conditional clause.
func (p *Policy) Unless(node Node) *Policy {
	return wrapPolicy(p.unwrap().Unless(node.Node))
}

// MarshalJSON encodes a single Policy statement in the JSON format specified by the [Cedar documentation].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/json-format.html
func (p *Policy) MarshalJSON() ([]byte, error) {
	jsonPolicy := (*json.Policy)(p)
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

	*p = (Policy)(jsonPolicy)
	return nil
}

// MarshalCedar encodes a single Policy statement in the human-readable format specified by the [Cedar documentation].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/syntax-grammar.html
func (p *Policy) MarshalCedar() []byte {
	cedarPolicy := (*parser.Policy)(p)

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
	*p = (Policy)(cedarPolicy)
	return nil
}
