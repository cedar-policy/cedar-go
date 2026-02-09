// Package json provides JSON marshaling and unmarshaling for Cedar schema ASTs.
package json

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

// Schema is a type alias of ast.Schema that provides JSON marshaling.
type Schema ast.Schema

// MarshalJSON encodes the schema as JSON.
func (s *Schema) MarshalJSON() ([]byte, error) {
	out := make(map[string]jsonNamespace)

	// Bare declarations go under the empty string key.
	if hasBareDecls((*ast.Schema)(s)) {
		ns, err := marshalNamespace("", ast.Namespace{
			Entities:    s.Entities,
			Enums:       s.Enums,
			Actions:     s.Actions,
			CommonTypes: s.CommonTypes,
		})
		if err != nil {
			return nil, err
		}
		out[""] = ns
	}

	for name, ns := range s.Namespaces {
		jns, err := marshalNamespace(name, ns)
		if err != nil {
			return nil, err
		}
		out[string(name)] = jns
	}
	return json.Marshal(out)
}

// UnmarshalJSON parses a JSON schema into the AST.
func (s *Schema) UnmarshalJSON(b []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	result := ast.Schema{}
	for name, data := range raw {
		var jns jsonNamespace
		if err := json.Unmarshal(data, &jns); err != nil {
			return fmt.Errorf("namespace %q: %w", name, err)
		}
		ns, err := unmarshalNamespace(jns)
		if err != nil {
			return fmt.Errorf("namespace %q: %w", name, err)
		}
		if name == "" {
			result.Entities = ns.Entities
			result.Enums = ns.Enums
			result.Actions = ns.Actions
			result.CommonTypes = ns.CommonTypes
		} else {
			if result.Namespaces == nil {
				result.Namespaces = ast.Namespaces{}
			}
			result.Namespaces[types.Path(name)] = ns
		}
	}
	*s = Schema(result)
	return nil
}

func hasBareDecls(s *ast.Schema) bool {
	return len(s.Entities) > 0 || len(s.Enums) > 0 || len(s.Actions) > 0 || len(s.CommonTypes) > 0
}

type jsonNamespace struct {
	EntityTypes map[string]jsonEntityType `json:"entityTypes"`
	Actions     map[string]jsonAction     `json:"actions"`
	CommonTypes map[string]jsonCommonType `json:"commonTypes,omitempty"`
	Annotations map[string]string         `json:"annotations,omitempty"`
}

type jsonEntityType struct {
	// Standard entity fields
	MemberOfTypes []string          `json:"memberOfTypes,omitempty"`
	Shape         *jsonType         `json:"shape,omitempty"`
	Tags          *jsonType         `json:"tags,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`

	// Enum entity field (mutually exclusive with standard fields)
	Enum []string `json:"enum,omitempty"`
}

type jsonAction struct {
	MemberOf    []jsonActionParent `json:"memberOf,omitempty"`
	AppliesTo   *jsonAppliesTo     `json:"appliesTo,omitempty"`
	Annotations map[string]string  `json:"annotations,omitempty"`
}

type jsonActionParent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type jsonAppliesTo struct {
	PrincipalTypes []string  `json:"principalTypes"`
	ResourceTypes  []string  `json:"resourceTypes"`
	Context        *jsonType `json:"context,omitempty"`
}

type jsonCommonType struct {
	jsonType
	Annotations map[string]string `json:"annotations,omitempty"`
}

type jsonType struct {
	Type                 string              `json:"type"`
	Element              *jsonType           `json:"element,omitempty"`
	Attributes           map[string]jsonAttr `json:"attributes,omitempty"`
	AdditionalAttributes bool                `json:"additionalAttributes,omitempty"`
	Name                 string              `json:"name,omitempty"`
}

