package cedar

import (
	"encoding/json"
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
