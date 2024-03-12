//go:build corpus

package integration_tests

import (
	"embed"
	"encoding/json"
	"io"
	"testing"

	"github.com/cedar-policy/cedar-go"
)

func fsLoad(t *testing.T, fs embed.FS, fn string) []byte {
	t.Helper()
	tf, err := fs.Open(fn)
	if err != nil {
		t.Fatal("error opening test", err)
	}
	b, err := io.ReadAll(tf)
	if err != nil {
		t.Fatal("error reading test", err)
	}
	return b
}

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

func newEntitiesFromJSON(b []byte) (cedar.Entities, error) {
	var res cedar.Entities
	return res, res.UnmarshalJSON(b)
}
