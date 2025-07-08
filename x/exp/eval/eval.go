// Package eval provides a simple interface for evaluating a policy node in a given environment.
package eval

import (
	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

// Env is the environment for evaluating a policy.
type Env = eval.Env

// Eval evaluates a policy node in the given environment.
func Eval(n ast.IsNode, env Env) (types.Value, error) {
	evaler := eval.ToEval(n)
	return evaler.Eval(env)
}

// PartialPolicyToNode returns node compiled from a partially evaluated version of the policy
// and a boolean indicating if the policy should be kept.
// (Policies that are determined to evaluate to false are not kept.)
var PartialPolicyToNode = eval.PartialPolicyToNode

// PartialError returns a node that represents a partial error.
var PartialError = eval.PartialError

// IsPartialError returns true if the node is a partial error.
var IsPartialError = eval.IsPartialError

// Variable is a variable in the policy.
var Variable = eval.Variable

// Ignore is a special value that is used to ignore a value.
var Ignore = eval.Ignore

// IsVariable checks if a value is a variable.
var IsVariable = eval.IsVariable

// IsIgnore checks if a value is an ignore value.
var IsIgnore = eval.IsIgnore

// ToVariable converts a value to a variable.
var ToVariable = eval.ToVariable

// TypeName returns the type name of a value.
var TypeName = eval.TypeName

// ErrType is the error type for type errors.
var ErrType = eval.ErrType
