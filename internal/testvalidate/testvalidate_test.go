package testvalidate

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
)

func boolPtr(b bool) *bool { return &b }

// buildSchema parses a cedar schema string and returns the resolved schema.
func buildSchema(t *testing.T, cedarSchema string) *schema.Schema {
	t.Helper()
	var s schema.Schema
	s.SetFilename("test.cedarschema")
	testutil.OK(t, s.UnmarshalCedar([]byte(cedarSchema)))
	return &s
}

const testSchema = `
entity User {
	name: String,
};
entity Document {
	title: String,
};
action "view" appliesTo {
	principal: User,
	resource: Document,
	context: {},
};
`

func TestParseValidation(t *testing.T) {
	t.Parallel()
	data := []byte(`{
		"policyValidation": {
			"strict": true,
			"permissive": true,
			"perPolicy": {
				"policy0": {"strict": true, "permissive": true}
			}
		},
		"entityValidation": {
			"perEntity": {
				"User::alice": {}
			}
		},
		"requestValidation": [
			{"description": "req1", "strict": true, "permissive": true}
		]
	}`)
	cv := ParseValidation(t, data)
	testutil.Equals(t, cv.PolicyValidation.Strict, true)
	testutil.Equals(t, cv.PolicyValidation.Permissive, true)
	testutil.Equals(t, len(cv.PolicyValidation.PerPolicy), 1)
	testutil.Equals(t, cv.PolicyValidation.PerPolicy["policy0"].Strict, true)
	testutil.Equals(t, len(cv.EntityValidation.PerEntity), 1)
	testutil.Equals(t, len(cv.EntityValidation.PerEntity["User::alice"].Errors), 0)
	testutil.Equals(t, len(cv.RequestValidation), 1)
	testutil.Equals(t, cv.RequestValidation[0].Description, "req1")
}

func TestRunPolicyChecks(t *testing.T) {
	t.Parallel()
	s := buildSchema(t, testSchema)
	rs, err := s.Resolve()
	testutil.OK(t, err)

	policySet, err := cedar.NewPolicySetFromBytes("test.cedar", []byte(`
		permit(principal is User, action == Action::"view", resource is Document);
	`))
	testutil.OK(t, err)

	cv := Validation{}
	cv.PolicyValidation.Strict = true
	cv.PolicyValidation.Permissive = true
	cv.PolicyValidation.PerPolicy = map[string]PerPolicyResult{
		"policy0": {Strict: true, Permissive: true},
	}

	RunPolicyChecks(t, rs, policySet, cv)
}

func TestRunPolicyChecksWithErrors(t *testing.T) {
	t.Parallel()
	s := buildSchema(t, testSchema)
	rs, err := s.Resolve()
	testutil.OK(t, err)

	policySet, err := cedar.NewPolicySetFromBytes("test.cedar", []byte(`
		permit(principal is User, action == Action::"nonexistent", resource is Document);
	`))
	testutil.OK(t, err)

	cv := Validation{}
	cv.PolicyValidation.Strict = false
	cv.PolicyValidation.Permissive = false
	cv.PolicyValidation.StrictErrors = []string{
		"for policy `policy0`, unrecognized action `Action::\"nonexistent\"`",
		"for policy `policy0`, unable to find an applicable action given the policy scope constraints",
	}
	cv.PolicyValidation.PermissiveErrors = []string{
		"for policy `policy0`, unrecognized action `Action::\"nonexistent\"`",
		"for policy `policy0`, unable to find an applicable action given the policy scope constraints",
	}
	cv.PolicyValidation.PerPolicy = map[string]PerPolicyResult{
		"policy0": {
			Strict: false, Permissive: false,
			StrictErrors: []string{
				"for policy `policy0`, unrecognized action `Action::\"nonexistent\"`",
				"for policy `policy0`, unable to find an applicable action given the policy scope constraints",
			},
			PermissiveErrors: []string{
				"for policy `policy0`, unrecognized action `Action::\"nonexistent\"`",
				"for policy `policy0`, unable to find an applicable action given the policy scope constraints",
			},
		},
	}

	RunPolicyChecks(t, rs, policySet, cv)
}

