package eval

import (
	"github.com/cedar-policy/cedar-go/internal/ast"
)

func Compile(p ast.Policy) Evaler {
	node := policyToNode(p).AsIsNode()
	return toEval(node)
}

func policyToNode(p ast.Policy) ast.Node {
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
