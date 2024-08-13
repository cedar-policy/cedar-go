package ast

import "github.com/cedar-policy/cedar-go/types"

func Principal() Node {
	return NewNode(NewPrincipalNode())
}

func Action() Node {
	return NewNode(NewActionNode())
}

func Resource() Node {
	return NewNode(NewResourceNode())
}

func Context() Node {
	return NewNode(NewContextNode())
}

func NewPrincipalNode() NodeTypeVariable {
	return NodeTypeVariable{Name: types.String("principal")}
}

func NewActionNode() NodeTypeVariable {
	return NodeTypeVariable{Name: types.String("action")}
}

func NewResourceNode() NodeTypeVariable {
	return NodeTypeVariable{Name: types.String("resource")}
}

func NewContextNode() NodeTypeVariable {
	return NodeTypeVariable{Name: types.String("context")}
}
