package cedar

import (
	"net/netip"
	"testing"
)

func TestEntityIsZero(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		uid  EntityUID
		want bool
	}{
		{"empty", EntityUID{}, true},
		{"empty-type", NewEntityUID("one", ""), false},
		{"empty-id", NewEntityUID("", "one"), false},
		{"not-empty", NewEntityUID("one", "two"), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutilEquals(t, tt.uid.IsZero(), tt.want)
		})
	}
}

func TestNewPolicySet(t *testing.T) {
	t.Parallel()
	t.Run("err-in-tokenize", func(t *testing.T) {
		t.Parallel()
		_, err := NewPolicySet("policy.cedar", []byte(`"`))
		testutilError(t, err)
	})
	t.Run("err-in-parse", func(t *testing.T) {
		t.Parallel()
		_, err := NewPolicySet("policy.cedar", []byte(`err`))
		testutilError(t, err)
	})
	t.Run("annotations", func(t *testing.T) {
		t.Parallel()
		ps, err := NewPolicySet("policy.cedar", []byte(`@key("value") permit (principal, action, resource);`))
		testutilOK(t, err)
		testutilEquals(t, ps[0].Annotations, Annotations{"key": "value"})
	})
}

func TestIsAuthorized(t *testing.T) {
	t.Parallel()
	tests := []struct {
		Name                        string
		Policy                      string
		Entities                    Entities
		Principal, Action, Resource EntityUID
		Context                     Record
		Want                        Decision
		DiagErr                     int
	}{
		{
			Name:      "simple-permit",
			Policy:    `permit(principal,action,resource);`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "simple-forbid",
			Policy:    `forbid(principal,action,resource);`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   0,
		},
		{
			Name:      "no-permit",
			Policy:    `permit(principal,action,resource in asdf::"1234");`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   0,
		},
		{
			Name:      "error-in-policy",
			Policy:    `permit(principal,action,resource) when { resource in "foo" };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name: "error-in-policy-continues",
			Policy: `permit(principal,action,resource) when { resource in "foo" };
			permit(principal,action,resource);
			`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   1,
		},
		{
			Name:      "permit-requires-context-success",
			Policy:    `permit(principal,action,resource) when { context.x == 42 };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{"x": Long(42)},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-requires-context-fail",
			Policy:    `permit(principal,action,resource) when { context.x == 42 };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{"x": Long(43)},
			Want:      false,
			DiagErr:   0,
		},
		{
			Name:   "permit-requires-entities-success",
			Policy: `permit(principal,action,resource) when { principal.x == 42 };`,
			Entities: entitiesFromSlice([]Entity{
				{
					UID:        EntityUID{"coder", "cuzco"},
					Attributes: Record{"x": Long(42)},
				},
			}),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-requires-entities-fail",
			Policy: `permit(principal,action,resource) when { principal.x == 42 };`,
			Entities: entitiesFromSlice([]Entity{
				{
					UID:        EntityUID{"coder", "cuzco"},
					Attributes: Record{"x": Long(43)},
				},
			}),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   0,
		},
		{
			Name:   "permit-requires-entities-parent-success",
			Policy: `permit(principal,action,resource) when { principal in parent::"bob" };`,
			Entities: entitiesFromSlice([]Entity{
				{
					UID:     EntityUID{"coder", "cuzco"},
					Parents: []EntityUID{{"parent", "bob"}},
				},
			}),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-principal-equals",
			Policy:    `permit(principal == coder::"cuzco",action,resource);`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-principal-in",
			Policy: `permit(principal in team::"osiris",action,resource);`,
			Entities: entitiesFromSlice([]Entity{
				{
					UID:     EntityUID{"coder", "cuzco"},
					Parents: []EntityUID{{"team", "osiris"}},
				},
			}),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-action-equals",
			Policy:    `permit(principal,action == table::"drop",resource);`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-action-in",
			Policy: `permit(principal,action in scary::"stuff",resource);`,
			Entities: entitiesFromSlice([]Entity{
				{
					UID:     EntityUID{"table", "drop"},
					Parents: []EntityUID{{"scary", "stuff"}},
				},
			}),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-action-in-set",
			Policy: `permit(principal,action in [scary::"stuff"],resource);`,
			Entities: entitiesFromSlice([]Entity{
				{
					UID:     EntityUID{"table", "drop"},
					Parents: []EntityUID{{"scary", "stuff"}},
				},
			}),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-resource-equals",
			Policy:    `permit(principal,action,resource == table::"whatever");`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-unless",
			Policy:    `permit(principal,action,resource) unless { false };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-if",
			Policy:    `permit(principal,action,resource) when { (if true then true else true) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-or",
			Policy:    `permit(principal,action,resource) when { (true || false) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-and",
			Policy:    `permit(principal,action,resource) when { (true && true) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-relations",
			Policy:    `permit(principal,action,resource) when { (1<2) && (1<=1) && (2>1) && (1>=1) && (1!=2) && (1==1)};`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-relations-in",
			Policy:    `permit(principal,action,resource) when { principal in principal };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "permit-when-relations-has",
			Policy: `permit(principal,action,resource) when { principal has name };`,
			Entities: entitiesFromSlice([]Entity{
				{
					UID:        EntityUID{"coder", "cuzco"},
					Attributes: Record{"name": String("bob")},
				},
			}),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-add-sub",
			Policy:    `permit(principal,action,resource) when { 40+3-1==42 };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-mul",
			Policy:    `permit(principal,action,resource) when { 6*7==42 };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-negate",
			Policy:    `permit(principal,action,resource) when { -42==-42 };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-not",
			Policy:    `permit(principal,action,resource) when { !(1+1==42) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set",
			Policy:    `permit(principal,action,resource) when { [1,2,3].contains(2) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-record",
			Policy:    `permit(principal,action,resource) when { {name:"bob"} has name };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-action",
			Policy:    `permit(principal,action,resource) when { action in action };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set-contains-ok",
			Policy:    `permit(principal,action,resource) when { [1,2,3].contains(2) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set-contains-error",
			Policy:    `permit(principal,action,resource) when { [1,2,3].contains(2,3) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-set-containsAll-ok",
			Policy:    `permit(principal,action,resource) when { [1,2,3].containsAll([2,3]) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set-containsAll-error",
			Policy:    `permit(principal,action,resource) when { [1,2,3].containsAll(2,3) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-set-containsAny-ok",
			Policy:    `permit(principal,action,resource) when { [1,2,3].containsAny([2,5]) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-set-containsAny-error",
			Policy:    `permit(principal,action,resource) when { [1,2,3].containsAny(2,5) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-record-attr",
			Policy:    `permit(principal,action,resource) when { {name:"bob"}["name"] == "bob" };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-unknown-method",
			Policy:    `permit(principal,action,resource) when { [1,2,3].shuffle() };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-like",
			Policy:    `permit(principal,action,resource) when { "bananas" like "*nan*" };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-unknown-ext-fun",
			Policy:    `permit(principal,action,resource) when { fooBar("10") };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name: "permit-when-decimal",
			Policy: `permit(principal,action,resource) when {
				decimal("10.0").lessThan(decimal("11.0")) &&
				decimal("10.0").lessThanOrEqual(decimal("11.0")) &&
				decimal("10.0").greaterThan(decimal("9.0")) &&
				decimal("10.0").greaterThanOrEqual(decimal("9.0")) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-decimal-fun-wrong-arity",
			Policy:    `permit(principal,action,resource) when { decimal(1, 2) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
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
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "permit-when-ip-fun-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip() };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isIpv4-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isIpv4(true) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isIpv6-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isIpv6(true) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isLoopback-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isLoopback(true) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isMulticast-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isMulticast(true) };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:      "permit-when-isInRange-wrong-arity",
			Policy:    `permit(principal,action,resource) when { ip("1.2.3.4").isInRange() };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"coder", "cuzco"},
			Action:    EntityUID{"table", "drop"},
			Resource:  EntityUID{"table", "whatever"},
			Context:   Record{},
			Want:      false,
			DiagErr:   1,
		},
		{
			Name:     "negative-unary-op",
			Policy:   `permit(principal,action,resource) when { -context.value > 0 };`,
			Entities: entitiesFromSlice(nil),
			Context:  Record{"value": Long(-42)},
			Want:     true,
			DiagErr:  0,
		},
		{
			Name:      "principal-is",
			Policy:    `permit(principal is Actor,action,resource);`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"Actor", "cuzco"},
			Action:    EntityUID{"Action", "drop"},
			Resource:  EntityUID{"Resource", "table"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "principal-is-in",
			Policy:    `permit(principal is Actor in Actor::"cuzco",action,resource);`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"Actor", "cuzco"},
			Action:    EntityUID{"Action", "drop"},
			Resource:  EntityUID{"Resource", "table"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "resource-is",
			Policy:    `permit(principal,action,resource is Resource);`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"Actor", "cuzco"},
			Action:    EntityUID{"Action", "drop"},
			Resource:  EntityUID{"Resource", "table"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "resource-is-in",
			Policy:    `permit(principal,action,resource is Resource in Resource::"table");`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"Actor", "cuzco"},
			Action:    EntityUID{"Action", "drop"},
			Resource:  EntityUID{"Resource", "table"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "when-is",
			Policy:    `permit(principal,action,resource) when { resource is Resource };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"Actor", "cuzco"},
			Action:    EntityUID{"Action", "drop"},
			Resource:  EntityUID{"Resource", "table"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:      "when-is-in",
			Policy:    `permit(principal,action,resource) when { resource is Resource in Resource::"table" };`,
			Entities:  entitiesFromSlice(nil),
			Principal: EntityUID{"Actor", "cuzco"},
			Action:    EntityUID{"Action", "drop"},
			Resource:  EntityUID{"Resource", "table"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
		{
			Name:   "when-is-in",
			Policy: `permit(principal,action,resource) when { resource is Resource in Parent::"id" };`,
			Entities: entitiesFromSlice([]Entity{
				{
					UID:     EntityUID{"Resource", "table"},
					Parents: []EntityUID{{"Parent", "id"}},
				},
			}),
			Principal: EntityUID{"Actor", "cuzco"},
			Action:    EntityUID{"Action", "drop"},
			Resource:  EntityUID{"Resource", "table"},
			Context:   Record{},
			Want:      true,
			DiagErr:   0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			ps, err := NewPolicySet("policy.cedar", []byte(tt.Policy))
			testutilOK(t, err)
			ok, diag := ps.IsAuthorized(tt.Entities, Request{
				Principal: tt.Principal,
				Action:    tt.Action,
				Resource:  tt.Resource,
				Context:   tt.Context,
			})
			testutilEquals(t, ok, tt.Want)
			testutilEquals(t, len(diag.Errors), tt.DiagErr)
		})
	}
}

func TestEntities(t *testing.T) {
	t.Parallel()
	t.Run("ToSlice", func(t *testing.T) {
		t.Parallel()
		s := []Entity{
			{
				UID: EntityUID{Type: "A", ID: "A"},
			},
			{
				UID: EntityUID{Type: "A", ID: "B"},
			},
			{
				UID: EntityUID{Type: "B", ID: "A"},
			},
			{
				UID: EntityUID{Type: "B", ID: "B"},
			},
		}
		entities := entitiesFromSlice(s)
		s2 := entities.toSlice()
		testutilEquals(t, s2, s)
	})
	t.Run("Clone", func(t *testing.T) {
		t.Parallel()
		s := []Entity{
			{
				UID: EntityUID{Type: "A", ID: "A"},
			},
			{
				UID: EntityUID{Type: "A", ID: "B"},
			},
			{
				UID: EntityUID{Type: "B", ID: "A"},
			},
			{
				UID: EntityUID{Type: "B", ID: "B"},
			},
		}
		entities := entitiesFromSlice(s)
		clone := entities.Clone()
		testutilEquals(t, clone, entities)
	})

}

func TestValueFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      Value
		outJSON string
	}{
		{
			name:    "string",
			in:      String("hello"),
			outJSON: `"hello"`,
		},
		{
			name:    "bool",
			in:      Boolean(true),
			outJSON: `true`,
		},
		{
			name:    "int64",
			in:      Long(42),
			outJSON: `42`,
		},
		{
			name:    "int64",
			in:      EntityUID{Type: "T", ID: "0"},
			outJSON: `{"__entity":{"type":"T","id":"0"}}`,
		},
		{
			name:    "record",
			in:      Record{"K": Boolean(true)},
			outJSON: `{"K":true}`,
		},
		{
			name:    "netipPrefix",
			in:      IPAddr(netip.MustParsePrefix("192.168.0.42/32")),
			outJSON: `{"__extn":{"fn":"ip","arg":"192.168.0.42"}}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := tt.in.ExplicitMarshalJSON()
			testutilOK(t, err)
			testutilEquals(t, string(out), tt.outJSON)
		})
	}
}

func TestError(t *testing.T) {
	t.Parallel()
	e := Error{Policy: 42, Message: "bad error"}
	testutilEquals(t, e.String(), "while evaluating policy `policy42`: bad error")
}

func TestInvalidPolicy(t *testing.T) {
	t.Parallel()
	// This case is very fabricated, it can't really happen
	ps := PolicySet{
		{
			Effect: Forbid,
			eval:   newLiteralEval(Long(42)),
		},
	}
	ok, diag := ps.IsAuthorized(Entities{}, Request{})
	testutilEquals(t, ok, Deny)
	testutilEquals(t, diag, Diagnostic{
		Errors: []Error{
			{
				Policy:  0,
				Message: "type error: expected bool, got long",
			},
		},
	})
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
			Request{Principal: NewEntityUID("a", "\u0000\u0000"), Action: NewEntityUID("Action", "action"), Resource: NewEntityUID("a", "\u0000\u0000")},
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
			Request{Principal: NewEntityUID("a", "\u0000\u0000"), Action: NewEntityUID("Action", "action"), Resource: NewEntityUID("a", "\u0000\u0000")},
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
			Request{Principal: NewEntityUID("a", "\u0000\u0000"), Action: NewEntityUID("Action", "action"), Resource: NewEntityUID("a", "\u0000\u0000")},
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
			Request{Principal: NewEntityUID("a", "\u0000\u0000"), Action: NewEntityUID("Action", "action"), Resource: NewEntityUID("a", "\u0000\u0000")},
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
			Request{Principal: NewEntityUID("a", "\u0000\b\u0011\u0000R"), Action: NewEntityUID("Action", "action"), Resource: NewEntityUID("a", "\u0000\b\u0011\u0000R")},
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
			request:  Request{Principal: NewEntityUID("a", "\u0000\u0000(W\u0000\u0000\u0000"), Action: NewEntityUID("Action", "action"), Resource: NewEntityUID("a", "")},
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
			testutilOK(t, err)
			ok, diag := policy.IsAuthorized(Entities{}, tt.request)
			testutilEquals(t, ok, tt.decision)
			var reasons []int
			for _, n := range diag.Reasons {
				reasons = append(reasons, n.Policy)
			}
			testutilEquals(t, reasons, tt.reasons)
			var errors []int
			for _, n := range diag.Errors {
				errors = append(errors, n.Policy)
			}
			testutilEquals(t, errors, tt.errors)
		})
	}
}
