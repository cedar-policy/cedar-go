package parser_test

import (
	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"testing"
)

func TestLinkTemplateToPolicy(t *testing.T) {
	linkTests := []struct {
		Name           string
		TemplateString string
		TemplateID     string
		LinkID         string
		Env            map[string]string
		Want           ast.LinkedPolicy
	}{
		{
			"principal variable",
			`@id("principal_test")
permit (
    principal == ?principal,
    action,
    resource
);`,
			"principal_test",
			"principal_link",
			map[string]string{"?principal": `User::"alice"`},
			ast.LinkedPolicy{
				TemplateID: "principal_test",
				LinkID:     "principal_link",
				Policy:     ast.Permit().Annotate("id", "principal_link").PrincipalEq(types.EntityUID{Type: "User", ID: "alice"}),
			},
		},
		{
			"resource variable",
			`@id("principal_test")
permit (
    principal,
    action,
    resource == ?resource
);`,
			"resource_test",
			"resource_link",
			map[string]string{"?resource": `Album::"trip"`},
			ast.LinkedPolicy{
				TemplateID: "principal_test",
				LinkID:     "principal_link",
				Policy:     ast.Permit().Annotate("id", "resource_link").ResourceEq(types.EntityUID{Type: "Album", ID: "trip"}),
			},
		},
	}

	for _, tt := range linkTests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			var templateBody parser.Policy
			testutil.OK(t, templateBody.UnmarshalCedar([]byte(tt.TemplateString)))
			template := ast.Template{
				Body: ast.Policy(templateBody),
			}

			linkedPolicy := parser.LinkTemplateToPolicy(template, tt.LinkID, tt.Env)

			testutil.Equals(t, tt.LinkID, linkedPolicy.LinkID)

			linkedPolicy.Policy.Position = ast.Position{}
			testutil.Equals(t, tt.Want.Policy, linkedPolicy.Policy)
		})
	}
}
