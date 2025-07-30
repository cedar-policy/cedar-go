// Package eval provides a simple interface for evaluating or partially evaluating a policy node in a given environment.
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

// PartialPolicy returns a partially evaluated version of the policy and a boolean indicating if the policy should be kept.
// (Policies that are determined to evaluate to false are not kept.)
//
// it is supposed to use `PartialPolicy` to partially evaluate a policy, and then use `PolicyToNode` to compile the policy to a node.
// but you can also use `PartialPolicy` directly.
//
// All the env parts (PARC) must be specified, but you can
// specify `Variable` as `Variable("principal")` or `Variable("action")` or `Variable("resource")` or `Variable("context")`.
// also you can specify part of Context to be a `Variable`, such as `key` in `Context` could be
// `
//
//	context := types.NewRecord(types.RecordMap{
//			"key": Variable("key"),
//	})
//
// `
//
// when the node is kept, it can be one of three kinds:
// 1. it is a `ValueNode`, and Must be `ast.True()` (e.g. `ast.True()`)
// 2. it is a `Node` contains `Variable` (e.g. `ast.Permit().When(ast.Context().Access("key").Equal(ast.Long(42)))`)
// 3. it is a `Node` contains `PartialError` (e.g. `ast.ExtensionCall(partialErrorName, ast.String("type error: expected comparable value, got string"))`)
//
// you can use the partial evaluation result `ast.Node` to do any additional work you want
// for example, you can convert it to an sql query.
// in which case the variable should be a column name and binary node should be an sql expression.
func PartialPolicy(env Env, p *ast.Policy) (policy *ast.Policy, keep bool) {
	return eval.PartialPolicy(env, p)
}

// PolicyToNode returns a node compiled from a policy.
func PolicyToNode(p *ast.Policy) ast.Node {
	return eval.PolicyToNode(p)
}

// PartialError returns a node that represents a partial error.
func PartialError(err error) ast.IsNode {
	return eval.PartialError(err)
}

// ToPartialError returns the error if the node is a partial error.
func ToPartialError(n ast.IsNode) (err error, ok bool) {
	return eval.ToPartialError(n)
}

// Variable is a variable in the policy.
func Variable(v types.String) types.Value {
	return eval.Variable(v)
}

// ToVariable converts a value to a variable.
func ToVariable(v types.Value) (types.String, bool) {
	if ent, ok := v.(types.EntityUID); ok {
		return eval.ToVariable(ent)
	}
	return "", false
}

// TypeName returns the type name of a value.
func TypeName(v types.Value) string {
	return eval.TypeName(v)
}

// ErrType is the error type for type errors.
var ErrType = eval.ErrType
