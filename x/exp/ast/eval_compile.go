package ast

type CompiledPolicySet map[string]CompiledPolicy

type CompiledPolicy struct {
	PolicySetEntry
	eval Evaler
}

func (p PolicySetEntry) TmpGetAnnotations() map[string]string {
	res := make(map[string]string, len(p.Policy.annotations))
	for _, e := range p.Policy.annotations {
		res[string(e.Key)] = string(e.Value)
	}
	return res
}
func (p PolicySetEntry) TmpGetEffect() bool {
	return bool(p.Policy.effect)
}

func Compile(p Policy) Evaler {
	node := policyToNode(p).v
	return toEval(node)
}

func policyToNode(p Policy) Node {
	nodes := make([]Node, 3+len(p.conditions))
	nodes[0] = p.principal.toNode()
	nodes[1] = p.action.toNode()
	nodes[2] = p.resource.toNode()
	for i, c := range p.conditions {
		if c.Condition == conditionUnless {
			nodes[i+3] = Not(newNode(c.Body))
			continue
		}
		nodes[i+3] = newNode(c.Body)
	}
	res := nodes[len(nodes)-1]
	for i := len(nodes) - 2; i >= 0; i-- {
		res = nodes[i].And(res)
	}
	return res
}
