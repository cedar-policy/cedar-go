package templates_test

import (
	"github.com/cedar-policy/cedar-go/x/exp/templates"
	"testing"

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

	policies, err := templates.NewPolicyListFromBytes("", []byte(policiesStr))
	testutil.OK(t, err)
	testutil.Equals(t, string(policies.MarshalCedar()), policiesStr)
}

func TestPolicyWithTemplateSlice(t *testing.T) {
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
);

permit (
    principal == ?principal,
    action,
    resource
);`

	policies, err := templates.NewPolicyListFromBytes("", []byte(policiesStr))
	testutil.OK(t, err)
	testutil.Equals(t, string(policies.MarshalCedar()), policiesStr)
}
