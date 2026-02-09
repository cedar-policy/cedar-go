package parser_test

import (
	"os"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	ast2 "github.com/cedar-policy/cedar-go/x/exp/schema/ast"
	parser2 "github.com/cedar-policy/cedar-go/x/exp/schema/internal/parser"
)

func TestParseEmpty(t *testing.T) {
	schema, err := parser2.ParseSchema("", []byte(""))
	testutil.OK(t, err)
	testutil.Equals(t, schema, &ast2.Schema{})
}

func TestParseBasicFile(t *testing.T) {
	src, err := os.ReadFile("testdata/basic.cedarschema")
	testutil.OK(t, err)
	schema, err := parser2.ParseSchema("basic.cedarschema", src)
	testutil.OK(t, err)

	ns := schema.Namespaces["PhotoApp"]
	testutil.Equals(t, len(ns.Entities), 3)
	testutil.Equals(t, len(ns.Actions), 2)
	testutil.Equals(t, len(ns.CommonTypes), 1)

	user := ns.Entities["User"]
	testutil.Equals(t, user.ParentTypes, []ast2.EntityTypeRef{"Group"})
	testutil.Equals(t, user.Shape != nil, true)
	testutil.Equals(t, len(*user.Shape), 2)
	testutil.Equals(t, (*user.Shape)["name"].Type, ast2.IsType(ast2.TypeRef("String")))
	testutil.Equals(t, (*user.Shape)["name"].Optional, false)
	testutil.Equals(t, (*user.Shape)["age"].Type, ast2.IsType(ast2.TypeRef("Long")))
	testutil.Equals(t, (*user.Shape)["age"].Optional, true)

	group := ns.Entities["Group"]
	testutil.Equals(t, group.Shape == nil, true)
	testutil.Equals(t, len(group.ParentTypes), 0)

	photo := ns.Entities["Photo"]
	testutil.Equals(t, photo.Shape != nil, true)
	testutil.Equals(t, photo.Tags, ast2.IsType(ast2.TypeRef("String")))

	viewPhoto := ns.Actions["viewPhoto"]
	testutil.Equals(t, viewPhoto.AppliesTo != nil, true)
	testutil.Equals(t, viewPhoto.AppliesTo.Principals, []ast2.EntityTypeRef{"User"})
	testutil.Equals(t, viewPhoto.AppliesTo.Resources, []ast2.EntityTypeRef{"Photo"})

	createPhoto := ns.Actions["createPhoto"]
	testutil.Equals(t, len(createPhoto.Parents), 1)
	testutil.Equals(t, createPhoto.Parents[0], ast2.ParentRefFromID("viewPhoto"))
}

func TestParseMultiNameEntity(t *testing.T) {
	src := `entity A, B, C { name: String };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Entities), 3)
	for _, name := range []types.Ident{"A", "B", "C"} {
		_, ok := schema.Entities[name]
		testutil.Equals(t, ok, true)
	}
}

func TestParseEnumEntity(t *testing.T) {
	src := `entity Status enum ["active", "inactive", "pending"];`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Enums), 1)
	status := schema.Enums["Status"]
	testutil.Equals(t, status.Values, []types.String{"active", "inactive", "pending"})
}

func TestParseMultiNameEnum(t *testing.T) {
	src := `entity A, B enum ["x", "y"];`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Enums), 2)
	testutil.Equals(t, schema.Enums["A"].Values, []types.String{"x", "y"})
	testutil.Equals(t, schema.Enums["B"].Values, []types.String{"x", "y"})
}

func TestParseAnnotations(t *testing.T) {
	src := `
@doc("user entity")
entity User;
`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.Annotations["doc"], types.String("user entity"))
}

func TestParseAnnotationNoValue(t *testing.T) {
	src := `
@deprecated
entity User;
`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	_, ok := user.Annotations["deprecated"]
	testutil.Equals(t, ok, true)
}

func TestParseNamespaceAnnotations(t *testing.T) {
	src := `
@doc("my namespace")
namespace Foo {
	entity Bar;
}
`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	ns := schema.Namespaces["Foo"]
	testutil.Equals(t, ns.Annotations["doc"], types.String("my namespace"))
}

func TestParseActionStringName(t *testing.T) {
	src := `action "view photo" appliesTo { principal: User, resource: Photo };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	_, ok := schema.Actions["view photo"]
	testutil.Equals(t, ok, true)
}

