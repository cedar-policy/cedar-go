package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

type Scope struct{}

func (s Scope) All() ScopeTypeAll {
	return ScopeTypeAll{}
}

func (s Scope) Eq(entity types.EntityReference) ScopeTypeEq {
	return ScopeTypeEq{Entity: entity}
}

func (s Scope) In(entity types.EntityReference) ScopeTypeIn {
	return ScopeTypeIn{Entity: entity}
}

func (s Scope) InSet(entities []types.EntityUID) ScopeTypeInSet {
	return ScopeTypeInSet{Entities: entities}
}

func (s Scope) Is(entityType types.EntityType) ScopeTypeIs {
	return ScopeTypeIs{Type: entityType}
}

func (s Scope) IsIn(entityType types.EntityType, entity types.EntityReference) ScopeTypeIsIn {
	return ScopeTypeIsIn{Type: entityType, Entity: entity}
}

func (p *Policy) PrincipalEq(entity types.EntityReference) *Policy {
	p.Principal = Scope{}.Eq(entity)
	return p.addSlot(entity)
}

func (p *Policy) PrincipalIn(entity types.EntityReference) *Policy {
	p.Principal = Scope{}.In(entity)
	return p.addSlot(entity)
}

func (p *Policy) PrincipalIs(entityType types.EntityType) *Policy {
	p.Principal = Scope{}.Is(entityType)

	return p
}

func (p *Policy) PrincipalIsIn(entityType types.EntityType, entity types.EntityReference) *Policy {
	p.Principal = Scope{}.IsIn(entityType, entity)

	return p.addSlot(entity)
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

func (p *Policy) ResourceEq(entity types.EntityReference) *Policy {
	p.Resource = Scope{}.Eq(entity)
	return p.addSlot(entity)
}

func (p *Policy) ResourceIn(entity types.EntityReference) *Policy {
	p.Resource = Scope{}.In(entity)
	return p.addSlot(entity)
}

func (p *Policy) ResourceIs(entityType types.EntityType) *Policy {
	p.Resource = Scope{}.Is(entityType)
	return p
}

func (p *Policy) ResourceIsIn(entityType types.EntityType, entity types.EntityReference) *Policy {
	p.Resource = Scope{}.IsIn(entityType, entity)
	return p.addSlot(entity)
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

func (t ScopeTypeAll) Slot() (slotID types.SlotID, found bool) {
	return "", false
}

type ScopeTypeEq struct {
	ScopeNode
	PrincipalScopeNode
	ActionScopeNode
	ResourceScopeNode
	Entity types.EntityReference
}

func (t ScopeTypeEq) Slot() (slotID types.SlotID, found bool) {
	switch et := t.Entity.(type) {
	case types.SlotID:
		slotID = et
		found = true
	}

	return
}

type ScopeTypeIn struct {
	ScopeNode
	PrincipalScopeNode
	ActionScopeNode
	ResourceScopeNode
	Entity types.EntityReference
}

func (t ScopeTypeIn) Slot() (slotID types.SlotID, found bool) {
	switch et := t.Entity.(type) {
	case types.SlotID:
		slotID = et
		found = true
	}

	return
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

func (t ScopeTypeIs) Slot() (slotID types.SlotID, found bool) {
	return "", false
}

type ScopeTypeIsIn struct {
	ScopeNode
	PrincipalScopeNode
	ResourceScopeNode
	Type   types.EntityType
	Entity types.EntityReference
}

func (t ScopeTypeIsIn) Slot() (slotID types.SlotID, found bool) {
	switch et := t.Entity.(type) {
	case types.SlotID:
		slotID = et
		found = true
	}

	return
}
