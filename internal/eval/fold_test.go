package eval

import (
	"net/netip"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func TestFoldNode(t *testing.T) {
	tests := []struct {
		name string
		in   ast.Node
		out  ast.Node
	}{
		{"record-bake",
			ast.Record(ast.Pairs{{Key: "key", Value: ast.True()}}),
			ast.Value(types.NewRecord(types.RecordMap{"key": types.True})),
		},
		{"set-bake",
			ast.Set(ast.True()),
			ast.Value(types.NewSet(types.True)),
		},
		{"record-fold",
			ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(6).Multiply(ast.Long(7))}}),
			ast.Value(types.NewRecord(types.RecordMap{"key": types.Long(42)})),
		},
		{"set-fold",
			ast.Set(ast.Long(6).Multiply(ast.Long(7))),
			ast.Value(types.NewSet(types.Long(42))),
		},
		{"record-blocked",
			ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(6).Multiply(ast.Context())}}),
			ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(6).Multiply(ast.Context())}}),
		},
		{"set-blocked",
			ast.Set(ast.Long(6).Multiply(ast.Context())),
			ast.Set(ast.Long(6).Multiply(ast.Context())),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			out := fold(tt.in.AsIsNode())
			testutil.Equals(t, out, tt.out.AsIsNode())
		})
	}
}

