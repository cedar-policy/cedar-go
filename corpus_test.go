package cedar_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/internal/testvalidate"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/batch"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
	schemaast "github.com/cedar-policy/cedar-go/x/exp/schema/ast"
	exptypes "github.com/cedar-policy/cedar-go/x/exp/types"
)

// jsonEntity is not part of entityValue as I can find
// no evidence this is part of the JSON spec.  It also
// requires creating a parser, so it's quite expensive.
type jsonEntity cedar.EntityUID

func (e *jsonEntity) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	var input cedar.EntityUID
	if err := json.Unmarshal(b, &input); err != nil {
		return err
	}
	*e = jsonEntity(input)
	return nil
}

type corpusTest struct {
	Schema         string `json:"schema"`
	Policies       string `json:"policies"`
	ShouldValidate bool   `json:"shouldValidate"`
	Entities       string `json:"entities"`
	Requests       []struct {
		Desc      string           `json:"description"`
		Principal jsonEntity       `json:"principal"`
		Action    jsonEntity       `json:"action"`
		Resource  jsonEntity       `json:"resource"`
		Context   cedar.Record     `json:"context"`
		Decision  cedar.Decision   `json:"decision"`
		Reasons   []cedar.PolicyID `json:"reason"`
		Errors    []cedar.PolicyID `json:"errors"`
	} `json:"requests"`
}

//go:embed corpus-tests.tar.gz
var corpusArchive []byte

//go:embed corpus-tests-json-schemas.tar.gz
var corpusJSONSchemasArchive []byte

//go:embed corpus-tests-validation.tar.gz
var corpusValidationArchive []byte

type tarFileDataPointer struct {
	Position int64
	Size     int64
}

type TarFileMap struct {
	buf   io.ReaderAt
	files map[string]tarFileDataPointer
}

func NewTarFileMap(buf io.ReaderAt) TarFileMap {
	return TarFileMap{
		buf:   buf,
		files: make(map[string]tarFileDataPointer),
	}
}

func (fdm *TarFileMap) AddFileData(path string, position int64, size int64) {
	fdm.files[path] = tarFileDataPointer{position, size}
}

func (fdm TarFileMap) GetFileData(path string) ([]byte, error) {
	fdp, ok := fdm.files[path]
	if !ok {
		return nil, fmt.Errorf("file not found in archive: %v", path)
	}
	content := make([]byte, fdp.Size)
	_, err := fdm.buf.ReadAt(content, fdp.Position)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func loadTarGz(t testing.TB, archive []byte) TarFileMap {
	t.Helper()

	gzipReader, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		t.Fatal("error reading compressed archive header", err)
	}
	defer gzipReader.Close() //nolint:errcheck

	buf, err := io.ReadAll(gzipReader)
	if err != nil {
		t.Fatal("error reading compressed archive", err)
	}

	bufReader := bytes.NewReader(buf)
	archiveReader := tar.NewReader(bufReader)

	fdm := NewTarFileMap(bufReader)
	for file, err := archiveReader.Next(); err == nil; file, err = archiveReader.Next() {
		if file.Typeflag != tar.TypeReg {
			continue
		}

		cursor, _ := bufReader.Seek(0, io.SeekCurrent)
		fdm.AddFileData(file.Name, cursor, file.Size)
	}

	return fdm
}

