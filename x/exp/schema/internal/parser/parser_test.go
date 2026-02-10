package parser_test

import (
	"os"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema/internal/parser"
)

func TestParseEmpty(t *testing.T) {
	schema, err := parser.ParseSchema("", []byte(""))
	testutil.OK(t, err)
	testutil.Equals(t, schema, &ast.Schema{})
}

func TestParseBasicFile(t *testing.T) {
	src, err := os.ReadFile("testdata/basic.cedarschema")
	testutil.OK(t, err)
	schema, err := parser.ParseSchema("basic.cedarschema", src)
	testutil.OK(t, err)

	ns := schema.Namespaces["PhotoApp"]
	testutil.Equals(t, len(ns.Entities), 3)
	testutil.Equals(t, len(ns.Actions), 2)
	testutil.Equals(t, len(ns.CommonTypes), 1)

	user := ns.Entities["User"]
	testutil.Equals(t, user.ParentTypes, []ast.EntityTypeRef{"Group"})
	testutil.Equals(t, user.Shape != nil, true)
	testutil.Equals(t, len(user.Shape), 2)
	testutil.Equals(t, user.Shape["name"].Type, ast.IsType(ast.TypeRef("String")))
	testutil.Equals(t, user.Shape["name"].Optional, false)
	testutil.Equals(t, user.Shape["age"].Type, ast.IsType(ast.TypeRef("Long")))
	testutil.Equals(t, user.Shape["age"].Optional, true)

	group := ns.Entities["Group"]
	testutil.Equals(t, group.Shape == nil, true)
	testutil.Equals(t, len(group.ParentTypes), 0)

	photo := ns.Entities["Photo"]
	testutil.Equals(t, photo.Shape != nil, true)
	testutil.Equals(t, photo.Tags, ast.IsType(ast.TypeRef("String")))

	viewPhoto := ns.Actions["viewPhoto"]
	testutil.Equals(t, viewPhoto.AppliesTo != nil, true)
	testutil.Equals(t, viewPhoto.AppliesTo.Principals, []ast.EntityTypeRef{"User"})
	testutil.Equals(t, viewPhoto.AppliesTo.Resources, []ast.EntityTypeRef{"Photo"})

	createPhoto := ns.Actions["createPhoto"]
	testutil.Equals(t, len(createPhoto.Parents), 1)
	testutil.Equals(t, createPhoto.Parents[0], ast.ParentRefFromID("viewPhoto"))
}

func TestParseMultiNameEntity(t *testing.T) {
	src := `entity A, B, C { name: String };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Entities), 3)
	for _, name := range []types.Ident{"A", "B", "C"} {
		_, ok := schema.Entities[name]
		testutil.Equals(t, ok, true)
	}
}

func TestParseEnumEntity(t *testing.T) {
	src := `entity Status enum ["active", "inactive", "pending"];`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Enums), 1)
	status := schema.Enums["Status"]
	testutil.Equals(t, status.Values, []types.String{"active", "inactive", "pending"})
}

func TestParseMultiNameEnum(t *testing.T) {
	src := `entity A, B enum ["x", "y"];`
	schema, err := parser.ParseSchema("", []byte(src))
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
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.Annotations["doc"], types.String("user entity"))
}

func TestParseAnnotationNoValue(t *testing.T) {
	src := `
@deprecated
entity User;
`
	schema, err := parser.ParseSchema("", []byte(src))
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
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	ns := schema.Namespaces["Foo"]
	testutil.Equals(t, ns.Annotations["doc"], types.String("my namespace"))
}

func TestParseActionStringName(t *testing.T) {
	src := `action "view photo" appliesTo { principal: User, resource: Photo };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	_, ok := schema.Actions["view photo"]
	testutil.Equals(t, ok, true)
}

