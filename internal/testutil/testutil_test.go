package testutil

import (
	"errors"
	"fmt"
	"testing"
)

func newTB() *TBMock {
	return &TBMock{
		HelperFunc: func() {},
		FatalfFunc: func(string, ...any) {},
	}
}

func TestEquals(t *testing.T) {
	t.Parallel()

	t.Run("Pass", func(t *testing.T) {
		tb := newTB()
		Equals(tb, 42, 42)

		// assertions
		Equals(t, len(tb.HelperCalls()), 1)
		Equals(t, len(tb.FatalfCalls()), 0)
	})
	t.Run("Fail", func(t *testing.T) {
		tb := newTB()
		Equals(tb, 42, 43)

		// assertions
		Equals(t, len(tb.HelperCalls()), 1)
		Equals(t, len(tb.FatalfCalls()), 1)
	})
}

func TestFatalIf(t *testing.T) {
	t.Parallel()

	t.Run("Pass", func(t *testing.T) {
		tb := newTB()
		FatalIf(tb, false, "unused")

		// assertions
		Equals(t, len(tb.HelperCalls()), 1)
		Equals(t, len(tb.FatalfCalls()), 0)
	})
	t.Run("Fail", func(t *testing.T) {
		tb := newTB()
		FatalIf(tb, true, "used")

		// assertions
		Equals(t, len(tb.HelperCalls()), 1)
		Equals(t, len(tb.FatalfCalls()), 1)
	})
}

func TestOK(t *testing.T) {
	t.Parallel()

	t.Run("Pass", func(t *testing.T) {
		tb := newTB()
		var err error
		OK(tb, err)

		// assertions
		Equals(t, len(tb.HelperCalls()), 1)
		Equals(t, len(tb.FatalfCalls()), 0)
	})
	t.Run("Fail", func(t *testing.T) {
		tb := newTB()
		err := fmt.Errorf("error")
		OK(tb, err)

		// assertions
		Equals(t, len(tb.HelperCalls()), 1)
		Equals(t, len(tb.FatalfCalls()), 1)
	})
}

func TestError(t *testing.T) {
	t.Parallel()

	t.Run("Pass", func(t *testing.T) {
		tb := newTB()
		err := fmt.Errorf("error")
		Error(tb, err)

		// assertions
		Equals(t, len(tb.HelperCalls()), 1)
		Equals(t, len(tb.FatalfCalls()), 0)
	})
	t.Run("Fail", func(t *testing.T) {
		tb := newTB()
		var err error
		Error(tb, err)

		// assertions
		Equals(t, len(tb.HelperCalls()), 1)
		Equals(t, len(tb.FatalfCalls()), 1)
	})
}

func TestErrorIs(t *testing.T) {
	t.Parallel()

	t.Run("Pass", func(t *testing.T) {
		tb := newTB()
		err := fmt.Errorf("error")
		ErrorIs(tb, err, err)

		// assertions
		Equals(t, len(tb.HelperCalls()), 1)
		Equals(t, len(tb.FatalfCalls()), 0)
	})
	t.Run("Fail", func(t *testing.T) {
		tb := newTB()
		err := fmt.Errorf("error")
		err2 := fmt.Errorf("error2")
		ErrorIs(tb, err, err2)

		// assertions
		Equals(t, len(tb.HelperCalls()), 1)
		Equals(t, len(tb.FatalfCalls()), 1)
	})
}

func TestPanic(t *testing.T) {
	t.Parallel()

	t.Run("Pass", func(t *testing.T) {
		tb := newTB()
		Panic(tb, func() {
			panic("panic")
		})

		// assertions
		Equals(t, len(tb.HelperCalls()), 1)
		Equals(t, len(tb.FatalfCalls()), 0)
	})
	t.Run("Fail", func(t *testing.T) {
		tb := newTB()
		Panic(tb, func() {
		})

		// assertions
		Equals(t, len(tb.HelperCalls()), 1)
		Equals(t, len(tb.FatalfCalls()), 1)
	})
}

func TestMust(t *testing.T) {
	t.Parallel()
	t.Run("Panic", func(t *testing.T) {
		tb := newTB()
		var x bool
		Panic(tb, func() {
			x = Must(true, fmt.Errorf("panic"))
		})
		// assertions
		Equals(t, x, false)
		Equals(t, len(tb.HelperCalls()), 1)
		Equals(t, len(tb.FatalfCalls()), 0)
	})
	t.Run("Okay", func(t *testing.T) {
		tb := newTB()
		var x bool
		Panic(tb, func() {
			x = Must(true, nil)
		})
		// assertions
		Equals(t, x, true)
	})
}

