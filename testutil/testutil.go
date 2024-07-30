package testutil

import (
	"errors"
	"reflect"
	"testing"
)

func Equals[T any](t testing.TB, a, b T) {
	t.Helper()
	if reflect.DeepEqual(a, b) {
		return
	}
	t.Fatalf("got %+v want %+v", a, b)
}

func FatalIf(t testing.TB, c bool, f string, args ...any) {
	t.Helper()
	if !c {
		return
	}
	t.Fatalf(f, args...)
}

func OK(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		return
	}
	t.Fatalf("got %v want nil", err)
}

func Error(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		return
	}
	t.Fatalf("got nil want error")
}

func AssertError(t *testing.T, got, want error) {
	t.Helper()
	FatalIf(t, !errors.Is(got, want), "err got %v want %v", got, want)
}
