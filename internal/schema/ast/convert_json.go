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
		parts := strings.Split(name, "::")
		idents := make([]*Ident, len(parts))
		for i, part := range parts {
			idents[i] = &Ident{Value: part}
		}
		ns.Name = &Path{Parts: idents}
	}

	// Convert common types
	for name, ct := range js.CommonTypes {
		ns.Decls = append(ns.Decls, &CommonTypeDecl{
			Name:  &Ident{Value: name},
			Value: convertJsonType(ct.JsonType),
		})
	}

	// Convert entity types
	for name, et := range js.EntityTypes {
		entity := &Entity{
			Names: []*Ident{{Value: name}},
		}

		// Convert memberOfTypes
		if len(et.MemberOfTypes) > 0 {
			entity.In = make([]*Path, len(et.MemberOfTypes))
			for i, t := range et.MemberOfTypes {
				parts := strings.Split(t, "::")
				idents := make([]*Ident, len(parts))
				for j, part := range parts {
					idents[j] = &Ident{Value: part}
				}
				entity.In[i] = &Path{Parts: idents}
			}
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

		ns.Decls = append(ns.Decls, entity)
	}

	// Convert actions
	for name, act := range js.Actions {
		action := &Action{
			Names: []Name{&String{QuotedVal: fmt.Sprintf("%q", name)}},
		}

		// Convert memberOf
		if len(act.MemberOf) > 0 {
			action.In = make([]*Ref, len(act.MemberOf))
			for i, m := range act.MemberOf {
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
				action.In[i] = ref
			}
		}

		// Convert appliesTo
		if act.AppliesTo != nil {
			action.AppliesTo = &AppliesTo{}

			// Convert principal types
			if len(act.AppliesTo.PrincipalTypes) > 0 {
				action.AppliesTo.Principal = make([]*Path, len(act.AppliesTo.PrincipalTypes))
				for i, t := range act.AppliesTo.PrincipalTypes {
					parts := strings.Split(t, "::")
					idents := make([]*Ident, len(parts))
					for j, part := range parts {
						idents[j] = &Ident{Value: part}
					}
					action.AppliesTo.Principal[i] = &Path{Parts: idents}
				}
			}

			// Convert resource types
			if len(act.AppliesTo.ResourceTypes) > 0 {
				action.AppliesTo.Resource = make([]*Path, len(act.AppliesTo.ResourceTypes))
				for i, t := range act.AppliesTo.ResourceTypes {
					parts := strings.Split(t, "::")
					idents := make([]*Ident, len(parts))
					for j, part := range parts {
						idents[j] = &Ident{Value: part}
					}
					action.AppliesTo.Resource[i] = &Path{Parts: idents}
				}
			}

			// Convert context
			if act.AppliesTo.Context != nil {
				if context, ok := convertJsonType(act.AppliesTo.Context).(*RecordType); ok {
					action.AppliesTo.Context = context
				}
			}
		}

		ns.Decls = append(ns.Decls, action)
	}

	return ns
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
