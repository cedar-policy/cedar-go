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

	user := ns.Entities["PhotoApp::User"]
	testutil.Equals(t, user.MemberOf, []ast2.EntityTypeRef{"Group"})
	testutil.Equals(t, user.Shape != nil, true)
	testutil.Equals(t, len(*user.Shape), 2)
	testutil.Equals(t, (*user.Shape)["name"].Type, ast2.IsType(ast2.TypeRef("String")))
	testutil.Equals(t, (*user.Shape)["name"].Optional, false)
	testutil.Equals(t, (*user.Shape)["age"].Type, ast2.IsType(ast2.TypeRef("Long")))
	testutil.Equals(t, (*user.Shape)["age"].Optional, true)

	group := ns.Entities["PhotoApp::Group"]
	testutil.Equals(t, group.Shape == nil, true)
	testutil.Equals(t, len(group.MemberOf), 0)

	photo := ns.Entities["PhotoApp::Photo"]
	testutil.Equals(t, photo.Shape != nil, true)
	testutil.Equals(t, photo.Tags, ast2.IsType(ast2.TypeRef("String")))

	viewPhoto := ns.Actions["viewPhoto"]
	testutil.Equals(t, viewPhoto.AppliesTo != nil, true)
	testutil.Equals(t, viewPhoto.AppliesTo.Principals, []ast2.EntityTypeRef{"User"})
	testutil.Equals(t, viewPhoto.AppliesTo.Resources, []ast2.EntityTypeRef{"Photo"})

	createPhoto := ns.Actions["createPhoto"]
	testutil.Equals(t, len(createPhoto.MemberOf), 1)
	testutil.Equals(t, createPhoto.MemberOf[0], ast2.ParentRefFromID("viewPhoto"))
}