func TestParseReservedWordAsStringName(t *testing.T) {
	// Reserved Cedar keywords are allowed when quoted as strings
	src := `action "true" appliesTo { principal: User, resource: Photo };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	_, ok := schema.Actions["true"]
	testutil.Equals(t, ok, true)
}

func TestParseReservedWordAsStringAttr(t *testing.T) {
	// Reserved Cedar keywords are allowed as quoted attribute names
	src := `entity Foo { "if": String };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	_, ok := schema.Entities["Foo"].Shape["if"]
	testutil.Equals(t, ok, true)
}

func TestParseReservedWordAsAnnotationName(t *testing.T) {
	src := `@in("group") entity Foo;`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, schema.Entities["Foo"].Annotations["in"], types.String("group"))
}

func TestParseCedarAsActionName(t *testing.T) {
	src := `action __cedar;`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	_, ok := schema.Actions["__cedar"]
	testutil.Equals(t, ok, true)
}

func TestParseCedarAsAttrName(t *testing.T) {
	src := `entity Foo { __cedar: String };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	_, ok := schema.Entities["Foo"].Shape["__cedar"]
	testutil.Equals(t, ok, true)
}

func TestParseActionMultipleNames(t *testing.T) {
	src := `action read, write appliesTo { principal: User, resource: Resource };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Actions), 2)
	_, ok := schema.Actions["read"]
	testutil.Equals(t, ok, true)
	_, ok = schema.Actions["write"]
	testutil.Equals(t, ok, true)
}

func TestParseActionQualifiedParent(t *testing.T) {
	src := `action view in [MyApp::Action::"readOnly"] appliesTo { principal: User, resource: Photo };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	view := schema.Actions["view"]
	testutil.Equals(t, len(view.Parents), 1)
	testutil.Equals(t, view.Parents[0], ast.NewParentRef("MyApp::Action", "readOnly"))
}

func TestParseActionBareParent(t *testing.T) {
	src := `action view in readOnly;`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	view := schema.Actions["view"]
	testutil.Equals(t, len(view.Parents), 1)
	testutil.Equals(t, view.Parents[0], ast.ParentRefFromID("readOnly"))
}

func TestParseActionNoAppliesTo(t *testing.T) {
	src := `action view;`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	view := schema.Actions["view"]
	testutil.Equals(t, view.AppliesTo == nil, true)
}

func TestParseEntityInList(t *testing.T) {
	src := `entity User in [Admin, Group];`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.ParentTypes, []ast.EntityTypeRef{"Admin", "Group"})
}

func TestParseEntityInSingle(t *testing.T) {
	src := `entity User in Admin;`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.ParentTypes, []ast.EntityTypeRef{"Admin"})
}

func TestParseEntityWithEquals(t *testing.T) {
	src := `entity User = { name: String };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.Shape != nil, true)
}

func TestParseSetOfSet(t *testing.T) {
	src := `entity User { tags: Set<Set<Long>> };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.Shape["tags"].Type, ast.IsType(ast.Set(ast.Set(ast.TypeRef("Long")))))
}

func TestParseTypeDecl(t *testing.T) {
	src := `type Context = { ip: ipaddr, name: String };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	ct := schema.CommonTypes["Context"]
	rec, ok := ct.Type.(ast.RecordType)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, len(rec), 2)
}

func TestParseReservedTypeName(t *testing.T) {
	tests := []string{"Bool", "Boolean", "Entity", "Extension", "Long", "Record", "Set", "String"}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			src := `type ` + name + ` = { x: Long };`
			_, err := parser.ParseSchema("", []byte(src))
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
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Entities), 2)
}

func TestParseOptionalAttribute(t *testing.T) {
	src := `entity User { name?: String };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, schema.Entities["User"].Shape["name"].Optional, true)
}

func TestParseBareDeclarations(t *testing.T) {
	src := `
entity User;
entity Group;
action view;
`
	schema, err := parser.ParseSchema("", []byte(src))
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
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	_, ok := schema.Entities["Global"]
	testutil.Equals(t, ok, true)
	ns := schema.Namespaces["Foo"]
	_, ok = ns.Entities["Bar"]
	testutil.Equals(t, ok, true)
}

func TestParseNestedNamespacePath(t *testing.T) {
	src := `namespace Foo::Bar { entity Baz; }`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	ns := schema.Namespaces["Foo::Bar"]
	_, ok := ns.Entities["Baz"]
	testutil.Equals(t, ok, true)
}

func TestParseCedarNamespace(t *testing.T) {
	src := `entity User { name: __cedar::String };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.Shape["name"].Type, ast.IsType(ast.TypeRef("__cedar::String")))
}

