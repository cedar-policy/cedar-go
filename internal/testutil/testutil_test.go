package testutil

import (
	"fmt"
	"testing"
)

func newTB() *TBMock {
	return &TBMock{
		HelperFunc: func() {},
		FatalfFunc: func(format string, args ...any) {},
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
