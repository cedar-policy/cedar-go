package ast

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

// newEqualTestNodes returns one node of each type, all different from each
// other. Each call builds the nodes anew, so that Equal can't succeed by
// comparing a node with itself.
func newEqualTestNodes() []Node {
	l1, l2 := Long(1), Long(2)
	return []Node{
		IfThenElse(l1, l2, l1),
		l1.Or(l2),
		l1.And(l2),
		l1.LessThan(l2),
		l1.LessThanOrEqual(l2),
		l1.GreaterThan(l2),
		l1.GreaterThanOrEqual(l2),
		l1.NotEqual(l2),
		l1.Equal(l2),
		l1.In(l2),
		l1.Has("a"),
		l1.HasTag(l2),
		l1.Like(types.NewPattern("a", types.Wildcard{})),
		l1.Is("T"),
		l1.IsIn("T", l2),
		l1.Subtract(l2),
		l1.Add(l2),
		l1.Multiply(l2),
		Negate(l1),
		Not(l1),
		l1.Access("a"),
		l1.GetTag(l2),
		ExtensionCall("ip", l1),
		l1.Contains(l2),
		l1.ContainsAll(l2),
		l1.ContainsAny(l2),
		l1.IsEmpty(),
		l1,
		Record(Pairs{{Key: "a", Value: l1}}),
		Set(l1),
		Principal(),
	}
}

func TestEqualNodeTypes(t *testing.T) {
	t.Parallel()
	as, bs := newEqualTestNodes(), newEqualTestNodes()
	for i, a := range as {
		for j, b := range bs {
			testutil.Equals(t, Equal(a, b), i == j)
		}
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()
	l1, l2 := Long(1), Long(2)
	tests := []struct {
		name string
		a, b Node
		want bool
	}{
		{"differentChild", l1.Add(l1), l1.Add(l2), false},
		{"differentIfThenElse", IfThenElse(l1, l1, l1), IfThenElse(l1, l1, l2), false},
		{"differentAccess", l1.Access("a"), l1.Access("b"), false},
		{"differentHasTag", Principal().HasTag(String("a")), Principal().HasTag(String("b")), false},
		{"sameHasTagExpression", Principal().HasTag(Context().Access("k")), Principal().HasTag(Context().Access("k")), true},
		{"differentLikePattern", l1.Like(types.NewPattern("a")), l1.Like(types.NewPattern("b")), false},
		{"differentIsType", l1.Is("T"), l1.Is("U"), false},
		{"differentIsInEntity", l1.IsIn("T", l1), l1.IsIn("T", l2), false},
		{"differentExtensionName", ExtensionCall("ip", l1), ExtensionCall("decimal", l1), false},
		{"differentExtensionArgs", ExtensionCall("ip", l1), ExtensionCall("ip", l1, l2), false},
		{"differentValue", l1, l2, false},
		{"differentValueType", l1, String("1"), false},
		{"equalRecordValues",
			Value(types.NewRecord(types.RecordMap{"a": types.Long(1), "b": types.Long(2)})),
			Value(types.NewRecord(types.RecordMap{"b": types.Long(2), "a": types.Long(1)})),
			true},
		{"recordOrder",
			Record(Pairs{{Key: "a", Value: l1}, {Key: "b", Value: l2}}),
			Record(Pairs{{Key: "b", Value: l2}, {Key: "a", Value: l1}}),
			true},
		{"recordLength", Record(Pairs{{Key: "a", Value: l1}}), Record(Pairs{{Key: "a", Value: l1}, {Key: "b", Value: l2}}), false},
		{"recordKey", Record(Pairs{{Key: "a", Value: l1}}), Record(Pairs{{Key: "b", Value: l1}}), false},
		{"recordValue", Record(Pairs{{Key: "a", Value: l1}}), Record(Pairs{{Key: "a", Value: l2}}), false},
		{"setOrder", Set(l1, l2), Set(l2, l1), false},
		{"differentVariable", Principal(), Resource(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, Equal(tt.a, tt.b), tt.want)
		})
	}
}

func TestEqualNilPanics(t *testing.T) {
	t.Parallel()
	testutil.Panic(t, func() {
		Equal(NewNode(nil), NewNode(nil))
	})
}
