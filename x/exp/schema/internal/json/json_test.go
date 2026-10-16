package json_test

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
	schemajson "github.com/cedar-policy/cedar-go/x/exp/schema/internal/json"
)

func TestRoundTripEmpty(t *testing.T) {
	s := ast.Schema{}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))
	testutil.Equals(t, (*ast.Schema)(&s2), &ast.Schema{})
}

func TestRoundTripEntity(t *testing.T) {
	s := ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Entities: ast.Entities{
					"User": ast.Entity{
						ParentTypes: []ast.EntityTypeRef{"Group"},
						Shape: ast.RecordType{
							"name": ast.Attribute{Type: ast.StringType{}},
							"age":  ast.Attribute{Type: ast.LongType{}, Optional: true},
						},
						Tags: ast.StringType{},
					},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	user := (*ast.Schema)(&s2).Namespaces["NS"].Entities["User"]
	testutil.Equals(t, user.ParentTypes, []ast.EntityTypeRef{"Group"})
	testutil.Equals(t, user.Shape != nil, true)
	testutil.Equals(t, user.Tags != nil, true)
}

func TestRoundTripEnum(t *testing.T) {
	s := ast.Schema{
		Enums: ast.Enums{
			"Status": ast.Enum{
				Values: []types.String{"active", "inactive"},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	status := (*ast.Schema)(&s2).Enums["Status"]
	testutil.Equals(t, status.Values, []types.String{"active", "inactive"})
}

func TestRoundTripAction(t *testing.T) {
	s := ast.Schema{
		Actions: ast.Actions{
			"view": ast.Action{
				Parents: []ast.ParentRef{
					ast.NewParentRef("NS::Action", "readOnly"),
					ast.ParentRefFromID("write"),
				},
				AppliesTo: &ast.AppliesTo{
					Principals: []ast.EntityTypeRef{"User"},
					Resources:  []ast.EntityTypeRef{"Photo"},
					Context:    ast.RecordType{},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	view := (*ast.Schema)(&s2).Actions["view"]
	testutil.Equals(t, len(view.Parents), 2)
	testutil.Equals(t, view.AppliesTo != nil, true)
	testutil.Equals(t, len(view.AppliesTo.Principals), 1)
}

func TestRoundTripCommonType(t *testing.T) {
	s := ast.Schema{
		CommonTypes: ast.CommonTypes{
			"Context": ast.CommonType{
				Annotations: ast.Annotations{"doc": "context type"},
				Type: ast.RecordType{
					"ip": ast.Attribute{Type: ast.ExtensionType("ipaddr")},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	ct := (*ast.Schema)(&s2).CommonTypes["Context"]
	testutil.Equals(t, ct.Annotations["doc"], types.String("context type"))
}

func TestRoundTripAllTypes(t *testing.T) {
	s := ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: ast.RecordType{
					"s":    ast.Attribute{Type: ast.StringType{}},
					"l":    ast.Attribute{Type: ast.LongType{}},
					"b":    ast.Attribute{Type: ast.BoolType{}},
					"ip":   ast.Attribute{Type: ast.ExtensionType("ipaddr")},
					"set":  ast.Attribute{Type: ast.Set(ast.LongType{})},
					"rec":  ast.Attribute{Type: ast.RecordType{}},
					"ref":  ast.Attribute{Type: ast.EntityTypeRef("Other")},
					"tref": ast.Attribute{Type: ast.TypeRef("CommonT")},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))
	testutil.Equals(t, len((*ast.Schema)(&s2).Entities["User"].Shape), 8)
}

func TestRoundTripAnnotations(t *testing.T) {
	s := ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Annotations: ast.Annotations{"doc": "user entity"},
				Shape: ast.RecordType{
					"name": ast.Attribute{
						Type:        ast.StringType{},
						Annotations: ast.Annotations{"doc": "user name"},
					},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	user := (*ast.Schema)(&s2).Entities["User"]
	testutil.Equals(t, user.Annotations["doc"], types.String("user entity"))
	testutil.Equals(t, user.Shape["name"].Annotations["doc"], types.String("user name"))
}

func TestRoundTripNamespaceAnnotations(t *testing.T) {
	s := ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Annotations: ast.Annotations{"doc": "my ns"},
				Entities:    ast.Entities{},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	ns := (*ast.Schema)(&s2).Namespaces["NS"]
	testutil.Equals(t, ns.Annotations["doc"], types.String("my ns"))
}

func TestRoundTripActionNoAppliesTo(t *testing.T) {
	s := ast.Schema{
		Actions: ast.Actions{
			"view": ast.Action{},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	view := (*ast.Schema)(&s2).Actions["view"]
	testutil.Equals(t, view.AppliesTo == nil, true)
}

func TestRoundTripActionEmptyLists(t *testing.T) {
	s := ast.Schema{
		Actions: ast.Actions{
			"view": ast.Action{
				AppliesTo: &ast.AppliesTo{
					Principals: []ast.EntityTypeRef{},
					Resources:  []ast.EntityTypeRef{},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	view := (*ast.Schema)(&s2).Actions["view"]
	testutil.Equals(t, view.AppliesTo != nil, true)
	testutil.Equals(t, len(view.AppliesTo.Principals), 0)
	testutil.Equals(t, len(view.AppliesTo.Resources), 0)
}

func TestUnmarshalBadJSON(t *testing.T) {
	var s schemajson.Schema
	testutil.Error(t, s.UnmarshalJSON([]byte(`not json`)))
}

func TestUnmarshalBadNamespace(t *testing.T) {
	var s schemajson.Schema
	testutil.Error(t, s.UnmarshalJSON([]byte(`{"NS": "bad"}`)))
}

func TestUnmarshalEmptyNamespaceAnnotations(t *testing.T) {
	var s schemajson.Schema
	testutil.Error(t, s.UnmarshalJSON([]byte(`{"": {"entityTypes": {}, "actions": {}, "annotations": {"doc": "x"}}}`)))
	testutil.OK(t, s.UnmarshalJSON([]byte(`{"": {"entityTypes": {}, "actions": {}, "annotations": {}}}`)))
}

func TestUnmarshalShorthandCommonTypeRef(t *testing.T) {
	// Per the Cedar JSON schema spec, the "type" field may directly hold a
	// common-type or entity-type name. The parser must accept this form;
	// undefined-name validation happens at resolve time.
	var s schemajson.Schema
	testutil.OK(t, s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"Foo": {"tags": {"type": "Unknown"}}}, "actions": {}}}`)))
}

func TestUnmarshalSetMissingElement(t *testing.T) {
	var s schemajson.Schema
	testutil.Error(t, s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"Foo": {"tags": {"type": "Set"}}}, "actions": {}}}`)))
}

func TestMarshalUnmarshalJSON(t *testing.T) {
	input := `{
		"NS": {
			"entityTypes": {
				"User": {
					"memberOfTypes": ["Group"],
					"shape": {
						"type": "Record",
						"attributes": {
							"name": {"type": "String"},
							"age": {"type": "Long", "required": false}
						}
					}
				},
				"Group": {}
			},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["Photo"],
						"context": {"type": "Record", "attributes": {}}
					}
				}
			},
			"commonTypes": {
				"Context": {"type": "Record", "attributes": {}}
			}
		}
	}`

	var s schemajson.Schema
	testutil.OK(t, json.Unmarshal([]byte(input), &s))

	ns := (*ast.Schema)(&s).Namespaces["NS"]
	testutil.Equals(t, len(ns.Entities), 2)
	testutil.Equals(t, len(ns.Actions), 1)
	testutil.Equals(t, len(ns.CommonTypes), 1)

	b, err := json.Marshal(&s)
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, json.Unmarshal(b, &s2))
	testutil.Equals(t, len((*ast.Schema)(&s2).Namespaces["NS"].Entities), 2)
}

func TestRoundTripEnumAnnotations(t *testing.T) {
	s := ast.Schema{
		Enums: ast.Enums{
			"Status": ast.Enum{
				Annotations: ast.Annotations{"doc": "status enum"},
				Values:      []types.String{"active"},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	status := (*ast.Schema)(&s2).Enums["Status"]
	testutil.Equals(t, status.Annotations["doc"], types.String("status enum"))
}

func TestRoundTripActionContext(t *testing.T) {
	s := ast.Schema{
		Actions: ast.Actions{
			"view": ast.Action{
				AppliesTo: &ast.AppliesTo{
					Principals: []ast.EntityTypeRef{"User"},
					Resources:  []ast.EntityTypeRef{"Photo"},
					Context: ast.RecordType{
						"ip": ast.Attribute{Type: ast.ExtensionType("ipaddr")},
					},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	view := (*ast.Schema)(&s2).Actions["view"]
	testutil.Equals(t, view.AppliesTo.Context != nil, true)
}

func TestMarshalActionParentWithoutType(t *testing.T) {
	// An action parent without a type is written without "type", rather than
	// with "type":"", which isn't a valid entity type.
	s := ast.Schema{
		Actions: ast.Actions{
			"view":  ast.Action{Parents: []ast.ParentRef{ast.ParentRefFromID("write")}},
			"write": ast.Action{},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)
	testutil.Equals(t, string(b), `{"":{"entityTypes":{},"actions":{"view":{"memberOf":[{"id":"write"}]},"write":{}}}}`)
}

func TestMarshalSorted(t *testing.T) {
	t.Parallel()
	s := ast.Schema{
		Entities: ast.Entities{
			"b": ast.Entity{
				Annotations: ast.Annotations{"z": "1", "y": "2"},
				Shape: ast.RecordType{
					"q": ast.Attribute{Type: ast.StringType{}},
					"p": ast.Attribute{Type: ast.LongType{}, Optional: true},
				},
			},
			"a": ast.Entity{},
		},
		Actions: ast.Actions{
			"view": ast.Action{AppliesTo: &ast.AppliesTo{}},
			"edit": ast.Action{},
		},
		Namespaces: ast.Namespaces{
			"NS2": ast.Namespace{},
			"NS1": ast.Namespace{CommonTypes: ast.CommonTypes{
				"T2": ast.CommonType{Type: ast.BoolType{}},
				"T1": ast.CommonType{Type: ast.StringType{}},
			}},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)
	// Every map is written sorted by key.
	testutil.Equals(t, string(b), `{"":{"entityTypes":{"a":{},"b":{"shape":{"type":"Record","attributes":{"p":{"type":"Long","required":false},"q":{"type":"String"}}},"annotations":{"y":"2","z":"1"}}},"actions":{"edit":{},"view":{"appliesTo":{"principalTypes":[],"resourceTypes":[]}}}},"NS1":{"entityTypes":{},"actions":{},"commonTypes":{"T1":{"type":"String"},"T2":{"type":"Boolean"}}},"NS2":{"entityTypes":{},"actions":{}}}`)
}

func TestUnmarshalErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
	}{
		{"duplicateNamespace", `{"NS":{"entityTypes":{},"actions":{}},"NS":{"entityTypes":{},"actions":{}}}`},
		{"duplicateEntityType", `{"":{"entityTypes":{"A":{},"A":{}},"actions":{}}}`},
		{"duplicateMember", `{"":{"entityTypes":{},"entityTypes":{},"actions":{}}}`},
		{"invalidUTF8", "{\"\":{\"entityTypes\":{\"\xff\":{}},\"actions\":{}}}"},
		{"trailingData", `{"":{"entityTypes":{},"actions":{}}} {}`},
		{"unknownNamespaceMember", `{"":{"entityTypes":{},"actions":{},"x":1}}`},
		{"unknownEntityTypeMember", `{"":{"entityTypes":{"A":{"x":1}},"actions":{}}}`},
		{"unknownActionMember", `{"":{"entityTypes":{},"actions":{"a":{"x":1}}}}`},
		{"unknownActionParentMember", `{"":{"entityTypes":{},"actions":{"a":{"memberOf":[{"id":"b","x":1}]},"b":{}}}}`},
		{"unknownAppliesToMember", `{"":{"entityTypes":{},"actions":{"a":{"appliesTo":{"principalTypes":[],"resourceTypes":[],"x":1}}}}}`},
		{"unknownTypeMember", `{"":{"entityTypes":{"A":{"tags":{"type":"String","x":1}}},"actions":{}}}`},
		{"unknownAttributeMember", `{"":{"entityTypes":{"A":{"shape":{"type":"Record","attributes":{"f":{"type":"String","x":1}}}}},"actions":{}}}`},
		{"unknownCommonTypeMember", `{"":{"entityTypes":{},"actions":{},"commonTypes":{"T":{"type":"String","required":false}}}}`},
		{"annotationsOnSetElement", `{"":{"entityTypes":{"A":{"tags":{"type":"Set","element":{"type":"String","annotations":{}}}}},"actions":{}}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var s schemajson.Schema
			testutil.Error(t, s.UnmarshalJSON([]byte(tt.input)))
		})
	}
}

func TestUnmarshalMiscasedMember(t *testing.T) {
	t.Parallel()
	// Member names are matched case-sensitively, so a miscased member is an
	// unknown member, which is rejected.
	var s schemajson.Schema
	testutil.Error(t, s.UnmarshalJSON([]byte(`{"":{"entityTypes":{},"EntityTypes":{"A":{}},"actions":{}}}`)))
}

func TestRecordAttributes(t *testing.T) {
	t.Parallel()
	t.Run("MarshalEmpty", func(t *testing.T) {
		t.Parallel()
		// Rust Cedar requires "attributes" in a Record type, so it's written
		// even when there are none.
		s := ast.Schema{
			Entities: ast.Entities{"A": ast.Entity{
				Shape: ast.RecordType{},
				Tags:  ast.RecordType{},
			}},
		}
		b, err := (*schemajson.Schema)(&s).MarshalJSON()
		testutil.OK(t, err)
		testutil.Equals(t, string(b), `{"":{"entityTypes":{"A":{"shape":{"type":"Record","attributes":{}},"tags":{"type":"Record","attributes":{}}}},"actions":{}}}`)
	})

	t.Run("UnmarshalMissing", func(t *testing.T) {
		t.Parallel()
		// cedar-go previously omitted "attributes" when there were none, so a
		// Record without it still decodes, unlike in Rust Cedar.
		var s schemajson.Schema
		testutil.OK(t, s.UnmarshalJSON([]byte(`{"":{"entityTypes":{"A":{"shape":{"type":"Record"},"tags":{"type":"Record"}}},"actions":{}}}`)))
		testutil.Equals(t, (*ast.Schema)(&s).Entities["A"], ast.Entity{Shape: ast.RecordType{}, Tags: ast.RecordType{}})
	})
}
