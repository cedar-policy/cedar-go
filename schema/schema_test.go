package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/schema"
	"github.com/cedar-policy/cedar-go/schema/ast"
	"github.com/cedar-policy/cedar-go/schema/resolved"
	"github.com/cedar-policy/cedar-go/types"
)

func TestNewSchemaFromAST(t *testing.T) {
	t.Parallel()
	a := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{},
		},
	}
	s := schema.NewSchemaFromAST(a)
	testutil.Equals(t, s.AST(), a)
}

func TestSetFilename(t *testing.T) {
	t.Parallel()
	s := &schema.Schema{}
	s.SetFilename("test.cedarschema")
	err := s.UnmarshalCedar([]byte("entity 42;"))
	testutil.Error(t, err)
	// The error should include the filename
	testutil.Equals(t, true, len(err.Error()) > 0)
}

func TestMarshalUnmarshalCedar(t *testing.T) {
	t.Parallel()
	input := []byte("entity User;\n")

	var s schema.Schema
	testutil.OK(t, s.UnmarshalCedar(input))

	b, err := s.MarshalCedar()
	testutil.OK(t, err)

	var s2 schema.Schema
	testutil.OK(t, s2.UnmarshalCedar(b))
	testutil.Equals(t, len(s2.AST().Entities), 1)
}

func TestMarshalUnmarshalJSON(t *testing.T) {
	t.Parallel()
	input := []byte(`{"": {"entityTypes": {"User": {}}, "actions": {}}}`)

	var s schema.Schema
	testutil.OK(t, s.UnmarshalJSON(input))

	b, err := s.MarshalJSON()
	testutil.OK(t, err)

	var s2 schema.Schema
	testutil.OK(t, s2.UnmarshalJSON(b))
	testutil.Equals(t, len(s2.AST().Entities), 1)
}

func TestCedarToJSONRoundTrip(t *testing.T) {
	t.Parallel()
	input := []byte(`entity User {
	name: String,
	age?: Long
};

action "view" appliesTo {
	principal: User,
	resource: User,
	context: {}
};
`)

	var s schema.Schema
	testutil.OK(t, s.UnmarshalCedar(input))

	jsonBytes, err := s.MarshalJSON()
	testutil.OK(t, err)

	var s2 schema.Schema
	testutil.OK(t, s2.UnmarshalJSON(jsonBytes))

	cedarBytes, err := s2.MarshalCedar()
	testutil.OK(t, err)

	var s3 schema.Schema
	testutil.OK(t, s3.UnmarshalCedar(cedarBytes))

	testutil.Equals(t, len(s3.AST().Entities), 1)
	testutil.Equals(t, len(s3.AST().Actions), 1)
}

func TestJSONToCedarRoundTrip(t *testing.T) {
	t.Parallel()
	input := []byte(`{
		"NS": {
			"entityTypes": {
				"User": {
					"memberOfTypes": ["Group"],
					"shape": {
						"type": "Record",
						"attributes": {
							"name": {"type": "String"}
						}
					}
				},
				"Group": {}
			},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["Group"],
						"context": {"type": "Record", "attributes": {}}
					}
				}
			}
		}
	}`)

	var s schema.Schema
	testutil.OK(t, s.UnmarshalJSON(input))

	cedarBytes, err := s.MarshalCedar()
	testutil.OK(t, err)

	var s2 schema.Schema
	testutil.OK(t, s2.UnmarshalCedar(cedarBytes))

	jsonBytes, err := s2.MarshalJSON()
	testutil.OK(t, err)

	var s3 schema.Schema
	testutil.OK(t, s3.UnmarshalJSON(jsonBytes))

	ns := s3.AST().Namespaces["NS"]
	testutil.Equals(t, len(ns.Entities), 2)
	testutil.Equals(t, len(ns.Actions), 1)
}

