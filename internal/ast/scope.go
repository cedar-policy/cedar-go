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
	p.Principal = Scope(newPrincipalNode()).Eq(entity)
	return p
}

func (p *Policy) PrincipalIn(entity types.EntityUID) *Policy {
	p.Principal = Scope(newPrincipalNode()).In(entity)
	return p
}

func (p *Policy) PrincipalIs(entityType types.Path) *Policy {
	p.Principal = Scope(newPrincipalNode()).Is(entityType)
	return p
}

func (p *Policy) PrincipalIsIn(entityType types.Path, entity types.EntityUID) *Policy {
	p.Principal = Scope(newPrincipalNode()).IsIn(entityType, entity)
	return p
}

func (p *Policy) ActionEq(entity types.EntityUID) *Policy {
	p.Action = Scope(newActionNode()).Eq(entity)
	return p
}

func (p *Policy) ActionIn(entity types.EntityUID) *Policy {
	p.Action = Scope(newActionNode()).In(entity)
	return p
}

func (p *Policy) ActionInSet(entities ...types.EntityUID) *Policy {
	p.Action = Scope(newActionNode()).InSet(entities)
	return p
}

func (p *Policy) ResourceEq(entity types.EntityUID) *Policy {
	p.Resource = Scope(newResourceNode()).Eq(entity)
	return p
}

func (p *Policy) ResourceIn(entity types.EntityUID) *Policy {
	p.Resource = Scope(newResourceNode()).In(entity)
	return p
}

func (p *Policy) ResourceIs(entityType types.Path) *Policy {
	p.Resource = Scope(newResourceNode()).Is(entityType)
	return p
}

func (p *Policy) ResourceIsIn(entityType types.Path, entity types.EntityUID) *Policy {
	p.Resource = Scope(newResourceNode()).IsIn(entityType, entity)
	return p
}

type IsScopeNode interface {
	isScope()
	MarshalCedar(*bytes.Buffer)
	toNode() Node
}

type ScopeNode struct {
	Variable NodeTypeVariable
}

func (n ScopeNode) isScope() {}

type ScopeTypeAll struct {
	ScopeNode
}

func (n ScopeTypeAll) toNode() Node {
	return newNode(True().v)
}

type ScopeTypeEq struct {
	ScopeNode
	Entity types.EntityUID
}

func (n ScopeTypeEq) toNode() Node {
	return newNode(newNode(n.Variable).Equals(EntityUID(n.Entity)).v)
}

type ScopeTypeIn struct {
	ScopeNode
	Entity types.EntityUID
}

func (n ScopeTypeIn) toNode() Node {
	return newNode(newNode(n.Variable).In(EntityUID(n.Entity)).v)
}

type ScopeTypeInSet struct {
	ScopeNode
	Entities []types.EntityUID
}

func (n ScopeTypeInSet) toNode() Node {
	set := make([]types.Value, len(n.Entities))
	for i, e := range n.Entities {
		set[i] = e
	}
	return newNode(newNode(n.Variable).In(Set(set)).v)
}

type ScopeTypeIs struct {
	ScopeNode
	Type types.Path
}

func (n ScopeTypeIs) toNode() Node {
	return newNode(newNode(n.Variable).Is(n.Type).v)
}

type ScopeTypeIsIn struct {
	ScopeNode
	Type   types.Path
	Entity types.EntityUID
}

func (n ScopeTypeIsIn) toNode() Node {
	return newNode(newNode(n.Variable).IsIn(n.Type, EntityUID(n.Entity)).v)
}
