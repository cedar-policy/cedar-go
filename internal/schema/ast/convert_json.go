package ast

import (
	"fmt"
	"strings"
)

// ConvertJSON2Human converts a JSON schema to a human-readable AST schema. The conversion process is lossy.
// Any information related to ordering, formatting, comments, etc... are lost completely.
func ConvertJSON2Human(js JSONSchema) *Schema {
	schema := &Schema{}

	// Handle anonymous namespace first (if it exists)
	if anon, ok := js[""]; ok {
		anonNamespace := convertJSONNamespace("", anon)
		// Append anonymous namespace declarations to the schema root
		schema.Decls = append(schema.Decls, anonNamespace.Decls...)
	}

	// Handle all other namespaces
	for name, ns := range js {
		if name != "" {
			schema.Decls = append(schema.Decls, convertJSONNamespace(name, ns))
		}
	}

	return schema
}

func convertJSONNamespace(name string, js *JSONNamespace) *Namespace {
	ns := &Namespace{}
	if name != "" {
		ns.Name = convertJSONNamespaceName(name)
	}

	// Convert annotations
	ns.Annotations = convertJSONAnnotations(js.Annotations)

	// Convert common types
	ns.Decls = append(ns.Decls, convertJSONCommonTypes(js.CommonTypes)...)

	// Convert entity types
	ns.Decls = append(ns.Decls, convertJSONEntityTypes(js.EntityTypes)...)

	// Convert actions
	ns.Decls = append(ns.Decls, convertJSONActions(js.Actions)...)

	return ns
}

func convertJSONAnnotations(annotations map[string]string) []*Annotation {
	var ans []*Annotation
	for k, v := range annotations {
		ans = append(ans, &Annotation{Key: &Ident{Value: k}, Value: &String{QuotedVal: fmt.Sprintf("%q", v)}})
	}
	return ans
}

func convertJSONNamespaceName(name string) *Path {
	parts := strings.Split(name, "::")
	idents := make([]*Ident, len(parts))
	for i, part := range parts {
		idents[i] = &Ident{Value: part}
	}
	return &Path{Parts: idents}
}

func convertJSONCommonTypes(types map[string]*JSONCommonType) []Declaration {
	decls := make([]Declaration, 0, len(types))
	for name, ct := range types {
		annotations := convertJSONAnnotations(ct.Annotations)

		decls = append(decls, &CommonTypeDecl{
			Annotations: annotations,
			Name:        &Ident{Value: name},
			Value:       convertJSONType(ct.JSONType),
		})
	}
	return decls
}

func convertJSONEntityTypes(types map[string]*JSONEntity) []Declaration {
	decls := make([]Declaration, 0, len(types))
	for name, et := range types {
		entity := &Entity{
			Names: []*Ident{{Value: name}},
		}

		// Convert annotations
		entity.Annotations = convertJSONAnnotations(et.Annotations)

		// Convert memberOfTypes
		if len(et.MemberOfTypes) > 0 {
			entity.In = convertJSONMemberOfTypes(et.MemberOfTypes)
		}

		// Convert shape
		if et.Shape != nil {
			if shape, ok := convertJSONType(et.Shape).(*RecordType); ok {
				entity.Shape = shape
			}
		}

		// Convert tags
		if et.Tags != nil {
			entity.Tags = convertJSONType(et.Tags)
		}

		// Convert enum
		for _, value := range et.Enum {
			entity.Enum = append(entity.Enum, &String{QuotedVal: fmt.Sprintf("%q", value)})
		}

		decls = append(decls, entity)
	}
	return decls
}

func convertJSONMemberOfTypes(types []string) []*Path {
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

func convertJSONActions(actions map[string]*JSONAction) []Declaration {
	decls := make([]Declaration, 0, len(actions))
	for name, act := range actions {
		action := &Action{
			Names: []Name{&String{QuotedVal: fmt.Sprintf("%q", name)}},
		}

		// Convert annotations
		action.Annotations = convertJSONAnnotations(act.Annotations)

		// Convert memberOf
		if len(act.MemberOf) > 0 {
			action.In = convertJSONMemberOf(act.MemberOf)
		}

		// Convert appliesTo
		if act.AppliesTo != nil {
			action.AppliesTo = convertJSONAppliesTo(act.AppliesTo)
		}

		decls = append(decls, action)
	}
	return decls
}

func convertJSONMemberOf(members []*JSONMember) []*Ref {
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

func convertJSONAppliesTo(appliesTo *JSONAppliesTo) *AppliesTo {
	at := &AppliesTo{}

	// Convert principal types
	if len(appliesTo.PrincipalTypes) > 0 {
		at.Principal = convertJSONMemberOfTypes(appliesTo.PrincipalTypes)
	}

	// Convert resource types
	if len(appliesTo.ResourceTypes) > 0 {
		at.Resource = convertJSONMemberOfTypes(appliesTo.ResourceTypes)
	}

	// Convert context
	if appliesTo.Context != nil {
		switch t := convertJSONType(appliesTo.Context).(type) {
		case *RecordType:
			at.ContextRecord = t
		case *Path:
			at.ContextPath = t
		}
	}

	return at
}

func convertJSONType(js *JSONType) Type {
	switch js.Type {
	case "Boolean":
		return &Path{Parts: []*Ident{{Value: "Boolean"}}}
	case "Long":
		return &Path{Parts: []*Ident{{Value: "Long"}}}
	case "String":
		return &Path{Parts: []*Ident{{Value: "String"}}}
	case "Set":
		return &SetType{
			Element: convertJSONType(js.Element),
		}
	case "Record":
		return convertJSONRecordType(js)
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

func convertJSONRecordType(js *JSONType) *RecordType {
	rt := &RecordType{
		Attributes: make([]*Attribute, 0, len(js.Attributes)),
	}

	for name, attr := range js.Attributes {
		annotations := convertJSONAnnotations(attr.Annotations)

		rt.Attributes = append(rt.Attributes, &Attribute{
			Annotations: annotations,
			Key:         &String{QuotedVal: fmt.Sprintf("%q", name)},
			IsRequired:  attr.Required,
			Type: convertJSONType(&JSONType{
				Type:       attr.Type,
				Element:    attr.Element,
				Name:       attr.Name,
				Attributes: attr.Attributes,
			}),
		})
	}

	return rt
}