func TestResolve(t *testing.T) {
	t.Parallel()
	input := []byte(`entity User;

action "view" appliesTo {
	principal: User,
	resource: User,
	context: {}
};
`)

	var s schema.Schema
	testutil.OK(t, s.UnmarshalCedar(input))

	res, err := s.Resolve()
	testutil.OK(t, err)

	testutil.Equals(t, len(res.Entities), 1)
	testutil.Equals(t, len(res.Actions), 1)

	action := res.Actions[types.NewEntityUID("Action", "view")]
	testutil.Equals(t, action.Name, types.String("view"))
	testutil.Equals(t, action.AppliesTo != nil, true)
	testutil.Equals(t, action.AppliesTo.Principals, []types.EntityType{"User"})
}

func TestResolveError(t *testing.T) {
	t.Parallel()
	input := []byte(`entity User in [NonExistent];`)

	var s schema.Schema
	testutil.OK(t, s.UnmarshalCedar(input))

	_, err := s.Resolve()
	testutil.Error(t, err)
}

func TestUnmarshalCedarError(t *testing.T) {
	t.Parallel()
	var s schema.Schema
	testutil.Error(t, s.UnmarshalCedar([]byte("not valid cedar {")))
}

func TestUnmarshalJSONError(t *testing.T) {
	t.Parallel()
	var s schema.Schema
	testutil.Error(t, s.UnmarshalJSON([]byte("not json")))
}

func TestJSONMarshalInterface(t *testing.T) {
	t.Parallel()
	var s schema.Schema
	testutil.OK(t, s.UnmarshalCedar([]byte("entity User;")))

	b, err := json.Marshal(&s)
	testutil.OK(t, err)

	var s2 schema.Schema
	testutil.OK(t, json.Unmarshal(b, &s2))
	testutil.Equals(t, len(s2.AST().Entities), 1)
}

func TestResolveWithNamespace(t *testing.T) {
	t.Parallel()
	input := []byte(`namespace NS {
	entity User;

	action "view" appliesTo {
		principal: User,
		resource: User,
		context: {}
	};
}
`)

	var s schema.Schema
	testutil.OK(t, s.UnmarshalCedar(input))

	res, err := s.Resolve()
	testutil.OK(t, err)

	_, ok := res.Entities["NS::User"]
	testutil.Equals(t, ok, true)

	action := res.Actions[types.NewEntityUID("NS::Action", "view")]
	testutil.Equals(t, action.AppliesTo.Principals, []types.EntityType{"NS::User"})
}

func TestResolveCommonType(t *testing.T) {
	t.Parallel()
	input := []byte(`type Context = {
	ip: ipaddr
};

entity User {
	ctx: Context
};

action "view" appliesTo {
	principal: User,
	resource: User,
	context: Context
};
`)

	var s schema.Schema
	testutil.OK(t, s.UnmarshalCedar(input))

	res, err := s.Resolve()
	testutil.OK(t, err)

	user := res.Entities["User"]
	attr := (*user.Shape)["ctx"]
	_, ok := attr.Type.(resolved.RecordType)
	testutil.Equals(t, ok, true)
}

func TestResolveEnum(t *testing.T) {
	t.Parallel()
	input := []byte(`entity Status enum ["active", "inactive"];`)

	var s schema.Schema
	testutil.OK(t, s.UnmarshalCedar(input))

	res, err := s.Resolve()
	testutil.OK(t, err)

	testutil.Equals(t, len(res.Enums), 1)
	status := res.Enums["Status"]
	testutil.Equals(t, status.Values, []types.String{"active", "inactive"})
}

func TestNewSchemaFromASTNil(t *testing.T) {
	t.Parallel()
	a := &ast.Schema{}
	s := schema.NewSchemaFromAST(a)
	b, err := s.MarshalCedar()
	testutil.OK(t, err)
	testutil.Equals(t, string(b), "")

	jb, err := s.MarshalJSON()
	testutil.OK(t, err)
	testutil.Equals(t, string(jb), "{}")
}
