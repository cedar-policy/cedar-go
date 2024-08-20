package testutil

import (
	"errors"
	"reflect"
)

type TB interface {
	Helper()
	Fatalf(format string, args ...any)
}

//go:generate moq -pkg testutil -fmt goimports -out mocks_test.go . TB

func Equals[T any](t TB, a, b T) {
	t.Helper()
	if reflect.DeepEqual(a, b) {
		return
	}
	t.Fatalf("got %+v want %+v", a, b)
}

func FatalIf(t TB, c bool, f string, args ...any) {
	t.Helper()
	if !c {
		return
	}
	t.Fatalf(f, args...)
}

func OK(t TB, err error) {
	t.Helper()
	if err == nil {
		return
	}
	t.Fatalf("got %v want nil", err)
}

func Error(t TB, err error) {
	t.Helper()
	if err != nil {
		return
	}
	t.Fatalf("got nil want error")
}

func ErrorIs(t TB, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("err got %v want %v", got, want)
	}
}

func Panic(t TB, f func()) {
	t.Helper()
	defer func() {
		if e := recover(); e == nil {
			t.Fatalf("got nil want panic")
		}
	}()
	f()
}