func TestParseMultiNameEntity(t *testing.T) {
	src := `entity A, B, C { name: String };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Entities), 3)
	for _, name := range []types.EntityType{"A", "B", "C"} {
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
	testutil.Equals(t, len(view.MemberOf), 1)
	testutil.Equals(t, view.MemberOf[0], ast2.NewParentRef("MyApp::Action", "readOnly"))
}

func TestParseActionBareParent(t *testing.T) {
	src := `action view in readOnly;`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	view := schema.Actions["view"]
	testutil.Equals(t, len(view.MemberOf), 1)
	testutil.Equals(t, view.MemberOf[0], ast2.ParentRefFromID("readOnly"))
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
	testutil.Equals(t, user.MemberOf, []ast2.EntityTypeRef{"Admin", "Group"})
}

func TestParseEntityInSingle(t *testing.T) {
	src := `entity User in Admin;`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	user := schema.Entities["User"]
	testutil.Equals(t, user.MemberOf, []ast2.EntityTypeRef{"Admin"})
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
	_, ok = ns.Entities["Foo::Bar"]
	testutil.Equals(t, ok, true)
}

func TestParseNestedNamespacePath(t *testing.T) {
	src := `namespace Foo::Bar { entity Baz; }`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	ns := schema.Namespaces["Foo::Bar"]
	_, ok := ns.Entities["Foo::Bar::Baz"]
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
	testutil.Equals(t, user.MemberOf, []ast2.EntityTypeRef{"NS::Group"})
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
	testutil.Equals(t, len(schema.Entities["User"].MemberOf), 2)
}

func TestParseTrailingCommaInAppliesTo(t *testing.T) {
	src := `action view appliesTo { principal: User, resource: Photo, };`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	testutil.Equals(t, len(schema.Actions["view"].AppliesTo.Principals), 1)
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

func TestParseErrorUnterminatedNamespace(t *testing.T) {
	src := `namespace Foo { entity Bar;`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorInvalidToken(t *testing.T) {
	src := `entity User $ {};`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
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
				MemberOf: []ast2.ParentRef{
					ast2.NewParentRef("NS::Action", "readOnly"),
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, schema2.Actions["view"].MemberOf[0], ast2.NewParentRef("NS::Action", "readOnly"))
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

func TestParseActionParentStringLiteral(t *testing.T) {
	src := `action view in ["readOnly"];`
	schema, err := parser2.ParseSchema("", []byte(src))
	testutil.OK(t, err)
	view := schema.Actions["view"]
	testutil.Equals(t, view.MemberOf[0], ast2.ParentRefFromID("readOnly"))
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
	_, ok := schema.Namespaces["A"].Entities["A::Foo"]
	testutil.Equals(t, ok, true)
	_, ok = schema.Namespaces["B"].Entities["B::Bar"]
	testutil.Equals(t, ok, true)
}

func TestParseErrorUnterminatedString(t *testing.T) {
	src := `entity User enum ["unterminated;`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorUnterminatedBlockComment(t *testing.T) {
	src := `/* unterminated`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorMissingSemicolon(t *testing.T) {
	src := `entity User`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorBadDeclaration(t *testing.T) {
	src := `foobar;`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorNonDeclKeyword(t *testing.T) {
	src := `namespace Foo { bogus; }`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

// Error path coverage tests

func TestParseErrorBadAnnotationName(t *testing.T) {
	src := `@ "bad" entity User;`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorBadAnnotationValue(t *testing.T) {
	src := `@doc(42) entity User;`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorEntityNotIdent(t *testing.T) {
	src := `entity "bad";`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorTypeDeclNotIdent(t *testing.T) {
	src := `type 42 = Long;`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorEnumNotString(t *testing.T) {
	src := `entity Foo enum [42];`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorEnumBadSeparator(t *testing.T) {
	src := `entity Foo enum ["a" "b"];`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorApplyToNotKeyword(t *testing.T) {
	src := `action view appliesTo { foo: User };`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorApplyToNotIdent(t *testing.T) {
	src := `action view appliesTo { 42: User };`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorApplyToEOF(t *testing.T) {
	src := `action view appliesTo {`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorRecordEOF(t *testing.T) {
	src := `entity User {`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorPathBadAfterDoubleColon(t *testing.T) {
	src := `entity User in Foo::42;`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorPathNotIdent(t *testing.T) {
	src := `entity User in 42;`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorEntityTypeListBadSep(t *testing.T) {
	src := `entity User in [Foo Bar];`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorActionParentListBadSep(t *testing.T) {
	src := `action view in [foo bar];`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorNameNotIdentOrString(t *testing.T) {
	src := `action 42;`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorDeclNotIdent(t *testing.T) {
	src := `42;`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorRecordBadAttrName(t *testing.T) {
	src := `entity User { 42: Long };`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorTypeDeclMissingEquals(t *testing.T) {
	src := `type Foo Long;`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorPathForRefBadAfterDoubleColon(t *testing.T) {
	src := `action view in [Foo::42];`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

// Marshal coverage: types that weren't covered

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
				MemberOf: []ast2.EntityTypeRef{"Admin", "Group"},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, len(schema2.Entities["User"].MemberOf), 2)
}

func TestMarshalMultipleActionParents(t *testing.T) {
	schema := &ast2.Schema{
		Actions: ast2.Actions{
			"view": ast2.Action{
				MemberOf: []ast2.ParentRef{
					ast2.ParentRefFromID("read"),
					ast2.ParentRefFromID("write"),
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, len(schema2.Actions["view"].MemberOf), 2)
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
					"Foo::Bar": ast2.Entity{},
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Namespaces["Foo"].Entities["Foo::Bar"]
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
					"Foo::Bar": ast2.Entity{},
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
				MemberOf: []ast2.ParentRef{
					ast2.ParentRefFromID("readOnly"),
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	testutil.Equals(t, schema2.Actions["view"].MemberOf[0], ast2.ParentRefFromID("readOnly"))
}

func TestParseErrorActionParentRefNotIdent(t *testing.T) {
	src := `action view in [42];`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorAppliesToMissingBrace(t *testing.T) {
	src := `action view appliesTo principal: User;`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
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

func TestParseErrorApplyToMissingColon(t *testing.T) {
	src := `action view appliesTo { principal User };`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorTypeNotIdentOrBrace(t *testing.T) {
	src := `entity User { name: 42 };`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorNamespaceEOF(t *testing.T) {
	src := `namespace Foo {`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorRecordMissingColon(t *testing.T) {
	src := `entity User { name String };`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorIdentsAfterComma(t *testing.T) {
	src := `entity A, 42 {};`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestParseErrorNamesAfterComma(t *testing.T) {
	src := `action foo, 42;`
	_, err := parser2.ParseSchema("", []byte(src))
	testutil.Error(t, err)
}

func TestMarshalBareAndNamespaced(t *testing.T) {
	schema := &ast2.Schema{
		Entities: ast2.Entities{
			"Global": ast2.Entity{},
		},
		Namespaces: ast2.Namespaces{
			"Foo": ast2.Namespace{
				Entities: ast2.Entities{
					"Foo::Bar": ast2.Entity{},
				},
			},
		},
	}
	out := parser2.MarshalSchema(schema)
	schema2, err := parser2.ParseSchema("", out)
	testutil.OK(t, err)
	_, ok := schema2.Entities["Global"]
	testutil.Equals(t, ok, true)
	_, ok = schema2.Namespaces["Foo"].Entities["Foo::Bar"]
	testutil.Equals(t, ok, true)
}
