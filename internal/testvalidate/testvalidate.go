// Package testvalidate provides shared validation test helpers used by both
// the top-level corpus tests and the x/exp/schema/validate corpus tests.
package testvalidate

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
	"github.com/cedar-policy/cedar-go/x/exp/schema/validate"
)

// PerPolicyResult holds expected per-policy validation results from Rust Cedar.
type PerPolicyResult struct {
	Strict           bool     `json:"strict"`
	Permissive       bool     `json:"permissive"`
	StrictErrors     []string `json:"strictErrors"`
	PermissiveErrors []string `json:"permissiveErrors"`
}

// PerEntityResult holds expected per-entity validation results from Rust Cedar.
type PerEntityResult struct {
	Errors []string `json:"errors"`
}

// RequestValidationResult holds expected per-request validation results.
type RequestValidationResult struct {
	Description string   `json:"description"`
	Strict      *bool    `json:"strict"`
	Permissive  *bool    `json:"permissive"`
	Errors      []string `json:"errors"`
}

// Validation holds the full expected validation results from Rust Cedar.
type Validation struct {
	PolicyValidation struct {
		Strict           bool                       `json:"strict"`
		Permissive       bool                       `json:"permissive"`
		StrictErrors     []string                   `json:"strictErrors"`
		PermissiveErrors []string                   `json:"permissiveErrors"`
		PerPolicy        map[string]PerPolicyResult `json:"perPolicy"`
	} `json:"policyValidation"`
	EntityValidation struct {
		PerEntity map[string]PerEntityResult `json:"perEntity"`
	} `json:"entityValidation"`
	RequestValidation []RequestValidationResult `json:"requestValidation"`
}

// ParseValidation unmarshals validation JSON data into a Validation struct.
func ParseValidation(t testing.TB, data []byte) Validation {
	t.Helper()
	var cv Validation
	testutil.OK(t, json.Unmarshal(data, &cv))
	return cv
}

// RunPolicyChecks runs aggregate and per-policy validation checks.
func RunPolicyChecks(t *testing.T, rs *resolved.Schema, policySet *cedar.PolicySet, cv Validation) {
	t.Helper()

	t.Run("validate-policy-strict", func(t *testing.T) {
		t.Parallel()
		v := validate.New(rs, validate.WithStrict())
		var allErrs []string
		for id, p := range policySet.All() {
			if err := v.Policy(string(id), (*ast.Policy)(p.AST())); err != nil {
				allErrs = append(allErrs, testutil.CollectErrors(err)...)
			}
		}
		testutil.Equals(t, len(allErrs) == 0, cv.PolicyValidation.Strict)
		testutil.CheckErrorStrings(t, allErrs, cv.PolicyValidation.StrictErrors)
	})

	t.Run("validate-policy-permissive", func(t *testing.T) {
		t.Parallel()
		v := validate.New(rs, validate.WithPermissive())
		var allErrs []string
		for id, p := range policySet.All() {
			if err := v.Policy(string(id), (*ast.Policy)(p.AST())); err != nil {
				allErrs = append(allErrs, testutil.CollectErrors(err)...)
			}
		}
		testutil.Equals(t, len(allErrs) == 0, cv.PolicyValidation.Permissive)
		testutil.CheckErrorStrings(t, allErrs, cv.PolicyValidation.PermissiveErrors)
	})

	for policyID, expected := range cv.PolicyValidation.PerPolicy {
		p := policySet.Get(cedar.PolicyID(policyID))
		testutil.FatalIf(t, p == nil, "policy %s not found in policy set", policyID)

		t.Run(fmt.Sprintf("validate-per-policy-strict/%s", policyID), func(t *testing.T) {
			t.Parallel()
			v := validate.New(rs, validate.WithStrict())
			err := v.Policy(policyID, (*ast.Policy)(p.AST()))
			errs := testutil.CollectErrors(err)
			testutil.Equals(t, len(errs) == 0, expected.Strict)
			testutil.CheckErrorStrings(t, errs, expected.StrictErrors)
		})

		t.Run(fmt.Sprintf("validate-per-policy-permissive/%s", policyID), func(t *testing.T) {
			t.Parallel()
			v := validate.New(rs, validate.WithPermissive())
			err := v.Policy(policyID, (*ast.Policy)(p.AST()))
			errs := testutil.CollectErrors(err)
			testutil.Equals(t, len(errs) == 0, expected.Permissive)
			testutil.CheckErrorStrings(t, errs, expected.PermissiveErrors)
		})
	}
}

// RunEntityChecks runs aggregate and per-entity validation checks.
func RunEntityChecks(t *testing.T, rs *resolved.Schema, entities cedar.EntityMap, cv Validation) {
	t.Helper()

	anyEntityErrors := false
	for _, pe := range cv.EntityValidation.PerEntity {
		if len(pe.Errors) > 0 {
			anyEntityErrors = true
			break
		}
	}

	t.Run("validate-entities", func(t *testing.T) {
		t.Parallel()
		v := validate.New(rs)
		err := v.Entities(entities)
		testutil.Equals(t, err == nil, !anyEntityErrors)
	})

	// Per-entity validation (pass/fail checked against Rust; error message
	// formats differ between Go and Rust so we only assert pass/fail here)
	for _, entity := range entities {
		uidStr := string(entity.UID.Type) + "::" + string(entity.UID.ID)
		expected, ok := cv.EntityValidation.PerEntity[uidStr]
		testutil.FatalIf(t, !ok, "entity %s not found in perEntity validation data", uidStr)
		t.Run(fmt.Sprintf("validate-per-entity/%s", uidStr), func(t *testing.T) {
			t.Parallel()
			v := validate.New(rs)
			err := v.Entity(entity)
			testutil.Equals(t, err == nil, len(expected.Errors) == 0)
		})
	}
}

// RunRequestChecks runs per-request validation checks.
func RunRequestChecks(t *testing.T, rs *resolved.Schema, cv Validation, requests []cedar.Request) {
	t.Helper()

	for i, reqVal := range cv.RequestValidation {
		if i >= len(requests) {
			break
		}
		req := requests[i]
		if reqVal.Strict != nil {
			t.Run(fmt.Sprintf("validate-request-strict/%s", reqVal.Description), func(t *testing.T) {
				t.Parallel()
				v := validate.New(rs, validate.WithStrict())
				err := v.Request(req)
				testutil.Equals(t, err == nil, *reqVal.Strict)
				testutil.CheckErrorStrings(t, testutil.CollectErrors(err), reqVal.Errors)
			})
		}
		if reqVal.Permissive != nil {
			t.Run(fmt.Sprintf("validate-request-permissive/%s", reqVal.Description), func(t *testing.T) {
				t.Parallel()
				v := validate.New(rs, validate.WithPermissive())
				err := v.Request(req)
				testutil.Equals(t, err == nil, *reqVal.Permissive)
				testutil.CheckErrorStrings(t, testutil.CollectErrors(err), reqVal.Errors)
			})
		}
	}
}