func TestParseEntityTypeQualified(t *testing.T) {
	src := `entity User in NS::Group;`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.ParentTypes, []ast.EntityTypeRef{"NS::Group"})
}

func TestParseActionAppliesToEmptyPrincipal(t *testing.T) {
	src := `action view appliesTo { principal: [], resource: Photo };`
	_, err := parser.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseContextTypeName(t *testing.T) {
	src := `action view appliesTo { principal: User, resource: Photo, context: MyContext };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	view := schema.Actions["view"]
	testutil.Equals(t, view.AppliesTo.Context, ast.IsType(ast.TypeRef("MyContext")))
}

func TestParseAttrAnnotations(t *testing.T) {
	src := `entity User {
	@doc("the name")
	name: String
};`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.Shape["name"].Annotations["doc"], types.String("the name"))
}

func TestParseUnicodeString(t *testing.T) {
	src := `entity User enum ["\u{1F600}"];`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, schema.Enums["User"].Values, []types.String{"\U0001F600"})
}

func TestParseTrailingCommaInRecord(t *testing.T) {
	src := `entity User { name: String, age: Long, };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Entities["User"].Shape), 2)
}

func TestParseTrailingCommaInEntityList(t *testing.T) {
	src := `entity User in [Admin, Group,];`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Entities["User"].ParentTypes), 2)
}

