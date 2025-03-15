package token

import (
	"errors"
	"testing"
)

func TestError_Error(t *testing.T) {
	pos := Position{Filename: "testfile", Line: 1, Column: 2}
	err := Error{Pos: pos, Err: errors.New("test error")}
	expected := "testfile:1:2: test error"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestErrList_Error(t *testing.T) {
	errs := Errors{
		errors.New("first error"),
		errors.New("second error"),
	}
	expected := "first error\nsecond error"
	if errs.Error() != expected {
		t.Errorf("expected %q, got %q", expected, errs.Error())
	}
}

func TestErrList_Sort(t *testing.T) {
	errs := Errors{
		Error{Pos: Position{Line: 1, Column: 2, Offset: 2}, Err: errors.New("second error")},
		Error{Pos: Position{Line: 1, Column: 1, Offset: 1}, Err: errors.New("first error")},
	}
	errs.Sort()
	if errs[0].Error() != "<input>:1:1: first error" || errs[1].Error() != "<input>:1:2: second error" {
		t.Errorf("errors not sorted correctly: %v", errs)
	}
}

func TestType_String(t *testing.T) {
	tests := []struct {
		tok      Type
		expected string
	}{
		{INVALID, "INVALID"},
		{NAMESPACE, "NAMESPACE"},
		{ENTITY, "ENTITY"},
		{ACTION, "ACTION"},
		{EOF, "EOF"},
		{Type(999), "Token(999)"},
	}

	for _, test := range tests {
		if test.tok.String() != test.expected {
			t.Errorf("expected %q, got %q", test.expected, test.tok.String())
		}
	}
}

func TestErrorsSort(t *testing.T) {
	// Test case for mixing Error and non-Error types
	reg := errors.New("regular error")
	reg2 := errors.New("another regular error")
	errs := Errors{
		reg,
		Error{Pos: Position{Offset: 10}, Err: errors.New("error at pos 10")},
		Error{Pos: Position{Offset: 5}, Err: errors.New("error at pos 5")},
		reg2,
	}
	errs.Sort()
	if len(errs) != 4 {
		t.Errorf("Expected length 4, got %d", len(errs))
	}
}