func TestRunEntityChecks(t *testing.T) {
	t.Parallel()
	s := buildSchema(t, testSchema)
	rs, err := s.Resolve()
	testutil.OK(t, err)

	var entities cedar.EntityMap
	testutil.OK(t, json.Unmarshal([]byte(`[
		{"uid": {"type": "User", "id": "alice"}, "attrs": {"name": "Alice"}, "parents": []},
		{"uid": {"type": "Document", "id": "doc1"}, "attrs": {"title": "Test"}, "parents": []}
	]`), &entities))

	cv := Validation{}
	cv.EntityValidation.PerEntity = map[string]PerEntityResult{
		"User::alice":    {},
		"Document::doc1": {},
	}

	RunEntityChecks(t, rs, entities, cv)
}

func TestRunEntityChecksWithErrors(t *testing.T) {
	t.Parallel()
	s := buildSchema(t, testSchema)
	rs, err := s.Resolve()
	testutil.OK(t, err)

	var entities cedar.EntityMap
	testutil.OK(t, json.Unmarshal([]byte(`[
		{"uid": {"type": "User", "id": "alice"}, "attrs": {"name": "Alice"}, "parents": [
			{"type": "Document", "id": "doc1"}
		]}
	]`), &entities))

	cv := Validation{}
	cv.EntityValidation.PerEntity = map[string]PerEntityResult{
		"User::alice": {Errors: []string{"some error"}},
	}

	RunEntityChecks(t, rs, entities, cv)
}

func TestRunRequestChecks(t *testing.T) {
	t.Parallel()
	s := buildSchema(t, testSchema)
	rs, err := s.Resolve()
	testutil.OK(t, err)

	requests := []cedar.Request{
		{
			Principal: cedar.NewEntityUID("User", "alice"),
			Action:    cedar.NewEntityUID("Action", "view"),
			Resource:  cedar.NewEntityUID("Document", "doc1"),
		},
	}

	cv := Validation{}
	cv.RequestValidation = []RequestValidationResult{
		{Description: "valid request", Strict: boolPtr(true), Permissive: boolPtr(true)},
	}

	RunRequestChecks(t, rs, cv, requests)
}

func TestRunRequestChecksWithErrors(t *testing.T) {
	t.Parallel()
	s := buildSchema(t, testSchema)
	rs, err := s.Resolve()
	testutil.OK(t, err)

	requests := []cedar.Request{
		{
			Principal: cedar.NewEntityUID("Document", "doc1"),
			Action:    cedar.NewEntityUID("Action", "view"),
			Resource:  cedar.NewEntityUID("Document", "doc1"),
		},
	}

	cv := Validation{}
	cv.RequestValidation = []RequestValidationResult{
		{
			Description: "wrong principal",
			Strict:      boolPtr(false),
			Permissive:  boolPtr(false),
			Errors:      []string{"principal type `Document` is not valid for `Action::\"view\"`"},
		},
	}

	RunRequestChecks(t, rs, cv, requests)
}

func TestRunRequestChecksSkipsNilModes(t *testing.T) {
	t.Parallel()
	s := buildSchema(t, testSchema)
	rs, err := s.Resolve()
	testutil.OK(t, err)

	requests := []cedar.Request{
		{
			Principal: cedar.NewEntityUID("User", "alice"),
			Action:    cedar.NewEntityUID("Action", "view"),
			Resource:  cedar.NewEntityUID("Document", "doc1"),
		},
	}

	cv := Validation{}
	cv.RequestValidation = []RequestValidationResult{
		{Description: "no validate", Strict: nil, Permissive: nil},
	}

	RunRequestChecks(t, rs, cv, requests)
}

func TestRunRequestChecksExtraValidation(t *testing.T) {
	t.Parallel()
	s := buildSchema(t, testSchema)
	rs, err := s.Resolve()
	testutil.OK(t, err)

	// More validation entries than requests - should not panic
	cv := Validation{}
	cv.RequestValidation = []RequestValidationResult{
		{Description: "req1", Strict: boolPtr(true), Permissive: boolPtr(true)},
		{Description: "req2", Strict: boolPtr(true), Permissive: boolPtr(true)},
	}

	RunRequestChecks(t, rs, cv, nil)
}
