package ast

import "github.com/cedar-policy/cedar-go/types"

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

func (p *Policy) PrincipalIs(entityType types.String) *Policy {
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
	p.resource = Resource().Equals(Entity(entity))
	return p
}

func (p *Policy) ResourceIn(entities ...types.EntityUID) *Policy {
	var entityValues []types.Value
	for _, e := range entities {
		entities = append(entities, e)
	}
	p.resource = Resource().In(Set(entityValues))
	return p
}

func (p *Policy) ResourceIs(entityType types.String) *Policy {
	p.resource = Resource().Is(EntityType(entityType))
	return p
}
