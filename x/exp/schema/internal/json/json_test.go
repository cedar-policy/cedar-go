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
