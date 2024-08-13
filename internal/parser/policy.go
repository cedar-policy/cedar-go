package parser

import "github.com/cedar-policy/cedar-go/internal/ast"

type PolicySet map[string]PolicySetEntry

type PolicySetEntry struct {
	Policy   Policy
	Position Position
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

type Policy struct {
	ast.Policy
}
