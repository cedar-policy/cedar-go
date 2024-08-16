package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

type Scope NodeTypeVariable

func (s Scope) All() IsScopeNode {
	return ScopeTypeAll{}
}

func (s Scope) Eq(entity types.EntityUID) IsScopeNode {
	return ScopeTypeEq{Entity: entity}
}

func (s Scope) In(entity types.EntityUID) IsScopeNode {
	return ScopeTypeIn{Entity: entity}
}

func (s Scope) InSet(entities []types.EntityUID) IsScopeNode {
	return ScopeTypeInSet{Entities: entities}
}

func (s Scope) Is(entityType types.EntityType) IsScopeNode {
	return ScopeTypeIs{Type: entityType}
}

func (s Scope) IsIn(entityType types.EntityType, entity types.EntityUID) IsScopeNode {
	return ScopeTypeIsIn{Type: entityType, Entity: entity}
}

func (p *Policy) PrincipalEq(entity types.EntityUID) *Policy {
	p.Principal = Scope(NewPrincipalNode()).Eq(entity)
	return p
}

func (p *Policy) PrincipalIn(entity types.EntityUID) *Policy {
	p.Principal = Scope(NewPrincipalNode()).In(entity)
	return p
}

func (p *Policy) PrincipalIs(entityType types.EntityType) *Policy {
	p.Principal = Scope(NewPrincipalNode()).Is(entityType)
	return p
}

func (p *Policy) PrincipalIsIn(entityType types.EntityType, entity types.EntityUID) *Policy {
	p.Principal = Scope(NewPrincipalNode()).IsIn(entityType, entity)
	return p
}

func (p *Policy) ActionEq(entity types.EntityUID) *Policy {
	p.Action = Scope(NewActionNode()).Eq(entity)
	return p
}

func (p *Policy) ActionIn(entity types.EntityUID) *Policy {
	p.Action = Scope(NewActionNode()).In(entity)
	return p
}

func (p *Policy) ActionInSet(entities ...types.EntityUID) *Policy {
	p.Action = Scope(NewActionNode()).InSet(entities)
	return p
}

func (p *Policy) ResourceEq(entity types.EntityUID) *Policy {
	p.Resource = Scope(NewResourceNode()).Eq(entity)
	return p
}

func (p *Policy) ResourceIn(entity types.EntityUID) *Policy {
	p.Resource = Scope(NewResourceNode()).In(entity)
	return p
}

func (p *Policy) ResourceIs(entityType types.EntityType) *Policy {
	p.Resource = Scope(NewResourceNode()).Is(entityType)
	return p
}

func (p *Policy) ResourceIsIn(entityType types.EntityType, entity types.EntityUID) *Policy {
	p.Resource = Scope(NewResourceNode()).IsIn(entityType, entity)
	return p
}

type IsScopeNode interface {
	isScope()
}

type ScopeNode struct {
}

func (n ScopeNode) isScope() {}

type ScopeTypeAll struct {
	ScopeNode
}

type ScopeTypeEq struct {
	ScopeNode
	Entity types.EntityUID
}

type ScopeTypeIn struct {
	ScopeNode
	Entity types.EntityUID
}

type ScopeTypeInSet struct {
	ScopeNode
	Entities []types.EntityUID
}

type ScopeTypeIs struct {
	ScopeNode
	Type types.EntityType
}

type ScopeTypeIsIn struct {
	ScopeNode
	Type   types.EntityType
	Entity types.EntityUID
}
