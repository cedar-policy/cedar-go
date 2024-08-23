package eval

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/types"
)

func Compile(p *ast.Policy) Evaler {
	p = bakePolicy(p)
	node := policyToNode(p).AsIsNode()
	return toEval(node)
}

func policyToNode(p *ast.Policy) ast.Node {
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

func scopeToNode(varNode ast.NodeTypeVariable, in ast.IsScopeNode) ast.Node {
	switch t := in.(type) {
	case ast.ScopeTypeAll:
		return ast.True()
	case ast.ScopeTypeEq:
		return ast.NewNode(varNode).Equal(ast.Value(t.Entity))
	case ast.ScopeTypeIn:
		return ast.NewNode(varNode).In(ast.Value(t.Entity))
	case ast.ScopeTypeInSet:
		set := make(types.Set, len(t.Entities))
		for i, e := range t.Entities {
			set[i] = e
		}
		return ast.NewNode(varNode).In(ast.Value(set))
	case ast.ScopeTypeIs:
		return ast.NewNode(varNode).Is(t.Type)

	case ast.ScopeTypeIsIn:
		return ast.NewNode(varNode).IsIn(t.Type, ast.Value(t.Entity))
	default:
		panic(fmt.Sprintf("unknown scope type %T", t))
	}
}
