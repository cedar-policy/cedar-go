package testutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
)

type TB interface {
	Helper()
	Errorf(format string, args ...any)
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

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

// CollectErrors recursively collects all leaf error strings from a (possibly joined) error.
func CollectErrors(err error) []string {
	if err == nil {
		return nil
	}
	if ue, ok := err.(interface{ Unwrap() []error }); ok {
		var result []string
		for _, e := range ue.Unwrap() {
			result = append(result, CollectErrors(e)...)
		}
		return result
	}
	return []string{err.Error()}
}

// CheckErrorStrings verifies that the actual error strings match the expected error strings.
// Both slices are sorted before comparison to handle ordering differences.
// Strings are normalized to remove cosmetic paren differences between Go and Rust Cedar formatting.
func CheckErrorStrings(t TB, got []string, expected []string) {
	t.Helper()
	Equals(t, len(got), len(expected))
	sortedGot := slices.Clone(got)
	slices.Sort(sortedGot)
	sortedExp := slices.Clone(expected)
	slices.Sort(sortedExp)
	for i := range sortedGot {
		if normalizeExprParens(sortedGot[i]) != normalizeExprParens(sortedExp[i]) {
			t.Errorf("error[%d] mismatch:\n  got:  %s\n  want: %s", i, sortedGot[i], sortedExp[i])
		}
	}
}

// normalizeExprParens strips cosmetic parentheses from Cedar expressions in error messages.
// Rust Cedar's maybe_with_parens wraps all non-primary subexpressions in parens,
// while our Go marshaler uses precedence-based parens. This function removes balanced
// parens that aren't part of function calls (i.e., not preceded by an identifier char)
// to normalize both styles for comparison.
func normalizeExprParens(s string) string {
	b := []byte(s)
	var out []byte
	for i := 0; i < len(b); i++ {
		if b[i] == '(' && !isFuncCallParen(b, i) {
			// Find matching close paren
			if j := findMatchingParen(b, i); j > i {
				out = append(out, b[i+1:j]...)
				i = j
				continue
			}
		}
		out = append(out, b[i])
	}
	return string(out)
}

// isFuncCallParen returns true if the paren at position i is a function call paren
// (preceded by an identifier character like a letter, digit, or underscore).
func isFuncCallParen(b []byte, i int) bool {
	if i == 0 {
		return false
	}
	c := b[i-1]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// findMatchingParen finds the index of the closing paren matching the open paren at position i.
// Returns -1 if no matching paren is found.
func findMatchingParen(b []byte, i int) int {
	depth := 1
	for j := i + 1; j < len(b); j++ {
		switch b[j] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return j
			}
		}
	}
	return -1
}

// JSONMarshalsTo asserts that obj marshals as JSON to the given string, allowing for formatting differences and
// displaying an easy-to-read diff.
func JSONMarshalsTo[T any](t TB, obj T, want string) {
	t.Helper()
	b, err := json.MarshalIndent(obj, "", "\t")
	OK(t, err)

	var wantBuf bytes.Buffer
	err = json.Indent(&wantBuf, []byte(want), "", "\t")
	OK(t, err)
	Equals(t, string(b), wantBuf.String())
}
