package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

func (p *Policy) PrincipalEq(entity types.EntityUID) *Policy {
	return &Policy{p.Policy.PrincipalEq(entity)}
}

func (p *Policy) PrincipalIn(entity types.EntityUID) *Policy {
	return &Policy{p.Policy.PrincipalIn(entity)}
}

func (p *Policy) PrincipalIs(entityType types.Path) *Policy {
	return &Policy{p.Policy.PrincipalIs(entityType)}
}

func (p *Policy) PrincipalIsIn(entityType types.Path, entity types.EntityUID) *Policy {
	return &Policy{p.Policy.PrincipalIsIn(entityType, entity)}
}

func (p *Policy) ActionEq(entity types.EntityUID) *Policy {
	return &Policy{p.Policy.ActionEq(entity)}
}

func (p *Policy) ActionIn(entity types.EntityUID) *Policy {
	return &Policy{p.Policy.ActionIn(entity)}
}

func (p *Policy) ActionInSet(entities ...types.EntityUID) *Policy {
	return &Policy{p.Policy.ActionInSet(entities...)}
}

func (p *Policy) ResourceEq(entity types.EntityUID) *Policy {
	return &Policy{p.Policy.ResourceEq(entity)}
}

func (p *Policy) ResourceIn(entity types.EntityUID) *Policy {
	return &Policy{p.Policy.ResourceIn(entity)}
}

func (p *Policy) ResourceIs(entityType types.Path) *Policy {
	return &Policy{p.Policy.ResourceIs(entityType)}
}

func (p *Policy) ResourceIsIn(entityType types.Path, entity types.EntityUID) *Policy {
	return &Policy{p.Policy.ResourceIsIn(entityType, entity)}
}
