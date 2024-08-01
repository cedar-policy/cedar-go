package ast

import "github.com/cedar-policy/cedar-go/types"

type scope Node

func (s scope) All() Node {
	return Node{nodeType: nodeTypeAll, args: []Node{Node(s)}}
}

func (s scope) Eq(entity types.EntityUID) Node {
	return Node(s).Equals(Entity(entity))
}

type scopeEqNode Node

func (n scopeEqNode) Entity() types.EntityUID {
	return n.args[1].value.(types.EntityUID)
}

func (s scope) In(entity types.EntityUID) Node {
	return Node(s).In(Entity(entity))
}

func (s scope) InSet(entities []types.EntityUID) Node {
	var entityValues []types.Value
	for _, e := range entities {
		entityValues = append(entityValues, e)
	}
	return Node(s).In(Set(entityValues))
}

type scopeInNode Node

func (n scopeInNode) IsSet() bool {
	return Node(n).args[1].nodeType == nodeTypeSet
}

func (n scopeInNode) Entity() types.EntityUID {
	return n.args[1].value.(types.EntityUID)
}

func (n scopeInNode) Set() []types.EntityUID {
	var res []types.EntityUID
	for _, a := range n.args[1].args {
		res = append(res, a.value.(types.EntityUID))
	}
	return res
}

func (s scope) Is(entityType types.String) Node {
	return Node(s).Is(entityType)
}

type scopeIsNode Node

func (n scopeIsNode) EntityType() types.String {
	return n.args[1].value.(types.String)
}

func (s scope) IsIn(entityType types.String, entity types.EntityUID) Node {
	return Node(s).IsIn(entityType, Entity(entity))
}

type scopeIsInNode Node

func (n scopeIsInNode) EntityType() types.String {
	return n.args[1].value.(types.String)
}

func (n scopeIsInNode) Entity() types.EntityUID {
	return n.args[2].value.(types.EntityUID)
}

func (p *Policy) PrincipalEq(entity types.EntityUID) *Policy {
	p.principal = scope(Principal()).Eq(entity)
	return p
}

func (p *Policy) PrincipalIn(entity types.EntityUID) *Policy {
	p.principal = scope(Principal()).In(entity)
	return p
}

func (p *Policy) PrincipalIs(entityType types.String) *Policy {
	p.principal = scope(Principal()).Is(entityType)
	return p
}

func (p *Policy) PrincipalIsIn(entityType types.String, entity types.EntityUID) *Policy {
	p.principal = scope(Principal()).IsIn(entityType, entity)
	return p
}

func (p *Policy) ActionEq(entity types.EntityUID) *Policy {
	p.action = scope(Action()).Eq(entity)
	return p
}

func (p *Policy) ActionIn(entity types.EntityUID) *Policy {
	p.action = scope(Action()).In(entity)
	return p
}

func (p *Policy) ActionInSet(entities ...types.EntityUID) *Policy {
	p.action = scope(Action()).InSet(entities)
	return p
}

func (p *Policy) ResourceEq(entity types.EntityUID) *Policy {
	p.resource = scope(Resource()).Eq(entity)
	return p
}

func (p *Policy) ResourceIn(entity types.EntityUID) *Policy {
	p.resource = scope(Resource()).In(entity)
	return p
}

func (p *Policy) ResourceIs(entityType types.String) *Policy {
	p.resource = scope(Resource()).Is(entityType)
	return p
}

func (p *Policy) ResourceIsIn(entityType types.String, entity types.EntityUID) *Policy {
	p.resource = scope(Resource()).IsIn(entityType, entity)
	return p
}
