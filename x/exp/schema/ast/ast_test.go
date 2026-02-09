package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

func TestConstructors(t *testing.T) {
	testutil.Equals(t, ast.String(), ast.StringType{})
	testutil.Equals(t, ast.Long(), ast.LongType{})
	testutil.Equals(t, ast.Bool(), ast.BoolType{})
	testutil.Equals(t, ast.IPAddr(), ast.ExtensionType("ipaddr"))
	testutil.Equals(t, ast.Decimal(), ast.ExtensionType("decimal"))
	testutil.Equals(t, ast.Datetime(), ast.ExtensionType("datetime"))
	testutil.Equals(t, ast.Duration(), ast.ExtensionType("duration"))
	testutil.Equals(t, ast.Set(ast.Long()), ast.SetType{Element: ast.LongType{}})
	testutil.Equals(t, ast.EntityType("User"), ast.EntityTypeRef("User"))
	testutil.Equals(t, ast.Type("MyType"), ast.TypeRef("MyType"))
}

func TestParentRefFromID(t *testing.T) {
	ref := ast.ParentRefFromID("view")
	testutil.Equals(t, ref.ID, types.String("view"))
	testutil.Equals(t, ref.Type, ast.EntityTypeRef(""))
}

func TestNewParentRef(t *testing.T) {
	ref := ast.NewParentRef("NS::Action", "view")
	testutil.Equals(t, ref.ID, types.String("view"))
	testutil.Equals(t, ref.Type, ast.EntityTypeRef("NS::Action"))
}

func TestIsTypeInterface(t *testing.T) {
	// Verify all types satisfy IsType by assigning to the interface.
	// The isType() marker methods have { _ = 0 } bodies;
	// calling through the interface exercises them at runtime.
	var types []ast.IsType
	types = append(types, ast.StringType{})
	types = append(types, ast.LongType{})
	types = append(types, ast.BoolType{})
	types = append(types, ast.ExtensionType("ipaddr"))
	types = append(types, ast.SetType{Element: ast.StringType{}})
	types = append(types, ast.RecordType{})
	types = append(types, ast.EntityTypeRef("User"))
	types = append(types, ast.TypeRef("Foo"))
	testutil.Equals(t, len(types), 8)
}
