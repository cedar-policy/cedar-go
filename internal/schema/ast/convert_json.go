package ast

import (
	"fmt"
	"strings"
)

// ConvertJSON2Human converts a JSON schema to a human-readable AST schema. The conversion process is lossy.
// Any information related to ordering, formatting, comments, etc... are lost completely.
func ConvertJSON2Human(js JsonSchema) *Schema {
	schema := &Schema{}

	// Handle anonymous namespace first (if it exists)
	if anon, ok := js[""]; ok {
		anonNamespace := convertJsonNamespace("", anon)
		// Append anonymous namespace declarations to the schema root
		schema.Decls = append(schema.Decls, anonNamespace.Decls...)
	}

	// Handle all other namespaces
	for name, ns := range js {
		if name != "" {
			schema.Decls = append(schema.Decls, convertJsonNamespace(name, ns))
		}
	}

	return schema
}

func convertJsonNamespace(name string, js *JsonNamespace) *Namespace {
	ns := &Namespace{}
	if name != "" {
		ns.Name = convertJsonNamespaceName(name)
	}

	// Convert common types
	ns.Decls = append(ns.Decls, convertJsonCommonTypes(js.CommonTypes)...)

	// Convert entity types
	ns.Decls = append(ns.Decls, convertJsonEntityTypes(js.EntityTypes)...)

	// Convert actions
	ns.Decls = append(ns.Decls, convertJsonActions(js.Actions)...)

	return ns
}

func convertJsonNamespaceName(name string) *Path {
	parts := strings.Split(name, "::")
	idents := make([]*Ident, len(parts))
	for i, part := range parts {
		idents[i] = &Ident{Value: part}
	}
	return &Path{Parts: idents}
}

func convertJsonCommonTypes(types map[string]*JsonCommonType) []Declaration {
	decls := make([]Declaration, 0, len(types))
	for name, ct := range types {
		decls = append(decls, &CommonTypeDecl{
			Name:  &Ident{Value: name},
			Value: convertJsonType(ct.JsonType),
		})
	}
	return decls
}

func convertJsonEntityTypes(types map[string]*JsonEntity) []Declaration {
	decls := make([]Declaration, 0, len(types))
	for name, et := range types {
		entity := &Entity{
			Names: []*Ident{{Value: name}},
		}

		// Convert memberOfTypes
		if len(et.MemberOfTypes) > 0 {
			entity.In = convertJsonMemberOfTypes(et.MemberOfTypes)
		}

		// Convert shape
		if et.Shape != nil {
			if shape, ok := convertJsonType(et.Shape).(*RecordType); ok {
				entity.Shape = shape
			}
		}

		// Convert tags
		if et.Tags != nil {
			entity.Tags = convertJsonType(et.Tags)
		}

		decls = append(decls, entity)
	}
	return decls
}

func convertJsonMemberOfTypes(types []string) []*Path {
	paths := make([]*Path, len(types))
	for i, t := range types {
		parts := strings.Split(t, "::")
		idents := make([]*Ident, len(parts))
		for j, part := range parts {
			idents[j] = &Ident{Value: part}
		}
		paths[i] = &Path{Parts: idents}
	}
	return paths
}

func convertJsonActions(actions map[string]*JsonAction) []Declaration {
	decls := make([]Declaration, 0, len(actions))
	for name, act := range actions {
		action := &Action{
			Names: []Name{&String{QuotedVal: fmt.Sprintf("%q", name)}},
		}

		// Convert memberOf
		if len(act.MemberOf) > 0 {
			action.In = convertJsonMemberOf(act.MemberOf)
		}

		// Convert appliesTo
		if act.AppliesTo != nil {
			action.AppliesTo = convertJsonAppliesTo(act.AppliesTo)
		}

		decls = append(decls, action)
	}
	return decls
}

func convertJsonMemberOf(members []*JsonMember) []*Ref {
	refs := make([]*Ref, len(members))
	for i, m := range members {
		ref := &Ref{
			Name: &String{QuotedVal: fmt.Sprintf("%q", m.ID)},
		}
		if m.Type != "" {
			parts := strings.Split(m.Type, "::")
			ref.Namespace = make([]*Ident, len(parts))
			for j, part := range parts {
				ref.Namespace[j] = &Ident{Value: part}
			}
		}
		refs[i] = ref
	}
	return refs
}

func convertJsonAppliesTo(appliesTo *JsonAppliesTo) *AppliesTo {
	at := &AppliesTo{}

	// Convert principal types
	if len(appliesTo.PrincipalTypes) > 0 {
		at.Principal = convertJsonMemberOfTypes(appliesTo.PrincipalTypes)
	}

	// Convert resource types
	if len(appliesTo.ResourceTypes) > 0 {
		at.Resource = convertJsonMemberOfTypes(appliesTo.ResourceTypes)
	}

	// Convert context
	if appliesTo.Context != nil {
		if context, ok := convertJsonType(appliesTo.Context).(*RecordType); ok {
			at.Context = context
		}
	}

	return at
}

func convertJsonType(js *JsonType) Type {
	switch js.Type {
	case "Boolean":
		return &Path{Parts: []*Ident{{Value: "Bool"}}}
	case "Long":
		return &Path{Parts: []*Ident{{Value: "Long"}}}
	case "String":
		return &Path{Parts: []*Ident{{Value: "String"}}}
	case "Set":
		return &SetType{
			Element: convertJsonType(js.Element),
		}
	case "Record":
		return convertJsonRecordType(js)
	case "EntityOrCommon":
		parts := strings.Split(js.Name, "::")
		idents := make([]*Ident, len(parts))
		for i, part := range parts {
			idents[i] = &Ident{Value: part}
		}
		return &Path{Parts: idents}
	default:
		panic(fmt.Sprintf("unknown JSON type: %s", js.Type))
	}
}

func convertJsonRecordType(js *JsonType) *RecordType {
	rt := &RecordType{
		Attributes: make([]*Attribute, 0, len(js.Attributes)),
	}

	for name, attr := range js.Attributes {
		rt.Attributes = append(rt.Attributes, &Attribute{
			Key:        &String{QuotedVal: fmt.Sprintf("%q", name)},
			IsRequired: attr.Required,
			Type: convertJsonType(&JsonType{
				Type:       attr.Type,
				Element:    attr.Element,
				Name:       attr.Name,
				Attributes: attr.Attributes,
			}),
		})
	}

	return rt
}
