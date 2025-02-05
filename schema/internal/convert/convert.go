package convert

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/schema/ast"
)

// Schema converts an AST schema to a JSON schema. The conversion process is lossy.
// Any information related to ordering, formatting, comments, etc... are lost completely.
//
// TODO: Add errors if the schema is invalid (references names that don't exist)
func Schema(n *ast.Schema) (ast.JsonSchema, error) {
	out := make(ast.JsonSchema)
	// In Cedar, all anonymous types (not under a namespace) are put into the "root" namespace,
	// which just has a name of "".
	anonymousNamespace := &ast.Namespace{}
	for _, decl := range n.Decls {
		switch decl := decl.(type) {
		case *ast.Namespace:
			out[decl.Name.String()] = convertNamespace(decl)
		default:
			anonymousNamespace.Decls = append(anonymousNamespace.Decls, decl)
		}
	}
	if len(anonymousNamespace.Decls) > 0 {
		out[""] = convertNamespace(anonymousNamespace)
	}
	return out, nil
}

func convertNamespace(n *ast.Namespace) *ast.JsonNamespace {
	jsNamespace := new(ast.JsonNamespace)
	jsNamespace.Actions = make(map[string]*ast.JsonAction)
	jsNamespace.EntityTypes = make(map[string]*ast.JsonEntity)
	jsNamespace.CommonTypes = make(map[string]*ast.JsonCommonType)

	for _, astDecl := range n.Decls {
		switch astDecl := astDecl.(type) {
		case *ast.Action:
			for _, astActionName := range astDecl.Names {
				jsAction := new(ast.JsonAction)
				jsNamespace.Actions[astActionName.String()] = jsAction
				for _, astMember := range astDecl.In {
					jsMember := &ast.JsonMember{
						ID: astMember.Name.String(),
					}
					if len(astMember.Namespace) > 0 {
						jsMember.Type = astMember.Namespace[0].String()
					}
					jsAction.MemberOf = append(jsAction.MemberOf, jsMember)
				}

				if astDecl.AppliesTo != nil {
					jsAction.AppliesTo = &ast.JsonAppliesTo{}
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
		case *ast.Entity:
			for _, name := range astDecl.Names {
				entity := new(ast.JsonEntity)
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
		case *ast.CommonTypeDecl:
			jsNamespace.CommonTypes[astDecl.Name.String()] = &ast.JsonCommonType{JsonType: convertType(astDecl.Value)}
		}
	}
	return jsNamespace
}

func convertType(t ast.Type) *ast.JsonType {
	switch t := t.(type) {
	case *ast.RecordType:
		return convertRecordType(t)
	case *ast.SetType:
		return &ast.JsonType{Type: "Set", Element: convertType(t.Element)}
	case *ast.Path:
		if len(t.Parts) == 1 {
			if t.Parts[0].Value == "Bool" {
				return &ast.JsonType{Type: "Boolean"}
			}
			if t.Parts[0].Value == "Long" {
				return &ast.JsonType{Type: "Long"}
			}
			if t.Parts[0].Value == "String" {
				return &ast.JsonType{Type: "String"}
			}
		}
		return &ast.JsonType{Type: "EntityOrCommon", Name: t.String()}
	default:
		panic(fmt.Sprintf("%T is not an AST type", t))
	}
}

func convertRecordType(t *ast.RecordType) *ast.JsonType {
	jt := new(ast.JsonType)
	jt.Type = "Record"
	jt.Attributes = make(map[string]*ast.JsonAttribute)
	for _, attr := range t.Attributes {
		jsAttr := &ast.JsonAttribute{
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
