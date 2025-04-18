package eval

import (
	"fmt"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go/internal"
	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

var errTest = fmt.Errorf("test error")

// not a real parser
func strEnt(v string) types.EntityUID {
	p := strings.Split(v, "::\"")
	return types.EntityUID{Type: types.EntityType(p[0]), ID: types.String(p[1][:len(p[1])-1])}
}

func AssertValue(t *testing.T, got, want types.Value) {
	t.Helper()
	testutil.FatalIf(
		t,
		(got != zeroValue() || want != zeroValue()) && (got == zeroValue() || want == zeroValue() || !got.Equal(want)),
		"got %v want %v", got, want)
}

func AssertBoolValue(t *testing.T, got types.Value, want bool) {
	t.Helper()
	testutil.Equals[types.Value](t, got, types.Boolean(want))
}

func AssertLongValue(t *testing.T, got types.Value, want int64) {
	t.Helper()
	testutil.Equals[types.Value](t, got, types.Long(want))
}

func AssertZeroValue(t *testing.T, got types.Value) {
	t.Helper()
	testutil.Equals(t, got, zeroValue())
}

func TestOrNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			lhs, rhs, result bool
		}{
			{false, false, false},
			{true, false, true},
			{false, true, true},
			{true, true, true},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%v%v", tt.lhs, tt.rhs), func(t *testing.T) {
				t.Parallel()
				n := newOrEval(newLiteralEval(types.Boolean(tt.lhs)), newLiteralEval(types.Boolean(tt.rhs)))
				v, err := n.Eval(Env{})
				testutil.OK(t, err)
				AssertBoolValue(t, v, tt.result)
			})
		}
	}

	t.Run("TrueXShortCircuit", func(t *testing.T) {
		t.Parallel()
		n := newOrEval(
			newLiteralEval(types.True), newLiteralEval(types.Long(1)))
		v, err := n.Eval(Env{})
		testutil.OK(t, err)
		AssertBoolValue(t, v, true)
	})

	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.True), errTest},
			{"LhsTypeError", newLiteralEval(types.Long(1)), newLiteralEval(types.True), ErrType},
			{"RhsError", newLiteralEval(types.False), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.False), newLiteralEval(types.Long(1)), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newOrEval(tt.lhs, tt.rhs)
				_, err := n.Eval(Env{})
				testutil.ErrorIs(t, err, tt.err)
			})
		}
	}
}

func TestAndNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			lhs, rhs, result bool
		}{
			{false, false, false},
			{true, false, false},
			{false, true, false},
			{true, true, true},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%v%v", tt.lhs, tt.rhs), func(t *testing.T) {
				t.Parallel()
				n := newAndEval(newLiteralEval(types.Boolean(tt.lhs)), newLiteralEval(types.Boolean(tt.rhs)))
				v, err := n.Eval(Env{})
				testutil.OK(t, err)
				AssertBoolValue(t, v, tt.result)
			})
		}
	}

	t.Run("FalseXShortCircuit", func(t *testing.T) {
		t.Parallel()
		n := newAndEval(
			newLiteralEval(types.False), newLiteralEval(types.Long(1)))
		v, err := n.Eval(Env{})
		testutil.OK(t, err)
		AssertBoolValue(t, v, false)
	})

	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.True), errTest},
			{"LhsTypeError", newLiteralEval(types.Long(1)), newLiteralEval(types.True), ErrType},
			{"RhsError", newLiteralEval(types.True), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Long(1)), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newAndEval(tt.lhs, tt.rhs)
				_, err := n.Eval(Env{})
				testutil.ErrorIs(t, err, tt.err)
			})
		}
	}
}

func TestNotNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			arg, result bool
		}{
			{false, true},
			{true, false},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%v", tt.arg), func(t *testing.T) {
				t.Parallel()
				n := newNotEval(newLiteralEval(types.Boolean(tt.arg)))
				v, err := n.Eval(Env{})
				testutil.OK(t, err)
				AssertBoolValue(t, v, tt.result)
			})
		}
	}

	{
		tests := []struct {
			name string
			arg  Evaler
			err  error
		}{
			{"Error", newErrorEval(errTest), errTest},
			{"TypeError", newLiteralEval(types.Long(1)), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newNotEval(tt.arg)
				_, err := n.Eval(Env{})
				testutil.ErrorIs(t, err, tt.err)
			})
		}
	}
}

func TestCheckedAddI64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		lhs, rhs, result types.Long
		ok               bool
	}{
		{1, 1, 2, true},

		{9_223_372_036_854_775_806, 1, 9_223_372_036_854_775_807, true},
		{9_223_372_036_854_775_807, 1, -9_223_372_036_854_775_808, false},
		{1, 9_223_372_036_854_775_806, 9_223_372_036_854_775_807, true},
		{1, 9_223_372_036_854_775_807, -9_223_372_036_854_775_808, false},
		{9_223_372_036_854_775_807, 9_223_372_036_854_775_807, -2, false},
		{4_611_686_018_427_387_904, 4_611_686_018_427_387_903, 9_223_372_036_854_775_807, true},
		{4_611_686_018_427_387_904, 4_611_686_018_427_387_904, -9_223_372_036_854_775_808, false},

		{-9_223_372_036_854_775_807, -1, -9_223_372_036_854_775_808, true},
		{-9_223_372_036_854_775_808, -1, 9_223_372_036_854_775_807, false},
		{-1, -9_223_372_036_854_775_807, -9_223_372_036_854_775_808, true},
		{-1, -9_223_372_036_854_775_808, 9_223_372_036_854_775_807, false},
		{-9_223_372_036_854_775_808, -9_223_372_036_854_775_808, 0, false},
		{-4_611_686_018_427_387_904, -4_611_686_018_427_387_904, -9_223_372_036_854_775_808, true},
		{-4_611_686_018_427_387_905, -4_611_686_018_427_387_904, 9_223_372_036_854_775_807, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%v+%v=%v(%v)", tt.lhs, tt.rhs, tt.result, tt.ok), func(t *testing.T) {
			t.Parallel()
			result, ok := checkedAddI64(tt.lhs, tt.rhs)
			testutil.Equals(t, ok, tt.ok)
			testutil.Equals(t, result, tt.result)
		})
	}
}

func TestCheckedSubI64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		lhs, rhs, result types.Long
		ok               bool
	}{
		{1, 1, 0, true},

		{9_223_372_036_854_775_806, -1, 9_223_372_036_854_775_807, true},
		{9_223_372_036_854_775_807, -1, -9_223_372_036_854_775_808, false},
		{1, -9_223_372_036_854_775_806, 9_223_372_036_854_775_807, true},
		{1, -9_223_372_036_854_775_807, -9_223_372_036_854_775_808, false},
		{9_223_372_036_854_775_807, -9_223_372_036_854_775_807, -2, false},
		{4_611_686_018_427_387_904, -4_611_686_018_427_387_903, 9_223_372_036_854_775_807, true},
		{4_611_686_018_427_387_904, -4_611_686_018_427_387_904, -9_223_372_036_854_775_808, false},

		{-9_223_372_036_854_775_807, 1, -9_223_372_036_854_775_808, true},
		{-9_223_372_036_854_775_808, 1, 9_223_372_036_854_775_807, false},
		{-1, 9_223_372_036_854_775_807, -9_223_372_036_854_775_808, true},
		{-2, 9_223_372_036_854_775_807, 9_223_372_036_854_775_807, false},
		{-9_223_372_036_854_775_808, 9_223_372_036_854_775_807, 1, false},
		{-4_611_686_018_427_387_904, 4_611_686_018_427_387_904, -9_223_372_036_854_775_808, true},
		{-4_611_686_018_427_387_905, 4_611_686_018_427_387_904, 9_223_372_036_854_775_807, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%v-%v=%v(%v)", tt.lhs, tt.rhs, tt.result, tt.ok), func(t *testing.T) {
			t.Parallel()
			result, ok := checkedSubI64(tt.lhs, tt.rhs)
			testutil.Equals(t, ok, tt.ok)
			testutil.Equals(t, result, tt.result)
		})
	}
}

