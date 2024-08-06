package ast

import (
	"bytes"

	"github.com/cedar-policy/cedar-go/types"
)

type scope nodeTypeVariable

func (s scope) All() isScopeNode {
	return scopeTypeAll{scopeNode: scopeNode{Variable: nodeTypeVariable(s)}}
}

func (s scope) Eq(entity types.EntityUID) isScopeNode {
	return scopeTypeEq{scopeNode: scopeNode{Variable: nodeTypeVariable(s)}, Entity: entity}
}

func (s scope) In(entity types.EntityUID) isScopeNode {
	return scopeTypeIn{scopeNode: scopeNode{Variable: nodeTypeVariable(s)}, Entity: entity}
}

func (s scope) InSet(entities []types.EntityUID) isScopeNode {
	return scopeTypeInSet{scopeNode: scopeNode{Variable: nodeTypeVariable(s)}, Entities: entities}
}

func (s scope) Is(entityType types.String) isScopeNode {
	return scopeTypeIs{scopeNode: scopeNode{Variable: nodeTypeVariable(s)}, Type: entityType}
}

func (s scope) IsIn(entityType types.String, entity types.EntityUID) isScopeNode {
	return scopeTypeIsIn{scopeNode: scopeNode{Variable: nodeTypeVariable(s)}, Type: entityType, Entity: entity}
}

func (p *Policy) PrincipalEq(entity types.EntityUID) *Policy {
	p.principal = scope(rawPrincipalNode()).Eq(entity)
	return p
}

func (p *Policy) PrincipalIn(entity types.EntityUID) *Policy {
	p.principal = scope(rawPrincipalNode()).In(entity)
	return p
}

func (p *Policy) PrincipalIs(entityType types.String) *Policy {
	p.principal = scope(rawPrincipalNode()).Is(entityType)
	return p
}

func (p *Policy) PrincipalIsIn(entityType types.String, entity types.EntityUID) *Policy {
	p.principal = scope(rawPrincipalNode()).IsIn(entityType, entity)
	return p
}

func (p *Policy) ActionEq(entity types.EntityUID) *Policy {
	p.action = scope(rawActionNode()).Eq(entity)
	return p
}

func (p *Policy) ActionIn(entity types.EntityUID) *Policy {
	p.action = scope(rawActionNode()).In(entity)
	return p
}

func (p *Policy) ActionInSet(entities ...types.EntityUID) *Policy {
	p.action = scope(rawActionNode()).InSet(entities)
	return p
}

func (p *Policy) ResourceEq(entity types.EntityUID) *Policy {
	p.resource = scope(rawResourceNode()).Eq(entity)
	return p
}

func (p *Policy) ResourceIn(entity types.EntityUID) *Policy {
	p.resource = scope(rawResourceNode()).In(entity)
	return p
}

func (p *Policy) ResourceIs(entityType types.String) *Policy {
	p.resource = scope(rawResourceNode()).Is(entityType)
	return p
}

func (p *Policy) ResourceIsIn(entityType types.String, entity types.EntityUID) *Policy {
	p.resource = scope(rawResourceNode()).IsIn(entityType, entity)
	return p
}

type isScopeNode interface {
	isScope()
	MarshalCedar(*bytes.Buffer)
}

type scopeNode struct {
	isScopeNode
	Variable nodeTypeVariable
}

type scopeTypeAll struct {
	scopeNode
}

type scopeTypeEq struct {
	scopeNode
	Entity types.EntityUID
}

type scopeTypeIn struct {
	scopeNode
	Entity types.EntityUID
}

type scopeTypeInSet struct {
	scopeNode
	Entities []types.EntityUID
}

type scopeTypeIs struct {
	scopeNode
	Type types.String
}

type scopeTypeIsIn struct {
	scopeNode
	Type   types.String
	Entity types.EntityUID
}
