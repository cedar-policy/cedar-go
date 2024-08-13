package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

func (p *Policy) PrincipalEq(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().PrincipalEq(entity))
}

func (p *Policy) PrincipalIn(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().PrincipalIn(entity))
}

func (p *Policy) PrincipalIs(entityType types.Path) *Policy {
	return wrapPolicy(p.unwrap().PrincipalIs(entityType))
}

func (p *Policy) PrincipalIsIn(entityType types.Path, entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().PrincipalIsIn(entityType, entity))
}

func (p *Policy) ActionEq(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ActionEq(entity))
}

func (p *Policy) ActionIn(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ActionIn(entity))
}

func (p *Policy) ActionInSet(entities ...types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ActionInSet(entities...))
}

func (p *Policy) ResourceEq(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ResourceEq(entity))
}

func (p *Policy) ResourceIn(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ResourceIn(entity))
}

func (p *Policy) ResourceIs(entityType types.Path) *Policy {
	return wrapPolicy(p.unwrap().ResourceIs(entityType))
}

func (p *Policy) ResourceIsIn(entityType types.Path, entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ResourceIsIn(entityType, entity))
}