func TestParseActionMultipleNames(t *testing.T) {
	src := `action read, write appliesTo { principal: User, resource: Resource };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Actions), 2)
	_, ok := schema.Actions["read"]
	testutil.Equals(t, ok, true)
	_, ok = schema.Actions["write"]
	testutil.Equals(t, ok, true)
}

func TestParseActionQualifiedParent(t *testing.T) {
	src := `action view in [MyApp::Action::"readOnly"] appliesTo { principal: User, resource: Photo };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	view := schema.Actions["view"]
	testutil.Equals(t, len(view.Parents), 1)
	testutil.Equals(t, view.Parents[0], ast2.NewParentRef("MyApp::Action", "readOnly"))
}

func TestParseActionBareParent(t *testing.T) {
	src := `action view in readOnly;`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	view := schema.Actions["view"]
	testutil.Equals(t, len(view.Parents), 1)
	testutil.Equals(t, view.Parents[0], ast2.ParentRefFromID("readOnly"))
}

func TestParseActionNoAppliesTo(t *testing.T) {
	src := `action view;`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	view := schema.Actions["view"]
	testutil.Equals(t, view.AppliesTo == nil, true)
}

func TestParseEntityInList(t *testing.T) {
	src := `entity User in [Admin, Group];`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.ParentTypes, []ast2.EntityTypeRef{"Admin", "Group"})
}

func TestParseEntityInSingle(t *testing.T) {
	src := `entity User in Admin;`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.ParentTypes, []ast2.EntityTypeRef{"Admin"})
}

func TestParseEntityWithEquals(t *testing.T) {
	src := `entity User = { name: String };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.Shape != nil, true)
}

func TestParseSetOfSet(t *testing.T) {
	src := `entity User { tags: Set<Set<Long>> };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, (*user.Shape)["tags"].Type, ast2.IsType(ast2.Set(ast2.Set(ast2.TypeRef("Long")))))
}

func TestParseTypeDecl(t *testing.T) {
	src := `type Context = { ip: ipaddr, name: String };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	ct := schema.CommonTypes["Context"]
	rec, ok := ct.Type.(ast2.RecordType)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, len(rec), 2)
}

func TestParseReservedTypeName(t *testing.T) {
	tests := []string{"Bool", "Boolean", "Entity", "Extension", "Long", "Record", "Set", "String"}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			src := `type ` + name + ` = { x: Long };`
			_, err := parser2.ParseSchema("", []byte(src))
			testutil.Error(t, err)
		})
	}
}

func TestParseComments(t *testing.T) {
	src := `
// This is a comment
entity User; // trailing comment
/* block
   comment */
entity Group;
`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Entities), 2)
}

func TestParseOptionalAttribute(t *testing.T) {
	src := `entity User { name?: String };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, (*schema.Entities["User"].Shape)["name"].Optional, true)
}

func TestParseBareDeclarations(t *testing.T) {
	src := `
entity User;
entity Group;
action view;
`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Entities), 2)
	testutil.Equals(t, len(schema.Actions), 1)
}

func TestParseMixedBareAndNamespaced(t *testing.T) {
	src := `
entity Global;
namespace Foo {
	entity Bar;
}
`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	_, ok := schema.Entities["Global"]
	testutil.Equals(t, ok, true)
	ns := schema.Namespaces["Foo"]
	_, ok = ns.Entities["Bar"]
	testutil.Equals(t, ok, true)
}

func TestParseNestedNamespacePath(t *testing.T) {
	src := `namespace Foo::Bar { entity Baz; }`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	ns := schema.Namespaces["Foo::Bar"]
	_, ok := ns.Entities["Baz"]
	testutil.Equals(t, ok, true)
}

