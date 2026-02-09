package resolved

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	ast2 "github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

func TestResolveTypeDefault(t *testing.T) {
	// Exercise the default branch of resolveType (unreachable with real AST types).
	r := &resolverState{
		entityTypes: make(map[types.EntityType]bool),
		enumTypes:   make(map[types.EntityType]bool),
		commonTypes: make(map[types.Path]ast2.IsType),
	}
	_, err := r.resolveType("", nil)
	testutil.Error(t, err)
}

func TestResolveTypePath(t *testing.T) {
	r := &resolverState{
		commonTypes: map[types.Path]ast2.IsType{
			"NS::A": ast2.StringType{},
			"B":     ast2.LongType{},
		},
		entityTypes: make(map[types.EntityType]bool),
		enumTypes:   make(map[types.EntityType]bool),
	}

	// __cedar:: prefix returns path unchanged
	p := r.resolveTypePath("NS", "__cedar::String")
	testutil.Equals(t, p, types.Path("__cedar::String"))

	// Already qualified (contains ::) returns path unchanged
	p = r.resolveTypePath("NS", "Other::Foo")
	testutil.Equals(t, p, types.Path("Other::Foo"))

	// Unqualified in namespace resolves to NS::A
	p = r.resolveTypePath("NS", "A")
	testutil.Equals(t, p, types.Path("NS::A"))
}

func TestResolveActionParentRef(t *testing.T) {
	// Exercise both branches of resolveActionParentRef.

	// Bare reference
	uid := resolveActionParentRef("NS", ast2.ParentRef{ID: "view"})
	testutil.Equals(t, uid, types.NewEntityUID("NS::Action", "view"))

	// Typed reference
	uid = resolveActionParentRef("NS", ast2.ParentRef{Type: "Other::Action", ID: "edit"})
	testutil.Equals(t, uid, types.NewEntityUID("Other::Action", "edit"))
}

func TestCollectTypeRefsDefault(t *testing.T) {
	// Exercise the default branch (non-TypeRef, non-Set, non-Record).
	refs := collectTypeRefs(ast2.StringType{})
	testutil.Equals(t, len(refs), 0)
}

func TestDetectCommonTypeCyclesBuiltinRef(t *testing.T) {
	// Verify cycle detection works correctly with __cedar:: refs.
	r := &resolverState{
		commonTypes: map[types.Path]ast2.IsType{
			"NS::A": ast2.TypeRef("__cedar::String"),
		},
		entityTypes: make(map[types.EntityType]bool),
		enumTypes:   make(map[types.EntityType]bool),
	}
	err := r.detectCommonTypeCycles()
	testutil.OK(t, err)
}