//nolint:revive // due to test cognitive complexity
func TestCorpus(t *testing.T) {
	t.Parallel()

	// Load corpus test files
	fdm := loadTarGz(t, corpusArchive)
	var testFiles []string
	for fileName := range fdm.files {
		if strings.HasSuffix(fileName, ".json") && !strings.HasSuffix(fileName, ".entities.json") {
			testFiles = append(testFiles, fileName)
		}
	}

	// Load JSON schemas for validation
	jsonSchemasFdm := loadTarGz(t, corpusJSONSchemasArchive)

	// Load validation results from Rust Cedar
	validationFdm := loadTarGz(t, corpusValidationArchive)

	for _, testFile := range testFiles {
		testFile := testFile
		t.Run(testFile, func(t *testing.T) {
			t.Parallel()
			testContent, err := fdm.GetFileData(testFile)
			if err != nil {
				t.Fatal("error reading test content", err)
			}

			var tt corpusTest
			if err := json.Unmarshal(testContent, &tt); err != nil {
				t.Fatal("error unmarshalling test", err)
			}

			schemaContent, err := fdm.GetFileData(tt.Schema)
			if err != nil {
				t.Fatal("error reading schema content", err)
			}

			var s schema.Schema
			s.SetFilename("test.schema")
			if err := s.UnmarshalCedar(schemaContent); err != nil {
				t.Fatal("error parsing schema", err, "\n===\n", string(schemaContent))
			}

			// Validate schema round-trip
			t.Run("schema-round-trip", func(t *testing.T) {
				t.Parallel()

				js, err := s.MarshalJSON()
				testutil.OK(t, err)

				var s2 schema.Schema
				err = s2.UnmarshalJSON(js)
				testutil.OK(t, err)

				sb, err := s2.MarshalCedar()
				testutil.OK(t, err)

				var s3 schema.Schema
				err = s3.UnmarshalCedar(sb)
				testutil.OK(t, err)

				j2, err := s3.MarshalJSON()
				testutil.OK(t, err)

				testutil.Equals(t, string(j2), string(js))
			})

			// Compare cedar-go's JSON for the schema with Rust Cedar's JSON
			// for the same Cedar schema.
			t.Run("schema-vs-rust", func(t *testing.T) {
				t.Parallel()

				// Extract schema filename from path (e.g., "corpus-tests/abc123.cedarschema" -> "abc123")
				schemaFilename := strings.TrimSuffix(strings.TrimPrefix(tt.Schema, "corpus-tests/"), ".cedarschema")
				jsonSchemaPath := fmt.Sprintf("corpus-tests-json-schemas/%s.cedarschema.json", schemaFilename)

				rustJSON, err := jsonSchemasFdm.GetFileData(jsonSchemaPath)
				testutil.OK(t, err)

				goJSON, err := schema.NewSchemaFromAST(rustCLISchema(s.AST())).MarshalJSON()
				testutil.OK(t, err)

				stringEquals(t, string(normalizeJSON(t, goJSON)), string(normalizeJSON(t, rustJSON)))
			})

			policyContent, err := fdm.GetFileData(tt.Policies)
			if err != nil {
				t.Fatal("error reading policy content", err)
			}

			policySet, err := cedar.NewPolicySetFromBytes("policy.cedar", policyContent)
			if err != nil {
				t.Fatal("error parsing policy set", err)
			}

			// Load Rust validation results for this test
			testBasename := strings.TrimSuffix(strings.TrimPrefix(testFile, "corpus-tests/"), ".json")
			validationPath := fmt.Sprintf("corpus-tests-validation/%s.validation.json", testBasename)
			validationContent, err := validationFdm.GetFileData(validationPath)
			if err != nil {
				t.Fatal("error reading validation data", err)
			}
			cv := testvalidate.ParseValidation(t, validationContent)

			rs, err := s.Resolve()
			testutil.OK(t, err)

			// Load and parse entities using schema-guided parsing
			entitiesContent, err := fdm.GetFileData(tt.Entities)
			if err != nil {
				t.Fatal("error reading entities content", err)
			}

			var expEntities exptypes.EntityMap
			if err := expEntities.UnmarshalJSONWithSchema(entitiesContent, rs); err != nil {
				t.Fatal("error unmarshalling entities", err)
			}
			entities := types.EntityMap(expEntities)

			// Build requests for validation checks
			var requests []cedar.Request
			for _, r := range tt.Requests {
				requests = append(requests, cedar.Request{
					Principal: cedar.EntityUID(r.Principal),
					Action:    cedar.EntityUID(r.Action),
					Resource:  cedar.EntityUID(r.Resource),
					Context:   r.Context,
				})
			}

			testvalidate.RunPolicyChecks(t, rs, policySet, cv)
			testvalidate.RunEntityChecks(t, rs, entities, cv)
			testvalidate.RunRequestChecks(t, rs, cv, requests)

			// Round-trip the policy set through cedar marshal/unmarshal
			cedarRTBytes := policySet.MarshalCedar()
			cedarRTPolicySet, err := cedar.NewPolicySetFromBytes("policy.cedar", cedarRTBytes)
			testutil.OK(t, wrapErr("parsing cedar round-tripped policy set", err))

			// Round-trip the policy set through JSON marshal/unmarshal
			jsonRTBytes, err := policySet.MarshalJSON()
			testutil.OK(t, wrapErr("marshalling policy set to JSON", err))
			var jsonRTPolicySet cedar.PolicySet
			err = jsonRTPolicySet.UnmarshalJSON(jsonRTBytes)
			testutil.OK(t, wrapErr("unmarshalling JSON round-tripped policy set", err))

			// Round-trip the entities through JSON and assert equality
			entitiesJSONBytes, err := entities.MarshalJSON()
			testutil.OK(t, wrapErr("marshalling entities to JSON", err))
			var rtEntities cedar.EntityMap
			err = rtEntities.UnmarshalJSON(entitiesJSONBytes)
			testutil.OK(t, wrapErr("unmarshalling JSON round-tripped entities", err))
			testutil.Equals(t, rtEntities, entities)

			policySets := map[string]*cedar.PolicySet{
				"":                 policySet,
				"/policy-cedar-rt": cedarRTPolicySet,
				"/policy-json-rt":  &jsonRTPolicySet,
			}

			for suffix, ps := range policySets {
				ps := ps
				suffix := suffix
				for _, request := range tt.Requests {
					if len(request.Reasons) == 0 && request.Reasons != nil {
						request.Reasons = nil
					}
					if len(request.Errors) == 0 && request.Errors != nil {
						request.Errors = nil
					}

					t.Run(request.Desc+suffix, func(t *testing.T) {
						t.Parallel()
						ok, diag := ps.IsAuthorized(
							entities,
							cedar.Request{
								Principal: cedar.EntityUID(request.Principal),
								Action:    cedar.EntityUID(request.Action),
								Resource:  cedar.EntityUID(request.Resource),
								Context:   request.Context,
							})

						testutil.Equals(t, ok, request.Decision)
						var errors []cedar.PolicyID
						for _, n := range diag.Errors {
							errors = append(errors, n.PolicyID)
						}
						testutil.Equals(t, errors, request.Errors)
						var reasons []cedar.PolicyID
						for _, n := range diag.Reasons {
							reasons = append(reasons, n.PolicyID)
						}
						testutil.Equals(t, reasons, request.Reasons)
					})

					t.Run(request.Desc+"/batch"+suffix, func(t *testing.T) {
						t.Parallel()
						ctx := context.Background()
						var res batch.Result
						var total int
						principal := cedar.EntityUID(request.Principal)
						action := cedar.EntityUID(request.Action)
						resource := cedar.EntityUID(request.Resource)
						context := request.Context
						err := batch.Authorize(ctx, ps, entities, batch.Request{
							Principal: batch.Variable("principal"),
							Action:    batch.Variable("action"),
							Resource:  batch.Variable("resource"),
							Context:   batch.Variable("context"),
							Variables: batch.Variables{
								"principal": []cedar.Value{principal},
								"action":    []cedar.Value{action},
								"resource":  []cedar.Value{resource},
								"context":   []cedar.Value{context},
							},
						}, func(r batch.Result) error {
							res = r
							total++
							return nil
						})
						testutil.OK(t, err)
						testutil.Equals(t, total, 1)
						testutil.Equals(t, res.Request.Principal, principal)
						testutil.Equals(t, res.Request.Action, action)
						testutil.Equals(t, res.Request.Resource, resource)
						testutil.Equals(t, res.Request.Context, context)

						ok, diag := res.Decision, res.Diagnostic
						testutil.Equals(t, ok, request.Decision)
						var errors []cedar.PolicyID
						for _, n := range diag.Errors {
							errors = append(errors, n.PolicyID)
						}
						testutil.Equals(t, errors, request.Errors)
						var reasons []cedar.PolicyID
						for _, n := range diag.Reasons {
							reasons = append(reasons, n.PolicyID)
						}
						testutil.Equals(t, reasons, request.Reasons)
					})
				}
			}
		})
	}
}

