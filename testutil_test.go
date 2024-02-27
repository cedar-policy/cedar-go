package cedar

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func testutilEquals[T any](t testing.TB, a, b T) {
	t.Helper()
	if reflect.DeepEqual(a, b) {
		return
	}
	t.Fatalf("got %+v want %+v", a, b)
}

func testutilFatalIf(t testing.TB, c bool, f string, args ...any) {
	t.Helper()
	if !c {
		return
	}
	t.Fatalf(f, args...)
}

func testutilOK(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		return
	}
	t.Fatalf("got %v want nil", err)
}

func testutilError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		return
	}
	t.Fatalf("got nil want error")
}

func assertError(t *testing.T, got, want error) {
	t.Helper()
	testutilFatalIf(t, !errors.Is(got, want), "err got %v want %v", got, want)
}

func assertValue(t *testing.T, got, want Value) {
	t.Helper()
	testutilFatalIf(
		t,
		!((got == zeroValue() && want == zeroValue()) ||
			(got != zeroValue() && want != zeroValue() && got.equal(want))),
		"got %v want %v", got, want)
}

func assertBoolValue(t *testing.T, got Value, want bool) {
	t.Helper()
	testutilEquals[Value](t, got, Boolean(want))
}

func assertLongValue(t *testing.T, got Value, want int64) {
	t.Helper()
	testutilEquals[Value](t, got, Long(want))
}

func assertZeroValue(t *testing.T, got Value) {
	t.Helper()
	testutilEquals(t, got, zeroValue())
}

func assertValueString(t *testing.T, v Value, want string) {
	t.Helper()
	testutilEquals(t, v.String(), want)
}

func safeDoErr(f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return f()
}
