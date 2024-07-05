package cedar

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
)

// jsonEntity is not part of entityValue as I can find
// no evidence this is part of the JSON spec.  It also
// requires creating a parser, so it's quite expensive.
type jsonEntity EntityUID

func (e *jsonEntity) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	var input EntityUID
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
		Desc      string     `json:"description"`
		Principal jsonEntity `json:"principal"`
		Action    jsonEntity `json:"action"`
		Resource  jsonEntity `json:"resource"`
		Context   Record     `json:"context"`
		Decision  string     `json:"decision"`
		Reasons   []string   `json:"reason"`
		Errors    []string   `json:"errors"`
	} `json:"requests"`
}

type FileDataPointer struct {
	Position int64
	Size     int64
}

type FileDataMap struct {
	buf   io.ReaderAt
	files map[string]FileDataPointer
}

func NewFileDataMap(buf io.ReaderAt) FileDataMap {
	return FileDataMap{
		buf:   buf,
		files: make(map[string]FileDataPointer),
	}
}

func (fdm *FileDataMap) AddFileData(path string, position int64, size int64) {
	fdm.files[path] = FileDataPointer{position, size}
}

func (fdm FileDataMap) GetFileData(path string) ([]byte, error) {
	fdp := fdm.files[path]
	content := make([]byte, fdp.Size)
	_, err := fdm.buf.ReadAt(content, fdp.Position)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func TestCorpus(t *testing.T) {
	t.Parallel()

	corpusArchive, err := os.Open("corpus-tests.tar.gz")
	if err != nil {
		t.Fatal("error opening corpus archive", err)
	}
	defer corpusArchive.Close()

	gzipReader, err := gzip.NewReader(corpusArchive)
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

	fdm := NewFileDataMap(bufReader)
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

			var entities Entities
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
							Principal: EntityUID(request.Principal),
							Action:    EntityUID(request.Action),
							Resource:  EntityUID(request.Resource),
							Context:   request.Context,
						})

					if ok != (request.Decision == "allow") {
						t.Fatalf("got %v want %v", ok, request.Decision)
					}
					var errors []string
					for _, n := range diag.Errors {
						errors = append(errors, fmt.Sprintf("policy%d", n.Policy))
					}
					if !slices.Equal(errors, request.Errors) {
						t.Errorf("errors got %v want %v", errors, request.Errors)
					}
					var reasons []string
					for _, n := range diag.Reasons {
						reasons = append(reasons, fmt.Sprintf("policy%d", n.Policy))
					}
					if !slices.Equal(reasons, request.Reasons) {
						t.Errorf("reasons got %v want %v", reasons, request.Reasons)
					}
				})
			}
		})
	}
}
