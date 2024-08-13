package ast

type CompiledPolicySet map[string]CompiledPolicy

type CompiledPolicy struct {
	PolicySetEntry
}

func (p PolicySetEntry) TmpGetAnnotations() map[string]string {
	res := make(map[string]string, len(p.Policy.Annotations))
	for _, e := range p.Policy.Annotations {
		res[string(e.Key)] = string(e.Value)
	}
	return res
}
func (p PolicySetEntry) TmpGetEffect() bool {
	return bool(p.Policy.Effect)
}

func Compile(p Policy) Evaler {
	node := policyToNode(p).v
	return toEval(node)
}

func policyToNode(p Policy) Node {
	nodes := make([]Node, 3+len(p.Conditions))
	nodes[0] = p.Principal.toNode()
	nodes[1] = p.Action.toNode()
	nodes[2] = p.Resource.toNode()
	for i, c := range p.Conditions {
		if c.Condition == ConditionUnless {
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
