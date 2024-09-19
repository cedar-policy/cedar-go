package extensions

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestExtensions(t *testing.T) {
	t.Parallel()
	testutil.Equals(t, len(ExtMap), 22)
}
