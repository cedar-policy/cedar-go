package ast

import "github.com/cedar-policy/cedar-go/x/exp/types"

func (p *Policy) PrincipalEq(entity types.EntityUID) *Policy {
	p.principal = Principal().Equals(Entity(entity))
	return p
}

func (p *Policy) PrincipalIn(entities ...types.EntityUID) *Policy {
	var entityValues []types.Value
	for _, e := range entities {
		entities = append(entities, e)
	}
	p.principal = Principal().In(Set(entityValues))
	return p
}

func (p *Policy) PrincipalIs(entityType types.EntityType) *Policy {
	p.principal = Principal().Is(EntityType(entityType))
	return p
}

func (p *Policy) ActionEq(entity types.EntityUID) *Policy {
	p.action = Action().Equals(Entity(entity))
	return p
}

func (p *Policy) ActionIn(entities ...types.EntityUID) *Policy {
	var entityValues []types.Value
	for _, e := range entities {
		entities = append(entities, e)
	}
	p.action = Action().In(Set(entityValues))
	return p
}

func (p *Policy) ResourceEq(entity types.EntityUID) *Policy {
	p.principal = Resource().Equals(Entity(entity))
	return p
}

func (p *Policy) ResourceIn(entities ...types.EntityUID) *Policy {
	var entityValues []types.Value
	for _, e := range entities {
		entities = append(entities, e)
	}
	p.principal = Resource().In(Set(entityValues))
	return p
}

func (p *Policy) ResourceIs(entityType types.EntityType) *Policy {
	p.principal = Resource().Is(EntityType(entityType))
	return p
}
