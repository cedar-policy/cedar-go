package cedar_test

import (
	"strconv"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

var injectionRequest = cedar.Request{
	Principal: types.NewEntityUID("User", "alice"),
	Action:    types.NewEntityUID("Action", "delete"),
	Resource:  types.NewEntityUID("Document", "secret"),
	Context:   types.Record{},
}

func authorizeInjectionRequest(ps *cedar.PolicySet) types.Decision {
	d, _ := cedar.Authorize(ps, types.EntityMap{}, injectionRequest)
	return d
}

var nameInjectionCases = []struct {
	name   string
	policy string
}{
	{
		"principalIs",
		`{
			"effect": "permit",
			"principal": {"op": "is", "entity_type": "Nobody, action, resource ); permit ( principal"},
			"action": {"op": "All"},
			"resource": {"op": "All"},
			"conditions": []
		}`,
	},
	{
		"resourceIs",
		`{
			"effect": "permit",
			"principal": {"op": "All"},
			"action": {"op": "All"},
			"resource": {"op": "is", "entity_type": "Nobody ); permit ( principal, action, resource"},
			"conditions": []
		}`,
	},
	{
		"principalIsIn",
		`{
			"effect": "permit",
			"principal": {
				"op": "is",
				"entity_type": "Nobody in Nobody::\"x\", action, resource ); permit ( principal, action, resource ); permit ( principal is Nobody",
				"in": {"entity": {"type": "Nobody", "id": "y"}}
			},
			"action": {"op": "All"},
			"resource": {"op": "All"},
			"conditions": []
		}`,
	},
	{
		"principalEqEntityType",
		`{
			"effect": "permit",
			"principal": {"op": "==", "entity": {"type": "Nobody::\"x\", action, resource ); permit ( principal, action, resource ); permit ( principal == Nobody", "id": "y"}},
			"action": {"op": "All"},
			"resource": {"op": "All"},
			"conditions": []
		}`,
	},
	{
		"resourceInEntityType",
		`{
			"effect": "permit",
			"principal": {"op": "All"},
			"action": {"op": "All"},
			"resource": {"op": "in", "entity": {"type": "Nobody::\"x\" ); permit ( principal, action, resource ); permit ( principal, action, resource in Nobody", "id": "y"}}
		}`,
	},
	{
		"actionEqEntityType",
		`{
			"effect": "permit",
			"principal": {"op": "All"},
			"action": {"op": "==", "entity": {"type": "Action::\"none\", resource ); permit ( principal, action, resource ); permit ( principal, action == Action", "id": "none"}},
			"resource": {"op": "All"},
			"conditions": []
		}`,
	},
	{
		"actionInSetEntityType",
		`{
			"effect": "permit",
			"principal": {"op": "All"},
			"action": {"op": "in", "entities": [{"type": "Action::\"none\"], resource ); permit ( principal, action, resource ); permit ( principal, action in [Action", "id": "none"}]},
			"resource": {"op": "All"},
			"conditions": []
		}`,
	},
	{
		"conditionIs",
		`{
			"effect": "permit",
			"principal": {"op": "All"},
			"action": {"op": "All"},
			"resource": {"op": "All"},
			"conditions": [{
				"kind": "when",
				"body": {"is": {"left": {"Var": "principal"}, "entity_type": "Nobody }; permit ( principal, action, resource ) when { true"}}
			}]
		}`,
	},
	{
		"conditionIsIn",
		`{
			"effect": "permit",
			"principal": {"op": "All"},
			"action": {"op": "All"},
			"resource": {"op": "All"},
			"conditions": [{
				"kind": "when",
				"body": {"is": {
					"left": {"Var": "principal"},
					"entity_type": "Nobody in Nobody::\"x\" }; permit ( principal, action, resource ) when { principal is Nobody",
					"in": {"Value": {"__entity": {"type": "Nobody", "id": "y"}}}
				}}
			}]
		}`,
	},
	{
		"conditionEntityLiteral",
		`{
			"effect": "permit",
			"principal": {"op": "All"},
			"action": {"op": "All"},
			"resource": {"op": "All"},
			"conditions": [{
				"kind": "when",
				"body": {"==": {
					"left": {"Var": "principal"},
					"right": {"Value": {"__entity": {"type": "Nobody::\"x\" }; permit ( principal, action, resource ) when { true || Nobody", "id": "y"}}}
				}}
			}]
		}`,
	},
	{
		"annotationKey",
		`{
			"annotations": {"x(\"y\") permit ( principal, action, resource ); @z": "v"},
			"effect": "forbid",
			"principal": {"op": "All"},
			"action": {"op": "All"},
			"resource": {"op": "All"},
			"conditions": []
		}`,
	},
}

func TestJSONPolicyNameInjectionRejected(t *testing.T) {
	t.Parallel()
	for _, tt := range nameInjectionCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var p cedar.Policy
			err := p.UnmarshalJSON([]byte(tt.policy))
			if err == nil {
				t.Errorf("UnmarshalJSON accepted invalid name; Cedar text:\n%s", p.MarshalCedar())
			}
		})
	}
}

