package types_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestString(t *testing.T) {
	t.Parallel()

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		hello := types.String("hello")
		hello2 := types.String("hello")
		goodbye := types.String("goodbye")
		testutil.FatalIf(t, !hello.Equal(hello), "%v not Equal to %v", hello, hello)
		testutil.FatalIf(t, !hello.Equal(hello2), "%v not Equal to %v", hello, hello2)
		testutil.FatalIf(t, hello.Equal(goodbye), "%v Equal to %v", hello, goodbye)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		AssertValueString(t, types.String("hello"), `hello`)
		AssertValueString(t, types.String("hello\ngoodbye"), "hello\ngoodbye")
	})

	t.Run("TypeName", func(t *testing.T) {
		t.Parallel()
		tn := types.String("hello").TypeName()
		testutil.Equals(t, tn, "string")
	})
}
