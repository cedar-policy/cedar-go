package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

type Scope struct{}

func (s Scope) All() ScopeTypeAll {
	return ScopeTypeAll{}
}

func (s Scope) Eq(entity types.EntityUID) ScopeTypeEq {
	return ScopeTypeEq{Entity: entity}
}

func (s Scope) In(entity types.EntityUID) ScopeTypeIn {
	return ScopeTypeIn{Entity: entity}
}

func (s Scope) InSet(entities []types.EntityUID) ScopeTypeInSet {
	return ScopeTypeInSet{Entities: entities}
}

func (s Scope) Is(entityType types.EntityType) ScopeTypeIs {
	return ScopeTypeIs{Type: entityType}
}

func (s Scope) IsIn(entityType types.EntityType, entity types.EntityUID) ScopeTypeIsIn {
	return ScopeTypeIsIn{Type: entityType, Entity: entity}
}

func (p *Policy) PrincipalEq(entity types.EntityUID) *Policy {
	p.Principal = Scope{}.Eq(entity)
	return p
}

func (p *Policy) PrincipalIn(entity types.EntityUID) *Policy {
	p.Principal = Scope{}.In(entity)
	return p
}

func (p *Policy) PrincipalIs(entityType types.EntityType) *Policy {
	p.Principal = Scope{}.Is(entityType)
	return p
}

func (p *Policy) PrincipalIsIn(entityType types.EntityType, entity types.EntityUID) *Policy {
	p.Principal = Scope{}.IsIn(entityType, entity)
	return p
}

func (p *Policy) ActionEq(entity types.EntityUID) *Policy {
	p.Action = Scope{}.Eq(entity)
	return p
}

func (p *Policy) ActionIn(entity types.EntityUID) *Policy {
	p.Action = Scope{}.In(entity)
	return p
}

func (p *Policy) ActionInSet(entities ...types.EntityUID) *Policy {
	p.Action = Scope{}.InSet(entities)
	return p
}

func (p *Policy) ResourceEq(entity types.EntityUID) *Policy {
	p.Resource = Scope{}.Eq(entity)
	return p
}

func (p *Policy) ResourceIn(entity types.EntityUID) *Policy {
	p.Resource = Scope{}.In(entity)
	return p
}

func (p *Policy) ResourceIs(entityType types.EntityType) *Policy {
	p.Resource = Scope{}.Is(entityType)
	return p
}

func (p *Policy) ResourceIsIn(entityType types.EntityType, entity types.EntityUID) *Policy {
	p.Resource = Scope{}.IsIn(entityType, entity)
	return p
}

type IsScopeNode interface {
	isScope()
}

type IsPrincipalScopeNode interface {
	IsScopeNode
	isPrincipalScope()
}

type IsActionScopeNode interface {
	IsScopeNode
	isActionScope()
}

type IsResourceScopeNode interface {
	IsScopeNode
	isResourceScope()
}

type ScopeNode struct{}

func (n ScopeNode) isScope() { _ = 0 } // No-op statement injected for code coverage instrumentation

type PrincipalScopeNode struct{}

func (n PrincipalScopeNode) isPrincipalScope() { _ = 0 } // No-op statement injected for code coverage instrumentation

type ActionScopeNode struct{}

func (n ActionScopeNode) isActionScope() { _ = 0 } // No-op statement injected for code coverage instrumentation

type ResourceScopeNode struct{}

func (n ResourceScopeNode) isResourceScope() { _ = 0 } // No-op statement injected for code coverage instrumentation

type ScopeTypeAll struct {
	ScopeNode
	PrincipalScopeNode
	ActionScopeNode
	ResourceScopeNode
}

type ScopeTypeEq struct {
	ScopeNode
	PrincipalScopeNode
	ActionScopeNode
	ResourceScopeNode
	Entity types.EntityUID
}

type ScopeTypeIn struct {
	ScopeNode
	PrincipalScopeNode
	ActionScopeNode
	ResourceScopeNode
	Entity types.EntityUID
}

type ScopeTypeInSet struct {
	ScopeNode
	ActionScopeNode
	Entities []types.EntityUID
}

type ScopeTypeIs struct {
	ScopeNode
	PrincipalScopeNode
	ResourceScopeNode
	Type types.EntityType
}

type ScopeTypeIsIn struct {
	ScopeNode
	PrincipalScopeNode
	ResourceScopeNode
	Type   types.EntityType
	Entity types.EntityUID
}
