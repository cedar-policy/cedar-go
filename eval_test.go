package cedar

import (
	"fmt"
	"net/netip"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/x/exp/parser"
)

var errTest = fmt.Errorf("test error")

// not a real parser
func strEnt(v string) EntityUID {
	p := strings.Split(v, "::\"")
	return EntityUID{Type: p[0], ID: p[1][:len(p[1])-1]}
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
				n := newOrNode(newLiteralEval(Boolean(tt.lhs)), newLiteralEval(Boolean(tt.rhs)))
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}

	t.Run("TrueXShortCircuit", func(t *testing.T) {
		t.Parallel()
		n := newOrNode(
			newLiteralEval(Boolean(true)), newLiteralEval(Long(1)))
		v, err := n.Eval(&evalContext{})
		testutilOK(t, err)
		assertBoolValue(t, v, true)
	})

	{
		tests := []struct {
			name     string
			lhs, rhs evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(Boolean(true)), errTest},
			{"LhsTypeError", newLiteralEval(Long(1)), newLiteralEval(Boolean(true)), errType},
			{"RhsError", newLiteralEval(Boolean(false)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(Boolean(false)), newLiteralEval(Long(1)), errType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newOrNode(tt.lhs, tt.rhs)
				_, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
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
				n := newAndEval(newLiteralEval(Boolean(tt.lhs)), newLiteralEval(Boolean(tt.rhs)))
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}

	t.Run("FalseXShortCircuit", func(t *testing.T) {
		t.Parallel()
		n := newAndEval(
			newLiteralEval(Boolean(false)), newLiteralEval(Long(1)))
		v, err := n.Eval(&evalContext{})
		testutilOK(t, err)
		assertBoolValue(t, v, false)
	})

	{
		tests := []struct {
			name     string
			lhs, rhs evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(Boolean(true)), errTest},
			{"LhsTypeError", newLiteralEval(Long(1)), newLiteralEval(Boolean(true)), errType},
			{"RhsError", newLiteralEval(Boolean(true)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Long(1)), errType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newAndEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
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
				n := newNotEval(newLiteralEval(Boolean(tt.arg)))
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}

	{
		tests := []struct {
			name string
			arg  evaler
			err  error
		}{
			{"Error", newErrorEval(errTest), errTest},
			{"TypeError", newLiteralEval(Long(1)), errType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newNotEval(tt.arg)
				_, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
			})
		}
	}
}

func TestCheckedAddI64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		lhs, rhs, result Long
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
			testutilEquals(t, ok, tt.ok)
			testutilEquals(t, result, tt.result)
		})
	}
}

func TestCheckedSubI64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		lhs, rhs, result Long
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
			testutilEquals(t, ok, tt.ok)
			testutilEquals(t, result, tt.result)
		})
	}
}

func TestCheckedMulI64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		lhs, rhs, result Long
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
			testutilEquals(t, ok, tt.ok)
			testutilEquals(t, result, tt.result)
		})
	}
}

func TestCheckedNegI64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		arg, result Long
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
			testutilEquals(t, ok, tt.ok)
			testutilEquals(t, result, tt.result)
		})
	}
}

func TestAddNode(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		n := newAddEval(newLiteralEval(Long(1)), newLiteralEval(Long(2)))
		v, err := n.Eval(&evalContext{})
		testutilOK(t, err)
		assertLongValue(t, v, 3)
	})

	tests := []struct {
		name     string
		lhs, rhs evaler
		err      error
	}{
		{"LhsError", newErrorEval(errTest), newLiteralEval(Long(0)), errTest},
		{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Long(0)), errType},
		{"RhsError", newLiteralEval(Long(0)), newErrorEval(errTest), errTest},
		{"RhsTypeError", newLiteralEval(Long(0)), newLiteralEval(Boolean(true)), errType},
		{"PositiveOverflow",
			newLiteralEval(Long(9_223_372_036_854_775_807)),
			newLiteralEval(Long(1)),
			errOverflow},
		{"NegativeOverflow",
			newLiteralEval(Long(-9_223_372_036_854_775_808)),
			newLiteralEval(Long(-1)),
			errOverflow},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newAddEval(tt.lhs, tt.rhs)
			_, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
		})
	}
}

