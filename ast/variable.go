package ast

import "github.com/cedar-policy/cedar-go/types"

func Principal() Node {
	return newPrincipalNode()
}

func Action() Node {
	return newActionNode()
}

func Resource() Node {
	return newResourceNode()
}

func Context() Node {
	return newContextNode()
}

func newPrincipalNode() Node {
	return newNode(rawPrincipalNode())
}

func newActionNode() Node {
	return newNode(rawActionNode())
}

func newResourceNode() Node {
	return newNode(rawResourceNode())
}

func newContextNode() Node {
	return newNode(rawContextNode())
}

func rawPrincipalNode() nodeTypeVariable {
	return nodeTypeVariable{Name: types.String("principal")}
}

func rawActionNode() nodeTypeVariable {
	return nodeTypeVariable{Name: types.String("action")}
}

func rawResourceNode() nodeTypeVariable {
	return nodeTypeVariable{Name: types.String("resource")}
}

func rawContextNode() nodeTypeVariable {
	return nodeTypeVariable{Name: types.String("context")}
}
