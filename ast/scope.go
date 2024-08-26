package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

// PrincipalEq replaces the principal scope condition.
func (p *Policy) PrincipalEq(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().PrincipalEq(entity))
}

// PrincipalIn replaces the principal scope condition.
func (p *Policy) PrincipalIn(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().PrincipalIn(entity))
}

// PrincipalIs replaces the principal scope condition.
func (p *Policy) PrincipalIs(entityType types.EntityType) *Policy {
	return wrapPolicy(p.unwrap().PrincipalIs(entityType))
}

// PrincipalIsIn replaces the principal scope condition.
func (p *Policy) PrincipalIsIn(entityType types.EntityType, entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().PrincipalIsIn(entityType, entity))
}

// ActionEq replaces the action scope condition.
func (p *Policy) ActionEq(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ActionEq(entity))
}

// ActionIn replaces the action scope condition.
func (p *Policy) ActionIn(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ActionIn(entity))
}

// ActionInSet replaces the action scope condition.
func (p *Policy) ActionInSet(entities ...types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ActionInSet(entities...))
}

// ResourceEq replaces the resource scope condition.
func (p *Policy) ResourceEq(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ResourceEq(entity))
}

// ResourceIn replaces the resource scope condition.
func (p *Policy) ResourceIn(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ResourceIn(entity))
}

// ResourceIs replaces the resource scope condition.
func (p *Policy) ResourceIs(entityType types.EntityType) *Policy {
	return wrapPolicy(p.unwrap().ResourceIs(entityType))
}

// ResourceIsIn replaces the resource scope condition.
func (p *Policy) ResourceIsIn(entityType types.EntityType, entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ResourceIsIn(entityType, entity))
}
