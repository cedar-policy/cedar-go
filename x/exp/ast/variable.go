package ast

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
	return newValueNode(nodeTypeVariable, "principal")
}

func newActionNode() Node {
	return newValueNode(nodeTypeVariable, "action")
}

func newResourceNode() Node {
	return newValueNode(nodeTypeVariable, "resource")
}

func newContextNode() Node {
	return newValueNode(nodeTypeVariable, "context")
}
