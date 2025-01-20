package cedar_test

//
//func TestLinkTemplateToPolicy(t *testing.T) {
//    linkTests := []struct {
//        Name           string
//        TemplateString string
//        TemplateID     string
//        LinkID         string
//        Env            map[string]string
//        Want           parser.LinkedPolicy
//    }{
//
//        {
//            "principal ScopeTypeEq",
//            `@id("scope_eq_test")
//permit (
//    principal == ?principal,
//    action,
//    resource
//);`,
//            "scope_eq_test",
//            "scope_eq_link",
//            map[string]string{"?principal": `User::"bob"`},
//            parser.LinkedPolicy{
//                TemplateID: "scope_eq_test",
//                LinkID:     "scope_eq_link",
//                Template: ast.Permit().
//                    Annotate("id", "scope_eq_link").
//                    PrincipalEq(types.EntityUID{Type: "User", ID: "bob"}).
//                    AddSlot(types.PrincipalSlot),
//            },
//        },
//        {
//            "resource ScopeTypeEq",
//            `@id("scope_eq_test")
//permit (
//    principal,
//    action,
//    resource == ?resource
//);`,
//            "scope_eq_test",
//            "scope_eq_link",
//            map[string]string{"?resource": `Album::"trip"`},
//            parser.LinkedPolicy{
//                TemplateID: "scope_eq_test",
//                LinkID:     "scope_eq_link",
//                Template: ast.Permit().
//                    Annotate("id", "scope_eq_link").
//                    ResourceEq(types.EntityUID{Type: "Album", ID: "trip"}).
//                    AddSlot(types.ResourceSlot),
//            },
//        },
//        {
//            "principal ScopeTypeIn",
//            `@id("scope_in_test")
//permit (
//    principal in ?principal,
//    action,
//    resource
//);`,
//            "scope_in_test",
//            "scope_in_link",
//            map[string]string{"?principal": `User::"charlie"`},
//            parser.LinkedPolicy{
//                TemplateID: "scope_in_test",
//                LinkID:     "scope_in_link",
//                Template: ast.Permit().
//                    Annotate("id", "scope_in_link").
//                    PrincipalIn(types.EntityUID{Type: "User", ID: "charlie"}).
//                    AddSlot(types.PrincipalSlot),
//            },
//        },
//        {
//            "principal ScopeTypeIsIn",
//            `@id("scope_isin_test")
//permit (
//    principal is User in ?principal,
//    action,
//    resource
//);`,
//            "scope_isin_test",
//            "scope_isin_link",
//            map[string]string{"?principal": `User::"dave"`},
//            parser.LinkedPolicy{
//                TemplateID: "scope_isin_test",
//                LinkID:     "scope_isin_link",
//                Template: ast.Permit().
//                    Annotate("id", "scope_isin_link").
//                    PrincipalIsIn(types.EntityType("User"), types.EntityUID{Type: "User", ID: "dave"}).
//                    AddSlot(types.PrincipalSlot),
//            },
//        },
//        {
//            "resource ScopeTypeEq",
//            `@id("resource_test")
//permit (
//    principal,
//    action,
//    resource == ?resource
//);`,
//            "resource_test",
//            "resource_link",
//            map[string]string{"?resource": `Album::"trip"`},
//            parser.LinkedPolicy{
//                TemplateID: "resource_test",
//                LinkID:     "resource_link",
//                Template: ast.Permit().
//                    Annotate("id", "resource_link").
//                    ResourceEq(types.EntityUID{Type: "Album", ID: "trip"}).
//                    AddSlot(types.ResourceSlot),
//            },
//        },
//        {
//            "resource ScopeTypeIn",
//            `@id("scope_in_test")
//permit (
//    principal,
//    action,
//    resource in ?resource
//);`,
//            "scope_in_test",
//            "scope_in_link",
//            map[string]string{"?resource": `Album::"trip"`},
//            parser.LinkedPolicy{
//                TemplateID: "scope_in_test",
//                LinkID:     "scope_in_link",
//                Template: ast.Permit().
//                    Annotate("id", "scope_in_link").
//                    ResourceIn(types.EntityUID{Type: "Album", ID: "trip"}).
//                    AddSlot(types.ResourceSlot),
//            },
//        },
//        {
//            "resource ScopeTypeIsIn",
//            `@id("scope_isin_test")
//permit (
//    principal,
//    action,
//    resource is Album in ?resource
//);`,
//            "scope_isin_test",
//            "scope_isin_link",
//            map[string]string{"?resource": `Album::"trip"`},
//            parser.LinkedPolicy{
//                TemplateID: "scope_isin_test",
//                LinkID:     "scope_isin_link",
//                Template: ast.Permit().
//                    Annotate("id", "scope_isin_link").
//                    ResourceIsIn(types.EntityType("Album"), types.EntityUID{Type: "Album", ID: "trip"}).
//                    AddSlot(types.ResourceSlot),
//            },
//        },
//    }
//
//    for _, tt := range linkTests {
//        t.Run(tt.Name, func(t *testing.T) {
//            t.Parallel()
//
//            var templateBody parser.Policy
//            testutil.OK(t, templateBody.UnmarshalCedar([]byte(tt.TemplateString)))
//            template := parser.Template(templateBody)
//
//            linkedPolicy, err := cedar.LinkTemplateToPolicy(template, tt.LinkID, tt.Env)
//            testutil.OK(t, err)
//
//            testutil.Equals(t, linkedPolicy.LinkID, tt.LinkID)
//
//            linkedPolicy.Template.Position = ast.Position{}
//            testutil.Equals(t, linkedPolicy.Template, tt.Want.Template)
//        })
//    }
//}