type jsonAttr struct {
	jsonType
	Required    *bool             `json:"required,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

func marshalNamespace(name types.Path, ns ast.Namespace) (jsonNamespace, error) {
	jns := jsonNamespace{
		EntityTypes: make(map[string]jsonEntityType),
		Actions:     make(map[string]jsonAction),
	}

	if len(ns.Annotations) > 0 {
		jns.Annotations = marshalAnnotations(ns.Annotations)
	}

	if len(ns.CommonTypes) > 0 {
		jns.CommonTypes = make(map[string]jsonCommonType)
		for ctName, ct := range ns.CommonTypes {
			jt, err := marshalIsType(ct.Type)
			if err != nil {
				return jsonNamespace{}, err
			}
			jct := jsonCommonType{jsonType: *jt}
			if len(ct.Annotations) > 0 {
				jct.Annotations = marshalAnnotations(ct.Annotations)
			}
			jns.CommonTypes[string(ctName)] = jct
		}
	}

	for etName, entity := range ns.Entities {
		jet := jsonEntityType{}
		if len(entity.Annotations) > 0 {
			jet.Annotations = marshalAnnotations(entity.Annotations)
		}
		if len(entity.ParentTypes) > 0 {
			for _, ref := range entity.ParentTypes {
				jet.MemberOfTypes = append(jet.MemberOfTypes, string(ref))
			}
			sort.Strings(jet.MemberOfTypes)
		}
		if entity.Shape != nil {
			jt, err := marshalRecordType(entity.Shape)
			if err != nil {
				return jsonNamespace{}, err
			}
			jet.Shape = jt
		}
		if entity.Tags != nil {
			jt, err := marshalIsType(entity.Tags)
			if err != nil {
				return jsonNamespace{}, err
			}
			jet.Tags = jt
		}
		jns.EntityTypes[string(etName)] = jet
	}

	for etName, enum := range ns.Enums {
		jet := jsonEntityType{}
		if len(enum.Annotations) > 0 {
			jet.Annotations = marshalAnnotations(enum.Annotations)
		}
		for _, v := range enum.Values {
			jet.Enum = append(jet.Enum, string(v))
		}
		jns.EntityTypes[string(etName)] = jet
	}

	for actionName, action := range ns.Actions {
		ja := jsonAction{}
		if len(action.Annotations) > 0 {
			ja.Annotations = marshalAnnotations(action.Annotations)
		}
		for _, ref := range action.Parents {
			parent := jsonActionParent{
				ID: string(ref.ID),
			}
			if types.EntityType(ref.Type) != "" {
				parent.Type = string(ref.Type)
			}
			ja.MemberOf = append(ja.MemberOf, parent)
		}
		if action.AppliesTo != nil {
			jat := &jsonAppliesTo{}
			for _, p := range action.AppliesTo.Principals {
				jat.PrincipalTypes = append(jat.PrincipalTypes, string(p))
			}
			for _, r := range action.AppliesTo.Resources {
				jat.ResourceTypes = append(jat.ResourceTypes, string(r))
			}
			if action.AppliesTo.Context != nil {
				jt, err := marshalIsType(action.AppliesTo.Context)
				if err != nil {
					return jsonNamespace{}, err
				}
				jat.Context = jt
			}
			ja.AppliesTo = jat
		}
		jns.Actions[string(actionName)] = ja
	}

	return jns, nil
}

func marshalIsType(t ast.IsType) (*jsonType, error) {
	switch t := t.(type) {
	case ast.StringType:
		return &jsonType{Type: "String"}, nil
	case ast.LongType:
		return &jsonType{Type: "Long"}, nil
	case ast.BoolType:
		return &jsonType{Type: "Boolean"}, nil
	case ast.ExtensionType:
		return &jsonType{Type: "Extension", Name: string(t)}, nil
	case ast.SetType:
		elem, err := marshalIsType(t.Element)
		if err != nil {
			return nil, err
		}
		return &jsonType{Type: "Set", Element: elem}, nil
	case ast.RecordType:
		return marshalRecordType(t)
	case ast.EntityTypeRef:
		return &jsonType{Type: "Entity", Name: string(t)}, nil
	case ast.TypeRef:
		return &jsonType{Type: "EntityOrCommon", Name: string(t)}, nil
	default:
		return nil, fmt.Errorf("unknown type: %T", t)
	}
}

func marshalRecordType(rec ast.RecordType) (*jsonType, error) {
	jt := &jsonType{
		Type:       "Record",
		Attributes: make(map[string]jsonAttr),
	}
	for name, attr := range rec {
		attrType, err := marshalIsType(attr.Type)
		if err != nil {
			return nil, err
		}
		ja := jsonAttr{jsonType: *attrType}
		if attr.Optional {
			f := false
			ja.Required = &f
		}
		if len(attr.Annotations) > 0 {
			ja.Annotations = marshalAnnotations(attr.Annotations)
		}
		jt.Attributes[string(name)] = ja
	}
	return jt, nil
}

func marshalAnnotations(annotations ast.Annotations) map[string]string {
	m := make(map[string]string, len(annotations))
	for k, v := range annotations {
		m[string(k)] = string(v)
	}
	return m
}

func unmarshalNamespace(jns jsonNamespace) (ast.Namespace, error) {
	ns := ast.Namespace{}

	if len(jns.Annotations) > 0 {
		ns.Annotations = unmarshalAnnotations(jns.Annotations)
	}

	for ctName, jct := range jns.CommonTypes {
		t, err := unmarshalType(&jct.jsonType)
		if err != nil {
			return ast.Namespace{}, fmt.Errorf("common type %q: %w", ctName, err)
		}
		if ns.CommonTypes == nil {
			ns.CommonTypes = ast.CommonTypes{}
		}
		ct := ast.CommonType{Type: t}
		if len(jct.Annotations) > 0 {
			ct.Annotations = unmarshalAnnotations(jct.Annotations)
		}
		ns.CommonTypes[types.Ident(ctName)] = ct
	}

	for etName, jet := range jns.EntityTypes {
		if len(jet.Enum) > 0 {
			enum := ast.Enum{}
			if len(jet.Annotations) > 0 {
				enum.Annotations = unmarshalAnnotations(jet.Annotations)
			}
			for _, v := range jet.Enum {
				enum.Values = append(enum.Values, types.String(v))
			}
			if ns.Enums == nil {
				ns.Enums = ast.Enums{}
			}
			ns.Enums[types.Ident(etName)] = enum
		} else {
			entity := ast.Entity{}
			if len(jet.Annotations) > 0 {
				entity.Annotations = unmarshalAnnotations(jet.Annotations)
			}
			for _, ref := range jet.MemberOfTypes {
				entity.ParentTypes = append(entity.ParentTypes, ast.EntityTypeRef(ref))
			}
			if jet.Shape != nil {
				rec, err := unmarshalRecordType(jet.Shape)
				if err != nil {
					return ast.Namespace{}, fmt.Errorf("entity %q shape: %w", etName, err)
				}
				entity.Shape = rec
			}
			if jet.Tags != nil {
				t, err := unmarshalType(jet.Tags)
				if err != nil {
					return ast.Namespace{}, fmt.Errorf("entity %q tags: %w", etName, err)
				}
				entity.Tags = t
			}
			if ns.Entities == nil {
				ns.Entities = ast.Entities{}
			}
			ns.Entities[types.Ident(etName)] = entity
		}
	}

	for actionName, ja := range jns.Actions {
		action := ast.Action{}
		if len(ja.Annotations) > 0 {
			action.Annotations = unmarshalAnnotations(ja.Annotations)
		}
		for _, parent := range ja.MemberOf {
			if parent.Type == "" {
				action.Parents = append(action.Parents, ast.ParentRefFromID(types.String(parent.ID)))
			} else {
				action.Parents = append(action.Parents, ast.NewParentRef(ast.EntityTypeRef(parent.Type), types.String(parent.ID)))
			}
		}
		if ja.AppliesTo != nil {
			at := &ast.AppliesTo{}
			for _, p := range ja.AppliesTo.PrincipalTypes {
				at.Principals = append(at.Principals, ast.EntityTypeRef(p))
			}
			for _, r := range ja.AppliesTo.ResourceTypes {
				at.Resources = append(at.Resources, ast.EntityTypeRef(r))
			}
			if ja.AppliesTo.Context != nil {
				t, err := unmarshalType(ja.AppliesTo.Context)
				if err != nil {
					return ast.Namespace{}, fmt.Errorf("action %q context: %w", actionName, err)
				}
				at.Context = t
			}
			action.AppliesTo = at
		}
		if ns.Actions == nil {
			ns.Actions = ast.Actions{}
		}
		ns.Actions[types.String(actionName)] = action
	}

	return ns, nil
}

func unmarshalType(jt *jsonType) (ast.IsType, error) {
	switch jt.Type {
	case "String":
		return ast.StringType{}, nil
	case "Long":
		return ast.LongType{}, nil
	case "Boolean":
		return ast.BoolType{}, nil
	case "Extension":
		return ast.ExtensionType(jt.Name), nil
	case "Set":
		if jt.Element == nil {
			return nil, fmt.Errorf("set type missing element")
		}
		elem, err := unmarshalType(jt.Element)
		if err != nil {
			return nil, err
		}
		return ast.Set(elem), nil
	case "Record":
		return unmarshalRecordType(jt)
	case "Entity":
		return ast.EntityTypeRef(jt.Name), nil
	case "EntityOrCommon":
		return ast.TypeRef(jt.Name), nil
	default:
		return nil, fmt.Errorf("unknown type %q", jt.Type)
	}
}

func unmarshalRecordType(jt *jsonType) (ast.RecordType, error) {
	rec := ast.RecordType{}
	for name, ja := range jt.Attributes {
		t, err := unmarshalType(&ja.jsonType)
		if err != nil {
			return nil, fmt.Errorf("attribute %q: %w", name, err)
		}
		attr := ast.Attribute{
			Type:     t,
			Optional: ja.Required != nil && !*ja.Required,
		}
		if len(ja.Annotations) > 0 {
			attr.Annotations = unmarshalAnnotations(ja.Annotations)
		}
		rec[types.String(name)] = attr
	}
	return rec, nil
}

func unmarshalAnnotations(m map[string]string) ast.Annotations {
	a := make(ast.Annotations, len(m))
	for k, v := range m {
		a[types.Ident(k)] = types.String(v)
	}
	return a
}
