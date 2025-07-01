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
