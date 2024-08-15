package eval

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/types"
)

func Compile(p *ast.Policy) Evaler {
	node := policyToNode(p).AsIsNode()
	return toEval(node)
}

func policyToNode(p *ast.Policy) ast.Node {
	nodes := make([]ast.Node, 3+len(p.Conditions))
	nodes[0] = scopeToNode(ast.NewPrincipalNode(), p.Principal)
	nodes[1] = scopeToNode(ast.NewActionNode(), p.Action)
	nodes[2] = scopeToNode(ast.NewResourceNode(), p.Resource)
	for i, c := range p.Conditions {
		if c.Condition == ast.ConditionUnless {
			nodes[i+3] = ast.Not(ast.NewNode(c.Body))
			continue
		}
		nodes[i+3] = ast.NewNode(c.Body)
	}
	res := nodes[len(nodes)-1]
	for i := len(nodes) - 2; i >= 0; i-- {
		res = nodes[i].And(res)
	}
	return res
}

func scopeToNode(varNode ast.NodeTypeVariable, in ast.IsScopeNode) ast.Node {
	switch t := in.(type) {
	case ast.ScopeTypeAll:
		return ast.True()
	case ast.ScopeTypeEq:
		return ast.NewNode(varNode).Equals(ast.EntityUID(t.Entity))
	case ast.ScopeTypeIn:
		return ast.NewNode(varNode).In(ast.EntityUID(t.Entity))
	case ast.ScopeTypeInSet:
		set := make([]types.Value, len(t.Entities))
		for i, e := range t.Entities {
			set[i] = e
		}
		return ast.NewNode(varNode).In(ast.Set(set))
	case ast.ScopeTypeIs:
		return ast.NewNode(varNode).Is(t.Type)

	case ast.ScopeTypeIsIn:
		return ast.NewNode(varNode).IsIn(t.Type, ast.EntityUID(t.Entity))
	default:
		panic(fmt.Sprintf("unknown scope type %T", t))
	}
}
