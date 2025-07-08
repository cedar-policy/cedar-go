package ast

import (
	"fmt"
	"strings"
)

// ConvertHuman2Json converts an AST schema to a JSON schema. The conversion process is lossy.
// Any information related to ordering, formatting, comments, etc... are lost completely.
//
// TODO: Add errors if the schema is invalid (references names that don't exist)
func ConvertHuman2Json(n *Schema) JsonSchema {
	out := make(JsonSchema)
	// In Cedar, all anonymous types (not under a namespace) are put into the "root" namespace,
	// which just has a name of "".
	anonymousNamespace := &Namespace{}
	for _, decl := range n.Decls {
		switch decl := decl.(type) {
		case *Namespace:
			out[decl.Name.String()] = convertNamespace(decl)
		default:
			anonymousNamespace.Decls = append(anonymousNamespace.Decls, decl)
		}
	}
	if len(anonymousNamespace.Decls) > 0 {
		out[""] = convertNamespace(anonymousNamespace)
	}
	return out
}

func convertNamespace(n *Namespace) *JsonNamespace {
	jsNamespace := new(JsonNamespace)
	jsNamespace.Actions = make(map[string]*JsonAction)
	jsNamespace.EntityTypes = make(map[string]*JsonEntity)
	jsNamespace.CommonTypes = make(map[string]*JsonCommonType)

	for _, astDecl := range n.Decls {
		switch astDecl := astDecl.(type) {
		case *Action:
			for _, astActionName := range astDecl.Names {
				jsAction := new(JsonAction)
				jsNamespace.Actions[astActionName.String()] = jsAction
				for _, astMember := range astDecl.In {
					jsMember := &JsonMember{
						ID: astMember.Name.String(),
					}
					if len(astMember.Namespace) > 0 {
						jsMember.Type = convertIdents(astMember.Namespace)
					}
					jsAction.MemberOf = append(jsAction.MemberOf, jsMember)
				}

				if astDecl.AppliesTo != nil {
					jsAction.AppliesTo = &JsonAppliesTo{}
					for _, princ := range astDecl.AppliesTo.Principal {
						jsAction.AppliesTo.PrincipalTypes = append(jsAction.AppliesTo.PrincipalTypes, princ.String())
					}
					for _, res := range astDecl.AppliesTo.Resource {
						jsAction.AppliesTo.ResourceTypes = append(jsAction.AppliesTo.ResourceTypes, res.String())
					}
					if astDecl.AppliesTo.Context != nil {
						jsAction.AppliesTo.Context = convertType(astDecl.AppliesTo.Context)
					}
				}
				jsNamespace.Actions[astActionName.String()] = jsAction
			}
		case *Entity:
			for _, name := range astDecl.Names {
				entity := new(JsonEntity)
				jsNamespace.EntityTypes[name.String()] = entity
				for _, member := range astDecl.In {
					entity.MemberOfTypes = append(entity.MemberOfTypes, member.String())
				}
				if astDecl.Shape != nil {
					entity.Shape = convertType(astDecl.Shape)
				}
				if astDecl.Tags != nil {
					entity.Tags = convertType(astDecl.Tags)
				}
				jsNamespace.EntityTypes[name.String()] = entity
			}
		case *CommonTypeDecl:
			jsNamespace.CommonTypes[astDecl.Name.String()] = &JsonCommonType{JsonType: convertType(astDecl.Value)}
		}
	}
	return jsNamespace
}

func convertType(t Type) *JsonType {
	switch t := t.(type) {
	case *RecordType:
		return convertRecordType(t)
	case *SetType:
		return &JsonType{Type: "Set", Element: convertType(t.Element)}
	case *Path:
		if len(t.Parts) == 1 {
			if t.Parts[0].Value == "Bool" {
				return &JsonType{Type: "Boolean"}
			}
			if t.Parts[0].Value == "Long" {
				return &JsonType{Type: "Long"}
			}
			if t.Parts[0].Value == "String" {
				return &JsonType{Type: "String"}
			}
		}
		return &JsonType{Type: "EntityOrCommon", Name: t.String()}
	default:
		panic(fmt.Sprintf("%T is not an AST type", t))
	}
}

func convertRecordType(t *RecordType) *JsonType {
	jt := new(JsonType)
	jt.Type = "Record"
	jt.Attributes = make(map[string]*JsonAttribute)
	for _, attr := range t.Attributes {
		jsAttr := &JsonAttribute{
			Required: attr.IsRequired,
		}
		inner := convertType(attr.Type)
		jsAttr.Type = inner.Type
		jsAttr.Element = inner.Element
		jsAttr.Name = inner.Name
		jsAttr.Attributes = inner.Attributes
		jt.Attributes[attr.Key.String()] = jsAttr
	}
	return jt
}

func convertIdents(ns []*Ident) string {
	var s []string
	for _, n := range ns {
		s = append(s, n.Value)
	}
	return strings.Join(s, "::")
}
