package eval

import (
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

type BoolEvaler struct {
	eval Evaler
}

func (e *BoolEvaler) Eval(env Env) (types.Boolean, error) {
	v, err := e.eval.Eval(env)
	if err != nil {
		return false, err
	}
	vb, err := ValueToBool(v)
	if err != nil {
		return false, err
	}
	return vb, nil
}

func Compile(p *ast.Policy) BoolEvaler {
	p = foldPolicy(p)
	node := PolicyToNode(p).AsIsNode()
	return BoolEvaler{eval: ToEval(node)}
}

func PolicyToNode(p *ast.Policy) ast.Node {
	var nodes []ast.Node
	_, principalAll := p.Principal.(ast.ScopeTypeAll)
	_, actionAll := p.Action.(ast.ScopeTypeAll)
	_, resourceAll := p.Resource.(ast.ScopeTypeAll)
	if principalAll && actionAll && resourceAll {
		nodes = append(nodes, ast.True())
	} else {
		if !principalAll {
			nodes = append(nodes, scopeToNode(ast.NewPrincipalNode(), p.Principal))
		}
		if !actionAll {
			nodes = append(nodes, scopeToNode(ast.NewActionNode(), p.Action))
		}
		if !resourceAll {
			nodes = append(nodes, scopeToNode(ast.NewResourceNode(), p.Resource))
		}
	}
	for _, c := range p.Conditions {
		if c.Condition == ast.ConditionUnless {
			nodes = append(nodes, ast.Not(ast.NewNode(c.Body)))
			continue
		}
		nodes = append(nodes, ast.NewNode(c.Body))
	}
	res := nodes[len(nodes)-1]
	for i := len(nodes) - 2; i >= 0; i-- {
		res = nodes[i].And(res)
	}
	return res
}

func scopeToNode(varNode ast.NodeTypeVariable, in ast.IsScopeNode) (out ast.Node) {
	switch t := in.(type) {
	case ast.ScopeTypeAll:
		out = ast.True()
	case ast.ScopeTypeEq:
		out = ast.NewNode(varNode).Equal(ast.Value(t.Entity))
	case ast.ScopeTypeIn:
		out = ast.NewNode(varNode).In(ast.Value(t.Entity))
	case ast.ScopeTypeInSet:
		vals := make([]types.Value, len(t.Entities))
		for i, e := range t.Entities {
			vals[i] = e
		}
		out = ast.NewNode(varNode).In(ast.Value(types.NewSet(vals...)))
	case ast.ScopeTypeIs:
		out = ast.NewNode(varNode).Is(t.Type)
	case ast.ScopeTypeIsIn:
		out = ast.NewNode(varNode).IsIn(t.Type, ast.Value(t.Entity))
	}
	return out
}
