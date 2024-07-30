package ast

import "github.com/cedar-policy/cedar-go/types"

func (p *Policy) Annotate(name, value types.String) *Policy {
	p.annotations = append(p.annotations, newAnnotationNode(name, value))
	return p
}

func newAnnotationNode(name, value types.String) Node {
	return Node{nodeType: nodeTypeAnnotation, args: []Node{String(name), String(value)}}
}
