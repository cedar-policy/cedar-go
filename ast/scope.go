package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

// This builder will replace the previous principal scope condition.
func (p *Policy) PrincipalEq(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().PrincipalEq(entity))
}

// This builder will replace the previous principal scope condition.
func (p *Policy) PrincipalIn(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().PrincipalIn(entity))
}

// This builder will replace the previous principal scope condition.
func (p *Policy) PrincipalIs(entityType types.Path) *Policy {
	return wrapPolicy(p.unwrap().PrincipalIs(entityType))
}

// This builder will replace the previous principal scope condition.
func (p *Policy) PrincipalIsIn(entityType types.Path, entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().PrincipalIsIn(entityType, entity))
}

// This builder will replace the previous action scope condition.
func (p *Policy) ActionEq(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ActionEq(entity))
}

// This builder will replace the previous action scope condition.
func (p *Policy) ActionIn(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ActionIn(entity))
}

// This builder will replace the previous action scope condition.
func (p *Policy) ActionInSet(entities ...types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ActionInSet(entities...))
}

// This builder will replace the previous resource scope condition.
func (p *Policy) ResourceEq(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ResourceEq(entity))
}

// This builder will replace the previous resource scope condition.
func (p *Policy) ResourceIn(entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ResourceIn(entity))
}

// This builder will replace the previous resource scope condition.
func (p *Policy) ResourceIs(entityType types.Path) *Policy {
	return wrapPolicy(p.unwrap().ResourceIs(entityType))
}

// This builder will replace the previous resource scope condition.
func (p *Policy) ResourceIsIn(entityType types.Path, entity types.EntityUID) *Policy {
	return wrapPolicy(p.unwrap().ResourceIsIn(entityType, entity))
}
