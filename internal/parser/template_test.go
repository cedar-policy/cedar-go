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
            "principal ScopeTypeEq",
            `@id("scope_eq_test")
permit (
    principal == ?principal,
    action,
    resource
);`,
            "scope_eq_test",
            "scope_eq_link",
            map[string]string{"?principal": `User::"bob"`},
            ast.LinkedPolicy{
                TemplateID: "scope_eq_test",
                LinkID:     "scope_eq_link",
                Policy: ast.Permit().
                    Annotate("id", "scope_eq_link").
                    PrincipalEq(types.EntityUID{Type: "User", ID: "bob"}).
                    AddSlot(types.PrincipalSlot),
            },
        },
        {
            "resource ScopeTypeEq",
            `@id("scope_eq_test")
permit (
    principal,
    action,
    resource == ?resource
);`,
            "scope_eq_test",
            "scope_eq_link",
            map[string]string{"?resource": `Album::"trip"`},
            ast.LinkedPolicy{
                TemplateID: "scope_eq_test",
                LinkID:     "scope_eq_link",
                Policy: ast.Permit().
                    Annotate("id", "scope_eq_link").
                    ResourceEq(types.EntityUID{Type: "Album", ID: "trip"}).
                    AddSlot(types.ResourceSlot),
            },
        },
        {
            "principal ScopeTypeIn",
            `@id("scope_in_test")
permit (
    principal in ?principal,
    action,
    resource
);`,
            "scope_in_test",
            "scope_in_link",
            map[string]string{"?principal": `User::"charlie"`},
            ast.LinkedPolicy{
                TemplateID: "scope_in_test",
                LinkID:     "scope_in_link",
                Policy: ast.Permit().
                    Annotate("id", "scope_in_link").
                    PrincipalIn(types.EntityUID{Type: "User", ID: "charlie"}).
                    AddSlot(types.PrincipalSlot),
            },
        },
        {
            "principal ScopeTypeIsIn",
            `@id("scope_isin_test")
permit (
    principal is User in ?principal,
    action,
    resource
);`,
            "scope_isin_test",
            "scope_isin_link",
            map[string]string{"?principal": `User::"dave"`},
            ast.LinkedPolicy{
                TemplateID: "scope_isin_test",
                LinkID:     "scope_isin_link",
                Policy: ast.Permit().
                    Annotate("id", "scope_isin_link").
                    PrincipalIsIn(types.EntityType("User"), types.EntityUID{Type: "User", ID: "dave"}).
                    AddSlot(types.PrincipalSlot),
            },
        },
        {
            "resource ScopeTypeEq",
            `@id("resource_test")
permit (
    principal,
    action,
    resource == ?resource
);`,
            "resource_test",
            "resource_link",
            map[string]string{"?resource": `Album::"trip"`},
            ast.LinkedPolicy{
                TemplateID: "resource_test",
                LinkID:     "resource_link",
                Policy: ast.Permit().
                    Annotate("id", "resource_link").
                    ResourceEq(types.EntityUID{Type: "Album", ID: "trip"}).
                    AddSlot(types.ResourceSlot),
            },
        },
        {
            "resource ScopeTypeIn",
            `@id("scope_in_test")
permit (
    principal,
    action,
    resource in ?resource
);`,
            "scope_in_test",
            "scope_in_link",
            map[string]string{"?resource": `Album::"trip"`},
            ast.LinkedPolicy{
                TemplateID: "scope_in_test",
                LinkID:     "scope_in_link",
                Policy: ast.Permit().
                    Annotate("id", "scope_in_link").
                    ResourceIn(types.EntityUID{Type: "Album", ID: "trip"}).
                    AddSlot(types.ResourceSlot),
            },
        },
        {
            "resource ScopeTypeIsIn",
            `@id("scope_isin_test")
permit (
    principal,
    action,
    resource is Album in ?resource
);`,
            "scope_isin_test",
            "scope_isin_link",
            map[string]string{"?resource": `Album::"trip"`},
            ast.LinkedPolicy{
                TemplateID: "scope_isin_test",
                LinkID:     "scope_isin_link",
                Policy: ast.Permit().
                    Annotate("id", "scope_isin_link").
                    ResourceIsIn(types.EntityType("Album"), types.EntityUID{Type: "Album", ID: "trip"}).
                    AddSlot(types.ResourceSlot),
            },
        },
    }

    for _, tt := range linkTests {
        t.Run(tt.Name, func(t *testing.T) {
            t.Parallel()

            var templateBody parser.Policy
            testutil.OK(t, templateBody.UnmarshalCedar([]byte(tt.TemplateString)))
            template := ast.Template(templateBody)

            linkedPolicy, err := parser.LinkTemplateToPolicy(template, tt.LinkID, tt.Env)
            testutil.OK(t, err)

            testutil.Equals(t, linkedPolicy.LinkID, tt.LinkID)

            linkedPolicy.Policy.Position = ast.Position{}
            testutil.Equals(t, linkedPolicy.Policy, tt.Want.Policy)
        })
    }
}
