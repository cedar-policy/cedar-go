package ast

import (
    "github.com/cedar-policy/cedar-go/types"
)

type Template Policy

func (p Template) ClonePolicy() Policy {
    clonedPolicy := Policy{
        Effect:      p.Effect,
        Annotations: append([]AnnotationType(nil), p.Annotations...),
        Principal:   p.Principal,
        Action:      p.Action,
        Resource:    p.Resource,
        Conditions:  append([]ConditionType(nil), p.Conditions...),
        Position:    p.Position,
        tplCtx: templateContext{
            slots: append([]types.SlotID(nil), p.tplCtx.slots...),
        },
    }

    return clonedPolicy
}

type LinkedPolicy struct {
    TemplateID string
    LinkID     string
    Policy     *Policy
}
