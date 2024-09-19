package types

import (
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestDeepClone(t *testing.T) {
	t.Parallel()
	t.Run("Boolean", func(t *testing.T) {
		t.Parallel()
		a := Boolean(true)
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a = Boolean(false)
		testutil.Equals(t, a, Boolean(false))
		testutil.Equals(t, b, Value(Boolean(true)))
	})
	t.Run("Long", func(t *testing.T) {
		t.Parallel()
		a := Long(42)
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a = Long(43)
		testutil.Equals(t, a, Long(43))
		testutil.Equals(t, b, Value(Long(42)))
	})
	t.Run("String", func(t *testing.T) {
		t.Parallel()
		a := String("cedar")
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a = String("policy")
		testutil.Equals(t, a, String("policy"))
		testutil.Equals(t, b, Value(String("cedar")))
	})
	t.Run("EntityUID", func(t *testing.T) {
		t.Parallel()
		a := NewEntityUID("Action", "test")
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a.ID = "bananas"
		testutil.Equals(t, a, NewEntityUID("Action", "bananas"))
		testutil.Equals(t, b, Value(NewEntityUID("Action", "test")))
	})
	t.Run("Set", func(t *testing.T) {
		t.Parallel()
		a := Set{Long(42)}
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a[0] = String("bananas")
		testutil.Equals(t, a, Set{String("bananas")})
		testutil.Equals(t, b, Value(Set{Long(42)}))
	})
	t.Run("NilSet", func(t *testing.T) {
		t.Parallel()
		var a Set
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
	})

	t.Run("Record", func(t *testing.T) {
		t.Parallel()
		a := Record{"key": Long(42)}
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a["key"] = String("bananas")
		testutil.Equals(t, a, Record{"key": String("bananas")})
		testutil.Equals(t, b, Value(Record{"key": Long(42)}))
	})

	t.Run("NilRecord", func(t *testing.T) {
		t.Parallel()
		var a Record
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
	})

	t.Run("Decimal", func(t *testing.T) {
		t.Parallel()
		a := UnsafeDecimal(42)
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a = UnsafeDecimal(43)
		testutil.Equals(t, a, UnsafeDecimal(43))
		testutil.Equals(t, b, Value(UnsafeDecimal(42)))
	})

	t.Run("IPAddr", func(t *testing.T) {
		t.Parallel()
		a := mustIPValue("127.0.0.42")
		b := a.deepClone()
		testutil.Equals(t, a.MarshalCedar(), b.MarshalCedar())
		a = mustIPValue("127.0.0.43")
		testutil.Equals(t, a.MarshalCedar(), mustIPValue("127.0.0.43").MarshalCedar())
		testutil.Equals(t, b.MarshalCedar(), mustIPValue("127.0.0.42").MarshalCedar())
	})

	t.Run("Datetime", func(t *testing.T) {
		t.Parallel()
		a := FromStdTime(time.UnixMilli(42))
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a = FromStdTime(time.UnixMilli(43))
		testutil.Equals(t, a, FromStdTime(time.UnixMilli(43)))
		testutil.Equals(t, b, Value(FromStdTime(time.UnixMilli(42))))
	})

	t.Run("Duration", func(t *testing.T) {
		t.Parallel()
		a := FromStdDuration(42 * time.Millisecond)
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a = FromStdDuration(43 * time.Millisecond)
		testutil.Equals(t, a, FromStdDuration(43*time.Millisecond))
		testutil.Equals(t, b, Value(FromStdDuration(42*time.Millisecond)))
	})

}
