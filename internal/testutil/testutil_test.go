package testutil

import (
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
