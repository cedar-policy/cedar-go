package resolved_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	ast2 "github.com/cedar-policy/cedar-go/x/exp/schema/ast"
	resolved2 "github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

func TestResolveEmpty(t *testing.T) {
	s := &ast2.Schema{}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, len(result.Entities), 0)
	testutil.Equals(t, len(result.Actions), 0)
}

func TestResolveBasicEntity(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"name": ast2.Attribute{Type: ast2.TypeRef("String")},
					"age":  ast2.Attribute{Type: ast2.TypeRef("Long"), Optional: true},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	user := result.Entities["User"]
	testutil.Equals(t, user.Shape["name"].Type, resolved2.IsType(resolved2.StringType{}))
	testutil.Equals(t, user.Shape["age"].Type, resolved2.IsType(resolved2.LongType{}))
	testutil.Equals(t, user.Shape["age"].Optional, true)
}

func TestResolveEntityMemberOf(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User":  ast2.Entity{ParentTypes: []ast2.EntityTypeRef{"Group"}},
			"Group": ast2.Entity{},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["User"].ParentTypes, []types.EntityType{"Group"})
}

func TestResolveEntityMemberOfUndefined(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{ParentTypes: []ast2.EntityTypeRef{"NonExistent"}},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveCommonType(t *testing.T) {
	s := &ast2.Schema{
		CommonTypes: ast2.CommonTypes{
			"Context": ast2.CommonType{
				Type: ast2.RecordType{
					"ip": ast2.Attribute{Type: ast2.TypeRef("ipaddr")},
				},
			},
		},
		Actions: ast2.Actions{
			"view": ast2.Action{
				AppliesTo: &ast2.AppliesTo{
					Context: ast2.TypeRef("Context"),
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("Action", "view")
	view := result.Actions[uid]
	testutil.Equals(t, view.AppliesTo != nil, true)
	_, ok := view.AppliesTo.Context["ip"]
	testutil.Equals(t, ok, true)
}

func TestResolveCommonTypeCycle(t *testing.T) {
	s := &ast2.Schema{
		CommonTypes: ast2.CommonTypes{
			"A": ast2.CommonType{Type: ast2.TypeRef("B")},
			"B": ast2.CommonType{Type: ast2.TypeRef("A")},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveCommonTypeIndirectCycle(t *testing.T) {
	s := &ast2.Schema{
		CommonTypes: ast2.CommonTypes{
			"A": ast2.CommonType{Type: ast2.TypeRef("B")},
			"B": ast2.CommonType{Type: ast2.TypeRef("C")},
			"C": ast2.CommonType{Type: ast2.TypeRef("A")},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveUndefinedType(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"x": ast2.Attribute{Type: ast2.TypeRef("NonExistent")},
				},
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveBuiltinTypes(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"s":   ast2.Attribute{Type: ast2.TypeRef("String")},
					"l":   ast2.Attribute{Type: ast2.TypeRef("Long")},
					"b":   ast2.Attribute{Type: ast2.TypeRef("Bool")},
					"b2":  ast2.Attribute{Type: ast2.TypeRef("Boolean")},
					"ip":  ast2.Attribute{Type: ast2.TypeRef("ipaddr")},
					"dec": ast2.Attribute{Type: ast2.TypeRef("decimal")},
					"dt":  ast2.Attribute{Type: ast2.TypeRef("datetime")},
					"dur": ast2.Attribute{Type: ast2.TypeRef("duration")},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	user := result.Entities["User"]
	testutil.Equals(t, user.Shape["s"].Type, resolved2.IsType(resolved2.StringType{}))
	testutil.Equals(t, user.Shape["l"].Type, resolved2.IsType(resolved2.LongType{}))
	testutil.Equals(t, user.Shape["b"].Type, resolved2.IsType(resolved2.BoolType{}))
	testutil.Equals(t, user.Shape["b2"].Type, resolved2.IsType(resolved2.BoolType{}))
	testutil.Equals(t, user.Shape["ip"].Type, resolved2.IsType(resolved2.ExtensionType("ipaddr")))
	testutil.Equals(t, user.Shape["dec"].Type, resolved2.IsType(resolved2.ExtensionType("decimal")))
	testutil.Equals(t, user.Shape["dt"].Type, resolved2.IsType(resolved2.ExtensionType("datetime")))
	testutil.Equals(t, user.Shape["dur"].Type, resolved2.IsType(resolved2.ExtensionType("duration")))
}

func TestResolveCedarNamespace(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"name": ast2.Attribute{Type: ast2.TypeRef("__cedar::String")},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["User"].Shape["name"].Type, resolved2.IsType(resolved2.StringType{}))
}

func TestResolveCedarNamespaceUndefined(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"x": ast2.Attribute{Type: ast2.TypeRef("__cedar::Bogus")},
				},
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveTypeDisambiguation(t *testing.T) {
	t.Parallel()

	t.Run("common_over_entity", func(t *testing.T) {
		t.Parallel()
		// Example from the Cedar spec: https://docs.cedarpolicy.com/schema/human-readable-schema.html#schema-typeDisambiguation
		// When "name" is declared as both a common type and an entity type in the same namespace,
		// the common type wins.
		s := &ast2.Schema{
			Namespaces: ast2.Namespaces{
				"NS": ast2.Namespace{
					CommonTypes: ast2.CommonTypes{
						"name": ast2.CommonType{
							Type: ast2.RecordType{
								"first": ast2.Attribute{Type: ast2.TypeRef("String")},
								"last":  ast2.Attribute{Type: ast2.TypeRef("String")},
							},
						},
					},
					Entities: ast2.Entities{
						"name": ast2.Entity{},
						"User": ast2.Entity{
							Shape: &ast2.RecordType{
								"n": ast2.Attribute{Type: ast2.TypeRef("name")},
							},
						},
					},
				},
			},
		}
		result, err := resolved2.Resolve(s)
		testutil.OK(t, err)
		user := result.Entities["NS::User"]
		// "name" should resolve to the common type (a record), not the entity type
		rec, ok := user.Shape["n"].Type.(resolved2.RecordType)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(rec), 2)
	})

	t.Run("entity_over_builtin", func(t *testing.T) {
		t.Parallel()
		// An entity type named "Long" shadows the built-in Long primitive.
		// A reference to "Long" should resolve to the entity type, not LongType.
		// The built-in is still accessible via __cedar::Long.
		s := &ast2.Schema{
			Namespaces: ast2.Namespaces{
				"NS": ast2.Namespace{
					Entities: ast2.Entities{
						"Long": ast2.Entity{},
						"User": ast2.Entity{
							Shape: &ast2.RecordType{
								"x": ast2.Attribute{Type: ast2.TypeRef("Long")},
								"y": ast2.Attribute{Type: ast2.TypeRef("__cedar::Long")},
							},
						},
					},
				},
			},
		}
		result, err := resolved2.Resolve(s)
		testutil.OK(t, err)
		user := result.Entities["NS::User"]
		// "Long" resolves to entity type NS::Long, not the built-in
		testutil.Equals(t, user.Shape["x"].Type, resolved2.IsType(resolved2.EntityType("NS::Long")))
		// "__cedar::Long" still resolves to the built-in
		testutil.Equals(t, user.Shape["y"].Type, resolved2.IsType(resolved2.LongType{}))
	})

	t.Run("common_over_builtin", func(t *testing.T) {
		t.Parallel()
		// A common type named "Long" shadows the built-in Long primitive.
		// A reference to "Long" should resolve to the common type, not LongType.
		s := &ast2.Schema{
			Namespaces: ast2.Namespaces{
				"NS": ast2.Namespace{
					CommonTypes: ast2.CommonTypes{
						"Long": ast2.CommonType{
							Type: ast2.RecordType{
								"value": ast2.Attribute{Type: ast2.TypeRef("__cedar::Long")},
							},
						},
					},
					Entities: ast2.Entities{
						"User": ast2.Entity{
							Shape: &ast2.RecordType{
								"x": ast2.Attribute{Type: ast2.TypeRef("Long")},
								"y": ast2.Attribute{Type: ast2.TypeRef("__cedar::Long")},
							},
						},
					},
				},
			},
		}
		result, err := resolved2.Resolve(s)
		testutil.OK(t, err)
		user := result.Entities["NS::User"]
		// "Long" resolves to the common type (a record), not the built-in
		rec, ok := user.Shape["x"].Type.(resolved2.RecordType)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(rec), 1)
		// "__cedar::Long" still resolves to the built-in
		testutil.Equals(t, user.Shape["y"].Type, resolved2.IsType(resolved2.LongType{}))
	})
}

func TestResolveNamespaceEntityRef(t *testing.T) {
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Entities: ast2.Entities{
					"User":  ast2.Entity{ParentTypes: []ast2.EntityTypeRef{"Group"}},
					"Group": ast2.Entity{},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["NS::User"].ParentTypes, []types.EntityType{"NS::Group"})
}

func TestResolveCrossNamespaceEntityRef(t *testing.T) {
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"A": ast2.Namespace{
				Entities: ast2.Entities{
					"User": ast2.Entity{ParentTypes: []ast2.EntityTypeRef{"B::Group"}},
				},
			},
			"B": ast2.Namespace{
				Entities: ast2.Entities{
					"Group": ast2.Entity{},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["A::User"].ParentTypes, []types.EntityType{"B::Group"})
}

func TestResolveAction(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User":  ast2.Entity{},
			"Photo": ast2.Entity{},
		},
		Actions: ast2.Actions{
			"view": ast2.Action{
				AppliesTo: &ast2.AppliesTo{
					Principals: []ast2.EntityTypeRef{"User"},
					Resources:  []ast2.EntityTypeRef{"Photo"},
					Context:    ast2.RecordType{},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("Action", "view")
	view := result.Actions[uid]
	testutil.Equals(t, view.AppliesTo.Principals, []types.EntityType{"User"})
	testutil.Equals(t, view.AppliesTo.Resources, []types.EntityType{"Photo"})
}

func TestResolveActionMemberOf(t *testing.T) {
	s := &ast2.Schema{
		Actions: ast2.Actions{
			"view":     ast2.Action{Parents: []ast2.ParentRef{ast2.ParentRefFromID("readOnly")}},
			"readOnly": ast2.Action{},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("Action", "view")
	view := result.Actions[uid]
	testutil.Equals(t, view.Parents, []types.EntityUID{types.NewEntityUID("Action", "readOnly")})
}

func TestResolveActionCycle(t *testing.T) {
	s := &ast2.Schema{
		Actions: ast2.Actions{
			"a": ast2.Action{Parents: []ast2.ParentRef{ast2.ParentRefFromID("b")}},
			"b": ast2.Action{Parents: []ast2.ParentRef{ast2.ParentRefFromID("a")}},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveActionUndefinedParent(t *testing.T) {
	s := &ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{Parents: []ast2.ParentRef{ast2.ParentRefFromID("nonExistent")}},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveEnum(t *testing.T) {
	s := &ast2.Schema{
		Enums: ast2.Enums{
			"Status": ast2.Enum{
				Values: []types.String{"active", "inactive"},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	status := result.Enums["Status"]
	testutil.Equals(t, status.Values, []types.String{"active", "inactive"})

	// Test EntityUIDs iterator
	count := 0
	for uid := range status.EntityUIDs() {
		testutil.Equals(t, uid.Type, types.EntityType("Status"))
		count++
	}
	testutil.Equals(t, count, 2)
}

func TestResolveEnumAsEntityType(t *testing.T) {
	s := &ast2.Schema{
		Enums: ast2.Enums{
			"Status": ast2.Enum{Values: []types.String{"active"}},
		},
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"s": ast2.Attribute{Type: ast2.TypeRef("Status")},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["User"].Shape["s"].Type, resolved2.IsType(resolved2.EntityType("Status")))
}

func TestResolveSetType(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"tags": ast2.Attribute{Type: ast2.Set(ast2.TypeRef("String"))},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	tags := result.Entities["User"].Shape["tags"]
	set, ok := tags.Type.(resolved2.SetType)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, set.Element, resolved2.IsType(resolved2.StringType{}))
}

func TestResolveEntityWithTags(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Tags: ast2.TypeRef("String"),
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["User"].Tags, resolved2.IsType(resolved2.StringType{}))
}

func TestResolveNamespacedAction(t *testing.T) {
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Actions: ast2.Actions{
					"view": ast2.Action{},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("NS::Action", "view")
	_, ok := result.Actions[uid]
	testutil.Equals(t, ok, true)
}

func TestResolveActionQualifiedParent(t *testing.T) {
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Actions: ast2.Actions{
					"view": ast2.Action{
						Parents: []ast2.ParentRef{
							ast2.NewParentRef("NS::Action", "readOnly"),
						},
					},
					"readOnly": ast2.Action{},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("NS::Action", "view")
	view := result.Actions[uid]
	testutil.Equals(t, view.Parents, []types.EntityUID{types.NewEntityUID("NS::Action", "readOnly")})
}

func TestResolveActionContextNull(t *testing.T) {
	s := &ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{
				AppliesTo: &ast2.AppliesTo{},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("Action", "view")
	view := result.Actions[uid]
	testutil.Equals(t, len(view.AppliesTo.Context), 0)
}

func TestResolveActionContextNonRecord(t *testing.T) {
	s := &ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{
				AppliesTo: &ast2.AppliesTo{
					Context: ast2.TypeRef("String"),
				},
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveActionPrincipalUndefined(t *testing.T) {
	s := &ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{
				AppliesTo: &ast2.AppliesTo{
					Principals: []ast2.EntityTypeRef{"NonExistent"},
				},
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveActionResourceUndefined(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{"User": ast2.Entity{}},
		Actions: ast2.Actions{
			"view": ast2.Action{
				AppliesTo: &ast2.AppliesTo{
					Principals: []ast2.EntityTypeRef{"User"},
					Resources:  []ast2.EntityTypeRef{"NonExistent"},
				},
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveCommonTypeChain(t *testing.T) {
	s := &ast2.Schema{
		CommonTypes: ast2.CommonTypes{
			"A": ast2.CommonType{Type: ast2.TypeRef("B")},
			"B": ast2.CommonType{Type: ast2.RecordType{
				"x": ast2.Attribute{Type: ast2.TypeRef("Long")},
			}},
		},
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"a": ast2.Attribute{Type: ast2.TypeRef("A")},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	a := result.Entities["User"].Shape["a"]
	rec, ok := a.Type.(resolved2.RecordType)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, len(rec), 1)
}

func TestResolveQualifiedCommonType(t *testing.T) {
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				CommonTypes: ast2.CommonTypes{
					"Ctx": ast2.CommonType{
						Type: ast2.RecordType{},
					},
				},
				Entities: ast2.Entities{
					"User": ast2.Entity{
						Shape: &ast2.RecordType{
							"c": ast2.Attribute{Type: ast2.TypeRef("NS::Ctx")},
						},
					},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	c := result.Entities["NS::User"].Shape["c"]
	_, ok := c.Type.(resolved2.RecordType)
	testutil.Equals(t, ok, true)
}

func TestResolveQualifiedUndefined(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"x": ast2.Attribute{Type: ast2.TypeRef("NS::NonExistent")},
				},
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveEntityTagsUndefined(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Tags: ast2.TypeRef("NonExistent"),
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveEntityShapeAttrUndefined(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"x": ast2.Attribute{Type: ast2.Set(ast2.TypeRef("NonExistent"))},
				},
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveNamespaceOutput(t *testing.T) {
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Annotations: ast2.Annotations{"doc": "test"},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	ns := result.Namespaces["NS"]
	testutil.Equals(t, ns.Name, types.Path("NS"))
	testutil.Equals(t, types.String(ns.Annotations["doc"]), types.String("test"))
}

func TestResolveEntityTypeRef(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"friend": ast2.Attribute{Type: ast2.EntityTypeRef("User")},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	friend := result.Entities["User"].Shape["friend"]
	testutil.Equals(t, friend.Type, resolved2.IsType(resolved2.EntityType("User")))
}

func TestResolveEntityTypeRefUndefined(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"x": ast2.Attribute{Type: ast2.EntityTypeRef("NonExistent")},
				},
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveEntityTypeRefQualified(t *testing.T) {
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"A": ast2.Namespace{
				Entities: ast2.Entities{
					"Foo": ast2.Entity{},
				},
			},
		},
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"x": ast2.Attribute{Type: ast2.EntityTypeRef("A::Foo")},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	x := result.Entities["User"].Shape["x"]
	testutil.Equals(t, x.Type, resolved2.IsType(resolved2.EntityType("A::Foo")))
}

func TestResolveEntityTypeRefQualifiedUndefined(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"x": ast2.Attribute{Type: ast2.EntityTypeRef("A::NonExistent")},
				},
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveDirectPrimitiveTypes(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"s": ast2.Attribute{Type: ast2.StringType{}},
					"l": ast2.Attribute{Type: ast2.LongType{}},
					"b": ast2.Attribute{Type: ast2.BoolType{}},
					"e": ast2.Attribute{Type: ast2.ExtensionType("ipaddr")},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	user := result.Entities["User"]
	testutil.Equals(t, user.Shape["s"].Type, resolved2.IsType(resolved2.StringType{}))
	testutil.Equals(t, user.Shape["l"].Type, resolved2.IsType(resolved2.LongType{}))
	testutil.Equals(t, user.Shape["b"].Type, resolved2.IsType(resolved2.BoolType{}))
	testutil.Equals(t, user.Shape["e"].Type, resolved2.IsType(resolved2.ExtensionType("ipaddr")))
}

func TestResolveEnumEntityUIDBrokenIterator(t *testing.T) {
	s := &ast2.Schema{
		Enums: ast2.Enums{
			"Status": ast2.Enum{
				Values: []types.String{"a", "b", "c"},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	status := result.Enums["Status"]
	// Break after first iteration to cover early return
	count := 0
	for range status.EntityUIDs() {
		count++
		break
	}
	testutil.Equals(t, count, 1)
}

func TestResolveNamespacedEntitiesError(t *testing.T) {
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Entities: ast2.Entities{
					"User": ast2.Entity{
						Shape: &ast2.RecordType{
							"x": ast2.Attribute{Type: ast2.TypeRef("NonExistent")},
						},
					},
				},
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveNamespacedEnumError(t *testing.T) {
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Enums: ast2.Enums{
					"Status": ast2.Enum{Values: []types.String{"a"}},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, len(result.Enums), 1)
}

func TestResolveNamespacedActionsError(t *testing.T) {
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Entities: ast2.Entities{
					"User": ast2.Entity{},
				},
				Actions: ast2.Actions{
					"view": ast2.Action{
						AppliesTo: &ast2.AppliesTo{
							Context: ast2.TypeRef("NonExistent"),
						},
					},
				},
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveEntityMemberOfValidationError(t *testing.T) {
	// Entity referencing an undefined parent type errors during resolution.
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				ParentTypes: []ast2.EntityTypeRef{"NonExistent"},
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveQualifiedEntityType(t *testing.T) {
	// Test resolving a qualified entity type ref
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Entities: ast2.Entities{
					"NS::User": ast2.Entity{},
					"Admin":    ast2.Entity{ParentTypes: []ast2.EntityTypeRef{"NS::User"}},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["NS::Admin"].ParentTypes, []types.EntityType{"NS::User"})
}

func TestResolveQualifiedEntityTypeRefAsType(t *testing.T) {
	// Test that a qualified entity type resolves when used through TypeRef
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Entities: ast2.Entities{
					"User": ast2.Entity{
						Shape: &ast2.RecordType{
							"ref": ast2.Attribute{Type: ast2.TypeRef("NS::Admin")},
						},
					},
					"Admin": ast2.Entity{},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	ref := result.Entities["NS::User"].Shape["ref"]
	testutil.Equals(t, ref.Type, resolved2.IsType(resolved2.EntityType("NS::Admin")))
}

func TestResolveEmptyNamespaceCommonType(t *testing.T) {
	// Test that empty namespace common types are found from a different namespace
	s := &ast2.Schema{
		CommonTypes: ast2.CommonTypes{
			"Ctx": ast2.CommonType{Type: ast2.RecordType{}},
		},
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Entities: ast2.Entities{
					"User": ast2.Entity{
						Shape: &ast2.RecordType{
							"c": ast2.Attribute{Type: ast2.TypeRef("Ctx")},
						},
					},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	c := result.Entities["NS::User"].Shape["c"]
	_, ok := c.Type.(resolved2.RecordType)
	testutil.Equals(t, ok, true)
}

func TestResolveEmptyNamespaceEntityType(t *testing.T) {
	// Test that empty namespace entity types are found from a different namespace
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"Global": ast2.Entity{},
		},
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Entities: ast2.Entities{
					"User": ast2.Entity{
						Shape: &ast2.RecordType{
							"g": ast2.Attribute{Type: ast2.TypeRef("Global")},
						},
					},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	g := result.Entities["NS::User"].Shape["g"]
	testutil.Equals(t, g.Type, resolved2.IsType(resolved2.EntityType("Global")))
}

func TestResolveAnnotationsOnAttributes(t *testing.T) {
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Annotations: ast2.Annotations{"doc": "user"},
				Shape: &ast2.RecordType{
					"name": ast2.Attribute{
						Type:        ast2.TypeRef("String"),
						Annotations: ast2.Annotations{"doc": "the name"},
					},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	user := result.Entities["User"]
	testutil.Equals(t, types.String(user.Annotations["doc"]), types.String("user"))
	testutil.Equals(t, types.String(user.Shape["name"].Annotations["doc"]), types.String("the name"))
}

func TestResolveEnumAnnotations(t *testing.T) {
	s := &ast2.Schema{
		Enums: ast2.Enums{
			"Status": ast2.Enum{
				Annotations: ast2.Annotations{"doc": "status"},
				Values:      []types.String{"a"},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, types.String(result.Enums["Status"].Annotations["doc"]), types.String("status"))
}

func TestResolveActionAnnotations(t *testing.T) {
	s := &ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{
				Annotations: ast2.Annotations{"doc": "view action"},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("Action", "view")
	testutil.Equals(t, types.String(result.Actions[uid].Annotations["doc"]), types.String("view action"))
}

func TestResolveBareEnumsError(t *testing.T) {
	// Test that bare enums with no errors pass through resolveEnums
	s := &ast2.Schema{
		Enums: ast2.Enums{
			"A": ast2.Enum{Values: []types.String{"x"}},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Enums["A"].Name, types.EntityType("A"))
}

func TestResolveBareActionsError(t *testing.T) {
	// Test bare actions with an error in context resolution
	s := &ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{
				AppliesTo: &ast2.AppliesTo{
					Context: ast2.TypeRef("NonExistent"),
				},
			},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveCommonTypeCycleInSet(t *testing.T) {
	s := &ast2.Schema{
		CommonTypes: ast2.CommonTypes{
			"A": ast2.CommonType{Type: ast2.Set(ast2.TypeRef("B"))},
			"B": ast2.CommonType{Type: ast2.TypeRef("A")},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveCommonTypeCycleInRecord(t *testing.T) {
	s := &ast2.Schema{
		CommonTypes: ast2.CommonTypes{
			"A": ast2.CommonType{Type: ast2.RecordType{
				"x": ast2.Attribute{Type: ast2.TypeRef("B")},
			}},
			"B": ast2.CommonType{Type: ast2.TypeRef("A")},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveNamespacedEntityTypeRef(t *testing.T) {
	// Exercise resolveTypeRef line 421-423: unqualified name resolves to
	// a namespaced entity type (not common type) via disambiguation rule 2.
	// Use a Set<User> attribute where "User" should resolve to NS::User entity type.
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Entities: ast2.Entities{
					"User": ast2.Entity{},
					"Group": ast2.Entity{
						Shape: &ast2.RecordType{
							"members": ast2.Attribute{Type: ast2.SetType{Element: ast2.TypeRef("User")}},
						},
					},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	members := result.Entities["NS::Group"].Shape["members"]
	setType, ok := members.Type.(resolved2.SetType)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, setType.Element, resolved2.IsType(resolved2.EntityType("NS::User")))
}

func TestResolveNamespacedEnumTypeRef(t *testing.T) {
	// Exercise resolveTypeRef line 421-423: unqualified name resolves to
	// a namespaced enum type via disambiguation rule 2.
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Enums: ast2.Enums{
					"Color": ast2.Enum{Values: []types.String{"red", "blue"}},
				},
				Entities: ast2.Entities{
					"Item": ast2.Entity{
						Shape: &ast2.RecordType{
							"color": ast2.Attribute{Type: ast2.TypeRef("Color")},
						},
					},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	color := result.Entities["NS::Item"].Shape["color"]
	testutil.Equals(t, color.Type, resolved2.IsType(resolved2.EntityType("NS::Color")))
}

func TestResolveUndefinedMemberOf(t *testing.T) {
	// Resolve returns an error when an entity references an undefined parent type.
	s := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{ParentTypes: []ast2.EntityTypeRef{"Nonexistent"}},
		},
	}
	_, err := resolved2.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveCedarBuiltinInTypePath(t *testing.T) {
	// Exercise resolveTypeRefPath line 462-464: __cedar:: prefix in cycle detection.
	s := &ast2.Schema{
		CommonTypes: ast2.CommonTypes{
			"A": ast2.CommonType{Type: ast2.TypeRef("__cedar::String")},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	_ = result
}

func TestResolveQualifiedTypePath(t *testing.T) {
	// Exercise resolveTypeRefPath line 465-467: qualified path with :: in cycle detection.
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				CommonTypes: ast2.CommonTypes{
					"A": ast2.CommonType{Type: ast2.TypeRef("NS::B")},
					"B": ast2.CommonType{Type: ast2.StringType{}},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	_ = result
}

func TestResolveNamespacedCommonTypePath(t *testing.T) {
	// Exercise resolveTypeRefPath line 470-472: namespaced common type ref in cycle detection.
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				CommonTypes: ast2.CommonTypes{
					"A": ast2.CommonType{Type: ast2.TypeRef("B")},
					"B": ast2.CommonType{Type: ast2.StringType{}},
				},
			},
		},
	}
	result, err := resolved2.Resolve(s)
	testutil.OK(t, err)
	_ = result
}
