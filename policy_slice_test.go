package cedar_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestPolicySlice(t *testing.T) {
	t.Parallel()

	policiesStr := `permit (
    principal,
    action == Action::"editPhoto",
    resource
)
when { resource.owner == principal };

forbid (
    principal in Groups::"bannedUsers",
    action,
    resource
);`

	var policies cedar.PolicySlice
	testutil.OK(t, policies.UnmarshalCedar([]byte(policiesStr)))

	testutil.Equals(t, string(policies.MarshalCedar()), policiesStr)
}
