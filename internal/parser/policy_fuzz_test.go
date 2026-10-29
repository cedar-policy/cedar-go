package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/json"
)

func fuzzRoundtripJSONSeeds(f *testing.F) [][]byte {
	paths, err := filepath.Glob(filepath.Join("testdata", "roundtrip-seeds", "*.json"))
	if err != nil {
		f.Fatal(err)
	}
	var res [][]byte
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			f.Fatal(err)
		}
		res = append(res, bytes.TrimSpace(b))
	}
	return res
}

func marshalCedarText(p *Policy) []byte {
	var buf bytes.Buffer
	p.MarshalCedar(&buf)
	return buf.Bytes()
}

func FuzzPolicyJSONTextRoundtrip(f *testing.F) {
	for _, text := range fuzzSeedPolicies {
		var p Policy
		if err := p.UnmarshalCedar([]byte(text)); err != nil {
			continue
		}
		b, err := (*json.Policy)(&p).MarshalJSON()
		if err != nil {
			f.Fatalf("seed %q: MarshalJSON: %v", text, err)
		}
		f.Add(b)
	}
	for _, b := range fuzzRoundtripJSONSeeds(f) {
		var p json.Policy
		if err := p.UnmarshalJSON(b); err != nil {
			f.Fatalf("seed %q: UnmarshalJSON: %v", b, err)
		}
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, in []byte) {
		var p0 json.Policy
		if err := p0.UnmarshalJSON(in); err != nil {
			return
		}
		textA := marshalCedarText((*Policy)(&p0))
		var p1 Policy
		if err := p1.UnmarshalCedar(textA); err != nil {
			t.Fatalf("UnmarshalCedar(%q): %v", textA, err)
		}
		textB := marshalCedarText(&p1)
		if !bytes.Equal(textA, textB) {
			t.Fatalf("text mismatch:\na: %s\nb: %s", textA, textB)
		}
	})
}
