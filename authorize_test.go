package cedar

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

//nolint:revive // due to table test function-length
func TestIsAuthorized(t *testing.T) {
	t.Parallel()
	cuzco := types.NewEntityUID("coder", "cuzco")
	dropTable := types.NewEntityUID("table", "drop")
	tests := []struct {
		Name                        string
		Policy                      string
		Entities                    types.Entities
		Principal, Action, Resource types.EntityUID
		Context                     types.Record
		Want                        Decision
		DiagErr                     int
		ParseErr                    bool
	}{
		{
			Name:      "simple-permit",
			Policy:    `permit(principal,action,resource);`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "simple-forbid",
			Policy:    `forbid(principal,action,resource);`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   0,
		},
		{
			Name:      "no-permit",
			Policy:    `permit(principal,action,resource in asdf::"1234");`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   0,
		},
		{
			Name:      "error-in-policy",
			Policy:    `permit(principal,action,resource) when { resource in "foo" };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name: "error-in-policy-continues",
			Policy: `permit(principal,action,resource) when { resource in "foo" };
			permit(principal,action,resource);
			`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   1,
		},
		{
			Name:      "permit-requires-context-success",
			Policy:    `permit(principal,action,resource) when { context.x == 42 };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{"x": types.Long(42)},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-requires-context-fail",
			Policy:    `permit(principal,action,resource) when { context.x == 42 };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{"x": types.Long(43)},
			Want:      false,
			DiagErr:   0,
		},
		{
			Name:   "permit-requires-entities-success",
			Policy: `permit(principal,action,resource) when { principal.x == 42 };`,
			Entities: types.Entities{
				cuzco: types.Entity{
					UID:        cuzco,
					Attributes: types.Record{"x": types.Long(42)},
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-requires-entities-fail",
			Policy: `permit(principal,action,resource) when { principal.x == 42 };`,
			Entities: types.Entities{
				cuzco: types.Entity{
					UID:        cuzco,
					Attributes: types.Record{"x": types.Long(43)},
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   0,
		},
		{
			Name:   "permit-requires-entities-parent-success",
			Policy: `permit(principal,action,resource) when { principal in parent::"bob" };`,
			Entities: types.Entities{
				cuzco: types.Entity{
					UID:     cuzco,
					Parents: []types.EntityUID{types.NewEntityUID("parent", "bob")},
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-principal-equals",
			Policy:    `permit(principal == coder::"cuzco",action,resource);`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-principal-in",
			Policy: `permit(principal in team::"osiris",action,resource);`,
			Entities: types.Entities{
				cuzco: types.Entity{
					UID:     cuzco,
					Parents: []types.EntityUID{types.NewEntityUID("team", "osiris")},
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-action-equals",
			Policy:    `permit(principal,action == table::"drop",resource);`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-action-in",
			Policy: `permit(principal,action in scary::"stuff",resource);`,
			Entities: types.Entities{
				dropTable: types.Entity{
					UID:     dropTable,
					Parents: []types.EntityUID{types.NewEntityUID("scary", "stuff")},
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-action-in-set",
			Policy: `permit(principal,action in [scary::"stuff"],resource);`,
			Entities: types.Entities{
				dropTable: types.Entity{
					UID:     dropTable,
					Parents: []types.EntityUID{types.NewEntityUID("scary", "stuff")},
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-resource-equals",
			Policy:    `permit(principal,action,resource == table::"whatever");`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-unless",
			Policy:    `permit(principal,action,resource) unless { false };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-if",
			Policy:    `permit(principal,action,resource) when { (if true then true else true) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-or",
			Policy:    `permit(principal,action,resource) when { (true || false) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-and",
			Policy:    `permit(principal,action,resource) when { (true && true) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-relations",
			Policy:    `permit(principal,action,resource) when { (1<2) && (1<=1) && (2>1) && (1>=1) && (1!=2) && (1==1)};`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-relations-in",
			Policy:    `permit(principal,action,resource) when { principal in principal };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-when-relations-has",
			Policy: `permit(principal,action,resource) when { principal has name };`,
			Entities: types.Entities{
				cuzco: types.Entity{
					UID:        cuzco,
					Attributes: types.Record{"name": types.String("bob")},
				},
			},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-add-sub",
			Policy:    `permit(principal,action,resource) when { 40+3-1==42 };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-mul",
			Policy:    `permit(principal,action,resource) when { 6*7==42 };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-negate",
			Policy:    `permit(principal,action,resource) when { -42==-42 };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-not",
			Policy:    `permit(principal,action,resource) when { !(1+1==42) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set",
			Policy:    `permit(principal,action,resource) when { [1,2,3].contains(2) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-record",
			Policy:    `permit(principal,action,resource) when { {name:"bob"} has name };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-action",
			Policy:    `permit(principal,action,resource) when { action in action };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set-contains-ok",
			Policy:    `permit(principal,action,resource) when { [1,2,3].contains(2) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set-contains-error",
			Policy:    `permit(principal,action,resource) when { [1,2,3].contains(2,3) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   0,
			ParseErr:  true,
		},
		{
			Name:      "permit-when-set-containsAll-ok",
			Policy:    `permit(principal,action,resource) when { [1,2,3].containsAll([2,3]) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set-containsAll-error",
			Policy:    `permit(principal,action,resource) when { [1,2,3].containsAll(2,3) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   0,
			ParseErr:  true,
		},
		{
			Name:      "permit-when-set-containsAny-ok",
			Policy:    `permit(principal,action,resource) when { [1,2,3].containsAny([2,5]) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set-containsAny-error",
			Policy:    `permit(principal,action,resource) when { [1,2,3].containsAny(2,5) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   0,
			ParseErr:  true,
		},
		{
			Name:      "permit-when-record-attr",
			Policy:    `permit(principal,action,resource) when { {name:"bob"}["name"] == "bob" };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-unknown-method",
			Policy:    `permit(principal,action,resource) when { [1,2,3].shuffle() };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   0,
			ParseErr:  true,
		},
		{
			Name:      "permit-when-like",
			Policy:    `permit(principal,action,resource) when { "bananas" like "*nan*" };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-unknown-ext-fun",
			Policy:    `permit(principal,action,resource) when { fooBar("10") };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   0,
			ParseErr:  true,
		},
		{
			Name: "permit-when-decimal",
			Policy: `permit(principal,action,resource) when {
				decimal("10.0").lessThan(decimal("11.0")) &&
				decimal("10.0").lessThanOrEqual(decimal("11.0")) &&
				decimal("10.0").greaterThan(decimal("9.0")) &&
				decimal("10.0").greaterThanOrEqual(decimal("9.0")) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-decimal-fun-wrong-arity",
			Policy:    `permit(principal,action,resource) when { decimal(1, 2) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name: "permit-when-ip",
			Policy: `permit(principal,action,resource) when {
				ip("1.2.3.4").isIpv4() &&
				ip("a:b:c:d::/16").isIpv6() &&
				ip("::1").isLoopback() &&
				ip("224.1.2.3").isMulticast() &&
				ip("127.0.0.1").isInRange(ip("127.0.0.0/16"))};`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-ip-fun-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip() };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isIpv4-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isIpv4(true) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isIpv6-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isIpv6(true) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isLoopback-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isLoopback(true) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isMulticast-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isMulticast(true) };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isInRange-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isInRange() };`,
			Entities:  types.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:     "negative-unary-op",
			Policy:   `permit(principal,action,resource) when { -context.value > 0 };`,
			Entities: types.Entities{},
			Context:  types.Record{"value": types.Long(-42)},
			Want:     true,
			DiagErr:  0,
		},
		{
			Name:      "principal-is",
			Policy:    `permit(principal is Actor,action,resource);`,
			Entities:  types.Entities{},
			Principal: types.NewEntityUID("Actor", "cuzco"),
			Action:    types.NewEntityUID("Action", "drop"),
			Resource:  types.NewEntityUID("Resource", "table"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "principal-is-in",
			Policy:    `permit(principal is Actor in Actor::"cuzco",action,resource);`,
			Entities:  types.Entities{},
			Principal: types.NewEntityUID("Actor", "cuzco"),
			Action:    types.NewEntityUID("Action", "drop"),
			Resource:  types.NewEntityUID("Resource", "table"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "resource-is",
			Policy:    `permit(principal,action,resource is Resource);`,
			Entities:  types.Entities{},
			Principal: types.NewEntityUID("Actor", "cuzco"),
			Action:    types.NewEntityUID("Action", "drop"),
			Resource:  types.NewEntityUID("Resource", "table"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "resource-is-in",
			Policy:    `permit(principal,action,resource is Resource in Resource::"table");`,
			Entities:  types.Entities{},
			Principal: types.NewEntityUID("Actor", "cuzco"),
			Action:    types.NewEntityUID("Action", "drop"),
			Resource:  types.NewEntityUID("Resource", "table"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "when-is",
			Policy:    `permit(principal,action,resource) when { resource is Resource };`,
			Entities:  types.Entities{},
			Principal: types.NewEntityUID("Actor", "cuzco"),
			Action:    types.NewEntityUID("Action", "drop"),
			Resource:  types.NewEntityUID("Resource", "table"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "when-is-in",
			Policy:    `permit(principal,action,resource) when { resource is Resource in Resource::"table" };`,
			Entities:  types.Entities{},
			Principal: types.NewEntityUID("Actor", "cuzco"),
			Action:    types.NewEntityUID("Action", "drop"),
			Resource:  types.NewEntityUID("Resource", "table"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "when-is-in",
			Policy: `permit(principal,action,resource) when { resource is Resource in Parent::"id" };`,
			Entities: types.Entities{
				types.NewEntityUID("Resource", "table"): types.Entity{
					UID:     types.NewEntityUID("Resource", "table"),
					Parents: []types.EntityUID{types.NewEntityUID("Parent", "id")},
				},
			},
			Principal: types.NewEntityUID("Actor", "cuzco"),
			Action:    types.NewEntityUID("Action", "drop"),
			Resource:  types.NewEntityUID("Resource", "table"),
			Context:   types.Record{},
			Want:      true,
			DiagErr:   0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			ps, err := NewPolicySetFromBytes("policy.cedar", []byte(tt.Policy))
			testutil.Equals(t, (err != nil), tt.ParseErr)
			ok, diag := ps.IsAuthorized(tt.Entities, Request{
				Principal: tt.Principal,
				Action:    tt.Action,
				Resource:  tt.Resource,
				Context:   tt.Context,
			})
			testutil.Equals(t, len(diag.Errors), tt.DiagErr)
			testutil.Equals(t, ok, tt.Want)
		})
	}
}

func TestError(t *testing.T) {
	t.Parallel()
	e := Error{PolicyID: "policy42", Message: "bad error"}
	testutil.Equals(t, e.String(), "while evaluating policy `policy42`: bad error")
}

func TestJSONDecision(t *testing.T) {
	t.Parallel()
	t.Run("MarshalAllow", func(t *testing.T) {
		t.Parallel()
		d := Allow
		b, err := d.MarshalJSON()
		testutil.OK(t, err)
		testutil.Equals(t, string(b), `"allow"`)
	})
	t.Run("MarshalDeny", func(t *testing.T) {
		t.Parallel()
		d := Deny
		b, err := d.MarshalJSON()
		testutil.OK(t, err)
		testutil.Equals(t, string(b), `"deny"`)
	})
	t.Run("UnmarshalAllow", func(t *testing.T) {
		t.Parallel()
		var d Decision
		err := json.Unmarshal([]byte(`"allow"`), &d)
		testutil.OK(t, err)
		testutil.Equals(t, d, Allow)
	})
	t.Run("UnmarshalDeny", func(t *testing.T) {
		t.Parallel()
		var d Decision
		err := json.Unmarshal([]byte(`"deny"`), &d)
		testutil.OK(t, err)
		testutil.Equals(t, d, Deny)
	})
}
