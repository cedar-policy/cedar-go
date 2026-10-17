package validate_test

import (
	"sync"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
	"github.com/cedar-policy/cedar-go/x/exp/schema/validate"
)

// TestPolicyConcurrent checks that one Validator can validate policies from
// many goroutines, including the first calls that build its shared state.
// Run with -race.
func TestPolicyConcurrent(t *testing.T) {
	t.Parallel()
	var s schema.Schema
	testutil.OK(t, s.UnmarshalCedar([]byte(`entity User;
action "Read" appliesTo { principal: User, resource: User, context: { ok: Bool } };`)))
	rs, err := s.Resolve()
	testutil.OK(t, err)
	ps, err := cedar.NewPolicySetFromBytes("", []byte(`
permit (principal, action == Action::"Read", resource) when { context.ok };
permit (principal, action == Action::"Read", resource) when { context.missing };`))
	testutil.OK(t, err)
	valid := (*ast.Policy)(ps.Get("policy0").AST())
	invalid := (*ast.Policy)(ps.Get("policy1").AST())

	v := validate.New(rs)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			testutil.OK(t, v.Policy("policy0", valid))
			testutil.Error(t, v.Policy("policy1", invalid))
		}()
	}
	wg.Wait()
}
