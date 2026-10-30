package validate_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
	"github.com/cedar-policy/cedar-go/x/exp/schema/validate"
)

// BenchmarkPolicySet validates n policies, each reading its own attribute of
// a context with n attributes, with one Validator. The cost per policy must
// not grow with the size of the schema's contexts.
func BenchmarkPolicySet(b *testing.B) {
	for _, n := range []int{100, 1000} {
		b.Run(fmt.Sprintf("policies=%d", n), func(b *testing.B) {
			var sch, pol strings.Builder
			sch.WriteString("entity User;\naction \"Read\" appliesTo { principal: User, resource: User, context: {\n")
			for i := range n {
				fmt.Fprintf(&sch, "  a%d: Bool,\n", i)
				fmt.Fprintf(&pol, "permit (principal, action == Action::\"Read\", resource) when { context.a%d };\n", i)
			}
			sch.WriteString("} };\n")

			var s schema.Schema
			if err := s.UnmarshalCedar([]byte(sch.String())); err != nil {
				b.Fatal(err)
			}
			rs, err := s.Resolve()
			if err != nil {
				b.Fatal(err)
			}
			ps, err := cedar.NewPolicySetFromBytes("", []byte(pol.String()))
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				v := validate.New(rs)
				for id, p := range ps.All() {
					if err := v.Policy(string(id), (*ast.Policy)(p.AST())); err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}