func TestSubtractNode(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		n := newSubtractEval(newLiteralEval(Long(1)), newLiteralEval(Long(2)))
		v, err := n.Eval(&evalContext{})
		testutilOK(t, err)
		assertLongValue(t, v, -1)
	})

	tests := []struct {
		name     string
		lhs, rhs evaler
		err      error
	}{
		{"LhsError", newErrorEval(errTest), newLiteralEval(Long(0)), errTest},
		{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Long(0)), errType},
		{"RhsError", newLiteralEval(Long(0)), newErrorEval(errTest), errTest},
		{"RhsTypeError", newLiteralEval(Long(0)), newLiteralEval(Boolean(true)), errType},
		{"PositiveOverflow",
			newLiteralEval(Long(9_223_372_036_854_775_807)),
			newLiteralEval(Long(-1)),
			errOverflow},
		{"NegativeOverflow",
			newLiteralEval(Long(-9_223_372_036_854_775_808)),
			newLiteralEval(Long(1)),
			errOverflow},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newSubtractEval(tt.lhs, tt.rhs)
			_, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
		})
	}
}

func TestMultiplyNode(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		n := newMultiplyEval(newLiteralEval(Long(-3)), newLiteralEval(Long(2)))
		v, err := n.Eval(&evalContext{})
		testutilOK(t, err)
		assertLongValue(t, v, -6)
	})

	tests := []struct {
		name     string
		lhs, rhs evaler
		err      error
	}{
		{"LhsError", newErrorEval(errTest), newLiteralEval(Long(0)), errTest},
		{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Long(0)), errType},
		{"RhsError", newLiteralEval(Long(0)), newErrorEval(errTest), errTest},
		{"RhsTypeError", newLiteralEval(Long(0)), newLiteralEval(Boolean(true)), errType},
		{"PositiveOverflow",
			newLiteralEval(Long(9_223_372_036_854_775_807)),
			newLiteralEval(Long(2)),
			errOverflow},
		{"NegativeOverflow",
			newLiteralEval(Long(-9_223_372_036_854_775_808)),
			newLiteralEval(Long(2)),
			errOverflow},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newMultiplyEval(tt.lhs, tt.rhs)
			_, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
		})
	}
}

func TestNegateNode(t *testing.T) {
	t.Parallel()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		n := newNegateEval(newLiteralEval(Long(-3)))
		v, err := n.Eval(&evalContext{})
		testutilOK(t, err)
		assertLongValue(t, v, 3)
	})

	tests := []struct {
		name string
		arg  evaler
		err  error
	}{
		{"Error", newErrorEval(errTest), errTest},
		{"TypeError", newLiteralEval(Boolean(true)), errType},
		{"Overflow", newLiteralEval(Long(-9_223_372_036_854_775_808)), errOverflow},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newNegateEval(tt.arg)
			_, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
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
					newLiteralEval(Long(tt.lhs)), newLiteralEval(Long(tt.rhs)))
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(Long(0)), errTest},
			{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Long(0)), errType},
			{"RhsError", newLiteralEval(Long(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(Long(0)), newLiteralEval(Boolean(true)), errType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newLongLessThanEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
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
					newLiteralEval(Long(tt.lhs)), newLiteralEval(Long(tt.rhs)))
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(Long(0)), errTest},
			{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Long(0)), errType},
			{"RhsError", newLiteralEval(Long(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(Long(0)), newLiteralEval(Boolean(true)), errType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newLongLessThanOrEqualEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
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
					newLiteralEval(Long(tt.lhs)), newLiteralEval(Long(tt.rhs)))
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(Long(0)), errTest},
			{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Long(0)), errType},
			{"RhsError", newLiteralEval(Long(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(Long(0)), newLiteralEval(Boolean(true)), errType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newLongGreaterThanEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
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
					newLiteralEval(Long(tt.lhs)), newLiteralEval(Long(tt.rhs)))
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(Long(0)), errTest},
			{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Long(0)), errType},
			{"RhsError", newLiteralEval(Long(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(Long(0)), newLiteralEval(Boolean(true)), errType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newLongGreaterThanOrEqualEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
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
				lhsd, err := ParseDecimal(tt.lhs)
				testutilOK(t, err)
				lhsv := lhsd
				rhsd, err := ParseDecimal(tt.rhs)
				testutilOK(t, err)
				rhsv := rhsd
				n := newDecimalLessThanEval(newLiteralEval(lhsv), newLiteralEval(rhsv))
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(Decimal(0)), errTest},
			{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Decimal(0)), errType},
			{"RhsError", newLiteralEval(Decimal(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(Decimal(0)), newLiteralEval(Boolean(true)), errType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newDecimalLessThanEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
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
				lhsd, err := ParseDecimal(tt.lhs)
				testutilOK(t, err)
				lhsv := lhsd
				rhsd, err := ParseDecimal(tt.rhs)
				testutilOK(t, err)
				rhsv := rhsd
				n := newDecimalLessThanOrEqualEval(newLiteralEval(lhsv), newLiteralEval(rhsv))
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(Decimal(0)), errTest},
			{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Decimal(0)), errType},
			{"RhsError", newLiteralEval(Decimal(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(Decimal(0)), newLiteralEval(Boolean(true)), errType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newDecimalLessThanOrEqualEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
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
				lhsd, err := ParseDecimal(tt.lhs)
				testutilOK(t, err)
				lhsv := lhsd
				rhsd, err := ParseDecimal(tt.rhs)
				testutilOK(t, err)
				rhsv := rhsd
				n := newDecimalGreaterThanEval(newLiteralEval(lhsv), newLiteralEval(rhsv))
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(Decimal(0)), errTest},
			{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Decimal(0)), errType},
			{"RhsError", newLiteralEval(Decimal(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(Decimal(0)), newLiteralEval(Boolean(true)), errType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newDecimalGreaterThanEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
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
				lhsd, err := ParseDecimal(tt.lhs)
				testutilOK(t, err)
				lhsv := lhsd
				rhsd, err := ParseDecimal(tt.rhs)
				testutilOK(t, err)
				rhsv := rhsd
				n := newDecimalGreaterThanOrEqualEval(newLiteralEval(lhsv), newLiteralEval(rhsv))
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}
	{
		tests := []struct {
			name     string
			lhs, rhs evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(Decimal(0)), errTest},
			{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Decimal(0)), errType},
			{"RhsError", newLiteralEval(Decimal(0)), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(Decimal(0)), newLiteralEval(Boolean(true)), errType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newDecimalGreaterThanOrEqualEval(tt.lhs, tt.rhs)
				_, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
			})
		}
	}
}

func TestIfThenElseNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		if_, then, else_ evaler
		result           Value
		err              error
	}{
		{"Then", newLiteralEval(Boolean(true)), newLiteralEval(Long(42)),
			newLiteralEval(Long(-1)), Long(42),
			nil},
		{"Else", newLiteralEval(Boolean(false)), newLiteralEval(Long(-1)),
			newLiteralEval(Long(42)), Long(42),
			nil},
		{"Err", newErrorEval(errTest), newLiteralEval(zeroValue()), newLiteralEval(zeroValue()), zeroValue(),
			errTest},
		{"ErrType", newLiteralEval(Long(123)), newLiteralEval(zeroValue()), newLiteralEval(zeroValue()), zeroValue(),
			errType},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newIfThenElseEval(tt.if_, tt.then, tt.else_)
			v, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
			testutilEquals(t, v, tt.result)
		})
	}
}

func TestEqualNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		lhs, rhs evaler
		result   Value
		err      error
	}{
		{"equals", newLiteralEval(Long(42)), newLiteralEval(Long(42)), Boolean(true), nil},
		{"notEquals", newLiteralEval(Long(42)), newLiteralEval(Long(1234)), Boolean(false), nil},
		{"leftErr", newErrorEval(errTest), newLiteralEval(zeroValue()), zeroValue(), errTest},
		{"rightErr", newLiteralEval(zeroValue()), newErrorEval(errTest), zeroValue(), errTest},
		{"typesNotEqual", newLiteralEval(Long(1)), newLiteralEval(Boolean(true)), Boolean(false), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newEqualEval(tt.lhs, tt.rhs)
			v, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
			assertValue(t, v, tt.result)
		})
	}
}

func TestNotEqualNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		lhs, rhs evaler
		result   Value
		err      error
	}{
		{"equals", newLiteralEval(Long(42)), newLiteralEval(Long(42)), Boolean(false), nil},
		{"notEquals", newLiteralEval(Long(42)), newLiteralEval(Long(1234)), Boolean(true), nil},
		{"leftErr", newErrorEval(errTest), newLiteralEval(zeroValue()), zeroValue(), errTest},
		{"rightErr", newLiteralEval(zeroValue()), newErrorEval(errTest), zeroValue(), errTest},
		{"typesNotEqual", newLiteralEval(Long(1)), newLiteralEval(Boolean(true)), Boolean(true), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newNotEqualEval(tt.lhs, tt.rhs)
			v, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
			assertValue(t, v, tt.result)
		})
	}
}

func TestSetLiteralNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		elems  []evaler
		result Value
		err    error
	}{
		{"empty", []evaler{}, Set{}, nil},
		{"errorNode", []evaler{newErrorEval(errTest)}, zeroValue(), errTest},
		{"nested",
			[]evaler{
				newLiteralEval(Boolean(true)),
				newLiteralEval(Set{
					Boolean(false),
					Long(1),
				}),
				newLiteralEval(Long(10)),
			},
			Set{
				Boolean(true),
				Set{
					Boolean(false),
					Long(1),
				},
				Long(10),
			},
			nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newSetLiteralEval(tt.elems)
			v, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
			assertValue(t, v, tt.result)
		})
	}
}

func TestContainsNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			name     string
			lhs, rhs evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(Long(0)), errTest},
			{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Long(0)), errType},
			{"RhsError", newLiteralEval(Set{}), newErrorEval(errTest), errTest},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newContainsEval(tt.lhs, tt.rhs)
				v, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
				assertZeroValue(t, v)
			})
		}
	}
	{
		empty := Set{}
		trueAndOne := Set{Boolean(true), Long(1)}
		nested := Set{trueAndOne, Boolean(false), Long(2)}

		tests := []struct {
			name     string
			lhs, rhs evaler
			result   bool
		}{
			{"empty", newLiteralEval(empty), newLiteralEval(Boolean(true)), false},
			{"trueAndOneContainsTrue", newLiteralEval(trueAndOne), newLiteralEval(Boolean(true)), true},
			{"trueAndOneContainsOne", newLiteralEval(trueAndOne), newLiteralEval(Long(1)), true},
			{"trueAndOneDoesNotContainTwo", newLiteralEval(trueAndOne), newLiteralEval(Long(2)), false},
			{"nestedContainsFalse", newLiteralEval(nested), newLiteralEval(Boolean(false)), true},
			{"nestedContainsSet", newLiteralEval(nested), newLiteralEval(trueAndOne), true},
			{"nestedDoesNotContainTrue", newLiteralEval(nested), newLiteralEval(Boolean(true)), false},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newContainsEval(tt.lhs, tt.rhs)
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}
}

func TestContainsAllNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			name     string
			lhs, rhs evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(Set{}), errTest},
			{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Set{}), errType},
			{"RhsError", newLiteralEval(Set{}), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(Set{}), newLiteralEval(Long(0)), errType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newContainsAllEval(tt.lhs, tt.rhs)
				v, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
				assertZeroValue(t, v)
			})
		}
	}
	{
		empty := Set{}
		trueOnly := Set{Boolean(true)}
		trueAndOne := Set{Boolean(true), Long(1)}
		nested := Set{trueAndOne, Boolean(false), Long(2)}

		tests := []struct {
			name     string
			lhs, rhs evaler
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
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}
}

func TestContainsAnyNode(t *testing.T) {
	t.Parallel()
	{
		tests := []struct {
			name     string
			lhs, rhs evaler
			err      error
		}{
			{"LhsError", newErrorEval(errTest), newLiteralEval(Set{}), errTest},
			{"LhsTypeError", newLiteralEval(Boolean(true)), newLiteralEval(Set{}), errType},
			{"RhsError", newLiteralEval(Set{}), newErrorEval(errTest), errTest},
			{"RhsTypeError", newLiteralEval(Set{}), newLiteralEval(Long(0)), errType},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				n := newContainsAnyEval(tt.lhs, tt.rhs)
				v, err := n.Eval(&evalContext{})
				assertError(t, err, tt.err)
				assertZeroValue(t, v)
			})
		}
	}
	{
		empty := Set{}
		trueOnly := Set{Boolean(true)}
		trueAndOne := Set{Boolean(true), Long(1)}
		trueAndTwo := Set{Boolean(true), Long(2)}
		nested := Set{trueAndOne, Boolean(false), Long(2)}

		tests := []struct {
			name     string
			lhs, rhs evaler
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
				v, err := n.Eval(&evalContext{})
				testutilOK(t, err)
				assertBoolValue(t, v, tt.result)
			})
		}
	}
}

func TestRecordLiteralNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		elems  map[string]evaler
		result Value
		err    error
	}{
		{"empty", map[string]evaler{}, Record{}, nil},
		{"errorNode", map[string]evaler{"foo": newErrorEval(errTest)}, zeroValue(), errTest},
		{"ok",
			map[string]evaler{
				"foo": newLiteralEval(Boolean(true)),
				"bar": newLiteralEval(String("baz")),
			}, Record{
				"foo": Boolean(true),
				"bar": String("baz"),
			}, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newRecordLiteralEval(tt.elems)
			v, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
			assertValue(t, v, tt.result)
		})
	}
}

func TestAttributeAccessNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		object    evaler
		attribute string
		result    Value
		err       error
	}{
		{"RecordError", newErrorEval(errTest), "foo", zeroValue(), errTest},
		{"RecordTypeError", newLiteralEval(Boolean(true)), "foo", zeroValue(), errType},
		{"UnknownAttribute",
			newLiteralEval(Record{}),
			"foo",
			zeroValue(),
			errAttributeAccess},
		{"KnownAttribute",
			newLiteralEval(Record{"foo": Long(42)}),
			"foo",
			Long(42),
			nil},
		{"KnownAttributeOnEntity",
			newLiteralEval(EntityUID{"knownType", "knownID"}),
			"knownAttr",
			Long(42),
			nil},
		{"UnknownEntity",
			newLiteralEval(EntityUID{"unknownType", "unknownID"}),
			"unknownAttr",
			zeroValue(),
			errEntityNotExist},
		{"UnspecifiedEntity",
			newLiteralEval(EntityUID{"", ""}),
			"knownAttr",
			zeroValue(),
			errUnspecifiedEntity},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newAttributeAccessEval(tt.object, tt.attribute)
			v, err := n.Eval(&evalContext{
				Entities: entitiesFromSlice([]Entity{
					{
						UID:        NewEntityUID("knownType", "knownID"),
						Attributes: Record{"knownAttr": Long(42)},
					},
				}),
			})
			assertError(t, err, tt.err)
			assertValue(t, v, tt.result)
		})
	}
}

func TestHasNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		record    evaler
		attribute string
		result    Value
		err       error
	}{
		{"RecordError", newErrorEval(errTest), "foo", zeroValue(), errTest},
		{"RecordTypeError", newLiteralEval(Boolean(true)), "foo", zeroValue(), errType},
		{"UnknownAttribute",
			newLiteralEval(Record{}),
			"foo",
			Boolean(false),
			nil},
		{"KnownAttribute",
			newLiteralEval(Record{"foo": Long(42)}),
			"foo",
			Boolean(true),
			nil},
		{"KnownAttributeOnEntity",
			newLiteralEval(EntityUID{"knownType", "knownID"}),
			"knownAttr",
			Boolean(true),
			nil},
		{"UnknownAttributeOnEntity",
			newLiteralEval(EntityUID{"knownType", "knownID"}),
			"unknownAttr",
			Boolean(false),
			nil},
		{"UnknownEntity",
			newLiteralEval(EntityUID{"unknownType", "unknownID"}),
			"unknownAttr",
			Boolean(false),
			nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newHasEval(tt.record, tt.attribute)
			v, err := n.Eval(&evalContext{
				Entities: entitiesFromSlice([]Entity{
					{
						UID:        NewEntityUID("knownType", "knownID"),
						Attributes: Record{"knownAttr": Long(42)},
					},
				}),
			})
			assertError(t, err, tt.err)
			assertValue(t, v, tt.result)
		})
	}
}

func TestLikeNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		str     evaler
		pattern string
		result  Value
		err     error
	}{
		{"leftError", newErrorEval(errTest), `"foo"`, zeroValue(), errTest},
		{"leftTypeError", newLiteralEval(Boolean(true)), `"foo"`, zeroValue(), errType},
		{"noMatch", newLiteralEval(String("test")), `"zebra"`, Boolean(false), nil},
		{"match", newLiteralEval(String("test")), `"*es*"`, Boolean(true), nil},

		{"case-1", newLiteralEval(String("eggs")), `"ham*"`, Boolean(false), nil},
		{"case-2", newLiteralEval(String("eggs")), `"*ham"`, Boolean(false), nil},
		{"case-3", newLiteralEval(String("eggs")), `"*ham*"`, Boolean(false), nil},
		{"case-4", newLiteralEval(String("ham and eggs")), `"ham*"`, Boolean(true), nil},
		{"case-5", newLiteralEval(String("ham and eggs")), `"*ham"`, Boolean(false), nil},
		{"case-6", newLiteralEval(String("ham and eggs")), `"*ham*"`, Boolean(true), nil},
		{"case-7", newLiteralEval(String("ham and eggs")), `"*h*a*m*"`, Boolean(true), nil},
		{"case-8", newLiteralEval(String("eggs and ham")), `"ham*"`, Boolean(false), nil},
		{"case-9", newLiteralEval(String("eggs and ham")), `"*ham"`, Boolean(true), nil},
		{"case-10", newLiteralEval(String("eggs, ham, and spinach")), `"ham*"`, Boolean(false), nil},
		{"case-11", newLiteralEval(String("eggs, ham, and spinach")), `"*ham"`, Boolean(false), nil},
		{"case-12", newLiteralEval(String("eggs, ham, and spinach")), `"*ham*"`, Boolean(true), nil},
		{"case-13", newLiteralEval(String("Gotham")), `"ham*"`, Boolean(false), nil},
		{"case-14", newLiteralEval(String("Gotham")), `"*ham"`, Boolean(true), nil},
		{"case-15", newLiteralEval(String("ham")), `"ham"`, Boolean(true), nil},
		{"case-16", newLiteralEval(String("ham")), `"ham*"`, Boolean(true), nil},
		{"case-17", newLiteralEval(String("ham")), `"*ham"`, Boolean(true), nil},
		{"case-18", newLiteralEval(String("ham")), `"*h*a*m*"`, Boolean(true), nil},
		{"case-19", newLiteralEval(String("ham and ham")), `"ham*"`, Boolean(true), nil},
		{"case-20", newLiteralEval(String("ham and ham")), `"*ham"`, Boolean(true), nil},
		{"case-21", newLiteralEval(String("ham")), `"*ham and eggs*"`, Boolean(false), nil},
		{"case-22", newLiteralEval(String("\\afterslash")), `"\\*"`, Boolean(true), nil},
		{"case-23", newLiteralEval(String("string\\with\\backslashes")), `"string\\with\\backslashes"`, Boolean(true), nil},
		{"case-24", newLiteralEval(String("string\\with\\backslashes")), `"string*with*backslashes"`, Boolean(true), nil},
		{"case-25", newLiteralEval(String("string*with*stars")), `"string\*with\*stars"`, Boolean(true), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pat, err := parser.NewPattern(tt.pattern)
			testutilOK(t, err)
			n := newLikeEval(tt.str, pat)
			v, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
			assertValue(t, v, tt.result)
		})
	}
}

func TestVariableNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		context  evalContext
		variable variableName
		result   Value
	}{
		{"principal",
			evalContext{Principal: String("foo")},
			variableNamePrincipal,
			String("foo")},
		{"action",
			evalContext{Action: String("bar")},
			variableNameAction,
			String("bar")},
		{"resource",
			evalContext{Resource: String("baz")},
			variableNameResource,
			String("baz")},
		{"context",
			evalContext{Context: String("frob")},
			variableNameContext,
			String("frob")},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newVariableEval(tt.variable)
			v, err := n.Eval(&tt.context)
			testutilOK(t, err)
			assertValue(t, v, tt.result)
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
			rhs := map[EntityUID]struct{}{}
			for _, v := range tt.rhs {
				rhs[strEnt(v)] = struct{}{}
			}
			entities := Entities{}
			for k, p := range tt.parents {
				var ps []EntityUID
				for _, pp := range p {
					ps = append(ps, strEnt(pp))
				}
				uid := strEnt(k)
				entities[uid] = Entity{
					UID:     uid,
					Parents: ps,
				}
			}
			res := entityIn(strEnt(tt.lhs), rhs, entities)
			testutilEquals(t, res, tt.result)
		})
	}
	t.Run("exponentialWithoutCaching", func(t *testing.T) {
		t.Parallel(
		// This test will run for a very long time (O(2^100)) if there isn't caching.
		)

		entities := Entities{}
		for i := 0; i < 100; i++ {
			p := []EntityUID{
				NewEntityUID(fmt.Sprint(i+1), "1"),
				NewEntityUID(fmt.Sprint(i+1), "2"),
			}
			uid1 := NewEntityUID(fmt.Sprint(i), "1")
			entities[uid1] = Entity{
				UID:     uid1,
				Parents: p,
			}
			uid2 := NewEntityUID(fmt.Sprint(i), "2")
			entities[uid2] = Entity{
				UID:     uid2,
				Parents: p,
			}

		}
		res := entityIn(NewEntityUID("0", "1"), map[EntityUID]struct{}{NewEntityUID("0", "3"): {}}, entities)
		testutilEquals(t, res, false)
	})
}

func TestIsNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		lhs, rhs evaler
		result   Value
		err      error
	}{
		{"happyEq", newLiteralEval(NewEntityUID("X", "z")), newLiteralEval(path("X")), Boolean(true), nil},
		{"happyNeq", newLiteralEval(NewEntityUID("X", "z")), newLiteralEval(path("Y")), Boolean(false), nil},
		{"badLhs", newLiteralEval(Long(42)), newLiteralEval(path("X")), zeroValue(), errType},
		{"badRhs", newLiteralEval(NewEntityUID("X", "z")), newLiteralEval(Long(42)), zeroValue(), errType},
		{"errLhs", newErrorEval(errTest), newLiteralEval(path("X")), zeroValue(), errTest},
		{"errRhs", newLiteralEval(NewEntityUID("X", "z")), newErrorEval(errTest), zeroValue(), errTest},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := newIsEval(tt.lhs, tt.rhs).Eval(&evalContext{})
			assertError(t, err, tt.err)
			assertValue(t, got, tt.result)
		})
	}
}

func TestInNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		lhs, rhs evaler
		parents  map[string][]string
		result   Value
		err      error
	}{
		{
			"LhsError",
			newErrorEval(errTest),
			newLiteralEval(Set{}),
			map[string][]string{},
			zeroValue(),
			errTest,
		},
		{
			"LhsTypeError",
			newLiteralEval(String("foo")),
			newLiteralEval(Set{}),
			map[string][]string{},
			zeroValue(),
			errType,
		},
		{
			"RhsError",
			newLiteralEval(EntityUID{"human", "joe"}),
			newErrorEval(errTest),
			map[string][]string{},
			zeroValue(),
			errTest,
		},
		{
			"RhsTypeError1",
			newLiteralEval(EntityUID{"human", "joe"}),
			newLiteralEval(String("foo")),
			map[string][]string{},
			zeroValue(),
			errType,
		},
		{
			"RhsTypeError2",
			newLiteralEval(EntityUID{"human", "joe"}),
			newLiteralEval(Set{
				String("foo"),
			}),
			map[string][]string{},
			zeroValue(),
			errType,
		},
		{
			"Reflexive1",
			newLiteralEval(EntityUID{"human", "joe"}),
			newLiteralEval(EntityUID{"human", "joe"}),
			map[string][]string{},
			Boolean(true),
			nil,
		},
		{
			"Reflexive2",
			newLiteralEval(EntityUID{"human", "joe"}),
			newLiteralEval(Set{
				EntityUID{"human", "joe"},
			}),
			map[string][]string{},
			Boolean(true),
			nil,
		},
		{
			"BasicTrue",
			newLiteralEval(EntityUID{"human", "joe"}),
			newLiteralEval(EntityUID{"kingdom", "animal"}),
			map[string][]string{
				`human::"joe"`:     {`species::"human"`},
				`species::"human"`: {`kingdom::"animal"`},
			},
			Boolean(true),
			nil,
		},
		{
			"BasicFalse",
			newLiteralEval(EntityUID{"human", "joe"}),
			newLiteralEval(EntityUID{"kingdom", "plant"}),
			map[string][]string{
				`human::"joe"`:     {`species::"human"`},
				`species::"human"`: {`kingdom::"animal"`},
			},
			Boolean(false),
			nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newInEval(tt.lhs, tt.rhs)
			entities := Entities{}
			for k, p := range tt.parents {
				var ps []EntityUID
				for _, pp := range p {
					ps = append(ps, strEnt(pp))
				}
				uid := strEnt(k)
				entities[uid] = Entity{
					UID:     uid,
					Parents: ps,
				}
			}
			evalContext := evalContext{Entities: entities}
			v, err := n.Eval(&evalContext)
			assertError(t, err, tt.err)
			assertValue(t, v, tt.result)
		})
	}
}

func TestDecimalLiteralNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		arg    evaler
		result Value
		err    error
	}{
		{"Error", newErrorEval(errTest), zeroValue(), errTest},
		{"TypeError", newLiteralEval(Long(1)), zeroValue(), errType},
		{"DecimalError", newLiteralEval(String("frob")), zeroValue(), errDecimal},
		{"Success", newLiteralEval(String("1.0")), Decimal(10000), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newDecimalLiteralEval(tt.arg)
			v, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
			assertValue(t, v, tt.result)
		})
	}
}

func TestIPLiteralNode(t *testing.T) {
	t.Parallel()
	ipv6Loopback, err := ParseIPAddr("::1")
	testutilOK(t, err)
	tests := []struct {
		name   string
		arg    evaler
		result Value
		err    error
	}{
		{"Error", newErrorEval(errTest), zeroValue(), errTest},
		{"TypeError", newLiteralEval(Long(1)), zeroValue(), errType},
		{"IPError", newLiteralEval(String("not-an-IP-address")), zeroValue(), errIP},
		{"Success", newLiteralEval(String("::1/128")), ipv6Loopback, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newIPLiteralEval(tt.arg)
			v, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
			assertValue(t, v, tt.result)
		})
	}
}