func TestFoldPolicy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   *ast.Policy
		out  *ast.Policy
	}{
		{
			"permit",
			ast.Permit(),
			ast.Permit(),
		},
		{
			"forbid",
			ast.Forbid(),
			ast.Forbid(),
		},
		{
			"annotationPermit",
			ast.Annotation("key", "value").Permit(),
			ast.Annotation("key", "value").Permit(),
		},
		{
			"annotationForbid",
			ast.Annotation("key", "value").Forbid(),
			ast.Annotation("key", "value").Forbid(),
		},
		{
			"annotations",
			ast.Annotation("key", "value").Annotation("abc", "xyz").Permit(),
			ast.Annotation("key", "value").Annotation("abc", "xyz").Permit(),
		},
		{
			"policyAnnotate",
			ast.Permit().Annotate("key", "value"),
			ast.Permit().Annotate("key", "value"),
		},
		{
			"when",
			ast.Permit().When(ast.True()),
			ast.Permit().When(ast.True()),
		},
		{
			"unless",
			ast.Permit().Unless(ast.True()),
			ast.Permit().Unless(ast.True()),
		},
		{
			"scopePrincipalEq",
			ast.Permit().PrincipalEq(types.NewEntityUID("T", "42")),
			ast.Permit().PrincipalEq(types.NewEntityUID("T", "42")),
		},
		{
			"scopePrincipalIn",
			ast.Permit().PrincipalIn(types.NewEntityUID("T", "42")),
			ast.Permit().PrincipalIn(types.NewEntityUID("T", "42")),
		},
		{
			"scopePrincipalIs",
			ast.Permit().PrincipalIs("T"),
			ast.Permit().PrincipalIs("T"),
		},
		{
			"scopePrincipalIsIn",
			ast.Permit().PrincipalIsIn("T", types.NewEntityUID("T", "42")),
			ast.Permit().PrincipalIsIn("T", types.NewEntityUID("T", "42")),
		},
		{
			"scopeActionEq",
			ast.Permit().ActionEq(types.NewEntityUID("T", "42")),
			ast.Permit().ActionEq(types.NewEntityUID("T", "42")),
		},
		{
			"scopeActionIn",
			ast.Permit().ActionIn(types.NewEntityUID("T", "42")),
			ast.Permit().ActionIn(types.NewEntityUID("T", "42")),
		},
		{
			"scopeActionInSet",
			ast.Permit().ActionInSet(types.NewEntityUID("T", "42"), types.NewEntityUID("T", "43")),
			ast.Permit().ActionInSet(types.NewEntityUID("T", "42"), types.NewEntityUID("T", "43")),
		},
		{
			"scopeResourceEq",
			ast.Permit().ResourceEq(types.NewEntityUID("T", "42")),
			ast.Permit().ResourceEq(types.NewEntityUID("T", "42")),
		},
		{
			"scopeResourceIn",
			ast.Permit().ResourceIn(types.NewEntityUID("T", "42")),
			ast.Permit().ResourceIn(types.NewEntityUID("T", "42")),
		},
		{
			"scopeResourceIs",
			ast.Permit().ResourceIs("T"),
			ast.Permit().ResourceIs("T"),
		},
		{
			"scopeResourceIsIn",
			ast.Permit().ResourceIsIn("T", types.NewEntityUID("T", "42")),
			ast.Permit().ResourceIsIn("T", types.NewEntityUID("T", "42")),
		},
		{
			"variablePrincipal",
			ast.Permit().When(ast.Principal()),
			ast.Permit().When(ast.Principal()),
		},
		{
			"variableAction",
			ast.Permit().When(ast.Action()),
			ast.Permit().When(ast.Action()),
		},
		{
			"variableResource",
			ast.Permit().When(ast.Resource()),
			ast.Permit().When(ast.Resource()),
		},
		{
			"variableContext",
			ast.Permit().When(ast.Context()),
			ast.Permit().When(ast.Context()),
		},
		{
			"valueBoolFalse",
			ast.Permit().When(ast.Boolean(false)),
			ast.Permit().When(ast.Boolean(false)),
		},
		{
			"valueBoolTrue",
			ast.Permit().When(ast.Boolean(true)),
			ast.Permit().When(ast.Boolean(true)),
		},
		{
			"valueTrue",
			ast.Permit().When(ast.True()),
			ast.Permit().When(ast.True()),
		},
		{
			"valueFalse",
			ast.Permit().When(ast.False()),
			ast.Permit().When(ast.False()),
		},
		{
			"valueString",
			ast.Permit().When(ast.String("cedar")),
			ast.Permit().When(ast.String("cedar")),
		},
		{
			"valueLong",
			ast.Permit().When(ast.Long(42)),
			ast.Permit().When(ast.Long(42)),
		},
		{
			"valueSetNodes",
			ast.Permit().When(ast.Set(ast.Long(42), ast.Long(43))),
			ast.Permit().When(ast.Value(types.NewSet(types.Long(42), types.Long(43)))),
		},
		{
			"valueRecordElements",
			ast.Permit().When(ast.Record(ast.Pairs{{Key: "key", Value: ast.Long(42)}})),
			ast.Permit().When(ast.Value(types.NewRecord(types.RecordMap{"key": types.Long(42)}))),
		},
		{
			"valueEntityUID",
			ast.Permit().When(ast.EntityUID("T", "42")),
			ast.Permit().When(ast.EntityUID("T", "42")),
		},
		{
			"valueIPAddr",
			ast.Permit().When(ast.IPAddr(netip.MustParsePrefix("127.0.0.1/16"))),
			ast.Permit().When(ast.IPAddr(netip.MustParsePrefix("127.0.0.1/16"))),
		},
		{
			"opEquals",
			ast.Permit().When(ast.Long(42).Equal(ast.Long(43))),
			ast.Permit().When(ast.False()),
		},
		{
			"opNotEquals",
			ast.Permit().When(ast.Long(42).NotEqual(ast.Long(43))),
			ast.Permit().When(ast.True()),
		},
		{
			"opEqualsUnfold",
			ast.Permit().When(ast.Long(42).Equal(ast.Context())),
			ast.Permit().When(ast.Long(42).Equal(ast.Context())),
		},
		{
			"opNotEqualsUnfold",
			ast.Permit().When(ast.Long(42).NotEqual(ast.Context())),
			ast.Permit().When(ast.Long(42).NotEqual(ast.Context())),
		},
		{
			"opLessThan",
			ast.Permit().When(ast.Long(42).LessThan(ast.Long(43))),
			ast.Permit().When(ast.True()),
		},
		{
			"opLessThanOrEqual",
			ast.Permit().When(ast.Long(42).LessThanOrEqual(ast.Long(43))),
			ast.Permit().When(ast.True()),
		},
		{
			"opGreaterThan",
			ast.Permit().When(ast.Long(42).GreaterThan(ast.Long(43))),
			ast.Permit().When(ast.False()),
		},
		{
			"opGreaterThanOrEqual",
			ast.Permit().When(ast.Long(42).GreaterThanOrEqual(ast.Long(43))),
			ast.Permit().When(ast.False()),
		},
		{
			"opLessThanErr",
			ast.Permit().When(ast.Long(42).LessThan(ast.String("test"))),
			ast.Permit().When(ast.Long(42).LessThan(ast.String("test"))),
		},
		{
			"opLessThanOrEqualErr",
			ast.Permit().When(ast.Long(42).LessThanOrEqual(ast.String("test"))),
			ast.Permit().When(ast.Long(42).LessThanOrEqual(ast.String("test"))),
		},
		{
			"opGreaterThanErr",
			ast.Permit().When(ast.Long(42).GreaterThan(ast.String("test"))),
			ast.Permit().When(ast.Long(42).GreaterThan(ast.String("test"))),
		},
		{
			"opGreaterThanOrEqualErr",
			ast.Permit().When(ast.Long(42).GreaterThanOrEqual(ast.String("test"))),
			ast.Permit().When(ast.Long(42).GreaterThanOrEqual(ast.String("test"))),
		},
		{
			"opLessThanComparable",
			ast.Permit().When(ast.Datetime(time.UnixMilli(42)).LessThan(ast.Datetime(time.UnixMilli(43)))),
			ast.Permit().When(ast.True()),
		},
		{
			"opLessThanOrEqualComparable",
			ast.Permit().When(ast.Datetime(time.UnixMilli(42)).LessThanOrEqual(ast.Datetime(time.UnixMilli(43)))),
			ast.Permit().When(ast.True()),
		},
		{
			"opGreaterThanComparable",
			ast.Permit().When(ast.Datetime(time.UnixMilli(42)).GreaterThan(ast.Datetime(time.UnixMilli(43)))),
			ast.Permit().When(ast.False()),
		},
		{
			"opGreaterThanOrEqualComparable",
			ast.Permit().When(ast.Datetime(time.UnixMilli(42)).GreaterThanOrEqual(ast.Datetime(time.UnixMilli(43)))),
			ast.Permit().When(ast.False()),
		},
		{
			"opLessThanExt",
			ast.Permit().When(ast.Value(testutil.Must(types.NewDecimalFromInt(42))).DecimalLessThan(ast.Value(testutil.Must(types.NewDecimalFromInt(43))))),
			ast.Permit().When(ast.True()),
		},
		{
			"opLessThanOrEqualExt",
			ast.Permit().When(ast.Value(testutil.Must(types.NewDecimalFromInt(42))).DecimalLessThanOrEqual(ast.Value(testutil.Must(types.NewDecimalFromInt(43))))),
			ast.Permit().When(ast.True()),
		},
		{
			"opGreaterThanExt",
			ast.Permit().When(ast.Value(testutil.Must(types.NewDecimalFromInt(42))).DecimalGreaterThan(ast.Value(testutil.Must(types.NewDecimalFromInt(43))))),
			ast.Permit().When(ast.False()),
		},
		{
			"opGreaterThanOrEqualExt",
			ast.Permit().When(ast.Value(testutil.Must(types.NewDecimalFromInt(42))).DecimalGreaterThanOrEqual(ast.Value(testutil.Must(types.NewDecimalFromInt(43))))),
			ast.Permit().When(ast.False()),
		},
		{
			"opLessThanExtErr",
			ast.Permit().When(ast.Long(42).DecimalLessThan(ast.Long(43))),
			ast.Permit().When(ast.Long(42).DecimalLessThan(ast.Long(43))),
		},
		{
			"opLessThanOrEqualExtErr",
			ast.Permit().When(ast.Long(42).DecimalLessThanOrEqual(ast.Long(43))),
			ast.Permit().When(ast.Long(42).DecimalLessThanOrEqual(ast.Long(43))),
		},
		{
			"opGreaterThanExtErr",
			ast.Permit().When(ast.Long(42).DecimalGreaterThan(ast.Long(43))),
			ast.Permit().When(ast.Long(42).DecimalGreaterThan(ast.Long(43))),
		},
		{
			"opGreaterThanOrEqualExtErr",
			ast.Permit().When(ast.Long(42).DecimalGreaterThanOrEqual(ast.Long(43))),
			ast.Permit().When(ast.Long(42).DecimalGreaterThanOrEqual(ast.Long(43))),
		},
		{
			"opLike",
			ast.Permit().When(ast.Long(42).Like(types.NewPattern(types.Wildcard{}))),
			ast.Permit().When(ast.Long(42).Like(types.NewPattern(types.Wildcard{}))),
		},
		{
			"opAnd",
			ast.Permit().When(ast.Long(42).And(ast.Long(43))),
			ast.Permit().When(ast.Long(42).And(ast.Long(43))),
		},
		{
			"opOr",
			ast.Permit().When(ast.Long(42).Or(ast.Long(43))),
			ast.Permit().When(ast.Long(42).Or(ast.Long(43))),
		},
		{
			"opNot",
			ast.Permit().When(ast.Not(ast.True())),
			ast.Permit().When(ast.False()),
		},
		{
			"opNotErr",
			ast.Permit().When(ast.Not(ast.String("test"))),
			ast.Permit().When(ast.Not(ast.String("test"))),
		},
		{
			"opNotUnfold",
			ast.Permit().When(ast.Not(ast.Context())),
			ast.Permit().When(ast.Not(ast.Context())),
		},
		{
			"opIf",
			ast.Permit().When(ast.IfThenElse(ast.True(), ast.Long(42), ast.Long(43))),
			ast.Permit().When(ast.Long(42)),
		},
		{
			"opIfUnfold",
			ast.Permit().When(ast.IfThenElse(ast.True(), ast.Principal(), ast.Action())),
			ast.Permit().When(ast.IfThenElse(ast.True(), ast.Principal(), ast.Action())),
		},
		{
			"opPlus",
			ast.Permit().When(ast.Long(42).Add(ast.Long(43))),
			ast.Permit().When(ast.Long(85)),
		},
		{
			"opMinus",
			ast.Permit().When(ast.Long(42).Subtract(ast.Long(43))),
			ast.Permit().When(ast.Long(-1)),
		},
		{
			"opTimes",
			ast.Permit().When(ast.Long(42).Multiply(ast.Long(43))),
			ast.Permit().When(ast.Long(1806)),
		},
		{
			"opNegate",
			ast.Permit().When(ast.Negate(ast.Long(42))),
			ast.Permit().When(ast.Long(-42)),
		},
		{
			"opPlusErr",
			ast.Permit().When(ast.Long(42).Add(ast.String("test"))),
			ast.Permit().When(ast.Long(42).Add(ast.String("test"))),
		},
		{
			"opMinusErr",
			ast.Permit().When(ast.Long(42).Subtract(ast.String("test"))),
			ast.Permit().When(ast.Long(42).Subtract(ast.String("test"))),
		},
		{
			"opTimesErr",
			ast.Permit().When(ast.Long(42).Multiply(ast.String("test"))),
			ast.Permit().When(ast.Long(42).Multiply(ast.String("test"))),
		},
		{
			"opNegateErr",
			ast.Permit().When(ast.Negate(ast.True())),
			ast.Permit().When(ast.Negate(ast.True())),
		},
		{
			"opIn",
			ast.Permit().When(ast.Long(42).In(ast.Long(43))),
			ast.Permit().When(ast.Long(42).In(ast.Long(43))),
		},
		{
			"opIs",
			ast.Permit().When(ast.Long(42).Is(types.EntityType("T"))),
			ast.Permit().When(ast.Long(42).Is(types.EntityType("T"))),
		},
		{
			"opIsIn",
			ast.Permit().When(ast.Long(42).IsIn(types.EntityType("T"), ast.Long(43))),
			ast.Permit().When(ast.Long(42).IsIn(types.EntityType("T"), ast.Long(43))),
		},
		{
			"opContains",
			ast.Permit().When(ast.Long(42).Contains(ast.Long(43))),
			ast.Permit().When(ast.Long(42).Contains(ast.Long(43))),
		},
		{
			"opContainsAll",
			ast.Permit().When(ast.Long(42).ContainsAll(ast.Long(43))),
			ast.Permit().When(ast.Long(42).ContainsAll(ast.Long(43))),
		},
		{
			"opContainsAny",
			ast.Permit().When(ast.Long(42).ContainsAny(ast.Long(43))),
			ast.Permit().When(ast.Long(42).ContainsAny(ast.Long(43))),
		},
		{
			"opIsEmpty",
			ast.Permit().When(ast.Long(42).IsEmpty()),
			ast.Permit().When(ast.Long(42).IsEmpty()),
		},
		{
			"opAccess",
			ast.Permit().When(ast.Long(42).Access("key")),
			ast.Permit().When(ast.Long(42).Access("key")),
		},
		{
			"opHas",
			ast.Permit().When(ast.Long(42).Has("key")),
			ast.Permit().When(ast.Long(42).Has("key")),
		},
		{
			"opAccessEntity",
			ast.Permit().When(ast.EntityUID("T", "1").Access("key")),
			ast.Permit().When(ast.EntityUID("T", "1").Access("key")),
		},
		{
			"opHasEntity",
			ast.Permit().When(ast.EntityUID("T", "1").Has("key")),
			ast.Permit().When(ast.EntityUID("T", "1").Has("key")),
		},
		{
			"opGetTagInvalidType",
			ast.Permit().When(ast.Long(42).GetTag(ast.String("key"))),
			ast.Permit().When(ast.Long(42).GetTag(ast.String("key"))),
		},
		{
			"opHasTagInvalidType",
			ast.Permit().When(ast.Long(42).HasTag(ast.String("key"))),
			ast.Permit().When(ast.Long(42).HasTag(ast.String("key"))),
		},
		{
			"opGetTag",
			ast.Permit().When(ast.EntityUID("T", "1").GetTag(ast.String("key"))),
			ast.Permit().When(ast.EntityUID("T", "1").GetTag(ast.String("key"))),
		},
		{
			"opHasTag",
			ast.Permit().When(ast.EntityUID("T", "1").HasTag(ast.String("key"))),
			ast.Permit().When(ast.EntityUID("T", "1").HasTag(ast.String("key"))),
		},
		{
			"opIsIpv4",
			ast.Permit().When(ast.Long(42).IsIpv4()),
			ast.Permit().When(ast.Long(42).IsIpv4()),
		},
		{
			"opIsIpv6",
			ast.Permit().When(ast.Long(42).IsIpv6()),
			ast.Permit().When(ast.Long(42).IsIpv6()),
		},
		{
			"opIsMulticast",
			ast.Permit().When(ast.Long(42).IsMulticast()),
			ast.Permit().When(ast.Long(42).IsMulticast()),
		},
		{
			"opIsLoopback",
			ast.Permit().When(ast.Long(42).IsLoopback()),
			ast.Permit().When(ast.Long(42).IsLoopback()),
		},
		{
			"opIsInRange",
			ast.Permit().When(ast.Long(42).IsInRange(ast.Long(43))),
			ast.Permit().When(ast.Long(42).IsInRange(ast.Long(43))),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := foldPolicy(tt.in)
			testutil.Equals(t, out, tt.out)
		})
	}
}

func TestFoldPanic(t *testing.T) {
	t.Parallel()
	testutil.Panic(t, func() {
		fold(nil)
	})
}