func wrapErr(msg string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// rustCLISchema emulates Rust Cedar's somewhat quirky behavior when converting
// Cedar-encoded schema to JSON. We use this in a test which performs its own
// conversion from Cedar-encoded schema to JSON and compares it to the Rust
// CLI's output for the same operation.
func rustCLISchema(in *schemaast.Schema) *schemaast.Schema {
	convert := func(actions schemaast.Actions) schemaast.Actions {
		out := make(schemaast.Actions, len(actions))
		for name, action := range actions {
			// The Rust Cedar CLI treats a missing appliesTo in Cedar as
			// containing two empty lists for resourceTypes and principalTypes
			// rather than being None. These are functionally equivalent.
			appliesTo := schemaast.AppliesTo{}
			if action.AppliesTo != nil {
				appliesTo = *action.AppliesTo
			}

			// The Rust Cedar CLI treats an empty context record in Cedar as
			// None rather than as a Record with nothing in it. Again, these
			// are functionally equivalent.
			if rec, ok := appliesTo.Context.(schemaast.RecordType); ok && len(rec) == 0 {
				appliesTo.Context = nil
			}
			action.AppliesTo = &appliesTo
			out[name] = action
		}
		return out
	}
	out := *in
	out.Actions = convert(in.Actions)
	out.Namespaces = make(schemaast.Namespaces, len(in.Namespaces))
	for name, ns := range in.Namespaces {
		ns.Actions = convert(ns.Actions)
		out.Namespaces[name] = ns
	}
	return &out
}

func normalizeJSON(t *testing.T, in []byte) []byte {
	t.Helper()
	v := jsontext.Value(bytes.Clone(in))
	testutil.OK(t, v.Canonicalize())
	testutil.OK(t, v.Indent(jsontext.WithIndent("  ")))
	return v
}

func stringEquals(t *testing.T, got, want string) {
	t.Helper()
	testutil.Equals(t, strings.TrimSpace(got), strings.TrimSpace(want))
}

// Specific corpus tests that have been extracted for easy regression testing purposes
func TestCorpusRelated(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		policy   string
		entities types.EntityGetter
		request  cedar.Request
		decision cedar.Decision
		reasons  []cedar.PolicyID
		errors   []cedar.PolicyID
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
			nil,
			cedar.Request{Principal: cedar.NewEntityUID("a", "\u0000\u0000"), Action: cedar.NewEntityUID("Action", "action"), Resource: cedar.NewEntityUID("a", "\u0000\u0000")},
			cedar.Deny,
			nil,
			[]cedar.PolicyID{"policy0"},
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
			nil,
			cedar.Request{Principal: cedar.NewEntityUID("a", "\u0000\u0000"), Action: cedar.NewEntityUID("Action", "action"), Resource: cedar.NewEntityUID("a", "\u0000\u0000")},
			cedar.Deny,
			nil,
			[]cedar.PolicyID{"policy0"},
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
			nil,
			cedar.Request{Principal: cedar.NewEntityUID("a", "\u0000\u0000"), Action: cedar.NewEntityUID("Action", "action"), Resource: cedar.NewEntityUID("a", "\u0000\u0000")},
			cedar.Deny,
			nil,
			[]cedar.PolicyID{"policy0"},
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
			nil,
			cedar.Request{Principal: cedar.NewEntityUID("a", "\u0000\u0000"), Action: cedar.NewEntityUID("Action", "action"), Resource: cedar.NewEntityUID("a", "\u0000\u0000")},
			cedar.Deny,
			nil,
			[]cedar.PolicyID{"policy0"},
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
			nil,
			cedar.Request{},
			cedar.Deny,
			nil,
			[]cedar.PolicyID{"policy0"},
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
			nil,
			cedar.Request{},
			cedar.Deny,
			nil,
			[]cedar.PolicyID{"policy0"},
		},

		{"48d0ba6537a3efe02112ba0f5a3daabdcad27b04",
			`forbid(
				principal,
				action in [Action::"action"],
				resource is a in a::"\0\u{8}\u{11}\0R"
			  ) when {
				true && ((if (principal in action) then (ip("")) else (if true then (ip("6b6b:f00::32ff:ffff:6368/00")) else (ip("7265:6c69:706d:6f43:5f74:6f70:7374:6f68")))).isMulticast())
			  };`,
			nil,
			cedar.Request{Principal: cedar.NewEntityUID("a", "\u0000\b\u0011\u0000R"), Action: cedar.NewEntityUID("Action", "action"), Resource: cedar.NewEntityUID("a", "\u0000\b\u0011\u0000R")},
			cedar.Deny,
			nil,
			[]cedar.PolicyID{"policy0"},
		},

		{"48d0ba6537a3efe02112ba0f5a3daabdcad27b04/simplified",
			`forbid(
			principal,
			action,
			resource
		  ) when {
			true && ip("6b6b:f00::32ff:ffff:6368/00").isMulticast()
		  };`,
			nil,
			cedar.Request{},
			cedar.Deny,
			nil,
			[]cedar.PolicyID{"policy0"},
		},

		{name: "e91da4e6af5c73e27f5fb610d723dfa21635d10b",
			policy: `forbid(
				principal is a in a::"\0\0(W\0\0\0",
				action,
				resource
			  ) when {
				true && (([ip("c5c5:c5c5:c5c5:c5c5:c5c5:c5c5:c5c5:c5c5/68")].containsAll([ip("c5c5:c5c5:c5c5:c5c5:c5c5:5cc5:c5c5:c5c5/68")])) || ((ip("")) == (ip(""))))
			  };`,
			request:  cedar.Request{Principal: cedar.NewEntityUID("a", "\u0000\u0000(W\u0000\u0000\u0000"), Action: cedar.NewEntityUID("Action", "action"), Resource: cedar.NewEntityUID("a", "")},
			decision: cedar.Deny,
			reasons:  nil,
			errors:   []cedar.PolicyID{"policy0"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			policy, err := cedar.NewPolicySetFromBytes("", []byte(tt.policy))
			testutil.OK(t, err)
			ok, diag := policy.IsAuthorized(tt.entities, tt.request)
			testutil.Equals(t, ok, tt.decision)
			var reasons []cedar.PolicyID
			for _, n := range diag.Reasons {
				reasons = append(reasons, n.PolicyID)
			}
			testutil.Equals(t, reasons, tt.reasons)
			var errors []cedar.PolicyID
			for _, n := range diag.Errors {
				errors = append(errors, n.PolicyID)
			}
			testutil.Equals(t, errors, tt.errors)
		})
	}
}