func TestParseCedarNamespace(t *testing.T) {
	src := `entity User { name: __cedar::String };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, (*user.Shape)["name"].Type, ast2.IsType(ast2.TypeRef("__cedar::String")))
}

func TestParseEntityTypeQualified(t *testing.T) {
	src := `entity User in NS::Group;`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.ParentTypes, []ast2.EntityTypeRef{"NS::Group"})
}

func TestParseActionAppliesToEmptyPrincipal(t *testing.T) {
	src := `action view appliesTo { principal: [], resource: Photo };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	view := schema.Actions["view"]
	testutil.Equals(t, len(view.AppliesTo.Principals), 0)
	testutil.Equals(t, view.AppliesTo.Resources, []ast2.EntityTypeRef{"Photo"})
}

func TestParseContextTypeName(t *testing.T) {
	src := `action view appliesTo { principal: User, resource: Photo, context: MyContext };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	view := schema.Actions["view"]
	testutil.Equals(t, view.AppliesTo.Context, ast2.IsType(ast2.TypeRef("MyContext")))
}

func TestParseAttrAnnotations(t *testing.T) {
	src := `entity User {
	@doc("the name")
	name: String
};`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, (*user.Shape)["name"].Annotations["doc"], types.String("the name"))
}

func TestParseUnicodeString(t *testing.T) {
	src := `entity User enum ["\u{1F600}"];`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, schema.Enums["User"].Values, []types.String{"\U0001F600"})
}

func TestParseTrailingCommaInRecord(t *testing.T) {
	src := `entity User { name: String, age: Long, };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(*schema.Entities["User"].Shape), 2)
}

func TestParseTrailingCommaInEntityList(t *testing.T) {
	src := `entity User in [Admin, Group,];`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Entities["User"].ParentTypes), 2)
}

func TestParseTrailingCommaInAppliesTo(t *testing.T) {
	src := `action view appliesTo { principal: User, resource: Photo, };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Actions["view"].AppliesTo.Principals), 1)
}

func TestParseActionParentStringLiteral(t *testing.T) {
	src := `action view in ["readOnly"];`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	view := schema.Actions["view"]
	testutil.Equals(t, view.Parents[0], ast2.ParentRefFromID("readOnly"))
}

func TestParseEntityEmptyRecord(t *testing.T) {
	src := `entity User {};`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.Shape != nil, true)
	testutil.Equals(t, len(*user.Shape), 0)
}

func TestParseEntityInlineRecord(t *testing.T) {
	src := `entity User { info: { name: String, age: Long } };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	rec, ok := (*user.Shape)["info"].Type.(ast2.RecordType)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, len(rec), 2)
}

func TestParseEntityWithTagsAndShape(t *testing.T) {
	src := `entity User { name: String } tags Long;`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.Shape != nil, true)
	testutil.Equals(t, user.Tags, ast2.IsType(ast2.TypeRef("Long")))
}

func TestParseEnumTrailingComma(t *testing.T) {
	src := `entity Status enum ["a", "b",];`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Enums["Status"].Values), 2)
}

func TestParseActionAttributesDeprecated(t *testing.T) {
	src := `action view appliesTo { principal: User, resource: Photo } attributes {};`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, schema.Actions["view"].AppliesTo != nil, true)
}

func TestParseMultipleNamespaces(t *testing.T) {
	src := `
namespace A {
	entity Foo;
}
namespace B {
	entity Bar;
}
`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Namespaces), 2)
	_, ok := schema.Namespaces["A"].Entities["Foo"]
	testutil.Equals(t, ok, true)
	_, ok = schema.Namespaces["B"].Entities["Bar"]
	testutil.Equals(t, ok, true)
}

func TestParseErrorPosition(t *testing.T) {
	src := `entity User {
	name String
};`
	_, err := parser2.ParseSchema("test.cedarschema", []byte(src))
	testutil.Error(t, err)
	errStr := err.Error()
	testutil.Equals(t, true, len(errStr) > 0)
}