func newTBWithErrorf() *TBMock {
	return &TBMock{
		HelperFunc: func() {},
		FatalfFunc: func(string, ...any) {},
		ErrorfFunc: func(string, ...any) {},
	}
}

func TestCollectErrors(t *testing.T) {
	t.Parallel()

	t.Run("Nil", func(t *testing.T) {
		got := CollectErrors(nil)
		Equals(t, got, []string(nil))
	})

	t.Run("SingleError", func(t *testing.T) {
		err := fmt.Errorf("single error")
		got := CollectErrors(err)
		Equals(t, len(got), 1)
		Equals(t, got[0], "single error")
	})

	t.Run("JoinedErrors", func(t *testing.T) {
		err := errors.Join(fmt.Errorf("error1"), fmt.Errorf("error2"))
		got := CollectErrors(err)
		Equals(t, len(got), 2)
		Equals(t, got[0], "error1")
		Equals(t, got[1], "error2")
	})

	t.Run("NestedJoinedErrors", func(t *testing.T) {
		inner := errors.Join(fmt.Errorf("a"), fmt.Errorf("b"))
		outer := errors.Join(inner, fmt.Errorf("c"))
		got := CollectErrors(outer)
		Equals(t, len(got), 3)
		Equals(t, got[0], "a")
		Equals(t, got[1], "b")
		Equals(t, got[2], "c")
	})
}

func TestCheckErrorStrings(t *testing.T) {
	t.Parallel()

	t.Run("Pass", func(t *testing.T) {
		tb := newTBWithErrorf()
		CheckErrorStrings(tb, []string{"a", "b"}, []string{"b", "a"})
		Equals(t, len(tb.FatalfCalls()), 0)
		Equals(t, len(tb.ErrorfCalls()), 0)
	})

	t.Run("Mismatch", func(t *testing.T) {
		tb := newTBWithErrorf()
		CheckErrorStrings(tb, []string{"a", "b"}, []string{"a", "c"})
		Equals(t, len(tb.ErrorfCalls()), 1)
	})

	t.Run("DifferentLengths", func(t *testing.T) {
		tb := newTBWithErrorf()
		CheckErrorStrings(tb, []string{"a"}, []string{"a", "b"})
		Equals(t, len(tb.FatalfCalls()), 1)
	})
}

func TestNormalizeExprParens(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"(1 + 2)", "1 + 2"},
		{"if (1 + 2) then x else y", "if 1 + 2 then x else y"},
		{"if true then (context.foo) else (context.foo)", "if true then context.foo else context.foo"},
		{"principal.getTag(\"x\")", "principal.getTag(\"x\")"},
		{"if true then (principal.getTag(\"x\")) else y", "if true then principal.getTag(\"x\") else y"},
		{"(if true then \"a\" else \"a\")", "if true then \"a\" else \"a\""},
		{"no parens here", "no parens here"},
		{"(unmatched", "(unmatched"},
	}
	for _, tt := range tests {
		got := normalizeExprParens(tt.input)
		Equals(t, got, tt.want)
	}
}

func TestJSONMarshalsTo(t *testing.T) {
	t.Parallel()
	t.Run("Okay", func(t *testing.T) {
		tb := newTB()
		JSONMarshalsTo(tb, "test", `"test"`)
		Equals(t, len(tb.HelperCalls()), 4)
		Equals(t, len(tb.FatalfCalls()), 0)
	})

	t.Run("ErrNotEqual", func(t *testing.T) {
		tb := newTB()
		JSONMarshalsTo(tb, "test", `"asdf"`)
		Equals(t, len(tb.HelperCalls()), 4)
		Equals(t, len(tb.FatalfCalls()), 1)
	})

	t.Run("ErrNotJSON", func(t *testing.T) {
		tb := newTB()
		JSONMarshalsTo(tb, "test", `asdf`)
		Equals(t, len(tb.HelperCalls()), 4)
		Equals(t, len(tb.FatalfCalls()), 2)
	})

	t.Run("ErrNotMarshalable", func(t *testing.T) {
		tb := newTB()
		cx := complex(0, 0)
		JSONMarshalsTo(tb, cx, `null`)
		Equals(t, len(tb.HelperCalls()), 4)
		Equals(t, len(tb.FatalfCalls()), 2)
	})

}
