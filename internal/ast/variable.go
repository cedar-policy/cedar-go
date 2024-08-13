package ast

import "github.com/cedar-policy/cedar-go/types"

func Principal() Node {
	return NewNode(newPrincipalNode())
}

func Action() Node {
	return NewNode(newActionNode())
}

func Resource() Node {
	return NewNode(newResourceNode())
}

func Context() Node {
	return NewNode(newContextNode())
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
