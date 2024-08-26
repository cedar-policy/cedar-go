package consts

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestConsts(t *testing.T) {
	t.Parallel()
	testutil.Equals(t, Principal, "principal")
	testutil.Equals(t, Action, "action")
	testutil.Equals(t, Resource, "resource")
	testutil.Equals(t, Context, "context")
}
