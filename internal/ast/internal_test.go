package ast

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestIsNode(t *testing.T) {
	t.Parallel()
	ScopeNode{}.isScope()

	StrOpNode{}.isNode()
	BinaryNode{}.isNode()
	NodeTypeIfThenElse{}.isNode()
	NodeTypeLike{}.isNode()
	NodeTypeIs{}.isNode()
	UnaryNode{}.isNode()
	NodeTypeExtensionCall{}.isNode()
	NodeValue{}.isNode()
	NodeTypeRecord{}.isNode()
	NodeTypeSet{}.isNode()
	NodeTypeVariable{}.isNode()

}

func TestAsNode(t *testing.T) {
	t.Parallel()
	n := NewNode(NodeValue{Value: types.Long(42)})
	v := n.AsIsNode()
	testutil.Equals(t, v, (IsNode)(NodeValue{Value: types.Long(42)}))
}
