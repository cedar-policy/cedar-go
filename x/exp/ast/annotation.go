package ast

func (p *Policy) Annotate(name string, value string) *Policy {
	p.annotations = append(p.annotations, newAnnotationNode(name, value))
	return p
}

func newAnnotationNode(name, value string) Node {
	return newValueNode(nodeTypeAnnotation, []string{name, value})
}
