package exptypes_test

import (
	"embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
	exptypes "github.com/cedar-policy/cedar-go/x/exp/types"
)

//go:embed testdata
var testdataFS embed.FS

type testManifest struct {
	Schema   string `json:"schema"`
	Entities string `json:"entities"`
}

type parsingResult struct {
	AllEntities struct {
		Success bool `json:"success"`
	} `json:"allEntities"`
	PerEntity map[string]struct {
		Success bool `json:"success"`
	} `json:"perEntity"`
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
		if strings.HasSuffix(name, ".entities.json") || strings.HasSuffix(name, ".parsing.json") {
			continue
		}

		testName := strings.TrimSuffix(name, ".json")
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			// Load manifest
			manifestData, err := testdataFS.ReadFile("testdata/" + name)
			testutil.OK(t, err)
			var manifest testManifest
			testutil.OK(t, json.Unmarshal(manifestData, &manifest))

			// Load expected results
			parsingData, err := testdataFS.ReadFile("testdata/" + testName + ".parsing.json")
			testutil.OK(t, err)
			var expected parsingResult
			testutil.OK(t, json.Unmarshal(parsingData, &expected))

			// Load and parse schema
			schemaContent, err := testdataFS.ReadFile("testdata/" + manifest.Schema)
			testutil.OK(t, err)
			var s schema.Schema
			testutil.OK(t, s.UnmarshalCedar(schemaContent))
			rs, err := s.Resolve()
			testutil.OK(t, err)

			// Load entities
			entitiesContent, err := testdataFS.ReadFile("testdata/" + manifest.Entities)
			testutil.OK(t, err)

			// Test EntityMap.UnmarshalJSONWithSchema
			t.Run("EntityMap", func(t *testing.T) {
				t.Parallel()
				var em exptypes.EntityMap
				err := em.UnmarshalJSONWithSchema(entitiesContent, rs)
				if expected.AllEntities.Success {
					testutil.OK(t, err)
				} else {
					testutil.Error(t, err)
				}
			})

			// Test Entity.UnmarshalJSONWithSchema for each entity
			var entities []json.RawMessage
			if err := json.Unmarshal(entitiesContent, &entities); err != nil {
				// Can't parse as array (e.g., bad JSON) - test Entity with raw bytes
				if len(expected.PerEntity) == 0 {
					t.Run("Entity/raw", func(t *testing.T) {
						t.Parallel()
						var e exptypes.Entity
						err := e.UnmarshalJSONWithSchema(entitiesContent, rs)
						testutil.Error(t, err)
					})
				}
				return
			}

			for i, entityJSON := range entities {
				var uid struct {
					UID struct {
						Type string `json:"type"`
						ID   string `json:"id"`
					} `json:"uid"`
				}
				testutil.OK(t, json.Unmarshal(entityJSON, &uid))
				uidStr := uid.UID.Type + "::" + uid.UID.ID

				expectedEntity, ok := expected.PerEntity[uidStr]
				if !ok {
					t.Fatalf("no expected result for entity %s (index %d)", uidStr, i)
				}

				t.Run("Entity/"+uidStr, func(t *testing.T) {
					t.Parallel()
					var e exptypes.Entity
					err := e.UnmarshalJSONWithSchema(entityJSON, rs)
					if expectedEntity.Success {
						testutil.OK(t, err)
					} else {
						testutil.Error(t, err)
					}
				})
			}
		})
	}
}
