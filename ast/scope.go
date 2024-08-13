package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

func (p *Policy) PrincipalEq(entity types.EntityUID) *Policy {
	return newPolicy(p.Policy.PrincipalEq(entity))
}

func (p *Policy) PrincipalIn(entity types.EntityUID) *Policy {
	return newPolicy(p.Policy.PrincipalIn(entity))
}

func (p *Policy) PrincipalIs(entityType types.Path) *Policy {
	return newPolicy(p.Policy.PrincipalIs(entityType))
}

func (p *Policy) PrincipalIsIn(entityType types.Path, entity types.EntityUID) *Policy {
	return newPolicy(p.Policy.PrincipalIsIn(entityType, entity))
}

func (p *Policy) ActionEq(entity types.EntityUID) *Policy {
	return newPolicy(p.Policy.ActionEq(entity))
}

func (p *Policy) ActionIn(entity types.EntityUID) *Policy {
	return newPolicy(p.Policy.ActionIn(entity))
}

func (p *Policy) ActionInSet(entities ...types.EntityUID) *Policy {
	return newPolicy(p.Policy.ActionInSet(entities...))
}

func (p *Policy) ResourceEq(entity types.EntityUID) *Policy {
	return newPolicy(p.Policy.ResourceEq(entity))
}

func (p *Policy) ResourceIn(entity types.EntityUID) *Policy {
	return newPolicy(p.Policy.ResourceIn(entity))
}

func (p *Policy) ResourceIs(entityType types.Path) *Policy {
	return newPolicy(p.Policy.ResourceIs(entityType))
}

func (p *Policy) ResourceIsIn(entityType types.Path, entity types.EntityUID) *Policy {
	return newPolicy(p.Policy.ResourceIsIn(entityType, entity))
}
