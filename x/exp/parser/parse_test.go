package parser

import (
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		err  bool
	}{
		// Success cases
		// Test cases from https://github.com/cedar-policy/cedar/blob/main/cedar-policy-core/src/parser/testfiles/policies.cedar
		{"empty", ``, false},
		{"ex1", `//@test_annotation("This is the annotation")
		permit(
		  principal == User::"alice",
		  action == PhotoOp::"view",
		  resource == Photo::"VacationPhoto94.jpg"
		);`, false},
		{"ex2", `permit(
				principal in Team::"admins",
				action in [PhotoOp::"view", PhotoOp::"edit", PhotoOp::"delete"],
				resource in Album::"jane_vacation"
			  );`, false},
		{"ex3", `permit(
				principal == User::"alice",
				action in PhotoflashRole::"admin",
				resource in Album::"jane_vacation"
			  );`, false},
		{"simplest", `permit(
				principal,
				action,
				resource
			  );`, false},
		{"in", `permit(
				principal in Team::"eng",
				action in PhotoflashRole::"admin",
				resource in Album::"jane_vacation"
			  );

			  permit(
				principal in Team::"eng",
				action in [PhotoflashRole::"admin"],
				resource in Album::"jane_vacation"
			  );

			  permit(
				principal in Team::"eng",
				action in [PhotoflashRole::"admin", PhotoflashRole::"operator"],
				resource in Album::"jane_vacation"
			  );
			  `, false},
		{"multipleIdentEntities", `permit(
				principal == Org::Team::User::"alice",
				action,
				resource
			  );`, false},
		{"multiplePolicies", `permit(
				principal,
				action,
				resource
			  );

			  forbid(
				principal in Team::"admins",
				action in [PhotoOp::"view", PhotoOp::"edit", PhotoOp::"delete"],
				resource in Album::"jane_vacation"
			  );
			  `, false},
		{"annotations", `@first_annotation("This is the annotation")
			  @second_annotation("This is another annotation")
			  permit(
				principal,
				action,
				resource
			  );`, false},

		// Additional success cases
		{"primaryInt", `permit(principal, action, resource) when { 1234 };`, false},
		{"primaryString", `permit(principal, action, resource) when { "test string" };`, false},
		{"primaryBool", `permit(principal, action, resource) when { true } unless { false };`, false},
		{"primaryVar", `permit(principal, action, resource)
				when { principal }
				unless { action }
				when { resource }
				unless { context };
			`, false},
		{"primaryEntity", `permit(principal, action, resource)
				when { Org::User::"alice" };
			`, false},
		{"primaryExtFun", `permit(principal, action, resource)
				when { foo() }
				unless { foo::bar::as() }
				when { foo("hello") }
				unless { foo::bar(true, 42, "forty two") };
			`, false},
		{"ifElseThen", `permit(principal, action, resource)
			when { if false then principal else principal };`, false},
		{"access", `permit(principal, action, resource)
			when { resource.foo }
			unless { resource.foo.bar }
			when { principal.foo() }
			unless { principal.bar(false) }
			when { action.foo["bar"].baz() }
			unless { principal.bar(false, 123, "foo") }
			when { principal["foo"] };`, false},
		{"unary", `permit(principal, action, resource)
			when { !resource.foo }
			unless { -resource.bar }
			when { !!resource.foo }
			unless { --resource.bar }
			when { !-!-resource.bar };`, false},
		{"mult", `permit(principal, action, resource)
			when { resource.foo * 42 }
			unless { 42 * resource.bar }
			when { 42 * resource.bar * 43 }
			when { resource.foo * principal.bar };`, false},
		{"add", `permit(principal, action, resource)
			when { resource.foo + 42 }
			unless { 42 - resource.bar }
			when { 42 + resource.bar - 43 }
			when { resource.foo + principal.bar };`, false},
		{"relations", `permit(principal, action, resource)
			when { foo() }
			unless { foo() < 3 }
			unless { foo() <= 3 }
			unless { foo() > 3 }
			unless { foo() >= 3 }
			unless { foo() != 3 }
			unless { foo() == 3 }
            unless { foo() in Domain::"value" }
            unless { foo() has blah }
            when { foo() has "bar" }
            when { foo() like "h*ll*" };`, false},
		{"foo-like-foo", `permit(principal, action, resource)
			when { "f*o" like "f\*o" };`, false},
		{"ands", `permit(principal, action, resource)
			when { foo() && bar() && 3};`, false},
		{"ors_and_ands", `permit(principal, action, resource)
			when { foo() && bar() || baz() || 1 < 2 && 2 < 3};`, false},
		{"primaryExpression", `permit(principal, action, resource)
			when { (true) }
			unless { ((if (foo() <= 234) then principal else principal) like "") };`, false},
		{"primaryExprList", `permit(principal, action, resource)
			when { [] }
			unless { [true] }
			when { [123, (principal has "name" && principal.name == "alice")]};`, false},
		{"primaryRecInits", `permit(principal, action, resource)
			when { {} }
			unless { {"key": "some value"} }
			when { {"key": "some value", id: "another value"} };`, false},
		{"most-positive-long",
			`permit(principal,action,resource) when { 9223372036854775807 == -(-9223372036854775807) };`,
			false},
		{"principal-is", `permit (principal is X, action, resource);`, false},
		{"principal-is-long", `permit (principal is X::Y, action, resource);`, false},
		{"principal-is-in", `permit (principal is X in X::"z", action, resource);`, false},
		{"resource-is", `permit (principal, action, resource is X);`, false},
		{"resource-is-long", `permit (principal, action, resource is X::Y);`, false},
		{"resource-is-in", `permit (principal, action, resource is X in X::"z");`, false},
		{"when-is", `permit (principal, action, resource) when { principal is X };`, false},
		{"when-is-long", `permit (principal, action, resource) when { principal is X::Y };`, false},
		{"when-is-in", `permit (principal, action, resource) when { principal is X in X::"z" };`, false},

		{"most-negative-long", `permit(principal,action,resource) when { -9223372036854775808 == -9223372036854775808 };`, false},
		{"most-negative-long2", `permit(principal,action,resource) when { -9223372036854775808 < -9223372036854775807 };`, false},

		// Error cases
		{"missingEffect", `@id("test")`, true},
		{"invalidEffect", `invalidEffect(principal, action, resource);`, true},
		{"missingSemicolon", `permit(principal, action, resource)`, true},
		{"missingScope", `permit;`, true},
		{"missingPrincipal", `permit(resource, action);`, true},
		{"missingResourceAndAction", `permit(principal);`, true},
		{"missingResource", `permit(principal, action);`, true},
		{"eofInScope", `permit(principal`, true},
		{"missingAction", `permit(principal, resource);`, true},
		{"invalidResource", `permit(principal, action, other);`, true},
		{"missingScopeEndParen", `permit(principal, action, resource;`, true},
		{"missingEntity", `permit(principal ==`, true},
		{"invalidEntity", `permit(principal == "alice", action, resource);`, true},
		{"invalidEntity2", `permit(principal == User::, action, resource);`, true},
		{"invalidEntity3", `permit(principal == User::123, action, resource);`, true},
		{"invalidEntity3", `permit(principal == User::`, true},
		{"invalidEntities", `permit(principal, action in [invalidEntity], resource);`, true},
		{"invalidEntities2", `permit(principal, action in [User::"alice", invalidEntity], resource);`, true},
		{"invalidEntities3", `permit(principal, action in [User::"alice";], resource);`, true},
		{"invalidEntities4", `permit(principal, action in [User::"alice"`, true},
		{"invalidAnnotation1", `@`, true},
		{"invalidAnnotation2", `@"annotate"`, true},
		{"invalidAnnotation3", `@annotate(`, true},
		{"invalidAnnotation4", `@annotate[""]`, true},
		{"invalidAnnotation5", `@annotate("test"]`, true},
		{"invalidAnnotation6", `@annotate(test)`, true},
		{"invalidCondition1", `permit(principal, action, resource) when`, true},
		{"invalidCondition2", `permit(principal, action, resource) when {`, true},
		{"invalidCondition3", `permit(principal, action, resource) when {}`, true},
		{"invalidCondition4", `permit(principal, action, resource) when { true`, true},
		{"invalidPrimaryInteger", `permit(principal, action, resource)
			when { 0xabcd };`, true},
		{"invalidPrimary", `permit(principal, action, resource)
			when { ( };`, true},
		{"invalidExtFun1", `permit(principal, action, resource)
			when { abcd`, true},
		{"invalidExtFun2", `permit(principal, action, resource)
			when { abcd(`, true},
		{"invalidExtFun3", `permit(principal, action, resource)
			when { abcd::`, true},
		{"invalidExtFun4", `permit(principal, action, resource)
			when { abcd::123`, true},
		{"invalidExtFun5", `permit(principal, action, resource)
			when { abcd(123`, true},
		{"invalidIfElseThen1", `permit(principal, action, resource)
			when { if }`, true},
		{"invalidIfElseThen2", `permit(principal, action, resource)
			when { if true }`, true},
		{"invalidIfElseThen3", `permit(principal, action, resource)
			when { if true then }`, true},
		{"invalidIfElseThen4", `permit(principal, action, resource)
			when { if true then principal }`, true},
		{"invalidIfElseThen5", `permit(principal, action, resource)
			when { if true then principal else }`, true},
		{"invalidAccess1", `permit(principal, action, resource)
			when { resource.`, true},
		{"invalidAccess2", `permit(principal, action, resource)
			when { resource.bar.123 };`, true},
		{"invalidAccess3", `permit(principal, action, resource)
			when { resource.bar(`, true},
		{"invalidAccess4", `permit(principal, action, resource)
			when { resource.bar(]`, true},
		{"invalidAccess5", `permit(principal, action, resource)
			when { resource.bar(,)`, true},
		{"invalidAccess6", `permit(principal, action, resource)
			when { resource.bar[`, true},
		{"invalidAccess7", `permit(principal, action, resource)
			when { resource.bar[baz]`, true},
		{"invalidAccess8", `permit(principal, action, resource)
			when { resource.bar["baz")`, true},
		{"invalidUnaryOp", `permit(principal, action, resource)
			when { +resource.bar };`, true},
		{"invalidAdd", `permit(principal, action, resource)
			when { resource.foo +`, true},
		{"invalidRelation", `permit(principal, action, resource)
			when { resource.name in`, true},
		{"invalidHas1", `permit(principal, action, resource)
			when { resource.name has`, true},
		{"invalidHas2", `permit(principal, action, resource)
			when { resource.name has 123`, true},
		{"invalidLike1", `permit(principal, action, resource)
			when { resource.name like`, true},
		{"invalidLike2", `permit(principal, action, resource)
			when { resource.name like foo`, true},
		{"invalidPrimaryExpr", `permit(principal, action, resource)
			when { (true`, true},
		{"invalidPrimaryExprList", `permit(principal, action, resource)
			when { [`, true},
		{"invalidActionEqRhs", `permit(principal, action == Foo, resource);`, true},
		{"invalidActionInRhs", `permit(principal, action in Foo, resource);`, true},
		{"invalidPrimaryRecInits1", `permit(principal, action, resource)
			when { {`, true},
		{"invalidPrimaryRecInits2", `permit(principal, action, resource)
			when { {123: "value"} };`, true},
		{"invalidPrimaryRecInits3", `permit(principal, action, resource)
			when { {"key" "value"} };`, true},
		{"invalidPrimaryRecInits4", `permit(principal, action, resource)
			when { {"key":`, true},
		{"invalidPrimaryRecInits5", `permit(principal, action, resource)
			when { {"key1": "value1" "key2": "value2" };`, true},

		{"invalidStringAnnotation", `@bananas("\*") permit (principal, action, resource);`, true},
		{"invalidStringEntityID", `permit(principal == User::"\*", action, resource);`, true},
		{"invalidStringHas", `permit(principal, action, resource) when { context has "\*" };`, true},
		{"invalidNumericLike", `permit(principal, action, resource) when { context.key like 42 };`, true},
		{"invalidPatternLike", `permit(principal, action, resource) when { context.key like "\u{DFFF}" };`, true},
		{"invalidStringPrimary", `permit(principal, action, resource) when { context.key == "\*" };`, true},
		{"invalidExtFun", `permit(principal, action, resource) when { principal == User::"\*" };`, true},
		{"invalidAccess", `permit(principal, action, resource) when { context["\*"] == 42 };`, true},
		{"invalidRecordKey", `permit(principal, action, resource) when { { "\*":42 } };`, true},
		{"invalidIs", `permit (principal is 1, action, resource);`, true},
		{"invalidIsLong", `permit (principal is X::1, action, resource);`, true},
		{"duplicateAnnotations", `@key("value") @key("value") permit (principal, action, resource);`, true},

		{"very-negative-long-bad", `permit(principal,action,resource) when { -9223372036823454775808 < -9224323372036854775807 };`, true},
		{"very-positive-long-bad", `permit(principal,action,resource) when { 9223372036823454775808 < 9224323372036854775807 };`, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tokens, err := Tokenize([]byte(tt.in))
			testutilOK(t, err)
			got, err := Parse(tokens)
			testutilEquals(t, err != nil, tt.err)
			if err != nil {
				testutilEquals(t, got, nil)
				return
			}

			gotTokens, err := Tokenize([]byte(got.String()))
			testutilOK(t, err)

			var tokenStrs []string
			for _, t := range tokens {
				tokenStrs = append(tokenStrs, t.toString())
			}

			var gotTokenStrs []string
			for _, t := range gotTokens {
				gotTokenStrs = append(gotTokenStrs, t.toString())
			}

			testutilEquals(t, gotTokenStrs, tokenStrs)
		})
	}
}

func TestParseTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		out  Policies
	}{
		{
			"first",
			"permit(principal, action, resource) when { 3 * 2 > 5 };",
			Policies{
				Policy{
					Position:    Position{Offset: 0, Line: 1, Column: 1},
					Annotations: []Annotation(nil),
					Effect:      "permit",
					Conditions: []Condition{
						{
							Type: "when",
							Expression: Expression{
								Type: ExpressionOr,
								Or: Or{
									Ands: []And{
										{
											Relations: []Relation{
												{
													Add: Add{
														Mults: []Mult{
															{
																Unaries: []Unary{
																	{
																		Ops: []UnaryOp(nil),
																		Member: Member{
																			Primary: Primary{
																				Type:    PrimaryLiteral,
																				Literal: Literal{Type: LiteralInt, Long: 3},
																			},
																			Accesses: []Access(nil),
																		},
																	},
																	{
																		Ops: []UnaryOp(nil),
																		Member: Member{
																			Primary: Primary{
																				Type:    PrimaryLiteral,
																				Literal: Literal{Type: LiteralInt, Long: 2},
																			},
																			Accesses: []Access(nil),
																		},
																	},
																},
															},
														},
													},
													Type:  "relop",
													RelOp: ">",
													RelOpRhs: Add{
														Mults: []Mult{
															{
																Unaries: []Unary{
																	{
																		Ops: []UnaryOp(nil),
																		Member: Member{
																			Primary: Primary{
																				Type:    PrimaryLiteral,
																				Literal: Literal{Type: LiteralInt, Long: 5},
																			},
																			Accesses: []Access(nil),
																		},
																	},
																},
															},
														},
													},
													Str: "",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tokens, err := Tokenize([]byte(tt.in))
			testutilOK(t, err)
			got, err := Parse(tokens)
			testutilOK(t, err)
			testutilEquals(t, got, tt.out)
		})
	}
}

func TestParseEntity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		out  Entity
		err  func(testing.TB, error)
	}{
		{"happy", `Action::"test"`, Entity{Path: []string{"Action", "test"}}, testutilOK},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			toks, err := Tokenize([]byte(tt.in))
			testutilOK(t, err)
			out, err := ParseEntity(toks)
			testutilEquals(t, out, tt.out)
			tt.err(t, err)
		})
	}
}

func TestPolicyPositions(t *testing.T) {
	t.Parallel()
	in := `// idk a comment
@blah("asdf")
permit( principal, action, resource );


// later on
  permit (principal, action, resource) ;

// annotation indent
 @test("1234") permit (principal, action, resource );
`
	toks, err := Tokenize([]byte(in))
	testutilOK(t, err)
	out, err := Parse(toks)
	testutilOK(t, err)
	testutilEquals(t, len(out), 3)
	testutilEquals(t, out[0].Position, Position{Offset: 17, Line: 2, Column: 1})
	testutilEquals(t, out[1].Position, Position{Offset: 86, Line: 7, Column: 3})
	testutilEquals(t, out[2].Position, Position{Offset: 148, Line: 10, Column: 2})
}
