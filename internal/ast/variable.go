package ast

import "github.com/cedar-policy/cedar-go/types"

func Principal() Node {
	return newNode(newPrincipalNode())
}

func Action() Node {
	return newNode(newActionNode())
}

func Resource() Node {
	return newNode(newResourceNode())
}

func Context() Node {
	return newNode(newContextNode())
}

func newPrincipalNode() NodeTypeVariable {
	return NodeTypeVariable{Name: types.String("principal")}
}

func newActionNode() NodeTypeVariable {
	return NodeTypeVariable{Name: types.String("action")}
}

func newResourceNode() NodeTypeVariable {
	return NodeTypeVariable{Name: types.String("resource")}
}

func newContextNode() NodeTypeVariable {
	return NodeTypeVariable{Name: types.String("context")}
}
