package types

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestRecordInternal(t *testing.T) {
	t.Parallel()

	t.Run("hash", func(t *testing.T) {
		t.Parallel()

		t.Run("independent of key order", func(t *testing.T) {
			t.Parallel()
			m1 := NewRecord(RecordMap{"foo": Long(42), "bar": Long(1337)})
			m2 := NewRecord(RecordMap{"bar": Long(1337), "foo": Long(42)})
			testutil.Equals(t, m1.hash(), m2.hash())
		})

		t.Run("empty record", func(t *testing.T) {
			t.Parallel()
			m1 := Record{}
			m2 := NewRecord(RecordMap{})
			testutil.Equals(t, m1.hash(), m2.hash())
		})

		// These tests don't necessarily hold for all values of Record, but we want to ensure we are considering
		// different aspects of the Record, which these particular tests demonstrate.

		t.Run("same keys, different values", func(t *testing.T) {
			t.Parallel()
			m1 := NewRecord(RecordMap{"foo": Long(42), "bar": Long(1337)})
			m2 := NewRecord(RecordMap{"foo": Long(1337), "bar": Long(42)})
			testutil.FatalIf(t, m1.hash() == m2.hash(), "unexpected hash collision")
		})

		t.Run("same values, different keys", func(t *testing.T) {
			t.Parallel()
			m1 := NewRecord(RecordMap{"foo": Long(42), "bar": Long(1337)})
			m2 := NewRecord(RecordMap{"foo2": Long(42), "bar2": Long(1337)})
			testutil.FatalIf(t, m1.hash() == m2.hash(), "unepxected hash collision")
		})

		t.Run("extra key", func(t *testing.T) {
			t.Parallel()
			m1 := NewRecord(RecordMap{"foo": Long(42), "bar": Long(1337)})
			m2 := NewRecord(
				RecordMap{"foo": Long(42), "bar": Long(1337), "baz": Long(0)},
			)
			testutil.FatalIf(t, m1.hash() == m2.hash(), "unepxected hash collision")
		})
	})
}
