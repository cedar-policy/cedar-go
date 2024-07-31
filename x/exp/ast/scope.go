package ast

import "github.com/cedar-policy/cedar-go/types"

func (p *Policy) PrincipalEq(entity types.EntityUID) *Policy {
	p.principal = Principal().Equals(Entity(entity))
	return p
}

func (p *Policy) PrincipalIn(entity types.EntityUID) *Policy {
	p.principal = Principal().In(Entity(entity))
	return p
}

func (p *Policy) PrincipalIs(entityType types.String) *Policy {
	p.principal = Principal().Is(entityType)
	return p
}

func (p *Policy) PrincipalIsIn(entityType types.String, entity types.EntityUID) *Policy {
	p.principal = Principal().IsIn(entityType, Entity(entity))
	return p
}

func (p *Policy) ActionEq(entity types.EntityUID) *Policy {
	p.action = Action().Equals(Entity(entity))
	return p
}

func (p *Policy) ActionIn(entities ...types.EntityUID) *Policy {
	if len(entities) == 1 {
		p.action = Action().In(Entity(entities[0]))
		return p
	}
	var entityValues []types.Value
	for _, e := range entities {
		entityValues = append(entityValues, e)
	}
	p.action = Action().In(Set(entityValues))
	return p
}

func (p *Policy) ResourceEq(entity types.EntityUID) *Policy {
	p.resource = Resource().Equals(Entity(entity))
	return p
}

func (p *Policy) ResourceIn(entity types.EntityUID) *Policy {
	p.resource = Resource().In(Entity(entity))
	return p
}

func (p *Policy) ResourceIs(entityType types.String) *Policy {
	p.resource = Resource().Is(entityType)
	return p
}

func (p *Policy) ResourceIsIn(entityType types.String, entity types.EntityUID) *Policy {
	p.resource = Resource().IsIn(entityType, Entity(entity))
	return p
}
