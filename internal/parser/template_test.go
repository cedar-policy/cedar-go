package parser_test

import (
	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"testing"
)

//package parser_test
//
//import (
//	"github.com/cedar-policy/cedar-go/internal/ast"
//	"github.com/cedar-policy/cedar-go/internal/parser"
//	"github.com/cedar-policy/cedar-go/internal/testutil"
//	"github.com/cedar-policy/cedar-go/types"
//	"testing"
//)

func TestLinkTemplateToPolicy(t *testing.T) {
	linkTests := []struct {
		Name           string
		TemplateString string
		TemplateID     string
		LinkID         string
		Env            map[types.SlotID]types.EntityUID
		Want           parser.Policy
	}{
		{
			"principal variable",
			`permit (
    principal == ?principal,
    action,
    resource
);`,
			"principal_test",
			"principal_link",
			map[types.SlotID]types.EntityUID{"?principal": types.NewEntityUID("User", "alice")},
			parserPolicy(ast.Permit().
				PrincipalEq(types.EntityUID{Type: "User", ID: "alice"}).
				AddSlot(types.PrincipalSlot)),
		},
		{
			"resource variable",
			`permit (
    principal,
    action,
    resource == ?resource
);`,
			"resource_test",
			"resource_link",
			map[types.SlotID]types.EntityUID{"?resource": types.NewEntityUID("Album", "trip")},
			parserPolicy(ast.Permit().
				ResourceEq(types.EntityUID{Type: "Album", ID: "trip"}).
				AddSlot(types.ResourceSlot)),
		},
	}

	for _, tt := range linkTests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			var templateBody parser.Policy
			testutil.OK(t, templateBody.UnmarshalCedar([]byte(tt.TemplateString)))
			template := parser.Template(templateBody)

			linkedPolicy := parser.NewLinkedPolicy(&template, tt.TemplateID, tt.LinkID, tt.Env)

			testutil.Equals(t, linkedPolicy.LinkID, tt.LinkID)

			newPolicy, err := linkedPolicy.Render()
			testutil.OK(t, err)

			newPolicy.Position = ast.Position{}
			testutil.Equals(t, newPolicy, tt.Want)
		})
	}
}

func parserPolicy(inAST *ast.Policy) parser.Policy {
	return parser.Policy(*inAST)
}
