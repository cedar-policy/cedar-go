package cedar

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	entities2 "github.com/cedar-policy/cedar-go/internal/entities"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

// jsonEntity is not part of entityValue as I can find
// no evidence this is part of the JSON spec.  It also
// requires creating a parser, so it's quite expensive.
type jsonEntity types.EntityUID

func (e *jsonEntity) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	var input types.EntityUID
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
		Desc      string       `json:"description"`
		Principal jsonEntity   `json:"principal"`
		Action    jsonEntity   `json:"action"`
		Resource  jsonEntity   `json:"resource"`
		Context   types.Record `json:"context"`
		Decision  string       `json:"decision"`
		Reasons   []string     `json:"reason"`
		Errors    []string     `json:"errors"`
	} `json:"requests"`
}

//go:embed corpus-tests.tar.gz
var corpusArchive []byte

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

func TestCorpus(t *testing.T) {
	t.Parallel()

	gzipReader, err := gzip.NewReader(bytes.NewReader(corpusArchive))
	if err != nil {
		t.Fatal("error reading corpus compressed archive header", err)
	}
	defer gzipReader.Close()

	buf, err := io.ReadAll(gzipReader)
	if err != nil {
		t.Fatal("error reading corpus compressed archive", err)
	}

	bufReader := bytes.NewReader(buf)
	archiveReader := tar.NewReader(bufReader)

	fdm := NewTarFileMap(bufReader)
	var testFiles []string
	for file, err := archiveReader.Next(); err == nil; file, err = archiveReader.Next() {
		if file.Typeflag != tar.TypeReg {
			continue
		}

		cursor, _ := bufReader.Seek(0, io.SeekCurrent)
		fdm.AddFileData(file.Name, cursor, file.Size)

		if strings.HasSuffix(file.Name, ".json") && !strings.HasSuffix(file.Name, ".entities.json") {
			testFiles = append(testFiles, file.Name)
		}
	}

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

			entitiesContent, err := fdm.GetFileData(tt.Entities)
			if err != nil {
				t.Fatal("error reading entities content", err)
			}

			var entities entities2.Entities
			if err := json.Unmarshal(entitiesContent, &entities); err != nil {
				t.Fatal("error unmarshalling test", err)
			}

			policyContent, err := fdm.GetFileData(tt.Policies)
			if err != nil {
				t.Fatal("error reading policy content", err)
			}

			policySet, err := NewPolicySet("policy.cedar", policyContent)
			if err != nil {
				t.Fatal("error parsing policy set", err)
			}

			for _, request := range tt.Requests {
				t.Run(request.Desc, func(t *testing.T) {
					t.Parallel()
					ok, diag := policySet.IsAuthorized(
						entities,
						Request{
							Principal: types.EntityUID(request.Principal),
							Action:    types.EntityUID(request.Action),
							Resource:  types.EntityUID(request.Resource),
							Context:   request.Context,
						})

					if ok != (request.Decision == "allow") {
						t.Fatalf("got %v want %v", ok, request.Decision)
					}
					var errors []string
					for _, n := range diag.Errors {
						errors = append(errors, string(n.PolicyID))
					}
					if !slices.Equal(errors, request.Errors) {
						t.Errorf("errors got %v want %v", errors, request.Errors)
					}
					var reasons []string
					for _, n := range diag.Reasons {
						reasons = append(reasons, string(n.PolicyID))
					}
					if !slices.Equal(reasons, request.Reasons) {
						t.Errorf("reasons got %v want %v", reasons, request.Reasons)
					}
				})
			}
		})
	}
}

// Specific corpus tests that have been extracted for easy regression testing purposes
func TestCorpusRelated(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		policy   string
		request  Request
		decision Decision
		reasons  []PolicyID
		errors   []PolicyID
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
			[]PolicyID{"policy0"},
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
			[]PolicyID{"policy0"},
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
			[]PolicyID{"policy0"},
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
			[]PolicyID{"policy0"},
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
			[]PolicyID{"policy0"},
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
			[]PolicyID{"policy0"},
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
			[]PolicyID{"policy0"},
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
			[]PolicyID{"policy0"},
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
			errors:   []PolicyID{"policy0"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			policy, err := NewPolicySet("", []byte(tt.policy))
			testutil.OK(t, err)
			ok, diag := policy.IsAuthorized(entities2.Entities{}, tt.request)
			testutil.Equals(t, ok, tt.decision)
			var reasons []PolicyID
			for _, n := range diag.Reasons {
				reasons = append(reasons, n.PolicyID)
			}
			testutil.Equals(t, reasons, tt.reasons)
			var errors []PolicyID
			for _, n := range diag.Errors {
				errors = append(errors, n.PolicyID)
			}
			testutil.Equals(t, errors, tt.errors)
		})
	}
}
