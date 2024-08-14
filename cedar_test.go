package cedar

import (
	"bytes"
	"encoding/json"
	"net/netip"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/entities"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestEntityIsZero(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		uid  types.EntityUID
		want bool
	}{
		{"empty", types.EntityUID{}, true},
		{"empty-type", types.NewEntityUID("one", ""), false},
		{"empty-id", types.NewEntityUID("", "one"), false},
		{"not-empty", types.NewEntityUID("one", "two"), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, tt.uid.IsZero(), tt.want)
		})
	}
}

func TestNewPolicySet(t *testing.T) {
	t.Parallel()
	t.Run("err-in-tokenize", func(t *testing.T) {
		t.Parallel()
		_, err := NewPolicySet("policy.cedar", []byte(`"`))
		testutil.Error(t, err)
	})
	t.Run("err-in-parse", func(t *testing.T) {
		t.Parallel()
		_, err := NewPolicySet("policy.cedar", []byte(`err`))
		testutil.Error(t, err)
	})
	t.Run("annotations", func(t *testing.T) {
		t.Parallel()
		ps, err := NewPolicySet("policy.cedar", []byte(`@key("value") permit (principal, action, resource);`))
		testutil.OK(t, err)
		testutil.Equals(t, ps[0].Annotations, Annotations{"key": "value"})
	})
}

