package ast

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestIsNode(t *testing.T) {
	t.Parallel()

	NodeTypeOr{}.isNode()
	NodeTypeAnd{}.isNode()
	NodeTypeLessThan{}.isNode()
	NodeTypeLessThanOrEqual{}.isNode()
	NodeTypeGreaterThan{}.isNode()
	NodeTypeGreaterThanOrEqual{}.isNode()
	NodeTypeNotEquals{}.isNode()
	NodeTypeEquals{}.isNode()
	NodeTypeIn{}.isNode()
	NodeTypeHas{}.isNode()
	NodeTypeHasTag{}.isNode()
	NodeTypeSub{}.isNode()
	NodeTypeAdd{}.isNode()
	NodeTypeMult{}.isNode()
	NodeTypeNegate{}.isNode()
	NodeTypeNot{}.isNode()
	NodeTypeAccess{}.isNode()
	NodeTypeGetTag{}.isNode()
	NodeTypeContains{}.isNode()
	NodeTypeContainsAll{}.isNode()
	NodeTypeContainsAny{}.isNode()
	NodeTypeIsEmpty{}.isNode()
	NodeTypeExtensionCall{}.isNode()
	NodeTypeIfThenElse{}.isNode()
	NodeTypeIs{}.isNode()
	NodeTypeLike{}.isNode()
	NodeTypeRecord{}.isNode()
	NodeTypeSet{}.isNode()
	NodeTypeVariable{}.isNode()
	NodeValue{}.isNode()
}

func TestAsNode(t *testing.T) {
	t.Parallel()
	n := NewNode(NodeValue{Value: types.Long(42)})
	v := n.AsIsNode()
	testutil.Equals(t, v, (IsNode)(NodeValue{Value: types.Long(42)}))
}

func TestIsScope(t *testing.T) {
	t.Parallel()
	ScopeTypeAll{}.isScope()
	ScopeTypeAll{}.isPrincipalScope()
	ScopeTypeAll{}.isActionScope()
	ScopeTypeAll{}.isResourceScope()

	ScopeTypeEq{}.isScope()
	ScopeTypeEq{}.isPrincipalScope()
	ScopeTypeEq{}.isActionScope()
	ScopeTypeEq{}.isResourceScope()

	ScopeTypeIn{}.isScope()
	ScopeTypeIn{}.isPrincipalScope()
	ScopeTypeIn{}.isActionScope()
	ScopeTypeIn{}.isResourceScope()

	ScopeTypeInSet{}.isScope()
	ScopeTypeInSet{}.isActionScope()

	ScopeTypeIs{}.isScope()
	ScopeTypeIs{}.isPrincipalScope()
	ScopeTypeIs{}.isResourceScope()

	ScopeTypeIsIn{}.isScope()
	ScopeTypeIsIn{}.isPrincipalScope()
	ScopeTypeIsIn{}.isResourceScope()
}

func TestStripNodes(t *testing.T) {
	t.Parallel()
	t.Run("preserveNil", func(t *testing.T) {
		t.Parallel()
		out := stripNodes(nil)
		testutil.Equals(t, out, nil)
	})
	t.Run("preserveNonNil", func(t *testing.T) {
		t.Parallel()
		out := stripNodes([]Node{})
		testutil.Equals(t, out, []IsNode{})
	})
}
