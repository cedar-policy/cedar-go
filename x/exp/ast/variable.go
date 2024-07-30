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
	return newValueNode(nodeTypeVariable, types.String("principal"))
}

func newActionNode() Node {
	return newValueNode(nodeTypeVariable, types.String("action"))
}

func newResourceNode() Node {
	return newValueNode(nodeTypeVariable, types.String("resource"))
}

func newContextNode() Node {
	return newValueNode(nodeTypeVariable, types.String("context"))
}
