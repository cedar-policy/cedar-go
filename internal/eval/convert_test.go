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
			ast.Record(types.Record{"key": types.Long(42)}).Access("key"),
			types.Long(42),
			testutil.OK,
		},
		{
			"has",
			ast.Record(types.Record{"key": types.Long(42)}).Has("key"),
			types.Boolean(true),
			testutil.OK,
		},
		{
			"like",
			ast.String("test").Like(types.Pattern{}),
			types.Boolean(false),
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
			ast.EntityUID(types.NewEntityUID("T", "42")).Is("T"),
			types.Boolean(true),
			testutil.OK,
		},
		{
			"isIn",
			ast.EntityUID(types.NewEntityUID("T", "42")).IsIn("T", ast.EntityUID(types.NewEntityUID("T", "42"))),
			types.Boolean(true),
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
			ast.RecordElements(ast.RecordElement{Key: "key", Value: ast.Long(42)}),
			types.Record{"key": types.Long(42)},
			testutil.OK,
		},
		{
			"set",
			ast.SetNodes(ast.Long(42)),
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
			types.Boolean(false),
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
			ast.EntityUID(types.NewEntityUID("T", "42")).In(ast.EntityUID(types.NewEntityUID("T", "43"))),
			types.Boolean(false),
			testutil.OK,
		},
		{
			"and",
			ast.True().And(ast.False()),
			types.Boolean(false),
			testutil.OK,
		},
		{
			"or",
			ast.True().Or(ast.False()),
			types.Boolean(true),
			testutil.OK,
		},
		{
			"equals",
			ast.Long(42).Equals(ast.Long(43)),
			types.Boolean(false),
			testutil.OK,
		},
		{
			"notEquals",
			ast.Long(42).NotEquals(ast.Long(43)),
			types.Boolean(true),
			testutil.OK,
		},
		{
			"greaterThan",
			ast.Long(42).GreaterThan(ast.Long(43)),
			types.Boolean(false),
			testutil.OK,
		},
		{
			"greaterThanOrEqual",
			ast.Long(42).GreaterThanOrEqual(ast.Long(43)),
			types.Boolean(false),
			testutil.OK,
		},
		{
			"lessThan",
			ast.Long(42).LessThan(ast.Long(43)),
			types.Boolean(true),
			testutil.OK,
		},
		{
			"lessThanOrEqual",
			ast.Long(42).LessThanOrEqual(ast.Long(43)),
			types.Boolean(true),
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
			ast.Set(types.Set{types.Long(42)}).Contains(ast.Long(42)),
			types.Boolean(true),
			testutil.OK,
		},
		{
			"containsAll",
			ast.Set(types.Set{types.Long(42), types.Long(43), types.Long(44)}).ContainsAll(ast.Set(types.Set{types.Long(42), types.Long(43)})),
			types.Boolean(true),
			testutil.OK,
		},
		{
			"containsAny",
			ast.Set(types.Set{types.Long(42), types.Long(43), types.Long(44)}).ContainsAny(ast.Set(types.Set{types.Long(1), types.Long(42)})),
			types.Boolean(true),
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
			ast.ExtensionCall("lessThan", ast.Decimal(420000), ast.Decimal(430000)),
			types.Boolean(true),
			testutil.OK,
		},
		{
			"lessThanOrEqual",
			ast.ExtensionCall("lessThanOrEqual", ast.Decimal(420000), ast.Decimal(430000)),
			types.Boolean(true),
			testutil.OK,
		},
		{
			"greaterThan",
			ast.ExtensionCall("greaterThan", ast.Decimal(420000), ast.Decimal(430000)),
			types.Boolean(false),
			testutil.OK,
		},
		{
			"greaterThanOrEqual",
			ast.ExtensionCall("greaterThanOrEqual", ast.Decimal(420000), ast.Decimal(430000)),
			types.Boolean(false),
			testutil.OK,
		},
		{
			"isIpv4",
			ast.ExtensionCall("isIpv4", ast.IPAddr(types.IPAddr(netip.MustParsePrefix("127.0.0.42/16")))),
			types.Boolean(true),
			testutil.OK,
		},
		{
			"isIpv6",
			ast.ExtensionCall("isIpv6", ast.IPAddr(types.IPAddr(netip.MustParsePrefix("::1/16")))),
			types.Boolean(true),
			testutil.OK,
		},
		{
			"isLoopback",
			ast.ExtensionCall("isLoopback", ast.IPAddr(types.IPAddr(netip.MustParsePrefix("127.0.0.1/32")))),
			types.Boolean(true),
			testutil.OK,
		},
		{
			"isMulticast",
			ast.ExtensionCall("isMulticast", ast.IPAddr(types.IPAddr(netip.MustParsePrefix("239.255.255.255/32")))),
			types.Boolean(true),
			testutil.OK,
		},
		{
			"isInRange",
			ast.ExtensionCall("isInRange", ast.IPAddr(types.IPAddr(netip.MustParsePrefix("127.0.0.42/32"))), ast.IPAddr(types.IPAddr(netip.MustParsePrefix("127.0.0.0/16")))),
			types.Boolean(true),
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

func TestToEvalPanics(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   ast.Node
	}{
		{
			"unknownNode",
			ast.Node{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.AssertPanic(t, func() {
				_ = toEval(tt.in.AsIsNode())
			})
		})
	}
}