//nolint:revive // due to table test function-length
func TestIsAuthorized(t *testing.T) {
	t.Parallel()
	cuzco := types.NewEntityUID("coder", "cuzco")
	dropTable := types.NewEntityUID("table", "drop")
	tests := []struct {
		Name                        string
		Policy                      string
		Entities                    entities.Entities
		Principal, Action, Resource types.EntityUID
		Context                     types.Record
		Want                        Decision
		DiagErr                     int
		ParseErr                    bool
	}{
		{
			Name:      "simple-permit",
			Policy:    `permit(principal,action,resource);`,
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities: entities.Entities{
				cuzco: entities.Entity{
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
			Entities: entities.Entities{
				cuzco: entities.Entity{
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
			Entities: entities.Entities{
				cuzco: entities.Entity{
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
			Entities:  entities.Entities{},
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
			Entities: entities.Entities{
				cuzco: entities.Entity{
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
			Entities:  entities.Entities{},
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
			Entities: entities.Entities{
				dropTable: entities.Entity{
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
			Entities: entities.Entities{
				dropTable: entities.Entity{
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities: entities.Entities{
				cuzco: entities.Entity{
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   1,
			ParseErr:  false,
		},
		{
			Name:      "permit-when-like",
			Policy:    `permit(principal,action,resource) when { "bananas" like "*nan*" };`,
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
			Principal: cuzco,
			Action:    dropTable,
			Resource:  types.NewEntityUID("table", "whatever"),
			Context:   types.Record{},
			Want:      false,
			DiagErr:   1,
			ParseErr:  false,
		},
		{
			Name: "permit-when-decimal",
			Policy: `permit(principal,action,resource) when {
				decimal("10.0").lessThan(decimal("11.0")) &&
				decimal("10.0").lessThanOrEqual(decimal("11.0")) &&
				decimal("10.0").greaterThan(decimal("9.0")) &&
				decimal("10.0").greaterThanOrEqual(decimal("9.0")) };`,
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities: entities.Entities{},
			Context:  types.Record{"value": types.Long(-42)},
			Want:     true,
			DiagErr:  0,
		},
		{
			Name:      "principal-is",
			Policy:    `permit(principal is Actor,action,resource);`,
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities:  entities.Entities{},
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
			Entities: entities.Entities{
				types.NewEntityUID("Resource", "table"): entities.Entity{
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
			ps, err := NewPolicySet("policy.cedar", []byte(tt.Policy))
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

func TestEntities(t *testing.T) {
	t.Parallel()
	t.Run("Clone", func(t *testing.T) {
		t.Parallel()
		e := entities.Entities{
			types.EntityUID{Type: "A", ID: "A"}: {},
			types.EntityUID{Type: "A", ID: "B"}: {},
			types.EntityUID{Type: "B", ID: "A"}: {},
			types.EntityUID{Type: "B", ID: "B"}: {},
		}
		clone := e.Clone()
		testutil.Equals(t, clone, e)
	})

}

func TestValueFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      types.Value
		outJSON string
	}{
		{
			name:    "string",
			in:      types.String("hello"),
			outJSON: `"hello"`,
		},
		{
			name:    "bool",
			in:      types.Boolean(true),
			outJSON: `true`,
		},
		{
			name:    "int64",
			in:      types.Long(42),
			outJSON: `42`,
		},
		{
			name:    "int64",
			in:      types.EntityUID{Type: "T", ID: "0"},
			outJSON: `{"__entity":{"type":"T","id":"0"}}`,
		},
		{
			name:    "record",
			in:      types.Record{"K": types.Boolean(true)},
			outJSON: `{"K":true}`,
		},
		{
			name:    "netipPrefix",
			in:      types.IPAddr(netip.MustParsePrefix("192.168.0.42/32")),
			outJSON: `{"__extn":{"fn":"ip","arg":"192.168.0.42"}}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := tt.in.ExplicitMarshalJSON()
			testutil.OK(t, err)
			testutil.Equals(t, string(out), tt.outJSON)
		})
	}
}

func TestError(t *testing.T) {
	t.Parallel()
	e := Error{Policy: 42, Message: "bad error"}
	testutil.Equals(t, e.String(), "while evaluating policy `policy42`: bad error")
}

func TestCorpusRelated(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		policy   string
		request  Request
		decision Decision
		reasons  []int
		errors   []int
	}{
		{
			"0cb1ad7042508e708f1999284b634ed0f334bc00",
			`forbid(
			principal in a::"\0\0",
			action == Action::"action",
			resource
		  ) when {
			(true && (((!870985681610) == principal) == principal)) && principal
		};`,
			Request{Principal: types.NewEntityUID("a", "\u0000\u0000"), Action: types.NewEntityUID("Action", "action"), Resource: types.NewEntityUID("a", "\u0000\u0000")},
			Deny,
			nil,
			[]int{0},
		},

		{
			"0cb1ad7042508e708f1999284b634ed0f334bc00/partial1",
			`forbid(
			principal in a::"\0\0",
			action == Action::"action",
			resource
		  ) when {
			(((!870985681610) == principal) == principal)
		};`,
			Request{Principal: types.NewEntityUID("a", "\u0000\u0000"), Action: types.NewEntityUID("Action", "action"), Resource: types.NewEntityUID("a", "\u0000\u0000")},
			Deny,
			nil,
			[]int{0},
		},
		{
			"0cb1ad7042508e708f1999284b634ed0f334bc00/partial2",
			`forbid(
			principal in a::"\0\0",
			action == Action::"action",
			resource
		  ) when {
			((!870985681610) == principal)
		};`,
			Request{Principal: types.NewEntityUID("a", "\u0000\u0000"), Action: types.NewEntityUID("Action", "action"), Resource: types.NewEntityUID("a", "\u0000\u0000")},
			Deny,
			nil,
			[]int{0},
		},

		{
			"0cb1ad7042508e708f1999284b634ed0f334bc00/partial3",
			`forbid(
			principal in a::"\0\0",
			action == Action::"action",
			resource
		  ) when {
			(!870985681610)
		};`,
			Request{Principal: types.NewEntityUID("a", "\u0000\u0000"), Action: types.NewEntityUID("Action", "action"), Resource: types.NewEntityUID("a", "\u0000\u0000")},
			Deny,
			nil,
			[]int{0},
		},

		{
			"0cb1ad7042508e708f1999284b634ed0f334bc00/partial2/simplified",
			`forbid(
			principal,
			action,
			resource
		  ) when {
			((!42) == principal)
		};`,
			Request{},
			Deny,
			nil,
			[]int{0},
		},

		{
			"0cb1ad7042508e708f1999284b634ed0f334bc00/partial2/simplified2",
			`forbid(
				principal,
				action,
				resource
			) when {
				(!42 == principal)
			};`,
			Request{},
			Deny,
			nil,
			[]int{0},
		},

		{"48d0ba6537a3efe02112ba0f5a3daabdcad27b04",
			`forbid(
				principal,
				action in [Action::"action"],
				resource is a in a::"\0\u{8}\u{11}\0R"
			  ) when {
				true && ((if (principal in action) then (ip("")) else (if true then (ip("6b6b:f00::32ff:ffff:6368/00")) else (ip("7265:6c69:706d:6f43:5f74:6f70:7374:6f68")))).isMulticast())
			  };`,
			Request{Principal: types.NewEntityUID("a", "\u0000\b\u0011\u0000R"), Action: types.NewEntityUID("Action", "action"), Resource: types.NewEntityUID("a", "\u0000\b\u0011\u0000R")},
			Deny,
			nil,
			[]int{0},
		},

		{"48d0ba6537a3efe02112ba0f5a3daabdcad27b04/simplified",
			`forbid(
			principal,
			action,
			resource
		  ) when {
			true && ip("6b6b:f00::32ff:ffff:6368/00").isMulticast()
		  };`,
			Request{},
			Deny,
			nil,
			[]int{0},
		},

		{name: "e91da4e6af5c73e27f5fb610d723dfa21635d10b",
			policy: `forbid(
				principal is a in a::"\0\0(W\0\0\0",
				action,
				resource
			  ) when {
				true && (([ip("c5c5:c5c5:c5c5:c5c5:c5c5:c5c5:c5c5:c5c5/68")].containsAll([ip("c5c5:c5c5:c5c5:c5c5:c5c5:5cc5:c5c5:c5c5/68")])) || ((ip("")) == (ip(""))))
			  };`,
			request:  Request{Principal: types.NewEntityUID("a", "\u0000\u0000(W\u0000\u0000\u0000"), Action: types.NewEntityUID("Action", "action"), Resource: types.NewEntityUID("a", "")},
			decision: Deny,
			reasons:  nil,
			errors:   []int{0},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			policy, err := NewPolicySet("", []byte(tt.policy))
			testutil.OK(t, err)
			ok, diag := policy.IsAuthorized(entities.Entities{}, tt.request)
			testutil.Equals(t, ok, tt.decision)
			var reasons []int
			for _, n := range diag.Reasons {
				reasons = append(reasons, n.Policy)
			}
			testutil.Equals(t, reasons, tt.reasons)
			var errors []int
			for _, n := range diag.Errors {
				errors = append(errors, n.Policy)
			}
			testutil.Equals(t, errors, tt.errors)
		})
	}
}

func prettifyJson(in []byte) []byte {
	var buf bytes.Buffer
	_ = json.Indent(&buf, in, "", "    ")
	return buf.Bytes()
}

func TestPolicyJSON(t *testing.T) {
	t.Parallel()

	// Taken from https://docs.cedarpolicy.com/policies/json-format.html
	jsonEncodedPolicy := prettifyJson([]byte(`
		{
			"effect": "permit",
			"principal": {
				"op": "==",
				"entity": { "type": "User", "id": "12UA45" }
			},
			"action": {
				"op": "==",
				"entity": { "type": "Action", "id": "view" }
			},
			"resource": {
				"op": "in",
				"entity": { "type": "Folder", "id": "abc" }
			},
			"conditions": [
				{
					"kind": "when",
					"body": {
						"==": {
							"left": {
								".": {
									"left": {
										"Var": "context"
									},
									"attr": "tls_version"
								}
							},
							"right": {
								"Value": "1.3"
							}
						}
					}
				}
			]
		}`,
	))

	var policy Policy
	testutil.OK(t, policy.UnmarshalJSON(jsonEncodedPolicy))

	output, err := policy.MarshalJSON()
	testutil.OK(t, err)

	testutil.Equals(t, string(prettifyJson(output)), string(jsonEncodedPolicy))
}

func TestEntitiesJSON(t *testing.T) {
	t.Parallel()
	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()
		e := entities.Entities{}
		ent := entities.Entity{
			UID:        types.NewEntityUID("Type", "id"),
			Parents:    []types.EntityUID{},
			Attributes: types.Record{"key": types.Long(42)},
		}
		e[ent.UID] = ent
		b, err := e.MarshalJSON()
		testutil.OK(t, err)
		testutil.Equals(t, string(b), `[{"uid":{"type":"Type","id":"id"},"attrs":{"key":42}}]`)
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()
		b := []byte(`[{"uid":{"type":"Type","id":"id"},"parents":[],"attrs":{"key":42}}]`)
		var e entities.Entities
		err := json.Unmarshal(b, &e)
		testutil.OK(t, err)
		want := entities.Entities{}
		ent := entities.Entity{
			UID:        types.NewEntityUID("Type", "id"),
			Parents:    []types.EntityUID{},
			Attributes: types.Record{"key": types.Long(42)},
		}
		want[ent.UID] = ent
		testutil.Equals(t, e, want)
	})

	t.Run("UnmarshalErr", func(t *testing.T) {
		t.Parallel()
		var e entities.Entities
		err := e.UnmarshalJSON([]byte(`!@#$`))
		testutil.Error(t, err)
	})
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