func TestCheckedMulI64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		lhs, rhs, result types.Long
		ok               bool
	}{
		{2, 3, 6, true},
		{-2, 3, -6, true},
		{2, -3, -6, true},
		{-2, -3, 6, true},
		{9_223_372_036_854_775_807, 0, 0, true},
		{0, 9_223_372_036_854_775_807, 0, true},
		{-9_223_372_036_854_775_808, 0, 0, true},
		{0, -9_223_372_036_854_775_808, 0, true},

		{9_223_372_036_854_775_807, 1, 9_223_372_036_854_775_807, true},
		{9_223_372_036_854_775_807, 2, -2, false},
		{9_223_372_036_854_775_807, 3, 9_223_372_036_854_775_805, false},
		{9_223_372_036_854_775_807, 4, -4, false},
		{9_223_372_036_854_775_807, 5, 9_223_372_036_854_775_803, false},

		{9_223_372_036_854_775_807, -1, -9_223_372_036_854_775_807, true},
		{9_223_372_036_854_775_807, -2, 2, false},
		{9_223_372_036_854_775_807, -3, -9_223_372_036_854_775_805, false},
		{9_223_372_036_854_775_807, -4, 4, false},
		{9_223_372_036_854_775_807, -5, -9_223_372_036_854_775_803, false},

		{1, 9_223_372_036_854_775_807, 9_223_372_036_854_775_807, true},
		{2, 9_223_372_036_854_775_807, -2, false},
		{3, 9_223_372_036_854_775_807, 9_223_372_036_854_775_805, false},
		{4, 9_223_372_036_854_775_807, -4, false},
		{5, 9_223_372_036_854_775_807, 9_223_372_036_854_775_803, false},

		{-1, 9_223_372_036_854_775_807, -9_223_372_036_854_775_807, true},
		{-2, 9_223_372_036_854_775_807, 2, false},
		{-3, 9_223_372_036_854_775_807, -9_223_372_036_854_775_805, false},
		{-4, 9_223_372_036_854_775_807, 4, false},
		{-5, 9_223_372_036_854_775_807, -9_223_372_036_854_775_803, false},

		{-9_223_372_036_854_775_808, 1, -9_223_372_036_854_775_808, true},
		{-9_223_372_036_854_775_808, 2, 0, false},
		{-9_223_372_036_854_775_808, 3, -9_223_372_036_854_775_808, false},
		{-9_223_372_036_854_775_808, 4, 0, false},
		{-9_223_372_036_854_775_808, 5, -9_223_372_036_854_775_808, false},

		{-9_223_372_036_854_775_808, -1, -9_223_372_036_854_775_808, false},
		{-9_223_372_036_854_775_808, -2, 0, false},
		{-9_223_372_036_854_775_808, -3, -9_223_372_036_854_775_808, false},
		{-9_223_372_036_854_775_808, -4, 0, false},
		{-9_223_372_036_854_775_808, -5, -9_223_372_036_854_775_808, false},

		{1, -9_223_372_036_854_775_808, -9_223_372_036_854_775_808, true},
		{2, -9_223_372_036_854_775_808, 0, false},
		{3, -9_223_372_036_854_775_808, -9_223_372_036_854_775_808, false},
		{4, -9_223_372_036_854_775_808, 0, false},
		{5, -9_223_372_036_854_775_808, -9_223_372_036_854_775_808, false},

		{-1, -9_223_372_036_854_775_808, -9_223_372_036_854_775_808, false},
		{-2, -9_223_372_036_854_775_808, 0, false},
		{-3, -9_223_372_036_854_775_808, -9_223_372_036_854_775_808, false},
		{-4, -9_223_372_036_854_775_808, 0, false},
		{-5, -9_223_372_036_854_775_808, -9_223_372_036_854_775_808, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%v*%v=%v(%v)", tt.lhs, tt.rhs, tt.result, tt.ok), func(t *testing.T) {
			t.Parallel()
			result, ok := checkedMulI64(tt.lhs, tt.rhs)
			testutil.Equals(t, ok, tt.ok)
			testutil.Equals(t, result, tt.result)
		})
	}
}

func TestCheckedNegI64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		arg, result types.Long
		ok          bool
	}{
		{2, -2, true},
		{-2, 2, true},
		{0, 0, true},
		{9_223_372_036_854_775_807, -9_223_372_036_854_775_807, true},
		{-9_223_372_036_854_775_807, 9_223_372_036_854_775_807, true},
		{-9_223_372_036_854_775_808, 0, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("-%v*=%v(%v)", tt.arg, tt.result, tt.ok), func(t *testing.T) {
			t.Parallel()
			result, ok := checkedNegI64(tt.arg)
			testutil.Equals(t, ok, tt.ok)
			testutil.Equals(t, result, tt.result)
		})
	}
}

func TestAddNode(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		n := newAddEval(newLiteralEval(types.Long(1)), newLiteralEval(types.Long(2)))
		v, err := n.Eval(Env{})
		testutil.OK(t, err)
		AssertLongValue(t, v, 3)
	})

	tests := []struct {
		name     string
		lhs, rhs Evaler
		err      error
	}{
		{"LhsError", newErrorEval(errTest), newLiteralEval(types.Long(0)), errTest},
		{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Long(0)), ErrType},
		{"RhsError", newLiteralEval(types.Long(0)), newErrorEval(errTest), errTest},
		{"RhsTypeError", newLiteralEval(types.Long(0)), newLiteralEval(types.True), ErrType},
		{"PositiveOverflow",
			newLiteralEval(types.Long(9_223_372_036_854_775_807)),
			newLiteralEval(types.Long(1)),
			errOverflow},
		{"NegativeOverflow",
			newLiteralEval(types.Long(-9_223_372_036_854_775_808)),
			newLiteralEval(types.Long(-1)),
			errOverflow},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newAddEval(tt.lhs, tt.rhs)
			_, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
		})
	}
}

func TestSubtractNode(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		n := newSubtractEval(newLiteralEval(types.Long(1)), newLiteralEval(types.Long(2)))
		v, err := n.Eval(Env{})
		testutil.OK(t, err)
		AssertLongValue(t, v, -1)
	})

	tests := []struct {
		name     string
		lhs, rhs Evaler
		err      error
	}{
		{"LhsError", newErrorEval(errTest), newLiteralEval(types.Long(0)), errTest},
		{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Long(0)), ErrType},
		{"RhsError", newLiteralEval(types.Long(0)), newErrorEval(errTest), errTest},
		{"RhsTypeError", newLiteralEval(types.Long(0)), newLiteralEval(types.True), ErrType},
		{"PositiveOverflow",
			newLiteralEval(types.Long(9_223_372_036_854_775_807)),
			newLiteralEval(types.Long(-1)),
			errOverflow},
		{"NegativeOverflow",
			newLiteralEval(types.Long(-9_223_372_036_854_775_808)),
			newLiteralEval(types.Long(1)),
			errOverflow},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newSubtractEval(tt.lhs, tt.rhs)
			_, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
		})
	}
}

func TestMultiplyNode(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		n := newMultiplyEval(newLiteralEval(types.Long(-3)), newLiteralEval(types.Long(2)))
		v, err := n.Eval(Env{})
		testutil.OK(t, err)
		AssertLongValue(t, v, -6)
	})

	tests := []struct {
		name     string
		lhs, rhs Evaler
		err      error
	}{
		{"LhsError", newErrorEval(errTest), newLiteralEval(types.Long(0)), errTest},
		{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Long(0)), ErrType},
		{"RhsError", newLiteralEval(types.Long(0)), newErrorEval(errTest), errTest},
		{"RhsTypeError", newLiteralEval(types.Long(0)), newLiteralEval(types.True), ErrType},
		{"PositiveOverflow",
			newLiteralEval(types.Long(9_223_372_036_854_775_807)),
			newLiteralEval(types.Long(2)),
			errOverflow},
		{"NegativeOverflow",
			newLiteralEval(types.Long(-9_223_372_036_854_775_808)),
			newLiteralEval(types.Long(2)),
			errOverflow},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newMultiplyEval(tt.lhs, tt.rhs)
			_, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
		})
	}
}

func TestNegateNode(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		n := newNegateEval(newLiteralEval(types.Long(-3)))
		v, err := n.Eval(Env{})
		testutil.OK(t, err)
		AssertLongValue(t, v, 3)
	})

	tests := []struct {
		name string
		arg  Evaler
		err  error
	}{
		{"Error", newErrorEval(errTest), errTest},
		{"TypeError", newLiteralEval(types.True), ErrType},
		{"Overflow", newLiteralEval(types.Long(-9_223_372_036_854_775_808)), errOverflow},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newNegateEval(tt.arg)
			_, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
		})
	}
}

func TestDecimalLessThanNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			lhs, rhs string
			result   bool
		}{
			{"-1.0", "-1.0", false},
			{"-1.0", "0.0", true},
			{"-1.0", "1.0", true},
			{"0.0", "-1.0", false},
			{"0.0", "0.0", false},
			{"0.0", "1.0", true},
			{"1.0", "-1.0", false},
			{"1.0", "0.0", false},
			{"1.0", "1.0", false},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%s<%s", tt.lhs, tt.rhs), func(t *testing.T) {
				t.Parallel()
				lhsd, err := types.ParseDecimal(tt.lhs)
				testutil.OK(t, err)
				lhsv := lhsd
				rhsd, err := types.ParseDecimal(tt.rhs)
				testutil.OK(t, err)
				rhsv := rhsd
				n := newDecimalLessThanEval(newLiteralEval(lhsv), newLiteralEval(rhsv))
				v, err := n.Eval(Env{})
				testutil.OK(t, err)
				AssertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Decimal{}), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Decimal{}), ErrType},
			{"RhsError", newLiteralEval(types.Decimal{}), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Decimal{}), newLiteralEval(types.True), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newDecimalLessThanEval(tt.lhs, tt.rhs)
				_, err := n.Eval(Env{})
				testutil.ErrorIs(t, err, tt.err)
			})
		}
	}
}

func TestDecimalLessThanOrEqualNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			lhs, rhs string
			result   bool
		}{
			{"-1.0", "-1.0", true},
			{"-1.0", "0.0", true},
			{"-1.0", "1.0", true},
			{"0.0", "-1.0", false},
			{"0.0", "0.0", true},
			{"0.0", "1.0", true},
			{"1.0", "-1.0", false},
			{"1.0", "0.0", false},
			{"1.0", "1.0", true},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%s<=%s", tt.lhs, tt.rhs), func(t *testing.T) {
				t.Parallel()
				lhsd, err := types.ParseDecimal(tt.lhs)
				testutil.OK(t, err)
				lhsv := lhsd
				rhsd, err := types.ParseDecimal(tt.rhs)
				testutil.OK(t, err)
				rhsv := rhsd
				n := newDecimalLessThanOrEqualEval(newLiteralEval(lhsv), newLiteralEval(rhsv))
				v, err := n.Eval(Env{})
				testutil.OK(t, err)
				AssertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Decimal{}), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Decimal{}), ErrType},
			{"RhsError", newLiteralEval(types.Decimal{}), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Decimal{}), newLiteralEval(types.True), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newDecimalLessThanOrEqualEval(tt.lhs, tt.rhs)
				_, err := n.Eval(Env{})
				testutil.ErrorIs(t, err, tt.err)
			})
		}
	}
}

func TestDecimalGreaterThanNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			lhs, rhs string
			result   bool
		}{
			{"-1.0", "-1.0", false},
			{"-1.0", "0.0", false},
			{"-1.0", "1.0", false},
			{"0.0", "-1.0", true},
			{"0.0", "0.0", false},
			{"0.0", "1.0", false},
			{"1.0", "-1.0", true},
			{"1.0", "0.0", true},
			{"1.0", "1.0", false},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%s>%s", tt.lhs, tt.rhs), func(t *testing.T) {
				t.Parallel()
				lhsd, err := types.ParseDecimal(tt.lhs)
				testutil.OK(t, err)
				lhsv := lhsd
				rhsd, err := types.ParseDecimal(tt.rhs)
				testutil.OK(t, err)
				rhsv := rhsd
				n := newDecimalGreaterThanEval(newLiteralEval(lhsv), newLiteralEval(rhsv))
				v, err := n.Eval(Env{})
				testutil.OK(t, err)
				AssertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Decimal{}), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Decimal{}), ErrType},
			{"RhsError", newLiteralEval(types.Decimal{}), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Decimal{}), newLiteralEval(types.True), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newDecimalGreaterThanEval(tt.lhs, tt.rhs)
				_, err := n.Eval(Env{})
				testutil.ErrorIs(t, err, tt.err)
			})
		}
	}
}

func TestDecimalGreaterThanOrEqualNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			lhs, rhs string
			result   bool
		}{
			{"-1.0", "-1.0", true},
			{"-1.0", "0.0", false},
			{"-1.0", "1.0", false},
			{"0.0", "-1.0", true},
			{"0.0", "0.0", true},
			{"0.0", "1.0", false},
			{"1.0", "-1.0", true},
			{"1.0", "0.0", true},
			{"1.0", "1.0", true},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%s>=%s", tt.lhs, tt.rhs), func(t *testing.T) {
				t.Parallel()
				lhsd, err := types.ParseDecimal(tt.lhs)
				testutil.OK(t, err)
				lhsv := lhsd
				rhsd, err := types.ParseDecimal(tt.rhs)
				testutil.OK(t, err)
				rhsv := rhsd
				n := newDecimalGreaterThanOrEqualEval(newLiteralEval(lhsv), newLiteralEval(rhsv))
				v, err := n.Eval(Env{})
				testutil.OK(t, err)
				AssertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Decimal{}), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Decimal{}), ErrType},
			{"RhsError", newLiteralEval(types.Decimal{}), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Decimal{}), newLiteralEval(types.True), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newDecimalGreaterThanOrEqualEval(tt.lhs, tt.rhs)
				_, err := n.Eval(Env{})
				testutil.ErrorIs(t, err, tt.err)
			})
		}
	}
}

// hack: newComparableValueLessThanEval and friends aren't func(Evaler, Evaler) Evaler, but we can turn them into it with generics.
func makeEvaler[T Evaler](fn func(Evaler, Evaler) T) func(Evaler, Evaler) Evaler {
	return func(lhs Evaler, rhs Evaler) Evaler {
		return fn(lhs, rhs)
	}
}

// toEvaler lifts a Value or error into an Evaler
func toEvaler(x any) Evaler {
	switch v := x.(type) {
	case types.Value:
		return newLiteralEval(v)
	case error:
		return newErrorEval(v)
	default:
		panic(fmt.Sprintf("can't convert %T into Evaler", v))
	}
}

func TestComparableValueComparisonNodes(t *testing.T) {
	t.Parallel()

	var (
		zero         = types.Long(0)
		neg1         = types.Long(-1)
		pos1         = types.Long(1)
		zeroDate     = types.NewDatetime(time.UnixMilli(0))
		futureDate   = types.NewDatetime(time.UnixMilli(1))
		pastDate     = types.NewDatetime(time.UnixMilli(-1))
		zeroDuration = types.NewDuration(time.Duration(0))
		negDuration  = types.NewDuration(-1 * time.Millisecond)
		posDuration  = types.NewDuration(1 * time.Millisecond)
	)

	type test struct {
		lhs, rhs any
		result   bool
		wantErr  error
	}
	type testNode struct {
		name   string
		evaler func(Evaler, Evaler) Evaler
		table  []test
	}

	tests := []testNode{
		{name: "<",
			evaler: makeEvaler(newComparableValueLessThanEval),
			table: []test{
				{neg1, neg1, false, nil},
				{neg1, zero, true, nil},
				{neg1, pos1, true, nil},
				{zero, neg1, false, nil},
				{zero, zero, false, nil},
				{zero, pos1, true, nil},
				{pos1, neg1, false, nil},
				{pos1, zero, false, nil},
				{pos1, pos1, false, nil},

				// Datetime
				{pastDate, pastDate, false, nil},
				{pastDate, zeroDate, true, nil},
				{pastDate, futureDate, true, nil},
				{zeroDate, pastDate, false, nil},
				{zeroDate, zeroDate, false, nil},
				{zeroDate, futureDate, true, nil},
				{futureDate, pastDate, false, nil},
				{futureDate, zeroDate, false, nil},
				{futureDate, futureDate, false, nil},

				// Duration
				{negDuration, negDuration, false, nil},
				{negDuration, zeroDuration, true, nil},
				{negDuration, posDuration, true, nil},
				{zeroDuration, negDuration, false, nil},
				{zeroDuration, zeroDuration, false, nil},
				{zeroDuration, posDuration, true, nil},
				{posDuration, negDuration, false, nil},
				{posDuration, zeroDuration, false, nil},
				{posDuration, posDuration, false, nil},

				// Errors
				{errTest, neg1, false, errTest},
				{neg1, errTest, false, errTest},
				{types.True, neg1, false, ErrType},
				{neg1, types.False, false, ErrType},
				{pastDate, neg1, false, ErrType},
				{negDuration, futureDate, false, ErrType},
				{neg1, negDuration, false, ErrType},
			},
		},
		{name: "<=",
			evaler: makeEvaler(newComparableValueLessThanOrEqualEval),
			table: []test{
				{neg1, neg1, true, nil},
				{neg1, zero, true, nil},
				{neg1, pos1, true, nil},
				{zero, neg1, false, nil},
				{zero, zero, true, nil},
				{zero, pos1, true, nil},
				{pos1, neg1, false, nil},
				{pos1, zero, false, nil},
				{pos1, pos1, true, nil},

				// Datetime
				{pastDate, pastDate, true, nil},
				{pastDate, zeroDate, true, nil},
				{pastDate, futureDate, true, nil},
				{zeroDate, pastDate, false, nil},
				{zeroDate, zeroDate, true, nil},
				{zeroDate, futureDate, true, nil},
				{futureDate, pastDate, false, nil},
				{futureDate, zeroDate, false, nil},
				{futureDate, futureDate, true, nil},

				// Duration
				{negDuration, negDuration, true, nil},
				{negDuration, zeroDuration, true, nil},
				{negDuration, posDuration, true, nil},
				{zeroDuration, negDuration, false, nil},
				{zeroDuration, zeroDuration, true, nil},
				{zeroDuration, posDuration, true, nil},
				{posDuration, negDuration, false, nil},
				{posDuration, zeroDuration, false, nil},
				{posDuration, posDuration, true, nil},

				// Errors
				{errTest, neg1, false, errTest},
				{neg1, errTest, false, errTest},
				{types.True, neg1, false, ErrType},
				{neg1, types.False, false, ErrType},
				{pastDate, neg1, false, ErrType},
				{negDuration, futureDate, false, ErrType},
				{neg1, negDuration, false, ErrType},
			},
		},
		{name: ">",
			evaler: makeEvaler(newComparableValueGreaterThanEval),
			table: []test{
				{neg1, neg1, false, nil},
				{neg1, zero, false, nil},
				{neg1, pos1, false, nil},
				{zero, neg1, true, nil},
				{zero, zero, false, nil},
				{zero, pos1, false, nil},
				{pos1, neg1, true, nil},
				{pos1, zero, true, nil},
				{pos1, pos1, false, nil},

				// Datetime
				{pastDate, pastDate, false, nil},
				{pastDate, zeroDate, false, nil},
				{pastDate, futureDate, false, nil},
				{zeroDate, pastDate, true, nil},
				{zeroDate, zeroDate, false, nil},
				{zeroDate, futureDate, false, nil},
				{futureDate, pastDate, true, nil},
				{futureDate, zeroDate, true, nil},
				{futureDate, futureDate, false, nil},

				// Duration
				{negDuration, negDuration, false, nil},
				{negDuration, zeroDuration, false, nil},
				{negDuration, posDuration, false, nil},
				{zeroDuration, negDuration, true, nil},
				{zeroDuration, zeroDuration, false, nil},
				{zeroDuration, posDuration, false, nil},
				{posDuration, negDuration, true, nil},
				{posDuration, zeroDuration, true, nil},
				{posDuration, posDuration, false, nil},

				// Errors
				{errTest, neg1, false, errTest},
				{neg1, errTest, false, errTest},
				{types.True, neg1, false, ErrType},
				{neg1, types.False, false, ErrType},
				{pastDate, neg1, false, ErrType},
				{negDuration, futureDate, false, ErrType},
				{neg1, negDuration, false, ErrType},
			},
		},
		{name: ">=",
			evaler: makeEvaler(newComparableValueGreaterThanOrEqualEval),
			table: []test{
				{neg1, neg1, true, nil},
				{neg1, zero, false, nil},
				{neg1, pos1, false, nil},
				{zero, neg1, true, nil},
				{zero, zero, true, nil},
				{zero, pos1, false, nil},
				{pos1, neg1, true, nil},
				{pos1, zero, true, nil},
				{pos1, pos1, true, nil},

				// Datetime
				{pastDate, pastDate, true, nil},
				{pastDate, zeroDate, false, nil},
				{pastDate, futureDate, false, nil},
				{zeroDate, pastDate, true, nil},
				{zeroDate, zeroDate, true, nil},
				{zeroDate, futureDate, false, nil},
				{futureDate, pastDate, true, nil},
				{futureDate, zeroDate, true, nil},
				{futureDate, futureDate, true, nil},

				// Duration
				{negDuration, negDuration, true, nil},
				{negDuration, zeroDuration, false, nil},
				{negDuration, posDuration, false, nil},
				{zeroDuration, negDuration, true, nil},
				{zeroDuration, zeroDuration, true, nil},
				{zeroDuration, posDuration, false, nil},
				{posDuration, negDuration, true, nil},
				{posDuration, zeroDuration, true, nil},
				{posDuration, posDuration, true, nil},

				// Errors
				{errTest, neg1, false, errTest},
				{neg1, errTest, false, errTest},
				{types.True, neg1, false, ErrType},
				{neg1, types.False, false, ErrType},
				{pastDate, neg1, false, ErrType},
				{negDuration, futureDate, false, ErrType},
				{neg1, negDuration, false, ErrType},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		for _, tt := range tc.table {
			t.Run(fmt.Sprintf("%v_%s_%v", tt.lhs, tc.name, tt.rhs), func(t *testing.T) {
				t.Parallel()
				n := tc.evaler(toEvaler(tt.lhs), toEvaler(tt.rhs))
				v, err := n.Eval(Env{})
				if tt.wantErr == nil {
					testutil.OK(t, err)
					AssertBoolValue(t, v, tt.result)
				} else {
					testutil.ErrorIs(t, err, tt.wantErr)
				}
			})
		}
	}
}

func TestIfThenElseNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                       string
		ifNode, thenNode, elseNode Evaler
		result                     types.Value
		err                        error
	}{
		{"Then", newLiteralEval(types.True), newLiteralEval(types.Long(42)),
			newLiteralEval(types.Long(-1)), types.Long(42),
			nil},
		{"Else", newLiteralEval(types.False), newLiteralEval(types.Long(-1)),
			newLiteralEval(types.Long(42)), types.Long(42),
			nil},
		{"Err", newErrorEval(errTest), newLiteralEval(zeroValue()), newLiteralEval(zeroValue()), zeroValue(),
			errTest},
		{"ErrType", newLiteralEval(types.Long(123)), newLiteralEval(zeroValue()), newLiteralEval(zeroValue()), zeroValue(),
			ErrType},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newIfThenElseEval(tt.ifNode, tt.thenNode, tt.elseNode)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			testutil.Equals(t, v, tt.result)
		})
	}
}

func TestEqualNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		lhs, rhs Evaler
		result   types.Value
		err      error
	}{
		{"equals", newLiteralEval(types.Long(42)), newLiteralEval(types.Long(42)), types.True, nil},
		{"notEquals", newLiteralEval(types.Long(42)), newLiteralEval(types.Long(1234)), types.False, nil},
		{"leftErr", newErrorEval(errTest), newLiteralEval(zeroValue()), zeroValue(), errTest},
		{"rightErr", newLiteralEval(zeroValue()), newErrorEval(errTest), zeroValue(), errTest},
		{"typesNotEqual", newLiteralEval(types.Long(1)), newLiteralEval(types.True), types.False, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newEqualEval(tt.lhs, tt.rhs)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestNotEqualNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		lhs, rhs Evaler
		result   types.Value
		err      error
	}{
		{"equals", newLiteralEval(types.Long(42)), newLiteralEval(types.Long(42)), types.False, nil},
		{"notEquals", newLiteralEval(types.Long(42)), newLiteralEval(types.Long(1234)), types.True, nil},
		{"leftErr", newErrorEval(errTest), newLiteralEval(zeroValue()), zeroValue(), errTest},
		{"rightErr", newLiteralEval(zeroValue()), newErrorEval(errTest), zeroValue(), errTest},
		{"typesNotEqual", newLiteralEval(types.Long(1)), newLiteralEval(types.True), types.True, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newNotEqualEval(tt.lhs, tt.rhs)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestSetLiteralNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		elems  []Evaler
		result types.Value
		err    error
	}{
		{"empty", []Evaler{}, types.Set{}, nil},
		{"errorNode", []Evaler{newErrorEval(errTest)}, zeroValue(), errTest},
		{"nested",
			[]Evaler{
				newLiteralEval(types.True),
				newLiteralEval(types.NewSet(
					types.False,
					types.Long(1),
				)),
				newLiteralEval(types.Long(10)),
			},
			types.NewSet(
				types.True,
				types.NewSet(
					types.False,
					types.Long(1),
				),
				types.Long(10),
			),
			nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newSetLiteralEval(tt.elems)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestContainsNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Long(0)), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Long(0)), ErrType},
			{"RhsError", newLiteralEval(types.Set{}), newErrorEval(errTest), errTest},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newContainsEval(tt.lhs, tt.rhs)
				v, err := n.Eval(Env{})
				testutil.ErrorIs(t, err, tt.err)
				AssertZeroValue(t, v)
			})
		}
	}
	{
		empty := types.Set{}
		trueAndOne := types.NewSet(types.True, types.Long(1))
		nested := types.NewSet(trueAndOne, types.False, types.Long(2))

		tests := []struct {
			name     string
			lhs, rhs Evaler
			result   bool
		}{
			{"empty", newLiteralEval(empty), newLiteralEval(types.True), false},
			{"trueAndOneContainsTrue", newLiteralEval(trueAndOne), newLiteralEval(types.True), true},
			{"trueAndOneContainsOne", newLiteralEval(trueAndOne), newLiteralEval(types.Long(1)), true},
			{"trueAndOneDoesNotContainTwo", newLiteralEval(trueAndOne), newLiteralEval(types.Long(2)), false},
			{"nestedContainsFalse", newLiteralEval(nested), newLiteralEval(types.False), true},
			{"nestedContainsSet", newLiteralEval(nested), newLiteralEval(trueAndOne), true},
			{"nestedDoesNotContainTrue", newLiteralEval(nested), newLiteralEval(types.True), false},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newContainsEval(tt.lhs, tt.rhs)
				v, err := n.Eval(Env{})
				testutil.OK(t, err)
				AssertBoolValue(t, v, tt.result)
			})
		}
	}
}

func TestContainsAllNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Set{}), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Set{}), ErrType},
			{"RhsError", newLiteralEval(types.Set{}), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Set{}), newLiteralEval(types.Long(0)), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newContainsAllEval(tt.lhs, tt.rhs)
				v, err := n.Eval(Env{})
				testutil.ErrorIs(t, err, tt.err)
				AssertZeroValue(t, v)
			})
		}
	}
	{
		empty := types.Set{}
		trueOnly := types.NewSet(types.True)
		trueAndOne := types.NewSet(types.True, types.Long(1))
		nested := types.NewSet(trueAndOne, types.False, types.Long(2))

		tests := []struct {
			name     string
			lhs, rhs Evaler
			result   bool
		}{
			{"emptyEmpty", newLiteralEval(empty), newLiteralEval(empty), true},
			{"trueAndOneEmpty", newLiteralEval(trueAndOne), newLiteralEval(empty), true},
			{"trueAndOneTrueOnly", newLiteralEval(trueAndOne), newLiteralEval(trueOnly), true},
			{"trueOnlyTrueAndOne", newLiteralEval(trueOnly), newLiteralEval(trueAndOne), false},
			{"nestedNested", newLiteralEval(nested), newLiteralEval(nested), true},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newContainsAllEval(tt.lhs, tt.rhs)
				v, err := n.Eval(Env{})
				testutil.OK(t, err)
				AssertBoolValue(t, v, tt.result)
			})
		}
	}
}

func TestContainsAnyNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Set{}), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Set{}), ErrType},
			{"RhsError", newLiteralEval(types.Set{}), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Set{}), newLiteralEval(types.Long(0)), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newContainsAnyEval(tt.lhs, tt.rhs)
				v, err := n.Eval(Env{})
				testutil.ErrorIs(t, err, tt.err)
				AssertZeroValue(t, v)
			})
		}
	}
	{
		empty := types.Set{}
		trueOnly := types.NewSet(types.True)
		trueAndOne := types.NewSet(types.True, types.Long(1))
		trueAndTwo := types.NewSet(types.True, types.Long(2))
		nested := types.NewSet(trueAndOne, types.False, types.Long(2))

		tests := []struct {
			name     string
			lhs, rhs Evaler
			result   bool
		}{
			{"emptyEmpty", newLiteralEval(empty), newLiteralEval(empty), false},
			{"emptyTrueAndOne", newLiteralEval(empty), newLiteralEval(trueAndOne), false},
			{"trueAndOneEmpty", newLiteralEval(trueAndOne), newLiteralEval(empty), false},
			{"trueAndOneTrueOnly", newLiteralEval(trueAndOne), newLiteralEval(trueOnly), true},
			{"trueOnlyTrueAndOne", newLiteralEval(trueOnly), newLiteralEval(trueAndOne), true},
			{"trueAndOneTrueAndTwo", newLiteralEval(trueAndOne), newLiteralEval(trueAndTwo), true},
			{"nestedTrueAndOne", newLiteralEval(nested), newLiteralEval(trueAndOne), false},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newContainsAnyEval(tt.lhs, tt.rhs)
				v, err := n.Eval(Env{})
				testutil.OK(t, err)
				AssertBoolValue(t, v, tt.result)
			})
		}
	}

	t.Run("not quadratic", func(t *testing.T) {
		t.Parallel()

		// Make two totally disjoint sets to force a worst case search
		setSize := 200000
		set1 := make([]types.Value, setSize)
		set2 := make([]types.Value, setSize)

		for i := 0; i < setSize; i++ {
			set1[i] = types.Long(i)
			set2[i] = types.Long(setSize + i)
		}

		n := newContainsAnyEval(newLiteralEval(types.NewSet(set1...)), newLiteralEval(types.NewSet(set2...)))

		// This call would take several minutes if the evaluation of ContainsAny was quadratic
		val, err := n.Eval(Env{})

		testutil.OK(t, err)
		testutil.Equals(t, val.(types.Boolean), types.False)
	})
}

func TestIsEmptyNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			name string
			lhs  Evaler
			err  error
		}{
			{"LhsError", newErrorEval(errTest), errTest},
			{"LhsTypeError", newLiteralEval(types.True), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newIsEmptyEval(tt.lhs)
				v, err := n.Eval(Env{})
				testutil.ErrorIs(t, err, tt.err)
				AssertZeroValue(t, v)
			})
		}
	}
	{
		empty := types.Set{}
		trueOnly := types.NewSet(types.True)

		tests := []struct {
			name   string
			lhs    Evaler
			result bool
		}{
			{"emptyEmpty", newLiteralEval(empty), true},
			{"trueAndOneEmpty", newLiteralEval(trueOnly), false},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newIsEmptyEval(tt.lhs)
				v, err := n.Eval(Env{})
				testutil.OK(t, err)
				AssertBoolValue(t, v, tt.result)
			})
		}
	}
}

func TestRecordLiteralNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		elems  map[types.String]Evaler
		result types.Value
		err    error
	}{
		{"empty", map[types.String]Evaler{}, types.Record{}, nil},
		{"errorNode", map[types.String]Evaler{"foo": newErrorEval(errTest)}, zeroValue(), errTest},
		{"ok",
			map[types.String]Evaler{
				"foo": newLiteralEval(types.True),
				"bar": newLiteralEval(types.String("baz")),
			}, types.NewRecord(types.RecordMap{
				"foo": types.True,
				"bar": types.String("baz"),
			}), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newRecordLiteralEval(tt.elems)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestAttributeAccessNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		object    Evaler
		attribute types.String
		result    types.Value
		err       error
	}{
		{"RecordError", newErrorEval(errTest), "foo", zeroValue(), errTest},
		{"RecordTypeError", newLiteralEval(types.True), "foo", zeroValue(), ErrType},
		{"UnknownAttribute",
			newLiteralEval(types.Record{}),
			"foo",
			zeroValue(),
			errAttributeAccess},
		{"KnownAttribute",
			newLiteralEval(types.NewRecord(types.RecordMap{"foo": types.Long(42)})),
			"foo",
			types.Long(42),
			nil},
		{"KnownAttributeOnEntity",
			newLiteralEval(types.NewEntityUID("knownType", "knownID")),
			"knownAttr",
			types.Long(42),
			nil},
		{"UnknownAttributeOnEntity",
			newLiteralEval(types.NewEntityUID("knownType", "knownID")),
			"unknownAttr",
			zeroValue(),
			errAttributeAccess},
		{"UnknownEntity",
			newLiteralEval(types.NewEntityUID("unknownType", "unknownID")),
			"unknownAttr",
			zeroValue(),
			errEntityNotExist},
		{"UnspecifiedEntity",
			newLiteralEval(types.NewEntityUID("", "")),
			"knownAttr",
			zeroValue(),
			errUnspecifiedEntity},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newAttributeAccessEval(tt.object, tt.attribute)
			entity := types.Entity{
				UID:        types.NewEntityUID("knownType", "knownID"),
				Attributes: types.NewRecord(types.RecordMap{"knownAttr": types.Long(42)}),
			}
			v, err := n.Eval(Env{
				Entities: types.EntityMap{
					entity.UID: entity,
				},
			})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestHasNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		record    Evaler
		attribute types.String
		result    types.Value
		err       error
	}{
		{"RecordError", newErrorEval(errTest), "foo", zeroValue(), errTest},
		{"RecordTypeError", newLiteralEval(types.True), "foo", zeroValue(), ErrType},
		{"UnknownAttribute",
			newLiteralEval(types.Record{}),
			"foo",
			types.False,
			nil},
		{"KnownAttribute",
			newLiteralEval(types.NewRecord(types.RecordMap{"foo": types.Long(42)})),
			"foo",
			types.True,
			nil},
		{"KnownAttributeOnEntity",
			newLiteralEval(types.NewEntityUID("knownType", "knownID")),
			"knownAttr",
			types.True,
			nil},
		{"UnknownAttributeOnEntity",
			newLiteralEval(types.NewEntityUID("knownType", "knownID")),
			"unknownAttr",
			types.False,
			nil},
		{"UnknownEntity",
			newLiteralEval(types.NewEntityUID("unknownType", "unknownID")),
			"unknownAttr",
			types.False,
			nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newHasEval(tt.record, tt.attribute)
			entity := types.Entity{
				UID:        types.NewEntityUID("knownType", "knownID"),
				Attributes: types.NewRecord(types.RecordMap{"knownAttr": types.Long(42)}),
			}
			v, err := n.Eval(Env{
				Entities: types.EntityMap{
					entity.UID: entity,
				},
			})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestGetTagNode(t *testing.T) {
	t.Parallel()

	const (
		knownType   = types.EntityType("knownType")
		unknownType = types.EntityType("")
		knownID     = "knownID"
		unknownID   = ""
		knownTag    = types.String("knownTag")
		unknownTag  = types.String("unknownTag")
		knownAttr   = "knownAttr"
		notString   = types.Long(42)
		value       = types.Long(42)
	)

	tests := []struct {
		name   string
		lhs    Evaler
		rhs    Evaler
		result types.Value
		err    error
	}{
		{"ObjectTypeError",
			newLiteralEval(types.True),
			newLiteralEval(knownTag),
			zeroValue(),
			ErrType},
		{"SubjectTypeError",
			newLiteralEval(types.NewEntityUID(knownType, knownID)),
			newLiteralEval(notString),
			zeroValue(),
			ErrType},
		{"TagOnRecord",
			newLiteralEval(types.NewRecord(nil)),
			newLiteralEval(knownTag), zeroValue(),
			ErrType},
		{"ProgrammaticTag",
			newLiteralEval(types.NewEntityUID(knownType, knownID)),
			newAttributeAccessEval(newLiteralEval(types.NewEntityUID(knownType, knownID)), knownAttr),
			value,
			nil,
		},
		{"KnownTag",
			newLiteralEval(types.NewEntityUID(knownType, knownID)),
			newLiteralEval(knownTag),
			value,
			nil},
		{"UnknownTag",
			newLiteralEval(types.NewEntityUID(knownType, knownID)),
			newLiteralEval(unknownTag),
			zeroValue(),
			errTagAccess},
		{"UnknownEntity",
			newLiteralEval(types.NewEntityUID("unknownType", unknownID)),
			newLiteralEval(knownTag),
			zeroValue(),
			errEntityNotExist},
		{"ZeroEntity",
			newLiteralEval(types.NewEntityUID("", "")),
			newLiteralEval(knownTag),
			zeroValue(),
			errUnspecifiedEntity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newGetTagEval(tt.lhs, tt.rhs)
			entity := types.Entity{
				UID:        types.NewEntityUID(knownType, knownID),
				Tags:       types.NewRecord(types.RecordMap{knownTag: value}),
				Attributes: types.NewRecord(types.RecordMap{knownAttr: knownTag}),
			}
			v, err := n.Eval(Env{
				Entities: types.EntityMap{
					entity.UID: entity,
				},
			})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})

	}
}

func TestHasTagNode(t *testing.T) {
	t.Parallel()

	const (
		knownType   = types.EntityType("knownType")
		unknownType = types.EntityType("unknownType")
		knownID     = "knownID"
		unkownID    = "unknownID"
		knownTag    = types.String("knownTag")
		unknownTag  = types.String("unknownTag")
		knownAttr   = types.String("knownAttr")
		notString   = types.Long(42)
		value       = types.Long(42)
	)

	tests := []struct {
		name   string
		lhs    Evaler
		rhs    Evaler
		result types.Value
		err    error
	}{
		{"ObjectTypeError",
			newLiteralEval(types.True),
			newLiteralEval(knownTag),
			zeroValue(),
			ErrType},
		{"SubjectTypeError",
			newLiteralEval(types.NewEntityUID(knownType, knownID)),
			newLiteralEval(notString),
			zeroValue(),
			ErrType},
		{"TagOnRecord",
			newLiteralEval(types.NewRecord(nil)),
			newLiteralEval(knownTag),
			zeroValue(),
			ErrType},
		{"ProgrammaticTag",
			newLiteralEval(types.NewEntityUID(knownType, knownID)),
			newAttributeAccessEval(newLiteralEval(types.NewEntityUID(knownType, knownID)), knownAttr),
			types.True,
			nil,
		},
		{"KnownTag",
			newLiteralEval(types.NewEntityUID(knownType, knownID)),
			newLiteralEval(knownTag),
			types.True,
			nil},
		{"UnknownTag",
			newLiteralEval(types.NewEntityUID(knownType, knownID)),
			newLiteralEval(unknownTag),
			types.False,
			nil},
		{"UnknownEntity",
			newLiteralEval(types.NewEntityUID(unknownType, unkownID)),
			newLiteralEval(knownTag),
			types.False,
			nil},
		{"UnspecifiedEntity",
			newLiteralEval(types.NewEntityUID("", "")),
			newLiteralEval(knownTag),
			types.False,
			nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newHasTagEval(tt.lhs, tt.rhs)
			entity := types.Entity{
				UID:        types.NewEntityUID(knownType, knownID),
				Tags:       types.NewRecord(types.RecordMap{knownTag: value}),
				Attributes: types.NewRecord(types.RecordMap{knownAttr: knownTag}),
			}
			v, err := n.Eval(Env{
				Entities: types.EntityMap{
					entity.UID: entity,
				},
			})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})

	}
}

func TestLikeNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		str     Evaler
		pattern string
		result  types.Value
		err     error
	}{
		{"leftError", newErrorEval(errTest), `"foo"`, zeroValue(), errTest},
		{"leftTypeError", newLiteralEval(types.True), `"foo"`, zeroValue(), ErrType},
		{"noMatch", newLiteralEval(types.String("test")), `"zebra"`, types.False, nil},
		{"match", newLiteralEval(types.String("test")), `"*es*"`, types.True, nil},

		{"case-1", newLiteralEval(types.String("eggs")), `"ham*"`, types.False, nil},
		{"case-2", newLiteralEval(types.String("eggs")), `"*ham"`, types.False, nil},
		{"case-3", newLiteralEval(types.String("eggs")), `"*ham*"`, types.False, nil},
		{"case-4", newLiteralEval(types.String("ham and eggs")), `"ham*"`, types.True, nil},
		{"case-5", newLiteralEval(types.String("ham and eggs")), `"*ham"`, types.False, nil},
		{"case-6", newLiteralEval(types.String("ham and eggs")), `"*ham*"`, types.True, nil},
		{"case-7", newLiteralEval(types.String("ham and eggs")), `"*h*a*m*"`, types.True, nil},
		{"case-8", newLiteralEval(types.String("eggs and ham")), `"ham*"`, types.False, nil},
		{"case-9", newLiteralEval(types.String("eggs and ham")), `"*ham"`, types.True, nil},
		{"case-10", newLiteralEval(types.String("eggs, ham, and spinach")), `"ham*"`, types.False, nil},
		{"case-11", newLiteralEval(types.String("eggs, ham, and spinach")), `"*ham"`, types.False, nil},
		{"case-12", newLiteralEval(types.String("eggs, ham, and spinach")), `"*ham*"`, types.True, nil},
		{"case-13", newLiteralEval(types.String("Gotham")), `"ham*"`, types.False, nil},
		{"case-14", newLiteralEval(types.String("Gotham")), `"*ham"`, types.True, nil},
		{"case-15", newLiteralEval(types.String("ham")), `"ham"`, types.True, nil},
		{"case-16", newLiteralEval(types.String("ham")), `"ham*"`, types.True, nil},
		{"case-17", newLiteralEval(types.String("ham")), `"*ham"`, types.True, nil},
		{"case-18", newLiteralEval(types.String("ham")), `"*h*a*m*"`, types.True, nil},
		{"case-19", newLiteralEval(types.String("ham and ham")), `"ham*"`, types.True, nil},
		{"case-20", newLiteralEval(types.String("ham and ham")), `"*ham"`, types.True, nil},
		{"case-21", newLiteralEval(types.String("ham")), `"*ham and eggs*"`, types.False, nil},
		{"case-22", newLiteralEval(types.String("\\afterslash")), `"\\*"`, types.True, nil},
		{"case-23", newLiteralEval(types.String("string\\with\\backslashes")), `"string\\with\\backslashes"`, types.True, nil},
		{"case-24", newLiteralEval(types.String("string\\with\\backslashes")), `"string*with*backslashes"`, types.True, nil},
		{"case-25", newLiteralEval(types.String("string*with*stars")), `"string\*with\*stars"`, types.True, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pat, err := parser.ParsePattern(tt.pattern[1 : len(tt.pattern)-1])
			testutil.OK(t, err)
			n := newLikeEval(tt.str, pat)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestVariableNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		env      Env
		variable types.String
		result   types.Value
	}{
		{"principal",
			Env{Principal: types.String("foo")},
			consts.Principal,
			types.String("foo")},
		{"action",
			Env{Action: types.String("bar")},
			consts.Action,
			types.String("bar")},
		{"resource",
			Env{Resource: types.String("baz")},
			consts.Resource,
			types.String("baz")},
		{"context",
			Env{Context: types.String("frob")},
			consts.Context,
			types.String("frob")},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newVariableEval(tt.variable)
			v, err := n.Eval(tt.env)
			testutil.OK(t, err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestEntityIn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		lhs     string
		rhs     []string
		parents map[string][]string
		result  bool
	}{
		{
			"reflexive",
			`person::"fred"`,
			[]string{`person::"fred"`},
			map[string][]string{},
			true,
		},
		{
			"simpleFalse",
			`person::"fred"`,
			[]string{`person::"jane"`},
			map[string][]string{},
			false,
		},
		{
			"oneLevelTrue",
			`person::"amy"`,
			[]string{`club::"running"`, `club::"rowing"`},
			map[string][]string{
				`person::"amy"`:    {`club::"dancing"`, `club::"rowing"`},
				`person::"fred"`:   {`club::"chess"`, `club::"rowing"`},
				`person::"brenda"`: {`club::"chess"`},
			},
			true,
		},
		{
			"oneLevelFalse",
			`person::"noah"`,
			[]string{`club::"chess"`, `club::"dancing"`},
			map[string][]string{
				`person::"amy"`:    {`club::"dancing"`, `club::"rowing"`},
				`person::"fred"`:   {`club::"chess"`, `club::"rowing"`},
				`person::"brenda"`: {`club::"chess"`},
			},
			false,
		},
		{
			"oneLevelFalse2",
			`person::"fred"`,
			[]string{`club::"sewing"`, `club::"dancing"`},
			map[string][]string{
				`person::"amy"`:    {`club::"dancing"`, `club::"rowing"`},
				`person::"fred"`:   {`club::"chess"`, `club::"rowing"`},
				`person::"brenda"`: {`club::"chess"`},
			},
			false,
		},
		{
			"twoLevelTrue",
			`person::"brenda"`,
			[]string{`category::"game"`},
			map[string][]string{
				`person::"amy"`:    {`club::"dancing"`, `club::"rowing"`},
				`person::"fred"`:   {`club::"chess"`, `club::"rowing"`},
				`person::"brenda"`: {`club::"chess"`},
				`club::"chess"`:    {`category::"game"`},
			},
			true,
		},
		{
			"loopFalse",
			`level::1::"a"`,
			[]string{`level::3::"z"`},
			map[string][]string{
				`level::1::"a"`: {`level::2::"a"`, `level::2::"b"`},
				`level::1::"b"`: {`level::2::"a"`, `level::2::"b"`},
				`level::2::"a"`: {`level::3::"a"`, `level::3::"b"`},
				`level::2::"b"`: {`level::3::"a"`, `level::3::"b"`},
				`level::3::"a"`: {`level::1::"a"`, `level::1::"b"`},
				`level::3::"b"`: {`level::1::"a"`, `level::1::"b"`},
			},
			false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var rhs []types.EntityUID
			for _, v := range tt.rhs {
				rhs = append(rhs, strEnt(v))
			}
			entityMap := types.EntityMap{}
			for k, p := range tt.parents {
				var ps []types.EntityUID
				for _, pp := range p {
					ps = append(ps, strEnt(pp))
				}
				uid := strEnt(k)
				entityMap[uid] = types.Entity{
					UID:     uid,
					Parents: types.NewEntityUIDSet(ps...),
				}
			}
			res := entityInSet(Env{Entities: entityMap}, strEnt(tt.lhs), types.NewEntityUIDSet(rhs...))
			testutil.Equals(t, res, tt.result)
		})
	}
	// This test will run for a very long time (O(2^100)) if there isn't caching.
	t.Run("exponentialWithoutCaching", func(t *testing.T) {
		t.Parallel()
		entityMap := types.EntityMap{}
		for i := 0; i < 100; i++ {
			p := types.NewEntityUIDSet(
				types.NewEntityUID(types.EntityType(fmt.Sprint(i+1)), "1"),
				types.NewEntityUID(types.EntityType(fmt.Sprint(i+1)), "2"),
			)
			uid1 := types.NewEntityUID(types.EntityType(fmt.Sprint(i)), "1")
			entityMap[uid1] = types.Entity{
				UID:     uid1,
				Parents: p,
			}
			uid2 := types.NewEntityUID(types.EntityType(fmt.Sprint(i)), "2")
			entityMap[uid2] = types.Entity{
				UID:     uid2,
				Parents: p,
			}

		}
		res := entityInSet(
			Env{Entities: entityMap},
			types.NewEntityUID("0", "1"),
			types.NewEntityUIDSet(types.NewEntityUID("0", "3")),
		)
		testutil.Equals(t, res, false)
	})
}

func TestIsNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		lhs    Evaler
		rhs    types.EntityType
		result types.Value
		err    error
	}{
		{"happyEq", newLiteralEval(types.NewEntityUID("X", "z")), types.EntityType("X"), types.True, nil},
		{"happyNeq", newLiteralEval(types.NewEntityUID("X", "z")), types.EntityType("Y"), types.False, nil},
		{"badLhs", newLiteralEval(types.Long(42)), types.EntityType("X"), zeroValue(), ErrType},
		{"errLhs", newErrorEval(errTest), types.EntityType("X"), zeroValue(), errTest},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := newIsEval(tt.lhs, tt.rhs).Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, got, tt.result)
		})
	}
}

func TestInNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		lhs, rhs Evaler
		parents  map[string][]string
		result   types.Value
		err      error
	}{
		{
			"LhsError",
			newErrorEval(errTest),
			newLiteralEval(types.Set{}),
			map[string][]string{},
			zeroValue(),
			errTest,
		},
		{
			"LhsTypeError",
			newLiteralEval(types.String("foo")),
			newLiteralEval(types.Set{}),
			map[string][]string{},
			zeroValue(),
			ErrType,
		},
		{
			"RhsError",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			newErrorEval(errTest),
			map[string][]string{},
			zeroValue(),
			errTest,
		},
		{
			"RhsTypeError1",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			newLiteralEval(types.String("foo")),
			map[string][]string{},
			zeroValue(),
			ErrType,
		},
		{
			"RhsTypeError2",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			newLiteralEval(types.NewSet(
				types.String("foo"),
			)),
			map[string][]string{},
			zeroValue(),
			ErrType,
		},
		{
			"Reflexive1",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			newLiteralEval(types.NewEntityUID("human", "joe")),
			map[string][]string{},
			types.True,
			nil,
		},
		{
			"Reflexive2",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			newLiteralEval(types.NewSet(
				types.NewEntityUID("human", "joe"),
			)),
			map[string][]string{},
			types.True,
			nil,
		},
		{
			"BasicTrue",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			newLiteralEval(types.NewEntityUID("kingdom", "animal")),
			map[string][]string{
				`human::"joe"`:     {`species::"human"`},
				`species::"human"`: {`kingdom::"animal"`},
			},
			types.True,
			nil,
		},
		{
			"BasicFalse",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			newLiteralEval(types.NewEntityUID("kingdom", "plant")),
			map[string][]string{
				`human::"joe"`:     {`species::"human"`},
				`species::"human"`: {`kingdom::"animal"`},
			},
			types.False,
			nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newInEval(tt.lhs, tt.rhs)
			entityMap := types.EntityMap{}
			for k, p := range tt.parents {
				var ps []types.EntityUID
				for _, pp := range p {
					ps = append(ps, strEnt(pp))
				}
				uid := strEnt(k)
				entityMap[uid] = types.Entity{
					UID:     uid,
					Parents: types.NewEntityUIDSet(ps...),
				}
			}
			ec := Env{Entities: entityMap}
			v, err := n.Eval(ec)
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestIsInNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		lhs     Evaler
		is      types.EntityType
		rhs     Evaler
		parents map[string][]string
		result  types.Value
		err     error
	}{
		{
			"LhsError",
			newErrorEval(errTest),
			"human",
			newLiteralEval(types.Set{}),
			map[string][]string{},
			zeroValue(),
			errTest,
		},
		{
			"LhsTypeError",
			newLiteralEval(types.String("foo")),
			"human",
			newLiteralEval(types.Set{}),
			map[string][]string{},
			zeroValue(),
			ErrType,
		},
		{
			"RhsError",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			"human",
			newErrorEval(errTest),
			map[string][]string{},
			zeroValue(),
			errTest,
		},
		{
			"RhsTypeError1",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			"human",
			newLiteralEval(types.String("foo")),
			map[string][]string{},
			zeroValue(),
			ErrType,
		},
		{
			"RhsTypeError2",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			"human",
			newLiteralEval(types.NewSet(
				types.String("foo"),
			)),
			map[string][]string{},
			zeroValue(),
			ErrType,
		},
		{
			"Reflexive1",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			"human",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			map[string][]string{},
			types.True,
			nil,
		},
		{
			"Reflexive2",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			"human",
			newLiteralEval(types.NewSet(
				types.NewEntityUID("human", "joe"),
			)),
			map[string][]string{},
			types.True,
			nil,
		},
		{
			"BasicTrue",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			"human",
			newLiteralEval(types.NewEntityUID("kingdom", "animal")),
			map[string][]string{
				`human::"joe"`:     {`species::"human"`},
				`species::"human"`: {`kingdom::"animal"`},
			},
			types.True,
			nil,
		},
		{
			"BasicFalse",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			"human",
			newLiteralEval(types.NewEntityUID("kingdom", "plant")),
			map[string][]string{
				`human::"joe"`:     {`species::"human"`},
				`species::"human"`: {`kingdom::"animal"`},
			},
			types.False,
			nil,
		},
		{
			"wrongType",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			"bananas",
			newLiteralEval(types.NewEntityUID("kingdom", "animal")),
			map[string][]string{
				`human::"joe"`:     {`species::"human"`},
				`species::"human"`: {`kingdom::"animal"`},
			},
			types.False,
			nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newIsInEval(tt.lhs, tt.is, tt.rhs)
			entityMap := types.EntityMap{}
			for k, p := range tt.parents {
				var ps []types.EntityUID
				for _, pp := range p {
					ps = append(ps, strEnt(pp))
				}
				uid := strEnt(k)
				entityMap[uid] = types.Entity{
					UID:     uid,
					Parents: types.NewEntityUIDSet(ps...),
				}
			}
			ec := Env{Entities: entityMap}
			v, err := n.Eval(ec)
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestDecimalLiteralNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		arg    Evaler
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), zeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), zeroValue(), ErrType},
		{"DecimalError", newLiteralEval(types.String("frob")), zeroValue(), internal.ErrDecimal},
		{"Success", newLiteralEval(types.String("1.0")), testutil.Must(types.NewDecimalFromInt(1)), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newDecimalLiteralEval(tt.arg)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestIPLiteralNode(t *testing.T) {
	t.Parallel()
	ipv6Loopback, err := types.ParseIPAddr("::1")
	testutil.OK(t, err)
	tests := []struct {
		name   string
		arg    Evaler
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), zeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), zeroValue(), ErrType},
		{"IPError", newLiteralEval(types.String("not-an-IP-address")), zeroValue(), internal.ErrIP},
		{"Success", newLiteralEval(types.String("::1/128")), ipv6Loopback, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newIPLiteralEval(tt.arg)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestIPTestNode(t *testing.T) {
	t.Parallel()
	ipv4Loopback, err := types.ParseIPAddr("127.0.0.1")
	testutil.OK(t, err)
	ipv6Loopback, err := types.ParseIPAddr("::1")
	testutil.OK(t, err)
	ipv4Multicast, err := types.ParseIPAddr("224.0.0.1")
	testutil.OK(t, err)
	tests := []struct {
		name   string
		lhs    Evaler
		rhs    ipTestType
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), ipTestIPv4, zeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), ipTestIPv4, zeroValue(), ErrType},
		{"IPv4True", newLiteralEval(ipv4Loopback), ipTestIPv4, types.True, nil},
		{"IPv4False", newLiteralEval(ipv6Loopback), ipTestIPv4, types.False, nil},
		{"IPv6True", newLiteralEval(ipv6Loopback), ipTestIPv6, types.True, nil},
		{"IPv6False", newLiteralEval(ipv4Loopback), ipTestIPv6, types.False, nil},
		{"LoopbackTrue", newLiteralEval(ipv6Loopback), ipTestLoopback, types.True, nil},
		{"LoopbackFalse", newLiteralEval(ipv4Multicast), ipTestLoopback, types.False, nil},
		{"MulticastTrue", newLiteralEval(ipv4Multicast), ipTestMulticast, types.True, nil},
		{"MulticastFalse", newLiteralEval(ipv6Loopback), ipTestMulticast, types.False, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newIPTestEval(tt.lhs, tt.rhs)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestIPIsInRangeNode(t *testing.T) {
	t.Parallel()
	ipv4A, err := types.ParseIPAddr("1.2.3.4")
	testutil.OK(t, err)
	ipv4B, err := types.ParseIPAddr("1.2.3.0/24")
	testutil.OK(t, err)
	ipv4C, err := types.ParseIPAddr("1.2.4.0/24")
	testutil.OK(t, err)
	tests := []struct {
		name     string
		lhs, rhs Evaler
		result   types.Value
		err      error
	}{
		{"LhsError", newErrorEval(errTest), newLiteralEval(ipv4A), zeroValue(), errTest},
		{"LhsTypeError", newLiteralEval(types.Long(1)), newLiteralEval(ipv4A), zeroValue(), ErrType},
		{"RhsError", newLiteralEval(ipv4A), newErrorEval(errTest), zeroValue(), errTest},
		{"RhsTypeError", newLiteralEval(ipv4A), newLiteralEval(types.Long(1)), zeroValue(), ErrType},
		{"AA", newLiteralEval(ipv4A), newLiteralEval(ipv4A), types.True, nil},
		{"AB", newLiteralEval(ipv4A), newLiteralEval(ipv4B), types.True, nil},
		{"BA", newLiteralEval(ipv4B), newLiteralEval(ipv4A), types.False, nil},
		{"AC", newLiteralEval(ipv4A), newLiteralEval(ipv4C), types.False, nil},
		{"CA", newLiteralEval(ipv4C), newLiteralEval(ipv4A), types.False, nil},
		{"BC", newLiteralEval(ipv4B), newLiteralEval(ipv4C), types.False, nil},
		{"CB", newLiteralEval(ipv4C), newLiteralEval(ipv4B), types.False, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newIPIsInRangeEval(tt.lhs, tt.rhs)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestCedarString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		in         types.Value
		wantString string
		wantCedar  string
	}{
		{"string", types.String("hello"), `hello`, `"hello"`},
		{"number", types.Long(42), `42`, `42`},
		{"bool", types.True, `true`, `true`},
		{"record", types.NewRecord(types.RecordMap{"a": types.Long(42), "b": types.Long(43)}), `{"a":42, "b":43}`, `{"a":42, "b":43}`},
		{"set", types.NewSet(types.Long(42), types.Long(43)), `[42, 43]`, `[42, 43]`},
		{"singleIP", types.IPAddr(netip.MustParsePrefix("192.168.0.42/32")), `192.168.0.42`, `ip("192.168.0.42")`},
		{"ipPrefix", types.IPAddr(netip.MustParsePrefix("192.168.0.42/24")), `192.168.0.42/24`, `ip("192.168.0.42/24")`},
		{"decimal", testutil.Must(types.NewDecimal(12345678, -4)), `1234.5678`, `decimal("1234.5678")`},
		{"duration", types.NewDuration(1 * time.Millisecond), `1ms`, `duration("1ms")`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotString := tt.in.String()
			testutil.Equals(t, gotString, tt.wantString)
			gotCedar := string(tt.in.MarshalCedar())
			testutil.Equals(t, gotCedar, tt.wantCedar)
		})
	}
}

func TestDatetimeLiteralNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		arg    Evaler
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), zeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), zeroValue(), ErrType},
		{"DatetimeError", newLiteralEval(types.String("frob")), zeroValue(), internal.ErrDatetime},
		{"Success", newLiteralEval(types.String("1970-01-01")), types.NewDatetime(time.UnixMilli(0)), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newDatetimeLiteralEval(tt.arg)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestDatetimeToDate(t *testing.T) {
	t.Parallel()
	aTime, err := types.ParseDatetime("1970-01-02T10:00:00Z")
	testutil.OK(t, err)

	tests := []struct {
		name   string
		arg    Evaler
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), zeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), zeroValue(), ErrType},
		{"Success", newLiteralEval(aTime), types.NewDatetime(time.UnixMilli(24 * 60 * 60 * 1000)), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newToDateEval(tt.arg)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestDatetimeDurationSince(t *testing.T) {
	t.Parallel()
	baseTime, err := types.ParseDatetime("1970-01-01T01:00:00Z")
	testutil.OK(t, err)
	endTime, err := types.ParseDatetime("1970-01-01T00:00:00Z")
	testutil.OK(t, err)
	dur := types.NewDuration(1 * time.Hour)
	bad := types.Long(1)

	tests := []struct {
		name   string
		lhs    Evaler
		rhs    Evaler
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), newLiteralEval(bad), zeroValue(), errTest},
		{"TypeError", newLiteralEval(bad), newLiteralEval(endTime), zeroValue(), ErrType},
		{"ArgTypeError", newLiteralEval(baseTime), newLiteralEval(bad), zeroValue(), ErrType},
		{"Success", newLiteralEval(baseTime), newLiteralEval(endTime), dur, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newDurationSinceEval(tt.lhs, tt.rhs)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestDatetimeOffset(t *testing.T) {
	t.Parallel()
	baseTime, err := types.ParseDatetime("1970-01-01T00:00:00Z")
	testutil.OK(t, err)
	endTime, err := types.ParseDatetime("1970-01-01T01:00:00Z")
	testutil.OK(t, err)
	dur := types.NewDuration(1 * time.Hour)
	bad := types.Long(1)

	tests := []struct {
		name   string
		lhs    Evaler
		rhs    Evaler
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), newLiteralEval(bad), zeroValue(), errTest},
		{"TypeError", newLiteralEval(bad), newLiteralEval(dur), zeroValue(), ErrType},
		{"ArgTypeError", newLiteralEval(baseTime), newLiteralEval(bad), zeroValue(), ErrType},
		{"Success", newLiteralEval(baseTime), newLiteralEval(dur), endTime, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newOffsetEval(tt.lhs, tt.rhs)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestDatetimeToTime(t *testing.T) {
	t.Parallel()
	aTime, err := types.ParseDatetime("1970-01-01T10:00:00Z")
	testutil.OK(t, err)

	tests := []struct {
		name   string
		arg    Evaler
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), zeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), zeroValue(), ErrType},
		{"Success", newLiteralEval(aTime), types.NewDuration(10 * time.Hour), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newToTimeEval(tt.arg)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestDurationLiteralNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		arg    Evaler
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), zeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), zeroValue(), ErrType},
		{"DurationError", newLiteralEval(types.String("frob")), zeroValue(), internal.ErrDuration},
		{"Success", newLiteralEval(types.String("1h")), types.NewDuration(1 * time.Hour), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newDurationLiteralEval(tt.arg)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestDurationToMilliseconds(t *testing.T) {
	t.Parallel()
	oneDay, err := types.ParseDuration("1d")
	testutil.OK(t, err)

	tests := []struct {
		name   string
		arg    Evaler
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), zeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), zeroValue(), ErrType},
		{"Success", newLiteralEval(oneDay), types.Long(24 * 60 * 60 * 1000), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newToMillisecondsEval(tt.arg)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestDurationToSeconds(t *testing.T) {
	t.Parallel()
	oneDay, err := types.ParseDuration("1d")
	testutil.OK(t, err)

	tests := []struct {
		name   string
		arg    Evaler
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), zeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), zeroValue(), ErrType},
		{"Success", newLiteralEval(oneDay), types.Long(24 * 60 * 60), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newToSecondsEval(tt.arg)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestDurationToMinutes(t *testing.T) {
	t.Parallel()
	oneDay, err := types.ParseDuration("1d")
	testutil.OK(t, err)

	tests := []struct {
		name   string
		arg    Evaler
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), zeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), zeroValue(), ErrType},
		{"Success", newLiteralEval(oneDay), types.Long(24 * 60), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newToMinutesEval(tt.arg)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestDurationToHours(t *testing.T) {
	t.Parallel()
	oneDay, err := types.ParseDuration("1d")
	testutil.OK(t, err)

	tests := []struct {
		name   string
		arg    Evaler
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), zeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), zeroValue(), ErrType},
		{"Success", newLiteralEval(oneDay), types.Long(24), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newToHoursEval(tt.arg)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}

func TestDurationToDays(t *testing.T) {
	t.Parallel()
	oneDay, err := types.ParseDuration("1d")
	testutil.OK(t, err)

	tests := []struct {
		name   string
		arg    Evaler
		result types.Value
		err    error
	}{
		{"Error", newErrorEval(errTest), zeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), zeroValue(), ErrType},
		{"Success", newLiteralEval(oneDay), types.Long(1), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newToDaysEval(tt.arg)
			v, err := n.Eval(Env{})
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, v, tt.result)
		})
	}
}
