//go:build corpus

package integration_tests

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"golang.org/x/exp/slices"
)

//go:embed tmp/**/*.json tmp/**/*.cedar
var integrationFS embed.FS

type corpusTest struct {
	Schema         string `json:"schema"`
	Policies       string `json:"policies"`
	ShouldValidate bool   `json:"should_validate"`
	Entities       string `json:"entities"`
	Requests       []struct {
		Desc      string       `json:"desc"`
		Principal jsonEntity   `json:"principal"`
		Action    jsonEntity   `json:"action"`
		Resource  jsonEntity   `json:"resource"`
		Context   cedar.Record `json:"context"`
		Decision  string       `json:"decision"`
		Reasons   []string     `json:"reason"`
		Errors    []string     `json:"errors"`
	} `json:"requests"`
}

func TestCorpus(t *testing.T) {
	t.Parallel()
	var tests []string
	for _, p := range []string{
		"tmp/corpus-tests/*.json",
	} {
		more, err := fs.Glob(integrationFS, p)
		if err != nil || len(more) == 0 {
			t.Fatal("err loading tests", p, err)
		}
		for _, tn := range more {
			if strings.Contains(tn, "schema_") {
				continue
			}
			if strings.Contains(tn, ".cedarschema.json") {
				continue
			}
			if strings.Contains(tn, ".entities.json") {
				continue
			}
			tests = append(tests, tn)
		}
	}

	total := 0
	problems := 0

	exclude := []string{}
	for _, x := range exclude {
		i := slices.Index(tests, x)
		if i == -1 {
			t.Fatal("exclusion not found", x)
		}
		tests = append(tests[:i], tests[i+1:]...)
	}
	slices.Sort(tests)
	prefix := func(v string) string {
		return "tmp/" + v
	}

	// detect possible corpus data pipeline failure. As of 2024/07/02, there were 1982 tests.
	if len(tests) < 100 {
		t.Fatalf("corpus test count too low: %v", len(tests))
	}

	for _, tn := range tests {
		tn := tn
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			b := fsLoad(t, integrationFS, tn)
			var tt corpusTest
			if err := json.Unmarshal(b, &tt); err != nil {
				t.Fatal("error unmarshalling test", err)
			}

			be := fsLoad(t, integrationFS, prefix(tt.Entities))
			ent, err := newEntitiesFromJSON(be)
			if err != nil {
				t.Fatal("error unmarshalling entities", err)
			}

			bp := fsLoad(t, integrationFS, prefix(tt.Policies))
			ps, err := cedar.NewPolicySet("policy.cedar", bp)
			if err != nil {
				t.Fatal("error parsing policy set", err)
			}

			todo := tt.Requests

			for _, q := range todo {
				q := q
				t.Run(q.Desc, func(t *testing.T) {
					t.Parallel()
					total++
					ok, diag := ps.IsAuthorized(ent, cedar.Request{
						Principal: cedar.EntityUID(q.Principal),
						Action:    cedar.EntityUID(q.Action),
						Resource:  cedar.EntityUID(q.Resource),
						Context:   q.Context,
					})
					if ok != (q.Decision == "Allow") {
						t.Fatalf("got %v want %v", ok, q.Decision)
					}
					var errors []string
					for _, n := range diag.Errors {
						errors = append(errors, fmt.Sprintf("policy%d", n.Policy))
					}
					isOkay := true
					if !slices.Equal(errors, q.Errors) {
						t.Errorf("errors got %v want %v", errors, q.Errors)
						isOkay = false
					}
					var reasons []string
					for _, n := range diag.Reasons {
						reasons = append(reasons, fmt.Sprintf("policy%d", n.Policy))
					}
					if !slices.Equal(reasons, q.Reasons) {
						t.Errorf("reasons got %v want %v", reasons, q.Reasons)
						isOkay = false
					}
					if !isOkay {
						problems++
					}
				})
			}
		})
	}
	_, _ = total, problems
	// t.Errorf("total: %v problems %v", total, problems)
	// fmt.Println("total", total)
	// panic(total)
}
