package eval

import (
	"net/netip"
	"testing"
	"time"

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
		err  func(testutil.TB, error)
	}{
		{
			"access",
			ast.Value(types.NewRecord(types.RecordMap{"key": types.Long(42)})).Access("key"),
			types.Long(42),
			testutil.OK,
		},
		{
			"has",
			ast.Value(types.NewRecord(types.RecordMap{"key": types.Long(42)})).Has("key"),
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
			ast.IfThenElse(ast.True(), ast.Long(42), ast.Long(43)),
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
			types.NewRecord(types.RecordMap{"key": types.Long(42)}),
			testutil.OK,
		},
		{
			"set",
			ast.Set(ast.Long(42)),
			types.NewSet([]types.Value{types.Long(42)}),
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
			ast.Long(42).Equal(ast.Long(43)),
			types.False,
			testutil.OK,
		},
		{
			"notEquals",
			ast.Long(42).NotEqual(ast.Long(43)),
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
			ast.Long(42).Subtract(ast.Long(2)),
			types.Long(40),
			testutil.OK,
		},
		{
			"add",
			ast.Long(40).Add(ast.Long(2)),
			types.Long(42),
			testutil.OK,
		},
		{
			"mult",
			ast.Long(6).Multiply(ast.Long(7)),
			types.Long(42),
			testutil.OK,
		},
		{
			"contains",
			ast.Value(types.NewSet([]types.Value{types.Long(42)})).Contains(ast.Long(42)),
			types.True,
			testutil.OK,
		},
		{
			"containsAll",
			ast.Value(types.NewSet([]types.Value{types.Long(42), types.Long(43), types.Long(44)})).ContainsAll(ast.Value(types.NewSet([]types.Value{types.Long(42), types.Long(43)}))),
			types.True,
			testutil.OK,
		},
		{
			"containsAny",
			ast.Value(types.NewSet([]types.Value{types.Long(42), types.Long(43), types.Long(44)})).ContainsAny(ast.Value(types.NewSet([]types.Value{types.Long(1), types.Long(42)}))),
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
			types.UnsafeDecimal(42.42),
			testutil.OK,
		},
		{
			"datetime",
			ast.ExtensionCall("datetime", ast.String("1970-01-01T00:00:00.001Z")),
			types.FromStdTime(time.UnixMilli(1)),
			testutil.OK,
		},
		{
			"duration",
			ast.ExtensionCall("duration", ast.String("1ms")),
			types.FromStdDuration(1 * time.Millisecond),
			testutil.OK,
		},
		{
			"toDate",
			ast.ExtensionCall("toDate", ast.Value(types.FromStdTime(time.UnixMilli(1)))),
			types.FromStdTime(time.UnixMilli(0)),
			testutil.OK,
		},
		{
			"toTime",
			ast.ExtensionCall("toTime", ast.Value(types.FromStdTime(time.UnixMilli(1)))),
			types.FromStdDuration(1 * time.Millisecond),
			testutil.OK,
		},
		{
			"toDays",
			ast.ExtensionCall("toDays", ast.Value(types.FromStdDuration(time.Duration(0)))),
			types.Long(0),
			testutil.OK,
		},
		{
			"toHours",
			ast.ExtensionCall("toHours", ast.Value(types.FromStdDuration(time.Duration(0)))),
			types.Long(0),
			testutil.OK,
		},
		{
			"toMinutes",
			ast.ExtensionCall("toMinutes", ast.Value(types.FromStdDuration(time.Duration(0)))),
			types.Long(0),
			testutil.OK,
		},
		{
			"toSeconds",
			ast.ExtensionCall("toSeconds", ast.Value(types.FromStdDuration(time.Duration(0)))),
			types.Long(0),
			testutil.OK,
		},
		{
			"toMilliseconds",
			ast.ExtensionCall("toMilliseconds", ast.Value(types.FromStdDuration(time.Duration(0)))),
			types.Long(0),
			testutil.OK,
		},
		{
			"offset",
			ast.ExtensionCall("offset", ast.Value(types.FromStdTime(time.UnixMilli(0))), ast.Value(types.FromStdDuration(1*time.Millisecond))),
			types.FromStdTime(time.UnixMilli(1)),
			testutil.OK,
		},
		{
			"durationSince",
			ast.ExtensionCall("durationSince", ast.Value(types.FromStdTime(time.UnixMilli(1))), ast.Value(types.FromStdTime(time.UnixMilli(1)))),
			types.FromStdDuration(time.Duration(0)),
			testutil.OK,
		},

		{
			"lessThan",
			ast.ExtensionCall("lessThan", ast.Value(types.UnsafeDecimal(42.0)), ast.Value(types.UnsafeDecimal(43))),
			types.True,
			testutil.OK,
		},
		{
			"lessThanOrEqual",
			ast.ExtensionCall("lessThanOrEqual", ast.Value(types.UnsafeDecimal(42.0)), ast.Value(types.UnsafeDecimal(43))),
			types.True,
			testutil.OK,
		},
		{
			"greaterThan",
			ast.ExtensionCall("greaterThan", ast.Value(types.UnsafeDecimal(42.0)), ast.Value(types.UnsafeDecimal(43))),
			types.False,
			testutil.OK,
		},
		{
			"greaterThanOrEqual",
			ast.ExtensionCall("greaterThanOrEqual", ast.Value(types.UnsafeDecimal(42.0)), ast.Value(types.UnsafeDecimal(43))),
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
			out, err := e.Eval(Env{
				Principal: types.NewEntityUID("Actor", "principal"),
				Action:    types.NewEntityUID("Action", "test"),
				Resource:  types.NewEntityUID("Resource", "database"),
				Context:   types.Record{},
				Entities:  types.Entities{},
			})
			tt.err(t, err)
			testutil.Equals(t, out, tt.out)
		})
	}

}

func TestToEvalPanic(t *testing.T) {
	t.Parallel()
	testutil.Panic(t, func() {
		_ = toEval(ast.Node{}.AsIsNode())
	})
}

func TestToEvalVariablePanic(t *testing.T) {
	t.Parallel()
	testutil.Panic(t, func() {
		_ = toEval(ast.NodeTypeVariable{Name: "bananas"})
	})
}
