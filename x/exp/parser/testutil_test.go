package parser

import (
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
