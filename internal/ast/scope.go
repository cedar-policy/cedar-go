package ast

import (
	"bytes"

	"github.com/cedar-policy/cedar-go/types"
)

type Scope NodeTypeVariable

func (s Scope) All() IsScopeNode {
	return ScopeTypeAll{ScopeNode: ScopeNode{Variable: NodeTypeVariable(s)}}
}

func (s Scope) Eq(entity types.EntityUID) IsScopeNode {
	return ScopeTypeEq{ScopeNode: ScopeNode{Variable: NodeTypeVariable(s)}, Entity: entity}
}

func (s Scope) In(entity types.EntityUID) IsScopeNode {
	return ScopeTypeIn{ScopeNode: ScopeNode{Variable: NodeTypeVariable(s)}, Entity: entity}
}

func (s Scope) InSet(entities []types.EntityUID) IsScopeNode {
	return ScopeTypeInSet{ScopeNode: ScopeNode{Variable: NodeTypeVariable(s)}, Entities: entities}
}

func (s Scope) Is(entityType types.Path) IsScopeNode {
	return ScopeTypeIs{ScopeNode: ScopeNode{Variable: NodeTypeVariable(s)}, Type: entityType}
}

func (s Scope) IsIn(entityType types.Path, entity types.EntityUID) IsScopeNode {
	return ScopeTypeIsIn{ScopeNode: ScopeNode{Variable: NodeTypeVariable(s)}, Type: entityType, Entity: entity}
}

func (p *Policy) PrincipalEq(entity types.EntityUID) *Policy {
	p.Principal = Scope(NewPrincipalNode()).Eq(entity)
	return p
}

func (p *Policy) PrincipalIn(entity types.EntityUID) *Policy {
	p.Principal = Scope(NewPrincipalNode()).In(entity)
	return p
}

func (p *Policy) PrincipalIs(entityType types.Path) *Policy {
	p.Principal = Scope(NewPrincipalNode()).Is(entityType)
	return p
}

func (p *Policy) PrincipalIsIn(entityType types.Path, entity types.EntityUID) *Policy {
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

func (p *Policy) ResourceIs(entityType types.Path) *Policy {
	p.Resource = Scope(NewResourceNode()).Is(entityType)
	return p
}

func (p *Policy) ResourceIsIn(entityType types.Path, entity types.EntityUID) *Policy {
	p.Resource = Scope(NewResourceNode()).IsIn(entityType, entity)
	return p
}

type IsScopeNode interface {
	isScope()
	MarshalCedar(*bytes.Buffer)
}

type ScopeNode struct {
	Variable NodeTypeVariable
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
	Type types.Path
}

type ScopeTypeIsIn struct {
	ScopeNode
	Type   types.Path
	Entity types.EntityUID
}