func TestIPTestNode(t *testing.T) {
	t.Parallel()
	ipv4Loopback, err := ParseIPAddr("127.0.0.1")
	testutilOK(t, err)
	ipv6Loopback, err := ParseIPAddr("::1")
	testutilOK(t, err)
	ipv4Multicast, err := ParseIPAddr("224.0.0.1")
	testutilOK(t, err)
	tests := []struct {
		name   string
		lhs    evaler
		rhs    ipTestType
		result Value
		err    error
	}{
		{"Error", newErrorEval(errTest), ipTestIPv4, zeroValue(), errTest},
		{"TypeError", newLiteralEval(Long(1)), ipTestIPv4, zeroValue(), errType},
		{"IPv4True", newLiteralEval(ipv4Loopback), ipTestIPv4, Boolean(true), nil},
		{"IPv4False", newLiteralEval(ipv6Loopback), ipTestIPv4, Boolean(false), nil},
		{"IPv6True", newLiteralEval(ipv6Loopback), ipTestIPv6, Boolean(true), nil},
		{"IPv6False", newLiteralEval(ipv4Loopback), ipTestIPv6, Boolean(false), nil},
		{"LoopbackTrue", newLiteralEval(ipv6Loopback), ipTestLoopback, Boolean(true), nil},
		{"LoopbackFalse", newLiteralEval(ipv4Multicast), ipTestLoopback, Boolean(false), nil},
		{"MulticastTrue", newLiteralEval(ipv4Multicast), ipTestMulticast, Boolean(true), nil},
		{"MulticastFalse", newLiteralEval(ipv6Loopback), ipTestMulticast, Boolean(false), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newIPTestEval(tt.lhs, tt.rhs)
			v, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
			assertValue(t, v, tt.result)
		})
	}
}

func TestIPIsInRangeNode(t *testing.T) {
	t.Parallel()
	ipv4A, err := ParseIPAddr("1.2.3.4")
	testutilOK(t, err)
	ipv4B, err := ParseIPAddr("1.2.3.0/24")
	testutilOK(t, err)
	ipv4C, err := ParseIPAddr("1.2.4.0/24")
	testutilOK(t, err)
	tests := []struct {
		name     string
		lhs, rhs evaler
		result   Value
		err      error
	}{
		{"LhsError", newErrorEval(errTest), newLiteralEval(ipv4A), zeroValue(), errTest},
		{"LhsTypeError", newLiteralEval(Long(1)), newLiteralEval(ipv4A), zeroValue(), errType},
		{"RhsError", newLiteralEval(ipv4A), newErrorEval(errTest), zeroValue(), errTest},
		{"RhsTypeError", newLiteralEval(ipv4A), newLiteralEval(Long(1)), zeroValue(), errType},
		{"AA", newLiteralEval(ipv4A), newLiteralEval(ipv4A), Boolean(true), nil},
		{"AB", newLiteralEval(ipv4A), newLiteralEval(ipv4B), Boolean(true), nil},
		{"BA", newLiteralEval(ipv4B), newLiteralEval(ipv4A), Boolean(false), nil},
		{"AC", newLiteralEval(ipv4A), newLiteralEval(ipv4C), Boolean(false), nil},
		{"CA", newLiteralEval(ipv4C), newLiteralEval(ipv4A), Boolean(false), nil},
		{"BC", newLiteralEval(ipv4B), newLiteralEval(ipv4C), Boolean(false), nil},
		{"CB", newLiteralEval(ipv4C), newLiteralEval(ipv4B), Boolean(false), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := newIPIsInRangeEval(tt.lhs, tt.rhs)
			v, err := n.Eval(&evalContext{})
			assertError(t, err, tt.err)
			assertValue(t, v, tt.result)
		})
	}
}

func TestCedarString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		in         Value
		wantString string
		wantCedar  string
	}{
		{"string", String("hello"), `hello`, `"hello"`},
		{"number", Long(42), `42`, `42`},
		{"bool", Boolean(true), `true`, `true`},
		{"record", Record{"a": Long(42), "b": Long(43)}, `{"a":42,"b":43}`, `{"a":42,"b":43}`},
		{"set", Set{Long(42), Long(43)}, `[42,43]`, `[42,43]`},
		{"singleIP", IPAddr(netip.MustParsePrefix("192.168.0.42/32")), `192.168.0.42`, `ip("192.168.0.42")`},
		{"ipPrefix", IPAddr(netip.MustParsePrefix("192.168.0.42/24")), `192.168.0.42/24`, `ip("192.168.0.42/24")`},
		{"decimal", Decimal(12345678), `1234.5678`, `decimal("1234.5678")`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotString := tt.in.String()
			testutilEquals(t, gotString, tt.wantString)
			gotCedar := tt.in.Cedar()
			testutilEquals(t, gotCedar, tt.wantCedar)
		})
	}
}
