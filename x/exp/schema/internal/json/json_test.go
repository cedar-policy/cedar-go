package json_test

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	ast2 "github.com/cedar-policy/cedar-go/x/exp/schema/ast"
	schemajson "github.com/cedar-policy/cedar-go/x/exp/schema/internal/json"
)

func TestRoundTripEmpty(t *testing.T) {
	s := ast2.Schema{}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))
	testutil.Equals(t, (*ast2.Schema)(&s2), &ast2.Schema{})
}

func TestRoundTripEntity(t *testing.T) {
	s := ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Entities: ast2.Entities{
					"NS::User": ast2.Entity{
						ParentTypes: []ast2.EntityTypeRef{"Group"},
						Shape: &ast2.RecordType{
							"name": ast2.Attribute{Type: ast2.StringType{}},
							"age":  ast2.Attribute{Type: ast2.LongType{}, Optional: true},
						},
						Tags: ast2.StringType{},
					},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	user := (*ast2.Schema)(&s2).Namespaces["NS"].Entities["NS::User"]
	testutil.Equals(t, user.ParentTypes, []ast2.EntityTypeRef{"Group"})
	testutil.Equals(t, user.Shape != nil, true)
	testutil.Equals(t, user.Tags != nil, true)
}

func TestRoundTripEnum(t *testing.T) {
	s := ast2.Schema{
		Enums: ast2.Enums{
			"Status": ast2.Enum{
				Values: []types.String{"active", "inactive"},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	status := (*ast2.Schema)(&s2).Enums["Status"]
	testutil.Equals(t, status.Values, []types.String{"active", "inactive"})
}

func TestRoundTripAction(t *testing.T) {
	s := ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{
				Parents: []ast2.ParentRef{
					ast2.NewParentRef("NS::Action", "readOnly"),
					ast2.ParentRefFromID("write"),
				},
				AppliesTo: &ast2.AppliesTo{
					Principals: []ast2.EntityTypeRef{"User"},
					Resources:  []ast2.EntityTypeRef{"Photo"},
					Context:    ast2.RecordType{},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	view := (*ast2.Schema)(&s2).Actions["view"]
	testutil.Equals(t, len(view.Parents), 2)
	testutil.Equals(t, view.AppliesTo != nil, true)
	testutil.Equals(t, len(view.AppliesTo.Principals), 1)
}

func TestRoundTripCommonType(t *testing.T) {
	s := ast2.Schema{
		CommonTypes: ast2.CommonTypes{
			"Context": ast2.CommonType{
				Annotations: ast2.Annotations{"doc": "context type"},
				Type: ast2.RecordType{
					"ip": ast2.Attribute{Type: ast2.ExtensionType("ipaddr")},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	ct := (*ast2.Schema)(&s2).CommonTypes["Context"]
	testutil.Equals(t, ct.Annotations["doc"], types.String("context type"))
}

func TestRoundTripAllTypes(t *testing.T) {
	s := ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"s":    ast2.Attribute{Type: ast2.StringType{}},
					"l":    ast2.Attribute{Type: ast2.LongType{}},
					"b":    ast2.Attribute{Type: ast2.BoolType{}},
					"ip":   ast2.Attribute{Type: ast2.ExtensionType("ipaddr")},
					"set":  ast2.Attribute{Type: ast2.Set(ast2.LongType{})},
					"rec":  ast2.Attribute{Type: ast2.RecordType{}},
					"ref":  ast2.Attribute{Type: ast2.EntityTypeRef("Other")},
					"tref": ast2.Attribute{Type: ast2.TypeRef("CommonT")},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))
	testutil.Equals(t, len(*(*ast2.Schema)(&s2).Entities["User"].Shape), 8)
}

func TestRoundTripAnnotations(t *testing.T) {
	s := ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Annotations: ast2.Annotations{"doc": "user entity"},
				Shape: &ast2.RecordType{
					"name": ast2.Attribute{
						Type:        ast2.StringType{},
						Annotations: ast2.Annotations{"doc": "user name"},
					},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	user := (*ast2.Schema)(&s2).Entities["User"]
	testutil.Equals(t, user.Annotations["doc"], types.String("user entity"))
	testutil.Equals(t, (*user.Shape)["name"].Annotations["doc"], types.String("user name"))
}

func TestRoundTripNamespaceAnnotations(t *testing.T) {
	s := ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Annotations: ast2.Annotations{"doc": "my ns"},
				Entities:    ast2.Entities{},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	ns := (*ast2.Schema)(&s2).Namespaces["NS"]
	testutil.Equals(t, ns.Annotations["doc"], types.String("my ns"))
}

func TestRoundTripActionNoAppliesTo(t *testing.T) {
	s := ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	view := (*ast2.Schema)(&s2).Actions["view"]
	testutil.Equals(t, view.AppliesTo == nil, true)
}

func TestRoundTripActionEmptyLists(t *testing.T) {
	s := ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{
				AppliesTo: &ast2.AppliesTo{
					Principals: []ast2.EntityTypeRef{},
					Resources:  []ast2.EntityTypeRef{},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	view := (*ast2.Schema)(&s2).Actions["view"]
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

func TestUnmarshalBadType(t *testing.T) {
	var s schemajson.Schema
	testutil.Error(t, s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"Foo": {"tags": {"type": "Unknown"}}}, "actions": {}}}`)))
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

	ns := (*ast2.Schema)(&s).Namespaces["NS"]
	testutil.Equals(t, len(ns.Entities), 2)
	testutil.Equals(t, len(ns.Actions), 1)
	testutil.Equals(t, len(ns.CommonTypes), 1)

	b, err := json.Marshal(&s)
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, json.Unmarshal(b, &s2))
	testutil.Equals(t, len((*ast2.Schema)(&s2).Namespaces["NS"].Entities), 2)
}

func TestRoundTripEnumAnnotations(t *testing.T) {
	s := ast2.Schema{
		Enums: ast2.Enums{
			"Status": ast2.Enum{
				Annotations: ast2.Annotations{"doc": "status enum"},
				Values:      []types.String{"active"},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	status := (*ast2.Schema)(&s2).Enums["Status"]
	testutil.Equals(t, status.Annotations["doc"], types.String("status enum"))
}

func TestRoundTripActionContext(t *testing.T) {
	s := ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{
				AppliesTo: &ast2.AppliesTo{
					Principals: []ast2.EntityTypeRef{"User"},
					Resources:  []ast2.EntityTypeRef{"Photo"},
					Context: ast2.RecordType{
						"ip": ast2.Attribute{Type: ast2.ExtensionType("ipaddr")},
					},
				},
			},
		},
	}
	b, err := (*schemajson.Schema)(&s).MarshalJSON()
	testutil.OK(t, err)

	var s2 schemajson.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))

	view := (*ast2.Schema)(&s2).Actions["view"]
	testutil.Equals(t, view.AppliesTo.Context != nil, true)
}
