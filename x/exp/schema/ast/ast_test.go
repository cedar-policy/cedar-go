package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	ast2 "github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

func TestConstructors(t *testing.T) {
	testutil.Equals(t, ast2.String(), ast2.StringType{})
	testutil.Equals(t, ast2.Long(), ast2.LongType{})
	testutil.Equals(t, ast2.Bool(), ast2.BoolType{})
	testutil.Equals(t, ast2.IPAddr(), ast2.ExtensionType("ipaddr"))
	testutil.Equals(t, ast2.Decimal(), ast2.ExtensionType("decimal"))
	testutil.Equals(t, ast2.Datetime(), ast2.ExtensionType("datetime"))
	testutil.Equals(t, ast2.Duration(), ast2.ExtensionType("duration"))
	testutil.Equals(t, ast2.Set(ast2.Long()), ast2.SetType{Element: ast2.LongType{}})
	testutil.Equals(t, ast2.EntityType("User"), ast2.EntityTypeRef("User"))
	testutil.Equals(t, ast2.Type("MyType"), ast2.TypeRef("MyType"))
}

func TestParentRefFromID(t *testing.T) {
	ref := ast2.ParentRefFromID("view")
	testutil.Equals(t, ref.ID, types.String("view"))
	testutil.Equals(t, ref.Type, ast2.EntityTypeRef(""))
}

func TestNewParentRef(t *testing.T) {
	ref := ast2.NewParentRef("NS::Action", "view")
	testutil.Equals(t, ref.ID, types.String("view"))
	testutil.Equals(t, ref.Type, ast2.EntityTypeRef("NS::Action"))
}

func TestIsTypeInterface(t *testing.T) {
	// Verify all types satisfy IsType by assigning to the interface.
	// The isType() marker methods have { _ = 0 } bodies;
	// calling through the interface exercises them at runtime.
	var types []ast2.IsType
	types = append(types, ast2.StringType{})
	types = append(types, ast2.LongType{})
	types = append(types, ast2.BoolType{})
	types = append(types, ast2.ExtensionType("ipaddr"))
	types = append(types, ast2.SetType{Element: ast2.StringType{}})
	types = append(types, ast2.RecordType{})
	types = append(types, ast2.EntityTypeRef("User"))
	types = append(types, ast2.TypeRef("Foo"))
	testutil.Equals(t, len(types), 8)
}