func TestParseTrailingCommaInAppliesTo(t *testing.T) {
	src := `action view appliesTo { principal: User, resource: Photo, };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Actions["view"].AppliesTo.Principals), 1)
}

func TestParseActionParentStringLiteral(t *testing.T) {
	src := `action view in ["readOnly"];`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	view := schema.Actions["view"]
	testutil.Equals(t, view.Parents[0], ast.ParentRefFromID("readOnly"))
}

func TestParseEntityEmptyRecord(t *testing.T) {
	src := `entity User {};`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.Shape != nil, true)
	testutil.Equals(t, len(user.Shape), 0)
}

func TestParseEntityInlineRecord(t *testing.T) {
	src := `entity User { info: { name: String, age: Long } };`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	rec, ok := user.Shape["info"].Type.(ast.RecordType)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, len(rec), 2)
}

func TestParseEntityWithTagsAndShape(t *testing.T) {
	src := `entity User { name: String } tags Long;`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.Shape != nil, true)
	testutil.Equals(t, user.Tags, ast.IsType(ast.TypeRef("Long")))
}

func TestParseEnumTrailingComma(t *testing.T) {
	src := `entity Status enum ["a", "b",];`
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Enums["Status"].Values), 2)
}

func TestParseActionAttributesDeprecated(t *testing.T) {
	src := `action view appliesTo { principal: User, resource: Photo } attributes {};`
	schema, err := parser.ParseSchema("", []byte(src))
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
	schema, err := parser.ParseSchema("", []byte(src))
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
	_, err := parser.ParseSchema("test.cedarschema", []byte(src))
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
	schema, err := parser.ParseSchema("", []byte(src))
	testutil.OK(t, err)

	out := parser.MarshalSchema(schema)

	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)

	out2 := parser.MarshalSchema(schema2)
	testutil.Equals(t, string(out), string(out2))
}

func TestMarshalEmpty(t *testing.T) {
	schema := &ast.Schema{}
	out := parser.MarshalSchema(schema)
	testutil.Equals(t, string(out), "")
}

func TestMarshalBareEntities(t *testing.T) {
	schema := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, len(schema2.Entities), 1)
}

func TestMarshalEnumEntity(t *testing.T) {
	schema := &ast.Schema{
		Enums: ast.Enums{
			"Status": ast.Enum{
				Values: []types.String{"active", "inactive"},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, schema2.Enums["Status"].Values, []types.String{"active", "inactive"})
}

func TestMarshalActionParentRef(t *testing.T) {
	schema := &ast.Schema{
		Actions: ast.Actions{
			"view": ast.Action{
				Parents: []ast.ParentRef{
					ast.NewParentRef("NS::Action", "readOnly"),
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, schema2.Actions["view"].Parents[0], ast.NewParentRef("NS::Action", "readOnly"))
}

func TestMarshalStringActionName(t *testing.T) {
	schema := &ast.Schema{
		Actions: ast.Actions{
			"view photo": ast.Action{},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Actions["view photo"]
	testutil.Equals(t, ok, true)
}

func TestMarshalAnnotations(t *testing.T) {
	schema := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Annotations: ast.Annotations{
					"doc": "user entity",
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, schema2.Entities["User"].Annotations["doc"], types.String("user entity"))
}

func TestMarshalAllTypes(t *testing.T) {
	schema := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: ast.RecordType{
					"s":   ast.Attribute{Type: ast.TypeRef("String")},
					"l":   ast.Attribute{Type: ast.TypeRef("Long")},
					"b":   ast.Attribute{Type: ast.TypeRef("Bool")},
					"ip":  ast.Attribute{Type: ast.TypeRef("ipaddr")},
					"dec": ast.Attribute{Type: ast.TypeRef("decimal")},
					"dt":  ast.Attribute{Type: ast.TypeRef("datetime")},
					"dur": ast.Attribute{Type: ast.TypeRef("duration")},
					"set": ast.Attribute{Type: ast.Set(ast.TypeRef("Long"))},
					"rec": ast.Attribute{Type: ast.RecordType{}},
					"ref": ast.Attribute{Type: ast.EntityTypeRef("NS::Foo")},
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	_, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
}

func TestMarshalMultipleEntityTypeRefs(t *testing.T) {
	schema := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				ParentTypes: []ast.EntityTypeRef{"Admin", "Group"},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, len(schema2.Entities["User"].ParentTypes), 2)
}

func TestMarshalMultipleActionParents(t *testing.T) {
	schema := &ast.Schema{
		Actions: ast.Actions{
			"view": ast.Action{
				Parents: []ast.ParentRef{
					ast.ParentRefFromID("read"),
					ast.ParentRefFromID("write"),
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, len(schema2.Actions["view"].Parents), 2)
}

func TestMarshalQuotedAttrName(t *testing.T) {
	schema := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: ast.RecordType{
					"has space": ast.Attribute{Type: ast.TypeRef("String")},
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Entities["User"].Shape["has space"]
	testutil.Equals(t, ok, true)
}

func TestMarshalReservedKeywordAttrName(t *testing.T) {
	schema := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: ast.RecordType{
					"true": ast.Attribute{Type: ast.TypeRef("String")},
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	// Verify the marshaled output quotes the reserved keyword
	testutil.Equals(t, strings.Contains(string(out), `"true"`), true)
	// Verify it round-trips
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Entities["User"].Shape["true"]
	testutil.Equals(t, ok, true)
}

func TestMarshalReservedKeywordActionName(t *testing.T) {
	schema := &ast.Schema{
		Actions: ast.Actions{
			"true": ast.Action{},
		},
	}
	out := parser.MarshalSchema(schema)
	// Verify the marshaled output quotes the reserved keyword
	testutil.Equals(t, strings.Contains(string(out), `"true"`), true)
	// Verify it round-trips
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Actions["true"]
	testutil.Equals(t, ok, true)
}

func TestMarshalAnnotationNoValue(t *testing.T) {
	schema := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Annotations: ast.Annotations{
					"deprecated": "",
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Entities["User"].Annotations["deprecated"]
	testutil.Equals(t, ok, true)
}

func TestMarshalNamespace(t *testing.T) {
	schema := &ast.Schema{
		Namespaces: ast.Namespaces{
			"Foo": ast.Namespace{
				Entities: ast.Entities{
					"Bar": ast.Entity{},
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Namespaces["Foo"].Entities["Bar"]
	testutil.Equals(t, ok, true)
}

func TestMarshalNamespaceWithAnnotations(t *testing.T) {
	schema := &ast.Schema{
		Namespaces: ast.Namespaces{
			"Foo": ast.Namespace{
				Annotations: ast.Annotations{
					"doc": "foo ns",
				},
				Entities: ast.Entities{
					"Bar": ast.Entity{},
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, schema2.Namespaces["Foo"].Annotations["doc"], types.String("foo ns"))
}

func TestMarshalActionBareParent(t *testing.T) {
	schema := &ast.Schema{
		Actions: ast.Actions{
			"view": ast.Action{
				Parents: []ast.ParentRef{
					ast.ParentRefFromID("readOnly"),
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, schema2.Actions["view"].Parents[0], ast.ParentRefFromID("readOnly"))
}

func TestMarshalPrimitiveTypes(t *testing.T) {
	schema := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{
				Shape: ast.RecordType{
					"s": ast.Attribute{Type: ast.StringType{}},
					"l": ast.Attribute{Type: ast.LongType{}},
					"b": ast.Attribute{Type: ast.BoolType{}},
					"e": ast.Attribute{Type: ast.ExtensionType("ipaddr")},
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, len(schema2.Entities["User"].Shape), 4)
}

func TestMarshalEmptyNamespaceMap(t *testing.T) {
	schema := &ast.Schema{
		Namespaces: ast.Namespaces{},
	}
	out := parser.MarshalSchema(schema)
	testutil.Equals(t, string(out), "")
}

func TestMarshalEmptyCommonTypes(t *testing.T) {
	schema := &ast.Schema{
		CommonTypes: ast.CommonTypes{},
	}
	out := parser.MarshalSchema(schema)
	testutil.Equals(t, string(out), "")
}

func TestMarshalNamespaceCommonTypes(t *testing.T) {
	schema := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				CommonTypes: ast.CommonTypes{
					"Ctx": ast.CommonType{
						Type: ast.RecordType{},
					},
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Namespaces["NS"].CommonTypes["Ctx"]
	testutil.Equals(t, ok, true)
}

func TestMarshalBareAndNamespaced(t *testing.T) {
	schema := &ast.Schema{
		Entities: ast.Entities{
			"Global": ast.Entity{},
		},
		Namespaces: ast.Namespaces{
			"Foo": ast.Namespace{
				Entities: ast.Entities{
					"Bar": ast.Entity{},
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	schema2, err := parser.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Entities["Global"]
	testutil.Equals(t, ok, true)
	_, ok = schema2.Namespaces["Foo"].Entities["Bar"]
	testutil.Equals(t, ok, true)
}

func TestMarshalMultipleDecls(t *testing.T) {
	s := &ast.Schema{
		Namespaces: ast.Namespaces{
			"NS1": ast.Namespace{},
			"NS2": ast.Namespace{
				CommonTypes: ast.CommonTypes{
					"A": ast.CommonType{Type: ast.StringType{}},
					"B": ast.CommonType{Type: ast.LongType{}},
				},
				Enums: ast.Enums{
					"Color": ast.Enum{Values: []types.String{"red"}},
					"Size":  ast.Enum{Values: []types.String{"small"}},
				},
			},
		},
	}
	result := parser.MarshalSchema(s)
	testutil.Equals(t, len(result) > 0, true)
	_, err := parser.ParseSchema("", result)
	testutil.OK(t, err)
}

func TestMarshalNamespaceQualifiedKeyRoundTripBreaks(t *testing.T) {
	// Entity keys must be bare Idents, not namespace-qualified.
	// A qualified key like "Foo::Bar" in namespace "Foo" marshals as
	// "entity Foo::Bar" inside the namespace block. Re-parsing fails
	// because "::" is not valid in a bare entity declaration name.
	schema := &ast.Schema{
		Namespaces: ast.Namespaces{
			"Foo": ast.Namespace{
				Entities: ast.Entities{
					"Foo::Bar": ast.Entity{},
				},
			},
		},
	}
	out := parser.MarshalSchema(schema)
	_, err := parser.ParseSchema("", out)
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

		// Reserved Cedar keywords as identifiers
		{"reserved identifier entity name true", `entity true;`},
		{"reserved identifier entity name false", `entity false;`},
		{"reserved identifier entity name if", `entity if;`},
		{"reserved identifier entity name then", `entity then;`},
		{"reserved identifier entity name else", `entity else;`},
		{"reserved identifier entity name in", `entity in;`},
		{"reserved identifier entity name is", `entity is;`},
		{"reserved identifier entity name like", `entity like;`},
		{"reserved identifier entity name has", `entity has;`},
		{"reserved identifier in namespace path", `namespace true {}`},
		{"reserved identifier in type reference", `entity Foo { x: true };`},
		{"reserved identifier type name", `type true = String;`},
		{"reserved identifier action name", `action true;`},
		{"reserved identifier attr name", `entity Foo { true: String };`},
		{"reserved identifier in path component", `entity Foo in [true::Bar];`},
		{"reserved identifier in path after double colon", `entity Foo in [Bar::true];`},
		{"reserved identifier second entity name", `entity A, true {};`},
		{"reserved identifier in action parent path", `action view in [true::Action::"foo"];`},
		{"reserved identifier in action parent path after double colon", `action view in [Foo::true::"bar"];`},

		// __cedar as definition name
		{"__cedar as namespace name", `namespace __cedar {}`},
		{"__cedar in namespace path", `namespace Foo::__cedar {}`},
		{"__cedar as entity name", `entity __cedar;`},
		{"__cedar as second entity name", `entity A, __cedar;`},
		{"__cedar as enum name", `entity __cedar enum ["x"];`},
		{"__cedar as type name", `type __cedar = String;`},

		// Duplicate annotations
		{"duplicate annotation key", `@doc("a") @doc("b") entity Foo;`},
		{"duplicate annotation key no value", `@deprecated @deprecated entity Foo;`},

		// AppliesTo semantic errors
		{"duplicate principal in appliesTo", `action view appliesTo { principal: A, principal: B };`},
		{"duplicate resource in appliesTo", `action view appliesTo { resource: A, principal: C, resource: B };`},
		{"duplicate context in appliesTo", `action view appliesTo { principal: A, resource: B, context: {}, context: {} };`},
		{"empty principal list", `action view appliesTo { principal: [], resource: Photo };`},
		{"empty resource list", `action view appliesTo { principal: Photo, resource: [] };`},
		{"missing principal in appliesTo", `action view appliesTo { resource: Photo };`},
		{"missing resource in appliesTo", `action view appliesTo { principal: Photo };`},
		{"empty appliesTo", `action view appliesTo {};`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ParseSchema("", []byte(tt.input))
			testutil.Error(t, err)
		})
	}
}

func TestParseDuplicateDeclarations(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"duplicate entity", `entity User; entity User;`},
		{"duplicate entity in namespace", `namespace Foo { entity User; entity User; }`},
		{"duplicate enum", `entity Status enum ["a"]; entity Status enum ["b"];`},
		{"duplicate action", `action view; action view;`},
		{"duplicate action in namespace", `namespace Foo { action view; action view; }`},
		{"duplicate common type", `type Ctx = { x: Long }; type Ctx = { y: Long };`},
		{"duplicate namespace", "namespace Foo { entity A; }\nnamespace Foo { entity B; }"},
		{"entity conflicts with enum", `entity User; entity User enum ["a"];`},
		{"enum conflicts with entity", `entity User enum ["a"]; entity User;`},
		{"duplicate multi-name entity", `entity A, A { name: String };`},
		{"duplicate multi-name action", `action read, read;`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ParseSchema("", []byte(tt.input))
			testutil.Error(t, err)
		})
	}
}
