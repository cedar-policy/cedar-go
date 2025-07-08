package ast

import (
	"fmt"
	"strings"
)

// ConvertHuman2JSON converts an AST schema to a JSON schema. The conversion process is lossy.
// Any information related to ordering, formatting, comments, etc... are lost completely.
//
// TODO: Add errors if the schema is invalid (references names that don't exist)
func ConvertHuman2JSON(n *Schema) JSONSchema {
	out := make(JSONSchema)
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

func convertNamespace(n *Namespace) *JSONNamespace {
	jsNamespace := new(JSONNamespace)
	jsNamespace.Actions = make(map[string]*JSONAction)
	jsNamespace.EntityTypes = make(map[string]*JSONEntity)
	jsNamespace.CommonTypes = make(map[string]*JSONCommonType)
	jsNamespace.Annotations = make(map[string]string)
	for _, a := range n.Annotations {
		jsNamespace.Annotations[a.Key.String()] = a.Value.String()
	}

	for _, astDecl := range n.Decls {
		switch astDecl := astDecl.(type) {
		case *Action:
			for _, astActionName := range astDecl.Names {
				jsAction := new(JSONAction)
				jsAction.Annotations = make(map[string]string)
				for _, a := range astDecl.Annotations {
					jsAction.Annotations[a.Key.String()] = a.Value.String()
				}
				jsNamespace.Actions[astActionName.String()] = jsAction
				for _, astMember := range astDecl.In {
					jsMember := &JSONMember{
						ID: astMember.Name.String(),
					}
					if len(astMember.Namespace) > 0 {
						jsMember.Type = convertIdents(astMember.Namespace)
					}
					jsAction.MemberOf = append(jsAction.MemberOf, jsMember)
				}

				if astDecl.AppliesTo != nil {
					jsAction.AppliesTo = &JSONAppliesTo{}
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
				entity := new(JSONEntity)
				entity.Annotations = make(map[string]string)
				for _, a := range astDecl.Annotations {
					entity.Annotations[a.Key.String()] = a.Value.String()
				}
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
				for _, value := range astDecl.Enum {
					entity.Enum = append(entity.Enum, value.String())
				}
				jsNamespace.EntityTypes[name.String()] = entity
			}
		case *CommonTypeDecl:
			commonType := new(JSONCommonType)
			commonType.JSONType = convertType(astDecl.Value)
			commonType.Annotations = make(map[string]string)
			for _, a := range astDecl.Annotations {
				commonType.Annotations[a.Key.String()] = a.Value.String()
			}
			jsNamespace.CommonTypes[astDecl.Name.String()] = commonType
		}
	}
	return jsNamespace
}

func convertType(t Type) *JSONType {
	switch t := t.(type) {
	case *RecordType:
		return convertRecordType(t)
	case *SetType:
		return &JSONType{Type: "Set", Element: convertType(t.Element)}
	case *Path:
		if len(t.Parts) == 1 {
			if t.Parts[0].Value == "Bool" || t.Parts[0].Value == "Boolean" {
				return &JSONType{Type: "Boolean"}
			}
			if t.Parts[0].Value == "Long" {
				return &JSONType{Type: "Long"}
			}
			if t.Parts[0].Value == "String" {
				return &JSONType{Type: "String"}
			}
		}
		return &JSONType{Type: "EntityOrCommon", Name: t.String()}
	default:
		panic(fmt.Sprintf("%T is not an AST type", t))
	}
}

func convertRecordType(t *RecordType) *JSONType {
	jt := new(JSONType)
	jt.Type = "Record"
	jt.Attributes = make(map[string]*JSONAttribute)
	for _, attr := range t.Attributes {
		jsAttr := &JSONAttribute{
			Required: attr.IsRequired,
		}
		inner := convertType(attr.Type)
		jsAttr.Type = inner.Type
		jsAttr.Element = inner.Element
		jsAttr.Name = inner.Name
		jsAttr.Attributes = inner.Attributes
		jt.Attributes[attr.Key.String()] = jsAttr
		jsAttr.Annotations = make(map[string]string)
		for _, a := range attr.Annotations {
			jsAttr.Annotations[a.Key.String()] = a.Value.String()
		}
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
