package ast

import (
	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/types"
)

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
	return NodeTypeVariable{Name: types.String(consts.Principal)}
}

func NewActionNode() NodeTypeVariable {
	return NodeTypeVariable{Name: types.String(consts.Action)}
}

func NewResourceNode() NodeTypeVariable {
	return NodeTypeVariable{Name: types.String(consts.Resource)}
}

func NewContextNode() NodeTypeVariable {
	return NodeTypeVariable{Name: types.String(consts.Context)}
}
