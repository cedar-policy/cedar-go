package resolved

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

func TestResolveTypeDefault(t *testing.T) {
	// Exercise the default branch of resolveType (unreachable with real AST types).
	r := &resolverState{
		entityTypes: make(map[types.EntityType]bool),
		enumTypes:   make(map[types.EntityType]bool),
		commonTypes: make(map[types.Path]ast.IsType),
	}
	testutil.Panic(t, func() {
		_, _ = r.resolveType("", nil)
	})
}

func TestResolveTypePath(t *testing.T) {
	r := &resolverState{
		commonTypes: map[types.Path]ast.IsType{
			"NS::A": ast.StringType{},
			"B":     ast.LongType{},
		},
		entityTypes: make(map[types.EntityType]bool),
		enumTypes:   make(map[types.EntityType]bool),
	}

	// __cedar:: prefix returns path unchanged
	p := r.resolveTypeRefPath("NS", "__cedar::String")
	testutil.Equals(t, p, types.Path("__cedar::String"))

	// Already qualified (contains ::) returns path unchanged
	p = r.resolveTypeRefPath("NS", "Other::Foo")
	testutil.Equals(t, p, types.Path("Other::Foo"))

	// Unqualified in namespace resolves to NS::A
	p = r.resolveTypeRefPath("NS", "A")
	testutil.Equals(t, p, types.Path("NS::A"))
}

func TestResolveActionParentRef(t *testing.T) {
	// Exercise both branches of resolveActionParentRef.

	// Bare reference
	uid := resolveActionParentRef("NS", ast.ParentRef{ID: "view"})
	testutil.Equals(t, uid, types.NewEntityUID("NS::Action", "view"))

	// Typed reference
	uid = resolveActionParentRef("NS", ast.ParentRef{Type: "Other::Action", ID: "edit"})
	testutil.Equals(t, uid, types.NewEntityUID("Other::Action", "edit"))
}

func TestCollectTypeRefsDefault(t *testing.T) {
	// Exercise the non-container type branch
	refs := collectTypeRefs(ast.StringType{})
	testutil.Equals(t, len(refs), 0)

	// Exercise the impossible to hit branch
	testutil.Panic(t, func() {
		collectTypeRefs(nil)
	})
}

func TestDetectCommonTypeCyclesBuiltinRef(t *testing.T) {
	// Verify cycle detection works correctly with __cedar:: refs.
	r := &resolverState{
		commonTypes: map[types.Path]ast.IsType{
			"NS::A": ast.TypeRef("__cedar::String"),
		},
		entityTypes: make(map[types.EntityType]bool),
		enumTypes:   make(map[types.EntityType]bool),
	}
	err := r.detectCommonTypeCycles()
	testutil.OK(t, err)
}
