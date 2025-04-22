package ast_test

import (
	"net/netip"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	internalast "github.com/cedar-policy/cedar-go/x/exp/ast"
)

func TestASTByTable(t *testing.T) {
	type CustomString string
	type CustomBool bool
	type CustomInt int
	type CustomInt64 int64
	t.Parallel()
	tests := []struct {
		name string
		in   *ast.Policy
		out  *internalast.Policy
	}{
		{
			"permit",
			ast.Permit(),
			internalast.Permit(),
		},
		{
			"forbid",
			ast.Forbid(),
			internalast.Forbid(),
		},
		{
			"annotationPermit",
			ast.Annotation("key", "value").Permit(),
			internalast.Annotation("key", "value").Permit(),
		},
		{
			"annotationForbid",
			ast.Annotation("key", "value").Forbid(),
			internalast.Annotation("key", "value").Forbid(),
		},
		{
			"annotations",
			ast.Annotation("key", "value").Annotation("abc", "xyz").Permit(),
			internalast.Annotation("key", "value").Annotation("abc", "xyz").Permit(),
		},
		{
			"policyAnnotate",
			ast.Permit().Annotate("key", "value"),
			internalast.Permit().Annotate("key", "value"),
		},
		{
			"when",
			ast.Permit().When(ast.True()),
			internalast.Permit().When(internalast.True()),
		},
		{
			"unless",
			ast.Permit().Unless(ast.True()),
			internalast.Permit().Unless(internalast.True()),
		},
		{
			"scopePrincipalEq",
			ast.Permit().PrincipalEq(types.NewEntityUID("T", "42")),
			internalast.Permit().PrincipalEq(types.NewEntityUID("T", "42")),
		},
		{
			"scopePrincipalIn",
			ast.Permit().PrincipalIn(types.NewEntityUID("T", "42")),
			internalast.Permit().PrincipalIn(types.NewEntityUID("T", "42")),
		},
		{
			"scopePrincipalIs",
			ast.Permit().PrincipalIs("T"),
			internalast.Permit().PrincipalIs("T"),
		},
		{
			"scopePrincipalIsIn",
			ast.Permit().PrincipalIsIn("T", types.NewEntityUID("T", "42")),
			internalast.Permit().PrincipalIsIn("T", types.NewEntityUID("T", "42")),
		},
		{
			"scopeActionEq",
			ast.Permit().ActionEq(types.NewEntityUID("T", "42")),
			internalast.Permit().ActionEq(types.NewEntityUID("T", "42")),
		},
		{
			"scopeActionIn",
			ast.Permit().ActionIn(types.NewEntityUID("T", "42")),
			internalast.Permit().ActionIn(types.NewEntityUID("T", "42")),
		},
		{
			"scopeActionInSet",
			ast.Permit().ActionInSet(types.NewEntityUID("T", "42"), types.NewEntityUID("T", "43")),
			internalast.Permit().ActionInSet(types.NewEntityUID("T", "42"), types.NewEntityUID("T", "43")),
		},
		{
			"scopeResourceEq",
			ast.Permit().ResourceEq(types.NewEntityUID("T", "42")),
			internalast.Permit().ResourceEq(types.NewEntityUID("T", "42")),
		},
		{
			"scopeResourceIn",
			ast.Permit().ResourceIn(types.NewEntityUID("T", "42")),
			internalast.Permit().ResourceIn(types.NewEntityUID("T", "42")),
		},
		{
			"scopeResourceIs",
			ast.Permit().ResourceIs("T"),
			internalast.Permit().ResourceIs("T"),
		},
		{
			"scopeResourceIsIn",
			ast.Permit().ResourceIsIn("T", types.NewEntityUID("T", "42")),
			internalast.Permit().ResourceIsIn("T", types.NewEntityUID("T", "42")),
		},
		{
			"variablePrincipal",
			ast.Permit().When(ast.Principal()),
			internalast.Permit().When(internalast.Principal()),
		},
		{
			"variableAction",
			ast.Permit().When(ast.Action()),
			internalast.Permit().When(internalast.Action()),
		},
		{
			"variableResource",
			ast.Permit().When(ast.Resource()),
			internalast.Permit().When(internalast.Resource()),
		},
		{
			"variableContext",
			ast.Permit().When(ast.Context()),
			internalast.Permit().When(internalast.Context()),
		},
		{
			"valueBoolFalse",
			ast.Permit().When(ast.Boolean(false)),
			internalast.Permit().When(internalast.Boolean(false)),
		},
		{
			"valueBoolTrue",
			ast.Permit().When(ast.Boolean(true)),
			internalast.Permit().When(internalast.Boolean(true)),
		},
		{
			"customValueBoolFalse",
			ast.Permit().When(ast.Boolean(CustomBool(false))),
			internalast.Permit().When(internalast.Boolean(false)),
		},
		{
			"customValueBoolTrue",
			ast.Permit().When(ast.Boolean(CustomBool(true))),
			internalast.Permit().When(internalast.Boolean(true)),
		},
		{
			"valueTrue",
			ast.Permit().When(ast.True()),
			internalast.Permit().When(internalast.True()),
		},
		{
			"valueFalse",
			ast.Permit().When(ast.False()),
			internalast.Permit().When(internalast.False()),
		},
		{
			"valueString",
			ast.Permit().When(ast.String("cedar")),
			internalast.Permit().When(internalast.String("cedar")),
		},
		{
			"customValueString",
			ast.Permit().When(ast.String(CustomString("cedar"))),
			internalast.Permit().When(internalast.String("cedar")),
		},
		{
			"customValueInt",
			ast.Permit().When(ast.Long(CustomInt(42))),
			internalast.Permit().When(internalast.Long(42)),
		},
		{
			"customValueInt64",
			ast.Permit().When(ast.Long(CustomInt64(42))),
			internalast.Permit().When(internalast.Long(42)),
		},
		{
			"valueLong",
			ast.Permit().When(ast.Long(42)),
			internalast.Permit().When(internalast.Long(42)),
		},
		{
			"valueSetNodes",
			ast.Permit().When(ast.Set(ast.Long(42), ast.Long(43))),
			internalast.Permit().When(internalast.Set(internalast.Long(42), internalast.Long(43))),
		},
		{
			"valueRecordElements",
			ast.Permit().When(ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(42)}})),
			internalast.Permit().When(internalast.Record(internalast.Pairs{{Key: "key", Value: internalast.Long(42)}})),
		},
		{
			"valueEntityUID",
			ast.Permit().When(ast.EntityUID("T", "42")),
			internalast.Permit().When(internalast.EntityUID("T", "42")),
		},
		{
			"valueIPAddr",
			ast.Permit().When(ast.IPAddr(netip.MustParsePrefix("127.0.0.1/16"))),
			internalast.Permit().When(internalast.IPAddr(netip.MustParsePrefix("127.0.0.1/16"))),
		},
		{
			"opEquals",
			ast.Permit().When(ast.Long(42).Equal(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).Equal(internalast.Long(43))),
		},
		{
			"opNotEquals",
			ast.Permit().When(ast.Long(42).NotEqual(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).NotEqual(internalast.Long(43))),
		},
		{
			"opLessThan",
			ast.Permit().When(ast.Long(42).LessThan(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).LessThan(internalast.Long(43))),
		},
		{
			"opLessThanOrEqual",
			ast.Permit().When(ast.Long(42).LessThanOrEqual(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).LessThanOrEqual(internalast.Long(43))),
		},
		{
			"opGreaterThan",
			ast.Permit().When(ast.Long(42).GreaterThan(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).GreaterThan(internalast.Long(43))),
		},
		{
			"opGreaterThanOrEqual",
			ast.Permit().When(ast.Long(42).GreaterThanOrEqual(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).GreaterThanOrEqual(internalast.Long(43))),
		},
		{
			"opLessThanExt",
			ast.Permit().When(ast.Long(42).DecimalLessThan(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).DecimalLessThan(internalast.Long(43))),
		},
		{
			"opLessThanOrEqualExt",
			ast.Permit().When(ast.Long(42).DecimalLessThanOrEqual(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).DecimalLessThanOrEqual(internalast.Long(43))),
		},
		{
			"opGreaterThanExt",
			ast.Permit().When(ast.Long(42).DecimalGreaterThan(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).DecimalGreaterThan(internalast.Long(43))),
		},
		{
			"opGreaterThanOrEqualExt",
			ast.Permit().When(ast.Long(42).DecimalGreaterThanOrEqual(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).DecimalGreaterThanOrEqual(internalast.Long(43))),
		},
		{
			"opLike",
			ast.Permit().When(ast.Long(42).Like(types.NewPattern(types.Wildcard{}))),
			internalast.Permit().When(internalast.Long(42).Like(types.NewPattern(types.Wildcard{}))),
		},
		{
			"opAnd",
			ast.Permit().When(ast.Long(42).And(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).And(internalast.Long(43))),
		},
		{
			"opOr",
			ast.Permit().When(ast.Long(42).Or(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).Or(internalast.Long(43))),
		},
		{
			"opNot",
			ast.Permit().When(ast.Not(ast.True())),
			internalast.Permit().When(internalast.Not(internalast.True())),
		},
		{
			"opIf",
			ast.Permit().When(ast.IfThenElse(ast.True(), ast.Long(42), ast.Long(43))),
			internalast.Permit().When(internalast.IfThenElse(internalast.True(), internalast.Long(42), internalast.Long(43))),
		},
		{
			"opPlus",
			ast.Permit().When(ast.Long(42).Add(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).Add(internalast.Long(43))),
		},
		{
			"opMinus",
			ast.Permit().When(ast.Long(42).Subtract(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).Subtract(internalast.Long(43))),
		},
		{
			"opTimes",
			ast.Permit().When(ast.Long(42).Multiply(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).Multiply(internalast.Long(43))),
		},
		{
			"opNegate",
			ast.Permit().When(ast.Negate(ast.True())),
			internalast.Permit().When(internalast.Negate(internalast.True())),
		},
		{
			"opIn",
			ast.Permit().When(ast.Long(42).In(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).In(internalast.Long(43))),
		},
		{
			"opIs",
			ast.Permit().When(ast.Long(42).Is(types.EntityType("T"))),
			internalast.Permit().When(internalast.Long(42).Is(types.EntityType("T"))),
		},
		{
			"opIsIn",
			ast.Permit().When(ast.Long(42).IsIn(types.EntityType("T"), ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).IsIn(types.EntityType("T"), internalast.Long(43))),
		},
		{
			"opContains",
			ast.Permit().When(ast.Long(42).Contains(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).Contains(internalast.Long(43))),
		},
		{
			"opContainsAll",
			ast.Permit().When(ast.Long(42).ContainsAll(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).ContainsAll(internalast.Long(43))),
		},
		{
			"opContainsAny",
			ast.Permit().When(ast.Long(42).ContainsAny(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).ContainsAny(internalast.Long(43))),
		},
		{
			"opContainsIsEmpty",
			ast.Permit().When(ast.Long(42).IsEmpty()),
			internalast.Permit().When(internalast.Long(42).IsEmpty()),
		},
		{
			"opAccess",
			ast.Permit().When(ast.Long(42).Access("key")),
			internalast.Permit().When(internalast.Long(42).Access("key")),
		},
		{
			"opHas",
			ast.Permit().When(ast.Long(42).Has("key")),
			internalast.Permit().When(internalast.Long(42).Has("key")),
		},
		{
			"opGetTag",
			ast.Permit().When(ast.EntityUID("T", "1").GetTag(ast.String("key"))),
			internalast.Permit().When(internalast.EntityUID("T", "1").GetTag(internalast.String("key"))),
		},
		{
			"opsHasTag",
			ast.Permit().When(ast.EntityUID("T", "1").HasTag(ast.String("key"))),
			internalast.Permit().When(internalast.EntityUID("T", "1").HasTag(internalast.String("key"))),
		},
		{
			"opIsIpv4",
			ast.Permit().When(ast.Long(42).IsIpv4()),
			internalast.Permit().When(internalast.Long(42).IsIpv4()),
		},
		{
			"opIsIpv6",
			ast.Permit().When(ast.Long(42).IsIpv6()),
			internalast.Permit().When(internalast.Long(42).IsIpv6()),
		},
		{
			"opIsMulticast",
			ast.Permit().When(ast.Long(42).IsMulticast()),
			internalast.Permit().When(internalast.Long(42).IsMulticast()),
		},
		{
			"opIsLoopback",
			ast.Permit().When(ast.Long(42).IsLoopback()),
			internalast.Permit().When(internalast.Long(42).IsLoopback()),
		},
		{
			"opIsInRange",
			ast.Permit().When(ast.Long(42).IsInRange(ast.Long(43))),
			internalast.Permit().When(internalast.Long(42).IsInRange(internalast.Long(43))),
		},
		{
			"opOffset",
			ast.Permit().When(ast.Datetime(time.Time{}).Offset(ast.Duration(time.Duration(100)))),
			internalast.Permit().When(internalast.Datetime(time.Time{}).Offset(internalast.Duration(time.Duration(100)))),
		},
		{
			"opDurationSince",
			ast.Permit().When(ast.Datetime(time.Time{}).DurationSince(ast.Datetime(time.Time{}))),
			internalast.Permit().When(internalast.Datetime(time.Time{}).DurationSince(internalast.Datetime(time.Time{}))),
		},
		{
			"opToDate",
			ast.Permit().When(ast.Datetime(time.Time{}).ToDate()),
			internalast.Permit().When(internalast.Datetime(time.Time{}).ToDate()),
		},
		{
			"opToTime",
			ast.Permit().When(ast.Datetime(time.Time{}).ToTime()),
			internalast.Permit().When(internalast.Datetime(time.Time{}).ToTime()),
		},
		{
			"opToDays",
			ast.Permit().When(ast.Duration(time.Duration(100)).ToDays()),
			internalast.Permit().When(internalast.Duration(100).ToDays()),
		},
		{
			"opToHours",
			ast.Permit().When(ast.Duration(time.Duration(100)).ToHours()),
			internalast.Permit().When(internalast.Duration(100).ToHours()),
		},
		{
			"opToMinutes",
			ast.Permit().When(ast.Duration(time.Duration(100)).ToMinutes()),
			internalast.Permit().When(internalast.Duration(100).ToMinutes()),
		},
		{
			"opToSeconds",
			ast.Permit().When(ast.Duration(time.Duration(100)).ToSeconds()),
			internalast.Permit().When(internalast.Duration(100).ToSeconds()),
		},
		{
			"opToMilliseconds",
			ast.Permit().When(ast.Duration(time.Duration(100)).ToMilliseconds()),
			internalast.Permit().When(internalast.Duration(100).ToMilliseconds()),
		},
		{
			"decimalExtension",
			ast.Permit().When(ast.DecimalExtensionCall(ast.Value(types.String("3.14")))),
			internalast.Permit().When(internalast.ExtensionCall("decimal", internalast.String("3.14"))),
		},
		{
			"ipExtension",
			ast.Permit().When(ast.IPExtensionCall(ast.Value(types.String("127.0.0.1")))),
			internalast.Permit().When(internalast.ExtensionCall("ip", internalast.String("127.0.0.1"))),
		},
		{
			"datetime",
			ast.Permit().When(ast.DatetimeExtensionCall(ast.Value(types.String("2006-01-02T15:04:05Z07:00")))),
			internalast.Permit().When(internalast.ExtensionCall("datetime", internalast.String("2006-01-02T15:04:05Z07:00"))),
		},
		{
			"duration",
			ast.Permit().When(ast.DurationExtensionCall(ast.Value(types.String("1d2h3m4s5ms")))),
			internalast.Permit().When(internalast.ExtensionCall("duration", internalast.String("1d2h3m4s5ms"))),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, (*internalast.Policy)(tt.in), tt.out)
		})
	}
}
