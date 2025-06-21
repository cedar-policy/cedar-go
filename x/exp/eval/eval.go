package eval

import (
	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func Eval(n ast.IsNode, env eval.Env) (types.Value, error) {
	evaler := eval.ToEval(n)
	return evaler.Eval(env)
}
