package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/internal/json"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

var fuzzEvalEnv = eval.Env{
	Entities:  types.EntityMap{},
	Principal: types.NewEntityUID("P", "p"),
	Action:    types.NewEntityUID("Action", "a"),
	Resource:  types.NewEntityUID("R", "r"),
	Context:   types.NewRecord(nil),
}

func evalPolicy(p *ast.Policy) (types.Boolean, bool) {
	e := eval.Compile(p)
	v, err := e.Eval(fuzzEvalEnv)
	return v, err == nil
}

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
		v0, ok0 := evalPolicy((*ast.Policy)(&p0))
		v1, ok1 := evalPolicy((*ast.Policy)(&p1))
		if v0 != v1 || ok0 != ok1 {
			t.Fatalf("eval mismatch:\njson: %s\ntext: %s\njson eval: %v (ok=%v)\ntext eval: %v (ok=%v)", in, textA, v0, ok0, v1, ok1)
		}
	})
}
