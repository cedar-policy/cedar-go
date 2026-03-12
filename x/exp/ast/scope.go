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

//sumtype:decl
type IsScopeNode interface {
	isScope()
}

//sumtype:decl
type IsPrincipalScopeNode interface {
	IsScopeNode
	isPrincipalScope()
}

//sumtype:decl
type IsActionScopeNode interface {
	IsScopeNode
	isActionScope()
}

//sumtype:decl
type IsResourceScopeNode interface {
	IsScopeNode
	isResourceScope()
}

type ScopeTypeAll struct{}

func (n ScopeTypeAll) isScope()          { _ = 0 }
func (n ScopeTypeAll) isPrincipalScope() { _ = 0 }
func (n ScopeTypeAll) isActionScope()    { _ = 0 }
func (n ScopeTypeAll) isResourceScope()  { _ = 0 }

type ScopeTypeEq struct {
	Entity types.EntityUID
}

func (n ScopeTypeEq) isScope()          { _ = 0 }
func (n ScopeTypeEq) isPrincipalScope() { _ = 0 }
func (n ScopeTypeEq) isActionScope()    { _ = 0 }
func (n ScopeTypeEq) isResourceScope()  { _ = 0 }

type ScopeTypeIn struct {
	Entity types.EntityUID
}

func (n ScopeTypeIn) isScope()          { _ = 0 }
func (n ScopeTypeIn) isPrincipalScope() { _ = 0 }
func (n ScopeTypeIn) isActionScope()    { _ = 0 }
func (n ScopeTypeIn) isResourceScope()  { _ = 0 }

type ScopeTypeInSet struct {
	Entities []types.EntityUID
}

func (n ScopeTypeInSet) isScope()       { _ = 0 }
func (n ScopeTypeInSet) isActionScope() { _ = 0 }

type ScopeTypeIs struct {
	Type types.EntityType
}

func (n ScopeTypeIs) isScope()          { _ = 0 }
func (n ScopeTypeIs) isPrincipalScope() { _ = 0 }
func (n ScopeTypeIs) isResourceScope()  { _ = 0 }

type ScopeTypeIsIn struct {
	Type   types.EntityType
	Entity types.EntityUID
}

func (n ScopeTypeIsIn) isScope()          { _ = 0 }
func (n ScopeTypeIsIn) isPrincipalScope() { _ = 0 }
func (n ScopeTypeIsIn) isResourceScope()  { _ = 0 }
