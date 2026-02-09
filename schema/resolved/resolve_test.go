package resolved_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/schema/ast"
	"github.com/cedar-policy/cedar-go/schema/resolved"
	"github.com/cedar-policy/cedar-go/types"
)

func TestResolveEmpty(t *testing.T) {
	s := &ast.Schema{}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, len(result.Entities), 0)
	testutil.Equals(t, len(result.Actions), 0)
}

func TestResolveBasicEntity(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"name": ast.Attribute{Type: ast.TypeRef("String")},
					"age":  ast.Attribute{Type: ast.TypeRef("Long"), Optional: true},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	user := result.Entities["User"]
	testutil.Equals(t, user.Shape["name"].Type, resolved.IsType(resolved.StringType{}))
	testutil.Equals(t, user.Shape["age"].Type, resolved.IsType(resolved.LongType{}))
	testutil.Equals(t, user.Shape["age"].Optional, true)
}

func TestResolveEntityMemberOf(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User":  ast.Entity{MemberOf: []ast.EntityTypeRef{"Group"}},
			"Group": ast.Entity{},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["User"].MemberOf, []types.EntityType{"Group"})
}

func TestResolveEntityMemberOfUndefined(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{MemberOf: []ast.EntityTypeRef{"NonExistent"}},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveCommonType(t *testing.T) {
	s := &ast.Schema{
		CommonTypes: ast.CommonTypes{
			"Context": ast.CommonType{
				Type: ast.RecordType{
					"ip": ast.Attribute{Type: ast.TypeRef("ipaddr")},
				},
			},
		},
		Actions: ast.Actions{
			"view": ast.Action{
				AppliesTo: &ast.AppliesTo{
					Context: ast.TypeRef("Context"),
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("Action", "view")
	view := result.Actions[uid]
	testutil.Equals(t, view.AppliesTo != nil, true)
	_, ok := view.AppliesTo.Context["ip"]
	testutil.Equals(t, ok, true)
}

func TestResolveCommonTypeCycle(t *testing.T) {
	s := &ast.Schema{
		CommonTypes: ast.CommonTypes{
			"A": ast.CommonType{Type: ast.TypeRef("B")},
			"B": ast.CommonType{Type: ast.TypeRef("A")},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveCommonTypeIndirectCycle(t *testing.T) {
	s := &ast.Schema{
		CommonTypes: ast.CommonTypes{
			"A": ast.CommonType{Type: ast.TypeRef("B")},
			"B": ast.CommonType{Type: ast.TypeRef("C")},
			"C": ast.CommonType{Type: ast.TypeRef("A")},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveUndefinedType(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"x": ast.Attribute{Type: ast.TypeRef("NonExistent")},
				},
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveBuiltinTypes(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"s":   ast.Attribute{Type: ast.TypeRef("String")},
					"l":   ast.Attribute{Type: ast.TypeRef("Long")},
					"b":   ast.Attribute{Type: ast.TypeRef("Bool")},
					"b2":  ast.Attribute{Type: ast.TypeRef("Boolean")},
					"ip":  ast.Attribute{Type: ast.TypeRef("ipaddr")},
					"dec": ast.Attribute{Type: ast.TypeRef("decimal")},
					"dt":  ast.Attribute{Type: ast.TypeRef("datetime")},
					"dur": ast.Attribute{Type: ast.TypeRef("duration")},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	user := result.Entities["User"]
	testutil.Equals(t, user.Shape["s"].Type, resolved.IsType(resolved.StringType{}))
	testutil.Equals(t, user.Shape["l"].Type, resolved.IsType(resolved.LongType{}))
	testutil.Equals(t, user.Shape["b"].Type, resolved.IsType(resolved.BoolType{}))
	testutil.Equals(t, user.Shape["b2"].Type, resolved.IsType(resolved.BoolType{}))
	testutil.Equals(t, user.Shape["ip"].Type, resolved.IsType(resolved.ExtensionType("ipaddr")))
	testutil.Equals(t, user.Shape["dec"].Type, resolved.IsType(resolved.ExtensionType("decimal")))
	testutil.Equals(t, user.Shape["dt"].Type, resolved.IsType(resolved.ExtensionType("datetime")))
	testutil.Equals(t, user.Shape["dur"].Type, resolved.IsType(resolved.ExtensionType("duration")))
}

func TestResolveCedarNamespace(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"name": ast.Attribute{Type: ast.TypeRef("__cedar::String")},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["User"].Shape["name"].Type, resolved.IsType(resolved.StringType{}))
}

func TestResolveCedarNamespaceUndefined(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"x": ast.Attribute{Type: ast.TypeRef("__cedar::Bogus")},
				},
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveTypeDisambiguation(t *testing.T) {
	t.Parallel()

	t.Run("common_over_entity", func(t *testing.T) {
		t.Parallel()
		// Example from the Cedar spec: https://docs.cedarpolicy.com/schema/human-readable-schema.html#schema-typeDisambiguation
		// When "name" is declared as both a common type and an entity type in the same namespace,
		// the common type wins.
		s := &ast.Schema{
			Namespaces: ast.Namespaces{
				"NS": ast.Namespace{
					CommonTypes: ast.CommonTypes{
						"name": ast.CommonType{
							Type: ast.RecordType{
								"first": ast.Attribute{Type: ast.TypeRef("String")},
								"last":  ast.Attribute{Type: ast.TypeRef("String")},
							},
						},
					},
					Entities: ast.Entities{
						"NS::name": ast.Entity{},
						"NS::User": ast.Entity{
							Shape: &ast.RecordType{
								"n": ast.Attribute{Type: ast.TypeRef("name")},
							},
						},
					},
				},
			},
		}
		result, err := resolved.Resolve(s)
		testutil.OK(t, err)
		user := result.Entities["NS::User"]
		// "name" should resolve to the common type (a record), not the entity type
		rec, ok := user.Shape["n"].Type.(resolved.RecordType)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(rec), 2)
	})

	t.Run("entity_over_builtin", func(t *testing.T) {
		t.Parallel()
		// An entity type named "Long" shadows the built-in Long primitive.
		// A reference to "Long" should resolve to the entity type, not LongType.
		// The built-in is still accessible via __cedar::Long.
		s := &ast.Schema{
			Namespaces: ast.Namespaces{
				"NS": ast.Namespace{
					Entities: ast.Entities{
						"NS::Long": ast.Entity{},
						"NS::User": ast.Entity{
							Shape: &ast.RecordType{
								"x": ast.Attribute{Type: ast.TypeRef("Long")},
								"y": ast.Attribute{Type: ast.TypeRef("__cedar::Long")},
							},
						},
					},
				},
			},
		}
		result, err := resolved.Resolve(s)
		testutil.OK(t, err)
		user := result.Entities["NS::User"]
		// "Long" resolves to entity type NS::Long, not the built-in
		testutil.Equals(t, user.Shape["x"].Type, resolved.IsType(resolved.EntityType("NS::Long")))
		// "__cedar::Long" still resolves to the built-in
		testutil.Equals(t, user.Shape["y"].Type, resolved.IsType(resolved.LongType{}))
	})

	t.Run("common_over_builtin", func(t *testing.T) {
		t.Parallel()
		// A common type named "Long" shadows the built-in Long primitive.
		// A reference to "Long" should resolve to the common type, not LongType.
		s := &ast.Schema{
			Namespaces: ast.Namespaces{
				"NS": ast.Namespace{
					CommonTypes: ast.CommonTypes{
						"Long": ast.CommonType{
							Type: ast.RecordType{
								"value": ast.Attribute{Type: ast.TypeRef("__cedar::Long")},
							},
						},
					},
					Entities: ast.Entities{
						"NS::User": ast.Entity{
							Shape: &ast.RecordType{
								"x": ast.Attribute{Type: ast.TypeRef("Long")},
								"y": ast.Attribute{Type: ast.TypeRef("__cedar::Long")},
							},
						},
					},
				},
			},
		}
		result, err := resolved.Resolve(s)
		testutil.OK(t, err)
		user := result.Entities["NS::User"]
		// "Long" resolves to the common type (a record), not the built-in
		rec, ok := user.Shape["x"].Type.(resolved.RecordType)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(rec), 1)
		// "__cedar::Long" still resolves to the built-in
		testutil.Equals(t, user.Shape["y"].Type, resolved.IsType(resolved.LongType{}))
	})
}

func TestResolveNamespaceEntityRef(t *testing.T) {
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Entities: ast.Entities{
					"NS::User":  ast.Entity{MemberOf: []ast.EntityTypeRef{"Group"}},
					"NS::Group": ast.Entity{},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["NS::User"].MemberOf, []types.EntityType{"NS::Group"})
}

func TestResolveCrossNamespaceEntityRef(t *testing.T) {
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"A": ast.Namespace{
				Entities: ast.Entities{
					"A::User": ast.Entity{MemberOf: []ast.EntityTypeRef{"B::Group"}},
				},
			},
			"B": ast.Namespace{
				Entities: ast.Entities{
					"B::Group": ast.Entity{},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["A::User"].MemberOf, []types.EntityType{"B::Group"})
}

func TestResolveAction(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User":  ast.Entity{},
			"Photo": ast.Entity{},
		},
		Actions: ast.Actions{
			"view": ast.Action{
				AppliesTo: &ast.AppliesTo{
					Principals: []ast.EntityTypeRef{"User"},
					Resources:  []ast.EntityTypeRef{"Photo"},
					Context:    ast.RecordType{},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("Action", "view")
	view := result.Actions[uid]
	testutil.Equals(t, view.AppliesTo.Principals, []types.EntityType{"User"})
	testutil.Equals(t, view.AppliesTo.Resources, []types.EntityType{"Photo"})
}

func TestResolveActionMemberOf(t *testing.T) {
	s := &ast.Schema{
		Actions: ast.Actions{
			"view":     ast.Action{MemberOf: []ast.ParentRef{ast.ParentRefFromID("readOnly")}},
			"readOnly": ast.Action{},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("Action", "view")
	view := result.Actions[uid]
	testutil.Equals(t, view.MemberOf, []types.EntityUID{types.NewEntityUID("Action", "readOnly")})
}

func TestResolveActionCycle(t *testing.T) {
	s := &ast.Schema{
		Actions: ast.Actions{
			"a": ast.Action{MemberOf: []ast.ParentRef{ast.ParentRefFromID("b")}},
			"b": ast.Action{MemberOf: []ast.ParentRef{ast.ParentRefFromID("a")}},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveActionUndefinedParent(t *testing.T) {
	s := &ast.Schema{
		Actions: ast.Actions{
			"view": ast.Action{MemberOf: []ast.ParentRef{ast.ParentRefFromID("nonExistent")}},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveEnum(t *testing.T) {
	s := &ast.Schema{
		Enums: ast.Enums{
			"Status": ast.Enum{
				Values: []types.String{"active", "inactive"},
			},
		},
	}
	result, err := resolved.Resolve(s)
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
	s := &ast.Schema{
		Enums: ast.Enums{
			"Status": ast.Enum{Values: []types.String{"active"}},
		},
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"s": ast.Attribute{Type: ast.TypeRef("Status")},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["User"].Shape["s"].Type, resolved.IsType(resolved.EntityType("Status")))
}

func TestResolveSetType(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"tags": ast.Attribute{Type: ast.Set(ast.TypeRef("String"))},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	tags := result.Entities["User"].Shape["tags"]
	set, ok := tags.Type.(resolved.SetType)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, set.Element, resolved.IsType(resolved.StringType{}))
}

func TestResolveEntityWithTags(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Tags: ast.TypeRef("String"),
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["User"].Tags, resolved.IsType(resolved.StringType{}))
}

func TestResolveNamespacedAction(t *testing.T) {
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Actions: ast.Actions{
					"view": ast.Action{},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("NS::Action", "view")
	_, ok := result.Actions[uid]
	testutil.Equals(t, ok, true)
}

func TestResolveActionQualifiedParent(t *testing.T) {
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Actions: ast.Actions{
					"view": ast.Action{
						MemberOf: []ast.ParentRef{
							ast.NewParentRef("NS::Action", "readOnly"),
						},
					},
					"readOnly": ast.Action{},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("NS::Action", "view")
	view := result.Actions[uid]
	testutil.Equals(t, view.MemberOf, []types.EntityUID{types.NewEntityUID("NS::Action", "readOnly")})
}

func TestResolveActionContextNull(t *testing.T) {
	s := &ast.Schema{
		Actions: ast.Actions{
			"view": ast.Action{
				AppliesTo: &ast.AppliesTo{},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("Action", "view")
	view := result.Actions[uid]
	testutil.Equals(t, len(view.AppliesTo.Context), 0)
}

func TestResolveActionContextNonRecord(t *testing.T) {
	s := &ast.Schema{
		Actions: ast.Actions{
			"view": ast.Action{
				AppliesTo: &ast.AppliesTo{
					Context: ast.TypeRef("String"),
				},
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveActionPrincipalUndefined(t *testing.T) {
	s := &ast.Schema{
		Actions: ast.Actions{
			"view": ast.Action{
				AppliesTo: &ast.AppliesTo{
					Principals: []ast.EntityTypeRef{"NonExistent"},
				},
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveActionResourceUndefined(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{"User": ast.Entity{}},
		Actions: ast.Actions{
			"view": ast.Action{
				AppliesTo: &ast.AppliesTo{
					Principals: []ast.EntityTypeRef{"User"},
					Resources:  []ast.EntityTypeRef{"NonExistent"},
				},
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveCommonTypeChain(t *testing.T) {
	s := &ast.Schema{
		CommonTypes: ast.CommonTypes{
			"A": ast.CommonType{Type: ast.TypeRef("B")},
			"B": ast.CommonType{Type: ast.RecordType{
				"x": ast.Attribute{Type: ast.TypeRef("Long")},
			}},
		},
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"a": ast.Attribute{Type: ast.TypeRef("A")},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	a := result.Entities["User"].Shape["a"]
	rec, ok := a.Type.(resolved.RecordType)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, len(rec), 1)
}

func TestResolveQualifiedCommonType(t *testing.T) {
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				CommonTypes: ast.CommonTypes{
					"Ctx": ast.CommonType{
						Type: ast.RecordType{},
					},
				},
				Entities: ast.Entities{
					"NS::User": ast.Entity{
						Shape: &ast.RecordType{
							"c": ast.Attribute{Type: ast.TypeRef("NS::Ctx")},
						},
					},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	c := result.Entities["NS::User"].Shape["c"]
	_, ok := c.Type.(resolved.RecordType)
	testutil.Equals(t, ok, true)
}

func TestResolveQualifiedUndefined(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"x": ast.Attribute{Type: ast.TypeRef("NS::NonExistent")},
				},
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveEntityTagsUndefined(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Tags: ast.TypeRef("NonExistent"),
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveEntityShapeAttrUndefined(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"x": ast.Attribute{Type: ast.Set(ast.TypeRef("NonExistent"))},
				},
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveNamespaceOutput(t *testing.T) {
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Annotations: ast.Annotations{"doc": "test"},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	ns := result.Namespaces["NS"]
	testutil.Equals(t, ns.Name, types.Path("NS"))
	testutil.Equals(t, types.String(ns.Annotations["doc"]), types.String("test"))
}

func TestResolveEntityTypeRef(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"friend": ast.Attribute{Type: ast.EntityTypeRef("User")},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	friend := result.Entities["User"].Shape["friend"]
	testutil.Equals(t, friend.Type, resolved.IsType(resolved.EntityType("User")))
}

func TestResolveEntityTypeRefUndefined(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"x": ast.Attribute{Type: ast.EntityTypeRef("NonExistent")},
				},
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveEntityTypeRefQualified(t *testing.T) {
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"A": ast.Namespace{
				Entities: ast.Entities{
					"A::Foo": ast.Entity{},
				},
			},
		},
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"x": ast.Attribute{Type: ast.EntityTypeRef("A::Foo")},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	x := result.Entities["User"].Shape["x"]
	testutil.Equals(t, x.Type, resolved.IsType(resolved.EntityType("A::Foo")))
}

func TestResolveEntityTypeRefQualifiedUndefined(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"x": ast.Attribute{Type: ast.EntityTypeRef("A::NonExistent")},
				},
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveDirectPrimitiveTypes(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: &ast.RecordType{
					"s": ast.Attribute{Type: ast.StringType{}},
					"l": ast.Attribute{Type: ast.LongType{}},
					"b": ast.Attribute{Type: ast.BoolType{}},
					"e": ast.Attribute{Type: ast.ExtensionType("ipaddr")},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	user := result.Entities["User"]
	testutil.Equals(t, user.Shape["s"].Type, resolved.IsType(resolved.StringType{}))
	testutil.Equals(t, user.Shape["l"].Type, resolved.IsType(resolved.LongType{}))
	testutil.Equals(t, user.Shape["b"].Type, resolved.IsType(resolved.BoolType{}))
	testutil.Equals(t, user.Shape["e"].Type, resolved.IsType(resolved.ExtensionType("ipaddr")))
}

func TestResolveEnumEntityUIDBrokenIterator(t *testing.T) {
	s := &ast.Schema{
		Enums: ast.Enums{
			"Status": ast.Enum{
				Values: []types.String{"a", "b", "c"},
			},
		},
	}
	result, err := resolved.Resolve(s)
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
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Entities: ast.Entities{
					"NS::User": ast.Entity{
						Shape: &ast.RecordType{
							"x": ast.Attribute{Type: ast.TypeRef("NonExistent")},
						},
					},
				},
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveNamespacedEnumError(t *testing.T) {
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Enums: ast.Enums{
					"NS::Status": ast.Enum{Values: []types.String{"a"}},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, len(result.Enums), 1)
}

func TestResolveNamespacedActionsError(t *testing.T) {
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Entities: ast.Entities{
					"NS::User": ast.Entity{},
				},
				Actions: ast.Actions{
					"view": ast.Action{
						AppliesTo: &ast.AppliesTo{
							Context: ast.TypeRef("NonExistent"),
						},
					},
				},
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveEntityMemberOfValidationError(t *testing.T) {
	// Entity referencing an undefined parent type errors during resolution.
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				MemberOf: []ast.EntityTypeRef{"NonExistent"},
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveQualifiedEntityType(t *testing.T) {
	// Test resolving a qualified entity type ref
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Entities: ast.Entities{
					"NS::User":  ast.Entity{},
					"NS::Admin": ast.Entity{MemberOf: []ast.EntityTypeRef{"NS::User"}},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Entities["NS::Admin"].MemberOf, []types.EntityType{"NS::User"})
}

func TestResolveQualifiedEntityTypeRefAsType(t *testing.T) {
	// Test that a qualified entity type resolves when used through TypeRef
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Entities: ast.Entities{
					"NS::User": ast.Entity{
						Shape: &ast.RecordType{
							"ref": ast.Attribute{Type: ast.TypeRef("NS::Admin")},
						},
					},
					"NS::Admin": ast.Entity{},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	ref := result.Entities["NS::User"].Shape["ref"]
	testutil.Equals(t, ref.Type, resolved.IsType(resolved.EntityType("NS::Admin")))
}

func TestResolveEmptyNamespaceCommonType(t *testing.T) {
	// Test that empty namespace common types are found from a different namespace
	s := &ast.Schema{
		CommonTypes: ast.CommonTypes{
			"Ctx": ast.CommonType{Type: ast.RecordType{}},
		},
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Entities: ast.Entities{
					"NS::User": ast.Entity{
						Shape: &ast.RecordType{
							"c": ast.Attribute{Type: ast.TypeRef("Ctx")},
						},
					},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	c := result.Entities["NS::User"].Shape["c"]
	_, ok := c.Type.(resolved.RecordType)
	testutil.Equals(t, ok, true)
}

func TestResolveEmptyNamespaceEntityType(t *testing.T) {
	// Test that empty namespace entity types are found from a different namespace
	s := &ast.Schema{
		Entities: ast.Entities{
			"Global": ast.Entity{},
		},
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Entities: ast.Entities{
					"NS::User": ast.Entity{
						Shape: &ast.RecordType{
							"g": ast.Attribute{Type: ast.TypeRef("Global")},
						},
					},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	g := result.Entities["NS::User"].Shape["g"]
	testutil.Equals(t, g.Type, resolved.IsType(resolved.EntityType("Global")))
}

func TestResolveAnnotationsOnAttributes(t *testing.T) {
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Annotations: ast.Annotations{"doc": "user"},
				Shape: &ast.RecordType{
					"name": ast.Attribute{
						Type:        ast.TypeRef("String"),
						Annotations: ast.Annotations{"doc": "the name"},
					},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	user := result.Entities["User"]
	testutil.Equals(t, types.String(user.Annotations["doc"]), types.String("user"))
	testutil.Equals(t, types.String(user.Shape["name"].Annotations["doc"]), types.String("the name"))
}

func TestResolveEnumAnnotations(t *testing.T) {
	s := &ast.Schema{
		Enums: ast.Enums{
			"Status": ast.Enum{
				Annotations: ast.Annotations{"doc": "status"},
				Values:      []types.String{"a"},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, types.String(result.Enums["Status"].Annotations["doc"]), types.String("status"))
}

func TestResolveActionAnnotations(t *testing.T) {
	s := &ast.Schema{
		Actions: ast.Actions{
			"view": ast.Action{
				Annotations: ast.Annotations{"doc": "view action"},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	uid := types.NewEntityUID("Action", "view")
	testutil.Equals(t, types.String(result.Actions[uid].Annotations["doc"]), types.String("view action"))
}

func TestResolveBareEnumsError(t *testing.T) {
	// Test that bare enums with no errors pass through resolveEnums
	s := &ast.Schema{
		Enums: ast.Enums{
			"A": ast.Enum{Values: []types.String{"x"}},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	testutil.Equals(t, result.Enums["A"].Name, types.EntityType("A"))
}

func TestResolveBareActionsError(t *testing.T) {
	// Test bare actions with an error in context resolution
	s := &ast.Schema{
		Actions: ast.Actions{
			"view": ast.Action{
				AppliesTo: &ast.AppliesTo{
					Context: ast.TypeRef("NonExistent"),
				},
			},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveCommonTypeCycleInSet(t *testing.T) {
	s := &ast.Schema{
		CommonTypes: ast.CommonTypes{
			"A": ast.CommonType{Type: ast.Set(ast.TypeRef("B"))},
			"B": ast.CommonType{Type: ast.TypeRef("A")},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveCommonTypeCycleInRecord(t *testing.T) {
	s := &ast.Schema{
		CommonTypes: ast.CommonTypes{
			"A": ast.CommonType{Type: ast.RecordType{
				"x": ast.Attribute{Type: ast.TypeRef("B")},
			}},
			"B": ast.CommonType{Type: ast.TypeRef("A")},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveNamespacedEntityTypeRef(t *testing.T) {
	// Exercise resolveTypeRef line 421-423: unqualified name resolves to
	// a namespaced entity type (not common type) via disambiguation rule 2.
	// Use a Set<User> attribute where "User" should resolve to NS::User entity type.
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Entities: ast.Entities{
					"NS::User": ast.Entity{},
					"NS::Group": ast.Entity{
						Shape: &ast.RecordType{
							"members": ast.Attribute{Type: ast.SetType{Element: ast.TypeRef("User")}},
						},
					},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	members := result.Entities["NS::Group"].Shape["members"]
	setType, ok := members.Type.(resolved.SetType)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, setType.Element, resolved.IsType(resolved.EntityType("NS::User")))
}

func TestResolveNamespacedEnumTypeRef(t *testing.T) {
	// Exercise resolveTypeRef line 421-423: unqualified name resolves to
	// a namespaced enum type via disambiguation rule 2.
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Enums: ast.Enums{
					"NS::Color": ast.Enum{Values: []types.String{"red", "blue"}},
				},
				Entities: ast.Entities{
					"NS::Item": ast.Entity{
						Shape: &ast.RecordType{
							"color": ast.Attribute{Type: ast.TypeRef("Color")},
						},
					},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	color := result.Entities["NS::Item"].Shape["color"]
	testutil.Equals(t, color.Type, resolved.IsType(resolved.EntityType("NS::Color")))
}

func TestResolveUndefinedMemberOf(t *testing.T) {
	// Resolve returns an error when an entity references an undefined parent type.
	s := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{MemberOf: []ast.EntityTypeRef{"Nonexistent"}},
		},
	}
	_, err := resolved.Resolve(s)
	testutil.Error(t, err)
}

func TestResolveCedarBuiltinInTypePath(t *testing.T) {
	// Exercise resolveTypePath line 462-464: __cedar:: prefix in cycle detection.
	s := &ast.Schema{
		CommonTypes: ast.CommonTypes{
			"A": ast.CommonType{Type: ast.TypeRef("__cedar::String")},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	_ = result
}

func TestResolveQualifiedTypePath(t *testing.T) {
	// Exercise resolveTypePath line 465-467: qualified path with :: in cycle detection.
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				CommonTypes: ast.CommonTypes{
					"A": ast.CommonType{Type: ast.TypeRef("NS::B")},
					"B": ast.CommonType{Type: ast.StringType{}},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	_ = result
}

func TestResolveNamespacedCommonTypePath(t *testing.T) {
	// Exercise resolveTypePath line 470-472: namespaced common type ref in cycle detection.
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				CommonTypes: ast.CommonTypes{
					"A": ast.CommonType{Type: ast.TypeRef("B")},
					"B": ast.CommonType{Type: ast.StringType{}},
				},
			},
		},
	}
	result, err := resolved.Resolve(s)
	testutil.OK(t, err)
	_ = result
}
