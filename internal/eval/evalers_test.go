package eval

import (
	"fmt"
	"net/netip"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/entities"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

var errTest = fmt.Errorf("test error")

// not a real parser
func strEnt(v string) types.EntityUID {
	p := strings.Split(v, "::\"")
	return types.EntityUID{Type: p[0], ID: p[1][:len(p[1])-1]}
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
				n := newOrNode(newLiteralEval(types.Boolean(tt.lhs)), newLiteralEval(types.Boolean(tt.rhs)))
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
			})
		}
	}

	t.Run("TrueXShortCircuit", func(t *testing.T) {
		t.Parallel()
		n := newOrNode(
			newLiteralEval(types.True), newLiteralEval(types.Long(1)))
		v, err := n.Eval(&Context{})
		testutil.OK(t, err)
		types.AssertBoolValue(t, v, true)
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
				n := newOrNode(tt.lhs, tt.rhs)
				_, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
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
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
			})
		}
	}

	t.Run("FalseXShortCircuit", func(t *testing.T) {
		t.Parallel()
		n := newAndEval(
			newLiteralEval(types.False), newLiteralEval(types.Long(1)))
		v, err := n.Eval(&Context{})
		testutil.OK(t, err)
		types.AssertBoolValue(t, v, false)
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
				_, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
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
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
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
				_, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
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
		v, err := n.Eval(&Context{})
		testutil.OK(t, err)
		types.AssertLongValue(t, v, 3)
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
			_, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
		})
	}
}

func TestSubtractNode(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		n := newSubtractEval(newLiteralEval(types.Long(1)), newLiteralEval(types.Long(2)))
		v, err := n.Eval(&Context{})
		testutil.OK(t, err)
		types.AssertLongValue(t, v, -1)
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
			_, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
		})
	}
}

func TestMultiplyNode(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		n := newMultiplyEval(newLiteralEval(types.Long(-3)), newLiteralEval(types.Long(2)))
		v, err := n.Eval(&Context{})
		testutil.OK(t, err)
		types.AssertLongValue(t, v, -6)
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
			_, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
		})
	}
}

func TestNegateNode(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		n := newNegateEval(newLiteralEval(types.Long(-3)))
		v, err := n.Eval(&Context{})
		testutil.OK(t, err)
		types.AssertLongValue(t, v, 3)
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
			_, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
		})
	}
}

func TestLongLessThanNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			lhs, rhs int64
			result   bool
		}{
			{-1, -1, false},
			{-1, 0, true},
			{-1, 1, true},
			{0, -1, false},
			{0, 0, false},
			{0, 1, true},
			{1, -1, false},
			{1, 0, false},
			{1, 1, false},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%v<%v", tt.lhs, tt.rhs), func(t *testing.T) {
				t.Parallel()
				n := newLongLessThanEval(
					newLiteralEval(types.Long(tt.lhs)), newLiteralEval(types.Long(tt.rhs)))
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Long(0)), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Long(0)), ErrType},
			{"RhsError", newLiteralEval(types.Long(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Long(0)), newLiteralEval(types.True), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newLongLessThanEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
			})
		}
	}
}

func TestLongLessThanOrEqualNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			lhs, rhs int64
			result   bool
		}{
			{-1, -1, true},
			{-1, 0, true},
			{-1, 1, true},
			{0, -1, false},
			{0, 0, true},
			{0, 1, true},
			{1, -1, false},
			{1, 0, false},
			{1, 1, true},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%v<=%v", tt.lhs, tt.rhs), func(t *testing.T) {
				t.Parallel()
				n := newLongLessThanOrEqualEval(
					newLiteralEval(types.Long(tt.lhs)), newLiteralEval(types.Long(tt.rhs)))
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Long(0)), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Long(0)), ErrType},
			{"RhsError", newLiteralEval(types.Long(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Long(0)), newLiteralEval(types.True), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newLongLessThanOrEqualEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
			})
		}
	}
}

func TestLongGreaterThanNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			lhs, rhs int64
			result   bool
		}{
			{-1, -1, false},
			{-1, 0, false},
			{-1, 1, false},
			{0, -1, true},
			{0, 0, false},
			{0, 1, false},
			{1, -1, true},
			{1, 0, true},
			{1, 1, false},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%v>%v", tt.lhs, tt.rhs), func(t *testing.T) {
				t.Parallel()
				n := newLongGreaterThanEval(
					newLiteralEval(types.Long(tt.lhs)), newLiteralEval(types.Long(tt.rhs)))
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Long(0)), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Long(0)), ErrType},
			{"RhsError", newLiteralEval(types.Long(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Long(0)), newLiteralEval(types.True), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newLongGreaterThanEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
			})
		}
	}
}

func TestLongGreaterThanOrEqualNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			lhs, rhs int64
			result   bool
		}{
			{-1, -1, true},
			{-1, 0, false},
			{-1, 1, false},
			{0, -1, true},
			{0, 0, true},
			{0, 1, false},
			{1, -1, true},
			{1, 0, true},
			{1, 1, true},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%v>=%v", tt.lhs, tt.rhs), func(t *testing.T) {
				t.Parallel()
				n := newLongGreaterThanOrEqualEval(
					newLiteralEval(types.Long(tt.lhs)), newLiteralEval(types.Long(tt.rhs)))
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Long(0)), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Long(0)), ErrType},
			{"RhsError", newLiteralEval(types.Long(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Long(0)), newLiteralEval(types.True), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newLongGreaterThanOrEqualEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
			})
		}
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
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Decimal(0)), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Decimal(0)), ErrType},
			{"RhsError", newLiteralEval(types.Decimal(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Decimal(0)), newLiteralEval(types.True), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newDecimalLessThanEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
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
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Decimal(0)), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Decimal(0)), ErrType},
			{"RhsError", newLiteralEval(types.Decimal(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Decimal(0)), newLiteralEval(types.True), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newDecimalLessThanOrEqualEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
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
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Decimal(0)), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Decimal(0)), ErrType},
			{"RhsError", newLiteralEval(types.Decimal(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Decimal(0)), newLiteralEval(types.True), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newDecimalGreaterThanEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
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
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs Evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(types.Decimal(0)), errTest},
			{"LhsTypeError", newLiteralEval(types.True), newLiteralEval(types.Decimal(0)), ErrType},
			{"RhsError", newLiteralEval(types.Decimal(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(types.Decimal(0)), newLiteralEval(types.True), ErrType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newDecimalGreaterThanOrEqualEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
			})
		}
	}
}

func TestIfThenElseNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		if_, then, else_ Evaler
		result           types.Value
		err              error
	}{
		{"Then", newLiteralEval(types.True), newLiteralEval(types.Long(42)),
			newLiteralEval(types.Long(-1)), types.Long(42),
			nil},
		{"Else", newLiteralEval(types.False), newLiteralEval(types.Long(-1)),
			newLiteralEval(types.Long(42)), types.Long(42),
			nil},
		{"Err", newErrorEval(errTest), newLiteralEval(types.ZeroValue()), newLiteralEval(types.ZeroValue()), types.ZeroValue(),
			errTest},
		{"ErrType", newLiteralEval(types.Long(123)), newLiteralEval(types.ZeroValue()), newLiteralEval(types.ZeroValue()), types.ZeroValue(),
			ErrType},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newIfThenElseEval(tt.if_, tt.then, tt.else_)
			v, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
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
		{"leftErr", newErrorEval(errTest), newLiteralEval(types.ZeroValue()), types.ZeroValue(), errTest},
		{"rightErr", newLiteralEval(types.ZeroValue()), newErrorEval(errTest), types.ZeroValue(), errTest},
		{"typesNotEqual", newLiteralEval(types.Long(1)), newLiteralEval(types.True), types.False, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newEqualEval(tt.lhs, tt.rhs)
			v, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
			types.AssertValue(t, v, tt.result)
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
		{"leftErr", newErrorEval(errTest), newLiteralEval(types.ZeroValue()), types.ZeroValue(), errTest},
		{"rightErr", newLiteralEval(types.ZeroValue()), newErrorEval(errTest), types.ZeroValue(), errTest},
		{"typesNotEqual", newLiteralEval(types.Long(1)), newLiteralEval(types.True), types.True, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newNotEqualEval(tt.lhs, tt.rhs)
			v, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
			types.AssertValue(t, v, tt.result)
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
		{"errorNode", []Evaler{newErrorEval(errTest)}, types.ZeroValue(), errTest},
		{"nested",
			[]Evaler{
				newLiteralEval(types.True),
				newLiteralEval(types.Set{
					types.False,
					types.Long(1),
				}),
				newLiteralEval(types.Long(10)),
			},
			types.Set{
				types.True,
				types.Set{
					types.False,
					types.Long(1),
				},
				types.Long(10),
			},
			nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newSetLiteralEval(tt.elems)
			v, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
			types.AssertValue(t, v, tt.result)
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
				v, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
				types.AssertZeroValue(t, v)
			})
		}
	}
	{
		empty := types.Set{}
		trueAndOne := types.Set{types.True, types.Long(1)}
		nested := types.Set{trueAndOne, types.False, types.Long(2)}

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
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
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
				v, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
				types.AssertZeroValue(t, v)
			})
		}
	}
	{
		empty := types.Set{}
		trueOnly := types.Set{types.True}
		trueAndOne := types.Set{types.True, types.Long(1)}
		nested := types.Set{trueAndOne, types.False, types.Long(2)}

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
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
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
				v, err := n.Eval(&Context{})
				testutil.AssertError(t, err, tt.err)
				types.AssertZeroValue(t, v)
			})
		}
	}
	{
		empty := types.Set{}
		trueOnly := types.Set{types.True}
		trueAndOne := types.Set{types.True, types.Long(1)}
		trueAndTwo := types.Set{types.True, types.Long(2)}
		nested := types.Set{trueAndOne, types.False, types.Long(2)}

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
				v, err := n.Eval(&Context{})
				testutil.OK(t, err)
				types.AssertBoolValue(t, v, tt.result)
			})
		}
	}
}

func TestRecordLiteralNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		elems  map[string]Evaler
		result types.Value
		err    error
	}{
		{"empty", map[string]Evaler{}, types.Record{}, nil},
		{"errorNode", map[string]Evaler{"foo": newErrorEval(errTest)}, types.ZeroValue(), errTest},
		{"ok",
			map[string]Evaler{
				"foo": newLiteralEval(types.True),
				"bar": newLiteralEval(types.String("baz")),
			}, types.Record{
				"foo": types.True,
				"bar": types.String("baz"),
			}, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newRecordLiteralEval(tt.elems)
			v, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
			types.AssertValue(t, v, tt.result)
		})
	}
}

func TestAttributeAccessNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		object    Evaler
		attribute string
		result    types.Value
		err       error
	}{
		{"RecordError", newErrorEval(errTest), "foo", types.ZeroValue(), errTest},
		{"RecordTypeError", newLiteralEval(types.True), "foo", types.ZeroValue(), ErrType},
		{"UnknownAttribute",
			newLiteralEval(types.Record{}),
			"foo",
			types.ZeroValue(),
			errAttributeAccess},
		{"KnownAttribute",
			newLiteralEval(types.Record{"foo": types.Long(42)}),
			"foo",
			types.Long(42),
			nil},
		{"KnownAttributeOnEntity",
			newLiteralEval(types.NewEntityUID("knownType", "knownID")),
			"knownAttr",
			types.Long(42),
			nil},
		{"UnknownEntity",
			newLiteralEval(types.NewEntityUID("unknownType", "unknownID")),
			"unknownAttr",
			types.ZeroValue(),
			errEntityNotExist},
		{"UnspecifiedEntity",
			newLiteralEval(types.NewEntityUID("", "")),
			"knownAttr",
			types.ZeroValue(),
			errUnspecifiedEntity},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newAttributeAccessEval(tt.object, tt.attribute)
			entity := entities.Entity{
				UID:        types.NewEntityUID("knownType", "knownID"),
				Attributes: types.Record{"knownAttr": types.Long(42)},
			}
			v, err := n.Eval(&Context{
				Entities: entities.Entities{
					entity.UID: entity,
				},
			})
			testutil.AssertError(t, err, tt.err)
			types.AssertValue(t, v, tt.result)
		})
	}
}

func TestHasNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		record    Evaler
		attribute string
		result    types.Value
		err       error
	}{
		{"RecordError", newErrorEval(errTest), "foo", types.ZeroValue(), errTest},
		{"RecordTypeError", newLiteralEval(types.True), "foo", types.ZeroValue(), ErrType},
		{"UnknownAttribute",
			newLiteralEval(types.Record{}),
			"foo",
			types.False,
			nil},
		{"KnownAttribute",
			newLiteralEval(types.Record{"foo": types.Long(42)}),
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
			entity := entities.Entity{
				UID:        types.NewEntityUID("knownType", "knownID"),
				Attributes: types.Record{"knownAttr": types.Long(42)},
			}
			v, err := n.Eval(&Context{
				Entities: entities.Entities{
					entity.UID: entity,
				},
			})
			testutil.AssertError(t, err, tt.err)
			types.AssertValue(t, v, tt.result)
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
		{"leftError", newErrorEval(errTest), `"foo"`, types.ZeroValue(), errTest},
		{"leftTypeError", newLiteralEval(types.True), `"foo"`, types.ZeroValue(), ErrType},
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
			var pat types.Pattern
			err := pat.UnmarshalCedar([]byte(tt.pattern[1 : len(tt.pattern)-1]))
			testutil.OK(t, err)
			n := newLikeEval(tt.str, pat)
			v, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
			types.AssertValue(t, v, tt.result)
		})
	}
}

func TestVariableNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		context  Context
		variable variableName
		result   types.Value
	}{
		{"principal",
			Context{Principal: types.String("foo")},
			variableNamePrincipal,
			types.String("foo")},
		{"action",
			Context{Action: types.String("bar")},
			variableNameAction,
			types.String("bar")},
		{"resource",
			Context{Resource: types.String("baz")},
			variableNameResource,
			types.String("baz")},
		{"context",
			Context{Context: types.String("frob")},
			variableNameContext,
			types.String("frob")},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newVariableEval(tt.variable)
			v, err := n.Eval(&tt.context)
			testutil.OK(t, err)
			types.AssertValue(t, v, tt.result)
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
			rhs := map[types.EntityUID]struct{}{}
			for _, v := range tt.rhs {
				rhs[strEnt(v)] = struct{}{}
			}
			entityMap := entities.Entities{}
			for k, p := range tt.parents {
				var ps []types.EntityUID
				for _, pp := range p {
					ps = append(ps, strEnt(pp))
				}
				uid := strEnt(k)
				entityMap[uid] = entities.Entity{
					UID:     uid,
					Parents: ps,
				}
			}
			res := entityIn(strEnt(tt.lhs), rhs, entityMap)
			testutil.Equals(t, res, tt.result)
		})
	}
	t.Run("exponentialWithoutCaching", func(t *testing.T) {
		t.Parallel(
		// This test will run for a very long time (O(2^100)) if there isn't caching.
		)

		entityMap := entities.Entities{}
		for i := 0; i < 100; i++ {
			p := []types.EntityUID{
				types.NewEntityUID(fmt.Sprint(i+1), "1"),
				types.NewEntityUID(fmt.Sprint(i+1), "2"),
			}
			uid1 := types.NewEntityUID(fmt.Sprint(i), "1")
			entityMap[uid1] = entities.Entity{
				UID:     uid1,
				Parents: p,
			}
			uid2 := types.NewEntityUID(fmt.Sprint(i), "2")
			entityMap[uid2] = entities.Entity{
				UID:     uid2,
				Parents: p,
			}

		}
		res := entityIn(types.NewEntityUID("0", "1"), map[types.EntityUID]struct{}{types.NewEntityUID("0", "3"): {}}, entityMap)
		testutil.Equals(t, res, false)
	})
}

func TestIsNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		lhs, rhs Evaler
		result   types.Value
		err      error
	}{
		{"happyEq", newLiteralEval(types.NewEntityUID("X", "z")), newLiteralEval(types.EntityType("X")), types.True, nil},
		{"happyNeq", newLiteralEval(types.NewEntityUID("X", "z")), newLiteralEval(types.EntityType("Y")), types.False, nil},
		{"badLhs", newLiteralEval(types.Long(42)), newLiteralEval(types.EntityType("X")), types.ZeroValue(), ErrType},
		{"badRhs", newLiteralEval(types.NewEntityUID("X", "z")), newLiteralEval(types.Long(42)), types.ZeroValue(), ErrType},
		{"errLhs", newErrorEval(errTest), newLiteralEval(types.EntityType("X")), types.ZeroValue(), errTest},
		{"errRhs", newLiteralEval(types.NewEntityUID("X", "z")), newErrorEval(errTest), types.ZeroValue(), errTest},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := newIsEval(tt.lhs, tt.rhs).Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
			types.AssertValue(t, got, tt.result)
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
			types.ZeroValue(),
			errTest,
		},
		{
			"LhsTypeError",
			newLiteralEval(types.String("foo")),
			newLiteralEval(types.Set{}),
			map[string][]string{},
			types.ZeroValue(),
			ErrType,
		},
		{
			"RhsError",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			newErrorEval(errTest),
			map[string][]string{},
			types.ZeroValue(),
			errTest,
		},
		{
			"RhsTypeError1",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			newLiteralEval(types.String("foo")),
			map[string][]string{},
			types.ZeroValue(),
			ErrType,
		},
		{
			"RhsTypeError2",
			newLiteralEval(types.NewEntityUID("human", "joe")),
			newLiteralEval(types.Set{
				types.String("foo"),
			}),
			map[string][]string{},
			types.ZeroValue(),
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
			newLiteralEval(types.Set{
				types.NewEntityUID("human", "joe"),
			}),
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
			entityMap := entities.Entities{}
			for k, p := range tt.parents {
				var ps []types.EntityUID
				for _, pp := range p {
					ps = append(ps, strEnt(pp))
				}
				uid := strEnt(k)
				entityMap[uid] = entities.Entity{
					UID:     uid,
					Parents: ps,
				}
			}
			ec := Context{Entities: entityMap}
			v, err := n.Eval(&ec)
			testutil.AssertError(t, err, tt.err)
			types.AssertValue(t, v, tt.result)
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
		{"Error", newErrorEval(errTest), types.ZeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), types.ZeroValue(), ErrType},
		{"DecimalError", newLiteralEval(types.String("frob")), types.ZeroValue(), types.ErrDecimal},
		{"Success", newLiteralEval(types.String("1.0")), types.Decimal(10000), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newDecimalLiteralEval(tt.arg)
			v, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
			types.AssertValue(t, v, tt.result)
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
		{"Error", newErrorEval(errTest), types.ZeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), types.ZeroValue(), ErrType},
		{"IPError", newLiteralEval(types.String("not-an-IP-address")), types.ZeroValue(), types.ErrIP},
		{"Success", newLiteralEval(types.String("::1/128")), ipv6Loopback, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newIPLiteralEval(tt.arg)
			v, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
			types.AssertValue(t, v, tt.result)
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
		{"Error", newErrorEval(errTest), ipTestIPv4, types.ZeroValue(), errTest},
		{"TypeError", newLiteralEval(types.Long(1)), ipTestIPv4, types.ZeroValue(), ErrType},
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
			v, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
			types.AssertValue(t, v, tt.result)
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
		{"LhsError", newErrorEval(errTest), newLiteralEval(ipv4A), types.ZeroValue(), errTest},
		{"LhsTypeError", newLiteralEval(types.Long(1)), newLiteralEval(ipv4A), types.ZeroValue(), ErrType},
		{"RhsError", newLiteralEval(ipv4A), newErrorEval(errTest), types.ZeroValue(), errTest},
		{"RhsTypeError", newLiteralEval(ipv4A), newLiteralEval(types.Long(1)), types.ZeroValue(), ErrType},
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
			v, err := n.Eval(&Context{})
			testutil.AssertError(t, err, tt.err)
			types.AssertValue(t, v, tt.result)
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
		{"record", types.Record{"a": types.Long(42), "b": types.Long(43)}, `{"a": 42, "b": 43}`, `{"a": 42, "b": 43}`},
		{"set", types.Set{types.Long(42), types.Long(43)}, `[42, 43]`, `[42, 43]`},
		{"singleIP", types.IPAddr(netip.MustParsePrefix("192.168.0.42/32")), `192.168.0.42`, `ip("192.168.0.42")`},
		{"ipPrefix", types.IPAddr(netip.MustParsePrefix("192.168.0.42/24")), `192.168.0.42/24`, `ip("192.168.0.42/24")`},
		{"decimal", types.Decimal(12345678), `1234.5678`, `decimal("1234.5678")`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotString := tt.in.String()
			testutil.Equals(t, gotString, tt.wantString)
			gotCedar := tt.in.Cedar()
			testutil.Equals(t, gotCedar, tt.wantCedar)
		})
	}
}
