package validate_test

import (
	"embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/internal/testvalidate"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
)

//go:embed testdata
var testdataFS embed.FS

type corpusTest struct {
	Schema         string `json:"schema"`
	Policies       string `json:"policies"`
	ShouldValidate bool   `json:"shouldValidate"`
	Entities       string `json:"entities"`
	Requests       []struct {
		Desc      string          `json:"description"`
		Principal cedar.EntityUID `json:"principal"`
		Action    cedar.EntityUID `json:"action"`
		Resource  cedar.EntityUID `json:"resource"`
		Context   cedar.Record    `json:"context"`
	} `json:"requests"`
}

func TestCorpus(t *testing.T) {
	t.Parallel()

	entries, err := testdataFS.ReadDir("testdata")
	testutil.OK(t, err)

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		if strings.HasSuffix(name, ".entities.json") || strings.HasSuffix(name, ".validation.json") {
			continue
		}

		testName := strings.TrimSuffix(name, ".json")
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			// Load test manifest
			manifestData, err := testdataFS.ReadFile("testdata/" + name)
			testutil.OK(t, err)
			var tt corpusTest
			testutil.OK(t, json.Unmarshal(manifestData, &tt))

			// Load validation expectations
			validationData, err := testdataFS.ReadFile("testdata/" + testName + ".validation.json")
			testutil.OK(t, err)
			cv := testvalidate.ParseValidation(t, validationData)

			// Load and parse schema
			schemaContent, err := testdataFS.ReadFile(tt.Schema)
			testutil.OK(t, err)
			var s schema.Schema
			s.SetFilename(testName + ".cedarschema")
			testutil.OK(t, s.UnmarshalCedar(schemaContent))
			rs, err := s.Resolve()
			testutil.OK(t, err)

			// Load and parse policies
			policyContent, err := testdataFS.ReadFile(tt.Policies)
			testutil.OK(t, err)
			policySet, err := cedar.NewPolicySetFromBytes(testName+".cedar", policyContent)
			testutil.OK(t, err)

			// Load entities
			entitiesContent, err := testdataFS.ReadFile(tt.Entities)
			testutil.OK(t, err)
			var entities cedar.EntityMap
			testutil.OK(t, json.Unmarshal(entitiesContent, &entities))

			// Build requests
			var requests []cedar.Request
			for _, r := range tt.Requests {
				requests = append(requests, cedar.Request{
					Principal: r.Principal,
					Action:    r.Action,
					Resource:  r.Resource,
					Context:   r.Context,
				})
			}

			testvalidate.RunPolicyChecks(t, rs, policySet, cv)
			testvalidate.RunEntityChecks(t, rs, entities, cv)
			testvalidate.RunRequestChecks(t, rs, cv, requests)
		})
	}
}