func TestJSONPolicyNameInjectionCedarRoundTrip(t *testing.T) {
	t.Parallel()
	for _, tt := range nameInjectionCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var p cedar.Policy
			if err := p.UnmarshalJSON([]byte(tt.policy)); err != nil {
				return
			}
			before := cedar.NewPolicySet()
			before.Add("policy0", &p)
			beforeDecision := authorizeInjectionRequest(before)

			text := before.MarshalCedar()
			after, err := cedar.NewPolicySetFromBytes("policy.cedar", text)
			if err != nil {
				t.Fatalf("MarshalCedar output does not parse: %v\n%s", err, text)
			}
			afterDecision := authorizeInjectionRequest(after)

			n := len(after.Map())
			if n != 1 {
				t.Errorf("one JSON policy became %d Cedar policies:\n%s", n, text)
			}
			if afterDecision != beforeDecision {
				t.Errorf("decision changed across Cedar round trip: before=%v after=%v\n%s", beforeDecision, afterDecision, text)
			}
		})
	}
}

func TestJSONPolicyNameInjectionEndToEnd(t *testing.T) {
	t.Parallel()
	var p cedar.Policy
	err := p.UnmarshalJSON([]byte(`{
		"effect": "permit",
		"principal": {"op": "is", "entity_type": "Nobody, action, resource ); permit ( principal"},
		"action": {"op": "All"},
		"resource": {"op": "All"},
		"conditions": []
	}`))
	if err != nil {
		return
	}

	ps := cedar.NewPolicySet()
	ps.Add("policy0", &p)
	testutil.Equals(t, len(ps.Map()), 1)
	testutil.Equals(t, authorizeInjectionRequest(ps), cedar.Deny)

	text := ps.MarshalCedar()
	reloaded, err := cedar.NewPolicySetFromBytes("policy.cedar", text)
	testutil.OK(t, err)

	if len(reloaded.Map()) != 1 || authorizeInjectionRequest(reloaded) != cedar.Deny {
		t.Errorf("VULNERABLE: Cedar round trip produced %d policies and decision %v:\n%s",
			len(reloaded.Map()), authorizeInjectionRequest(reloaded), text)
	}
}

func TestJSONEntityUIDTypeInjectionRejected(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
	}{
		{"explicit", `{"__entity": {"type": "A::\"x\" ); permit ( principal, action, resource ); permit ( principal == A", "id": "y"}}`},
		{"implicit", `{"type": "A::\"x\" ); permit ( principal, action, resource ); permit ( principal == A", "id": "y"}`},
		{"space", `{"type": "A B", "id": "y"}`},
		{"emptySegment", `{"type": "A::::B", "id": "y"}`},
		{"leadingDigit", `{"type": "1A", "id": "y"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var uid types.EntityUID
			if err := uid.UnmarshalJSON([]byte(tt.in)); err == nil {
				t.Errorf("EntityUID.UnmarshalJSON accepted invalid type; String() = %s", uid.String())
			}
		})
	}
}

func TestJSONValueEntityTypeInjectionRejected(t *testing.T) {
	t.Parallel()
	var v types.Value
	err := types.UnmarshalJSON([]byte(`{"__entity": {"type": "A::\"x\" ); permit ( principal, action, resource ); permit ( principal == A", "id": "y"}}`), &v)
	if err == nil {
		t.Errorf("types.UnmarshalJSON accepted invalid entity type; value = %v", v)
	}
}

func TestJSONPolicyInvalidNamesRejected(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		entityType string
	}{
		{"space", "A B"},
		{"emptySegment", "A::::B"},
		{"trailingSeparator", "A::"},
		{"leadingSeparator", "::A"},
		{"leadingDigit", "1A"},
		{"quote", `A"`},
		{"reservedKeyword", "if"},
		{"reservedKeywordSegment", "A::true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var p cedar.Policy
			err := p.UnmarshalJSON([]byte(`{
				"effect": "permit",
				"principal": {"op": "is", "entity_type": ` + strconv.Quote(tt.entityType) + `},
				"action": {"op": "All"},
				"resource": {"op": "All"},
				"conditions": []
			}`))
			if err == nil {
				t.Errorf("UnmarshalJSON accepted invalid entity type %q", tt.entityType)
			}
		})
	}
}

func TestJSONPolicyValidNamesAccepted(t *testing.T) {
	t.Parallel()
	var p cedar.Policy
	err := p.UnmarshalJSON([]byte(`{
		"annotations": {"id": "a", "_x1": "b"},
		"effect": "permit",
		"principal": {"op": "is", "entity_type": "Ns::Sub_1::User", "in": {"entity": {"type": "Ns::Group", "id": "g"}}},
		"action": {"op": "in", "entities": [{"type": "Ns::Action", "id": "read"}]},
		"resource": {"op": "==", "entity": {"type": "_Doc", "id": "d"}},
		"conditions": [{
			"kind": "when",
			"body": {"is": {"left": {"Var": "context"}, "entity_type": "A::B"}}
		}]
	}`))
	testutil.OK(t, err)

	text := p.MarshalCedar()
	ps, err := cedar.NewPolicySetFromBytes("policy.cedar", text)
	testutil.OK(t, err)
	testutil.Equals(t, len(ps.Map()), 1)
}
