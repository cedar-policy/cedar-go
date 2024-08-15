package eval

import (
	"net/netip"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestToEval(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   ast.Node
		out  types.Value
		err  func(testing.TB, error)
	}{
		{
			"access",
			ast.Value(types.Record{"key": types.Long(42)}).Access("key"),
			types.Long(42),
			testutil.OK,
		},
		{
			"has",
			ast.Value(types.Record{"key": types.Long(42)}).Has("key"),
			types.True,
			testutil.OK,
		},
		{
			"like",
			ast.String("test").Like(types.Pattern{}),
			types.False,
			testutil.OK,
		},
		{
			"if",
			ast.If(ast.True(), ast.Long(42), ast.Long(43)),
			types.Long(42),
			testutil.OK,
		},
		{
			"is",
			ast.EntityUID("T", "42").Is("T"),
			types.True,
			testutil.OK,
		},
		{
			"isIn",
			ast.EntityUID("T", "42").IsIn("T", ast.EntityUID("T", "42")),
			types.True,
			testutil.OK,
		},
		{
			"value",
			ast.Long(42),
			types.Long(42),
			testutil.OK,
		},
		{
			"record",
			ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(42)}}),
			types.Record{"key": types.Long(42)},
			testutil.OK,
		},
		{
			"set",
			ast.Set(ast.Long(42)),
			types.Set{types.Long(42)},
			testutil.OK,
		},
		{
			"negate",
			ast.Negate(ast.Long(42)),
			types.Long(-42),
			testutil.OK,
		},
		{
			"not",
			ast.Not(ast.True()),
			types.False,
			testutil.OK,
		},
		{
			"principal",
			ast.Principal(),
			types.NewEntityUID("Actor", "principal"),
			testutil.OK,
		},
		{
			"action",
			ast.Action(),
			types.NewEntityUID("Action", "test"),
			testutil.OK,
		},
		{
			"resource",
			ast.Resource(),
			types.NewEntityUID("Resource", "database"),
			testutil.OK,
		},
		{
			"context",
			ast.Context(),
			types.Record{},
			testutil.OK,
		},
		{
			"in",
			ast.EntityUID("T", "42").In(ast.EntityUID("T", "43")),
			types.False,
			testutil.OK,
		},
		{
			"and",
			ast.True().And(ast.False()),
			types.False,
			testutil.OK,
		},
		{
			"or",
			ast.True().Or(ast.False()),
			types.True,
			testutil.OK,
		},
		{
			"equals",
			ast.Long(42).Equals(ast.Long(43)),
			types.False,
			testutil.OK,
		},
		{
			"notEquals",
			ast.Long(42).NotEquals(ast.Long(43)),
			types.True,
			testutil.OK,
		},
		{
			"greaterThan",
			ast.Long(42).GreaterThan(ast.Long(43)),
			types.False,
			testutil.OK,
		},
		{
			"greaterThanOrEqual",
			ast.Long(42).GreaterThanOrEqual(ast.Long(43)),
			types.False,
			testutil.OK,
		},
		{
			"lessThan",
			ast.Long(42).LessThan(ast.Long(43)),
			types.True,
			testutil.OK,
		},
		{
			"lessThanOrEqual",
			ast.Long(42).LessThanOrEqual(ast.Long(43)),
			types.True,
			testutil.OK,
		},
		{
			"sub",
			ast.Long(42).Minus(ast.Long(2)),
			types.Long(40),
			testutil.OK,
		},
		{
			"add",
			ast.Long(40).Plus(ast.Long(2)),
			types.Long(42),
			testutil.OK,
		},
		{
			"mult",
			ast.Long(6).Times(ast.Long(7)),
			types.Long(42),
			testutil.OK,
		},
		{
			"contains",
			ast.SetDeprecated(types.Set{types.Long(42)}).Contains(ast.Long(42)),
			types.True,
			testutil.OK,
		},
		{
			"containsAll",
			ast.SetDeprecated(types.Set{types.Long(42), types.Long(43), types.Long(44)}).ContainsAll(ast.SetDeprecated(types.Set{types.Long(42), types.Long(43)})),
			types.True,
			testutil.OK,
		},
		{
			"containsAny",
			ast.SetDeprecated(types.Set{types.Long(42), types.Long(43), types.Long(44)}).ContainsAny(ast.SetDeprecated(types.Set{types.Long(1), types.Long(42)})),
			types.True,
			testutil.OK,
		},
		{
			"ip",
			ast.ExtensionCall("ip", ast.String("127.0.0.42/16")),
			types.IPAddr(netip.MustParsePrefix("127.0.0.42/16")),
			testutil.OK,
		},
		{
			"decimal",
			ast.ExtensionCall("decimal", ast.String("42.42")),
			types.Decimal(424200),
			testutil.OK,
		},
		{
			"lessThan",
			ast.ExtensionCall("lessThan", ast.Value(types.Decimal(420000)), ast.Value(types.Decimal(430000))),
			types.True,
			testutil.OK,
		},
		{
			"lessThanOrEqual",
			ast.ExtensionCall("lessThanOrEqual", ast.Value(types.Decimal(420000)), ast.Value(types.Decimal(430000))),
			types.True,
			testutil.OK,
		},
		{
			"greaterThan",
			ast.ExtensionCall("greaterThan", ast.Value(types.Decimal(420000)), ast.Value(types.Decimal(430000))),
			types.False,
			testutil.OK,
		},
		{
			"greaterThanOrEqual",
			ast.ExtensionCall("greaterThanOrEqual", ast.Value(types.Decimal(420000)), ast.Value(types.Decimal(430000))),
			types.False,
			testutil.OK,
		},
		{
			"isIpv4",
			ast.ExtensionCall("isIpv4", ast.IPAddr(netip.MustParsePrefix("127.0.0.42/16"))),
			types.True,
			testutil.OK,
		},
		{
			"isIpv6",
			ast.ExtensionCall("isIpv6", ast.IPAddr(netip.MustParsePrefix("::1/16"))),
			types.True,
			testutil.OK,
		},
		{
			"isLoopback",
			ast.ExtensionCall("isLoopback", ast.IPAddr(netip.MustParsePrefix("127.0.0.1/32"))),
			types.True,
			testutil.OK,
		},
		{
			"isMulticast",
			ast.ExtensionCall("isMulticast", ast.IPAddr(netip.MustParsePrefix("239.255.255.255/32"))),
			types.True,
			testutil.OK,
		},
		{
			"isInRange",
			ast.ExtensionCall("isInRange", ast.IPAddr(netip.MustParsePrefix("127.0.0.42/32")), ast.IPAddr(netip.MustParsePrefix("127.0.0.0/16"))),
			types.True,
			testutil.OK,
		},
		{
			"extUnknown",
			ast.ExtensionCall("unknown", ast.String("hello")),
			nil,
			testutil.Error,
		},
		{
			"extArgs",
			ast.ExtensionCall("ip", ast.String("1"), ast.String("2")),
			nil,
			testutil.Error,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := toEval(tt.in.AsIsNode())
			out, err := e.Eval(&Context{
				Principal: types.NewEntityUID("Actor", "principal"),
				Action:    types.NewEntityUID("Action", "test"),
				Resource:  types.NewEntityUID("Resource", "database"),
				Context:   types.Record{},
			})
			tt.err(t, err)
			testutil.Equals(t, out, tt.out)
		})
	}

}

func TestToEvalPanic(t *testing.T) {
	t.Parallel()
	testutil.AssertPanic(t, func() {
		_ = toEval(ast.Node{}.AsIsNode())
	})
}