func TestMarshalRoundTrip(t *testing.T) {
	src := `namespace PhotoApp {
	entity User in [Group] {
		name: String,
		age?: Long
	};

	entity Group;

	entity Photo {
		owner: User,
		tags: Set<String>
	} tags String;

	action viewPhoto appliesTo {
		principal: User,
		resource: Photo,
		context: {}
	};

	type Context = {
		ip: ipaddr
	};
}
`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)

	out := parser2.MarshalSchema(schema)

	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)

	out2 := parser2.MarshalSchema(schema2)
	testutil.Equals(t, string(out), string(out2))
}

func TestMarshalEmpty(t *testing.T) {
	schema := &ast2.Schema{}
	out := parser2.MarshalSchema(schema)
	testutil.Equals(t, string(out), "")
}

func TestMarshalBareEntities(t *testing.T) {
	schema := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, len(schema2.Entities), 1)
}

func TestMarshalEnumEntity(t *testing.T) {
	schema := &ast2.Schema{
		Enums: ast2.Enums{
			"Status": ast2.Enum{
				Values: []types.String{"active", "inactive"},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, schema2.Enums["Status"].Values, []types.String{"active", "inactive"})
}

func TestMarshalActionParentRef(t *testing.T) {
	schema := &ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{
				Parents: []ast2.ParentRef{
					ast2.NewParentRef("NS::Action", "readOnly"),
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, schema2.Actions["view"].Parents[0], ast2.NewParentRef("NS::Action", "readOnly"))
}

func TestMarshalStringActionName(t *testing.T) {
	schema := &ast2.Schema{
		Actions: ast2.Actions{
			"view photo": ast2.Action{},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Actions["view photo"]
	testutil.Equals(t, ok, true)
}

func TestMarshalAnnotations(t *testing.T) {
	schema := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Annotations: ast2.Annotations{
					"doc": "user entity",
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, schema2.Entities["User"].Annotations["doc"], types.String("user entity"))
}

func TestMarshalAllTypes(t *testing.T) {
	schema := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"s":   ast2.Attribute{Type: ast2.TypeRef("String")},
					"l":   ast2.Attribute{Type: ast2.TypeRef("Long")},
					"b":   ast2.Attribute{Type: ast2.TypeRef("Bool")},
					"ip":  ast2.Attribute{Type: ast2.TypeRef("ipaddr")},
					"dec": ast2.Attribute{Type: ast2.TypeRef("decimal")},
					"dt":  ast2.Attribute{Type: ast2.TypeRef("datetime")},
					"dur": ast2.Attribute{Type: ast2.TypeRef("duration")},
					"set": ast2.Attribute{Type: ast2.Set(ast2.TypeRef("Long"))},
					"rec": ast2.Attribute{Type: ast2.RecordType{}},
					"ref": ast2.Attribute{Type: ast2.EntityTypeRef("NS::Foo")},
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	_, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
}

func TestMarshalMultipleEntityTypeRefs(t *testing.T) {
	schema := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				ParentTypes: []ast2.EntityTypeRef{"Admin", "Group"},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, len(schema2.Entities["User"].ParentTypes), 2)
}

func TestMarshalMultipleActionParents(t *testing.T) {
	schema := &ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{
				Parents: []ast2.ParentRef{
					ast2.ParentRefFromID("read"),
					ast2.ParentRefFromID("write"),
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, len(schema2.Actions["view"].Parents), 2)
}

func TestMarshalQuotedAttrName(t *testing.T) {
	schema := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Shape: &ast2.RecordType{
					"has space": ast2.Attribute{Type: ast2.TypeRef("String")},
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := (*schema2.Entities["User"].Shape)["has space"]
	testutil.Equals(t, ok, true)
}

func TestMarshalAnnotationNoValue(t *testing.T) {
	schema := &ast2.Schema{
		Entities: ast2.Entities{
			"User": ast2.Entity{
				Annotations: ast2.Annotations{
					"deprecated": "",
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Entities["User"].Annotations["deprecated"]
	testutil.Equals(t, ok, true)
}

func TestMarshalNamespace(t *testing.T) {
	schema := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"Foo": ast2.Namespace{
				Entities: ast2.Entities{
					"Bar": ast2.Entity{},
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Namespaces["Foo"].Entities["Bar"]
	testutil.Equals(t, ok, true)
}

func TestMarshalNamespaceWithAnnotations(t *testing.T) {
	schema := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"Foo": ast2.Namespace{
				Annotations: ast2.Annotations{
					"doc": "foo ns",
				},
				Entities: ast2.Entities{
					"Bar": ast2.Entity{},
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, schema2.Namespaces["Foo"].Annotations["doc"], types.String("foo ns"))
}

func TestMarshalActionBareParent(t *testing.T) {
	schema := &ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{
				Parents: []ast2.ParentRef{
					ast2.ParentRefFromID("readOnly"),
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, schema2.Actions["view"].Parents[0], ast2.ParentRefFromID("readOnly"))
}

func TestMarshalPrimitiveTypes(t *testing.T) {
	schema := &ast2.Schema{
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
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, len(*schema2.Entities["User"].Shape), 4)
}

func TestMarshalEmptyNamespaceMap(t *testing.T) {
	schema := &ast2.Schema{
		Namespaces: ast2.Namespaces{},
	}
	out := parser2.MarshalSchema(schema)
	testutil.Equals(t, string(out), "")
}

func TestMarshalEmptyCommonTypes(t *testing.T) {
	schema := &ast2.Schema{
		CommonTypes: ast2.CommonTypes{},
	}
	out := parser2.MarshalSchema(schema)
	testutil.Equals(t, string(out), "")
}

func TestMarshalNamespaceCommonTypes(t *testing.T) {
	schema := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				CommonTypes: ast2.CommonTypes{
					"Ctx": ast2.CommonType{
						Type: ast2.RecordType{},
					},
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Namespaces["NS"].CommonTypes["Ctx"]
	testutil.Equals(t, ok, true)
}

func TestMarshalBareAndNamespaced(t *testing.T) {
	schema := &ast2.Schema{
		Entities: ast2.Entities{
			"Global": ast2.Entity{},
		},
		Namespaces: ast2.Namespaces{
			"Foo": ast2.Namespace{
				Entities: ast2.Entities{
					"Bar": ast2.Entity{},
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Entities["Global"]
	testutil.Equals(t, ok, true)
	_, ok = schema2.Namespaces["Foo"].Entities["Bar"]
	testutil.Equals(t, ok, true)
}

func TestMarshalMultipleDecls(t *testing.T) {
	s := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"NS1": ast2.Namespace{},
			"NS2": ast2.Namespace{
				CommonTypes: ast2.CommonTypes{
					"A": ast2.CommonType{Type: ast2.StringType{}},
					"B": ast2.CommonType{Type: ast2.LongType{}},
				},
				Enums: ast2.Enums{
					"Color": ast2.Enum{Values: []types.String{"red"}},
					"Size":  ast2.Enum{Values: []types.String{"small"}},
				},
			},
		},
	}
	result := parser2.MarshalSchema(s)
	testutil.Equals(t, len(result) > 0, true)
	_, err := parser2.ParseSchema("", result)
	testutil.OK(t, err)
}

func TestMarshalNamespaceQualifiedKeyRoundTripBreaks(t *testing.T) {
	// Entity keys must be bare Idents, not namespace-qualified.
	// A qualified key like "Foo::Bar" in namespace "Foo" marshals as
	// "entity Foo::Bar" inside the namespace block. Re-parsing fails
	// because "::" is not valid in a bare entity declaration name.
	schema := &ast2.Schema{
		Namespaces: ast2.Namespaces{
			"Foo": ast2.Namespace{
				Entities: ast2.Entities{
					"Foo::Bar": ast2.Entity{},
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	_, err := parser2.ParseSchema("", out)
	testutil.Error(t, err)
}

func TestParseSchemaErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// Lexer error coverage: '$' is invalid and causes a lexer error
		{"readToken after namespace keyword", `namespace $`},
		{"readToken in inner annotations", `namespace Foo { @$`},
		{"readToken after namespace close brace", `namespace Foo { entity Bar; }$`},
		{"readToken after entity keyword", `entity $`},
		{"readToken after action keyword", `action $`},
		{"readToken after type keyword", `type $`},
		{"readToken after enum keyword", `entity Foo enum $`},
		{"readToken after in keyword", `entity Foo in $`},
		{"readToken after equals", `entity Foo = $`},
		{"parseRecordType error after equals", `entity Foo = { $ }`},
		{"readToken after tags keyword", `entity Foo tags $`},
		{"readToken after string in enum", `entity Foo enum ["a"$`},
		{"readToken after comma in enum", `entity Foo enum ["a",$`},
		{"readToken after bracket in enum", `entity Foo enum ["a"]$`},
		{"expect semicolon after enum", `entity Foo enum ["a"]}`},
		{"readToken after action in", `action view in $`},
		{"readToken after appliesTo", `action view appliesTo $`},
		{"readToken after attributes keyword", `action view attributes $`},
		{"expect lbrace after attributes", `action view attributes foo`},
		{"expect rbrace after attributes lbrace", `action view attributes { foo`},
		{"expect semicolon after action", `action view}`},
		{"readToken after type name", `type Foo$`},
		{"parseType error in type decl", `type Foo = $`},
		{"expect semicolon after type decl", `type Foo = Long}`},
		{"readToken after at sign", `@$`},
		{"readToken after annotation name", `@doc$`},
		{"readToken after annotation lparen", `@doc($`},
		{"readToken after annotation string value", `@doc("x"$`},
		{"expect rparen after annotation value", `@doc("x"}`},
		{"readToken after path ident", `entity Foo in Bar$`},
		{"readToken after path double colon", `entity Foo in Bar::$`},
		{"readToken after path second ident", `entity Foo in Bar::Baz$`},
		{"readToken after ref ident", `action view in foo$`},
		{"readToken after ref double colon", `action view in [foo::$]`},
		{"readToken after ref string", `action view in [Foo::"bar"$]`},
		{"readToken after ref second ident", `action view in [Foo::Bar$]`},
		{"readToken after first ident in list", `entity Foo$`},
		{"readToken after comma in ident list", `entity Foo,$`},
		{"readToken after second ident in list", `entity Foo, Bar$`},
		{"readToken after comma in names", `action foo,$`},
		{"readToken after ident name", `action foo in [bar$]`},
		{"readToken after string name", `action "foo"$`},
		{"readToken after entity type lbracket", `entity Foo in [$`},
		{"readToken after entity type comma", `entity Foo in [Bar,$`},
		{"readToken after entity type rbracket", `entity Foo in [Bar]$`},
		{"readToken after action parent lbracket", `action view in [$`},
		{"readToken after action parent comma", `action view in [foo,$`},
		{"parseQualName error for single parent", `action view in 42`},
		{"readToken after qual name string", `action view in ["foo"$]`},
		{"EOF inside appliesTo", `action view appliesTo { principal: User`},
		{"readToken after principal", `action view appliesTo { principal$`},
		{"expect colon after principal", `action view appliesTo { principal User }`},
		{"parseEntityTypes error after principal", `action view appliesTo { principal: $ }`},
		{"readToken after resource", `action view appliesTo { resource$`},
		{"expect colon after resource", `action view appliesTo { resource User }`},
		{"parseEntityTypes error after resource", `action view appliesTo { resource: $ }`},
		{"readToken after context", `action view appliesTo { context$`},
		{"expect colon after context", `action view appliesTo { context User }`},
		{"parseType error after context", `action view appliesTo { context: $ }`},
		{"readToken after comma in appliesTo", `action view appliesTo { principal: User,$`},
		{"parseRecordType error in type", `entity Foo { x: { $ } };`},
		{"readToken after Set", `entity Foo { x: Set$`},
		{"expect langle after Set", `entity Foo { x: Set(Long) };`},
		{"parseType error inside Set", `entity Foo { x: Set<$> };`},
		{"expect rangle after Set element", `entity Foo { x: Set<Long; };`},
		{"parsePath error in type", `entity Foo { x: 42 };`},
		{"parseAnnotations error in record", `entity Foo { @$ name: Long };`},
		{"parseName error for attr name", `entity Foo { 42: Long };`},
		{"readToken after question mark", `entity Foo { x?$`},
		{"parseType error in record attr", `entity Foo { x: $ };`},
		{"readToken after comma in record", `entity Foo { x: Long,$`},
		{"readToken after ident in action name", `action foo$`},

		// Semantic error paths
		{"namespace path not ident", `namespace "bad" {}`},
		{"non-ident in namespace decl", `namespace Foo { ; }`},
		{"parseType error after tags", `entity Foo tags ;`},
		{"enum value not string", `entity Foo enum [Bar];`},
		{"parseType error in type decl semantic", `type Foo = ;`},
		{"annotation value not string", `@doc(;) entity Foo;`},
		{"path not ident at start", `namespace Foo { entity Bar in "bad"; }`},
		{"path not ident after double colon", `namespace Foo { entity Bar in Baz::"bad"; }`},
		{"ref not ident or string at start", `action view in [;];`},
		{"ref not ident or string after double colon", `action view in [Foo::;];`},
		{"non-ident after comma in ident list", `entity Foo, "bar" {};`},
		{"parseName error after comma in names", `action foo, ;`},
		{"name neither ident nor string", `action ;`},
		{"parsePath error in entity type list", `entity Foo in ["bad"];`},
		{"non-ident in appliesTo", `action view appliesTo { "bad" };`},
		{"parseEntityTypes error for principal type", `action view appliesTo { principal: "bad" };`},
		{"parseEntityTypes error for resource type", `action view appliesTo { resource: "bad" };`},
		{"parseType error for context type", `action view appliesTo { context: ; };`},
		{"parseType error inside Set angle brackets", `entity Foo { x: Set<;> };`},
		{"EOF inside record", `entity Foo {`},
		{"parseName error for attr non-ident non-string", `entity Foo { ;: Long };`},
		{"expect lbrace after namespace path", `namespace Foo entity Bar;`},
		{"type name not ident", `type "bad" = Long;`},

		// General parse error coverage
		{"unterminated namespace", `namespace Foo { entity Bar;`},
		{"invalid token", `entity User $ {};`},
		{"unterminated string", `entity User enum ["unterminated;`},
		{"unterminated block comment", `/* unterminated`},
		{"missing semicolon", `entity User`},
		{"bad declaration keyword", `foobar;`},
		{"non-decl keyword in namespace", `namespace Foo { bogus; }`},
		{"bad annotation name", `@ "bad" entity User;`},
		{"bad annotation value type", `@doc(42) entity User;`},
		{"entity name not ident", `entity "bad";`},
		{"type decl name not ident", `type 42 = Long;`},
		{"enum value not string literal", `entity Foo enum [42];`},
		{"enum bad separator", `entity Foo enum ["a" "b"];`},
		{"appliesTo unknown keyword", `action view appliesTo { foo: User };`},
		{"appliesTo not ident", `action view appliesTo { 42: User };`},
		{"appliesTo EOF", `action view appliesTo {`},
		{"record EOF", `entity User {`},
		{"path bad after double colon", `entity User in Foo::42;`},
		{"path not ident", `entity User in 42;`},
		{"entity type list bad separator", `entity User in [Foo Bar];`},
		{"action parent list bad separator", `action view in [foo bar];`},
		{"action name not ident or string", `action 42;`},
		{"decl not ident", `42;`},
		{"record bad attr name", `entity User { 42: Long };`},
		{"type decl missing equals", `type Foo Long;`},
		{"ref bad after double colon", `action view in [Foo::42];`},
		{"action parent ref not ident", `action view in [42];`},
		{"appliesTo missing brace", `action view appliesTo principal: User;`},
		{"appliesTo missing colon", `action view appliesTo { principal User };`},
		{"type not ident or brace", `entity User { name: 42 };`},
		{"namespace EOF", `namespace Foo {`},
		{"record missing colon", `entity User { name String };`},
		{"idents non-ident after comma", `entity A, 42 {};`},
		{"names non-ident after comma", `action foo, 42;`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser2.ParseSchema("", []byte(tt.input))
			testutil.Error(t, err)
		})
	}
}
