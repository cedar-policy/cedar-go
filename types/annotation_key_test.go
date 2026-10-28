package types_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestNewAnnotationKey(t *testing.T) {
	t.Parallel()

	for _, s := range []string{"id", "_x9", "if", "true", "__cedar"} {
		t.Run(s, func(t *testing.T) {
			t.Parallel()
			got, err := types.NewAnnotationKey(s)
			testutil.OK(t, err)
			testutil.Equals(t, got, types.AnnotationKey(s))
		})
	}

	for _, s := range []string{"", "9id", "NS::id", "a b", `id("x")`, "id) permit (principal"} {
		t.Run(s, func(t *testing.T) {
			t.Parallel()
			got, err := types.NewAnnotationKey(s)
			testutil.Error(t, err)
			testutil.Equals(t, got, types.AnnotationKey(""))
		})
	}
}
