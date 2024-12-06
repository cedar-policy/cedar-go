package parser_test

import (
	"bytes"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/parser"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
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
				when { ip() }
				when { ip("hello") }
				when { ip(context.someString) };
			`, false},
		{"ifElseThen", `permit(principal, action, resource)
			when { if false then principal else principal };`, false},
		{"access", `permit(principal, action, resource)
			when { resource.foo }
			unless { resource.foo.bar }
			when { principal.isIpv4() }
			unless { principal.isIpv4(false) }
			when { action.foo["bar"].isIpv4() }
			unless { principal.isIpv4(false, 123, "foo") }
			when { principal["foo"] };`, false},
		{"tags", `permit(principal, action, resource)
			when { resource.hasTag("blue") };

			permit(principal, action, resource)
			when { resource.getTag("blue") };

			permit(principal, action, resource)
			when { resource.hasTag(context.color) };

			permit(principal, action, resource)
			when { resource.getTag(context.color) };
			`, false},
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
			when { ip() }
			unless { ip() < 3 }
			unless { ip() <= 3 }
			unless { ip() > 3 }
			unless { ip() >= 3 }
			unless { ip() != 3 }
			unless { ip() == 3 }
            unless { ip() in Domain::"value" }
            unless { ip() has blah }
            when { ip() has "bar" }
            when { ip() like "h*ll*" };`, false},
		{"foo-like-foo", `permit(principal, action, resource)
			when { "f*o" like "f\*o" };`, false},
		{"ands", `permit(principal, action, resource)
			when { ip() && decimal() && 3};`, false},
		{"ors_and_ands", `permit(principal, action, resource)
			when { ip() && decimal() || ip() || 1 < 2 && 2 < 3};`, false},
		{"primaryExpression", `permit(principal, action, resource)
			when { (true) }
			unless { ((if (ip() <= 234) then principal else principal) like "") };`, false},
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
		{"invalidTag1", `permit(principal, action, resource)
			when { resource.getTag(42)}`, true},
		{"invalidTag2", `permit(principal, action, resource)
			when { resource.hasTag(42)}`, true},
		{"invalidTag3", `permit(principal, action, resource)
			when { resource.hasTag(12.1 + 3.6)}`, true},
		{"invalidTag4", `permit(principal, action, resource)
			when { resource.hasTag(true)}`, true},
		{"invalidTag5", `permit(principal, action, resource)
			when { "blue".hasTag("true")}`, true},
		{"invalidTag6", `permit(principal, action, resource)
			when { 42.hasTag("true")}`, true},
		{"invalidTag7", `permit(principal, action, resource)
			when { true.hasTag("true")}`, true},
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

			var policies parser.PolicySlice
			err := policies.UnmarshalCedar([]byte(tt.in))
			if tt.err {
				testutil.Error(t, err)
				return
			}
			testutil.OK(t, err)

			// N.B. Until we support the re-rendering of comments, we have to ignore the position for the purposes of
			// these tests (see test "ex1")
			for _, pp := range policies {
				pp.Position = ast.Position{Offset: 0, Line: 1, Column: 1}

				var buf bytes.Buffer
				pp.MarshalCedar(&buf)

				var p2 parser.PolicySlice
				err = p2.UnmarshalCedar(buf.Bytes())
				testutil.OK(t, err)

				testutil.Equals(t, p2[0], pp)
			}
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

	var out parser.PolicySlice
	err := out.UnmarshalCedar([]byte(in))
	testutil.OK(t, err)
	testutil.Equals(t, len(out), 3)
	testutil.Equals(t, out[0].Position, ast.Position{Offset: 17, Line: 2, Column: 1})
	testutil.Equals(t, out[1].Position, ast.Position{Offset: 86, Line: 7, Column: 3})
	testutil.Equals(t, out[2].Position, ast.Position{Offset: 148, Line: 10, Column: 2})
}
