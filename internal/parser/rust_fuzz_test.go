package parser

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/json"
)

type rustCedar struct {
	mu  sync.Mutex
	in  io.Writer
	out *bufio.Reader
}

var (
	rustCedarOnce sync.Once
	rustCedarInst *rustCedar
	errRustCedar  error
)

func rustCedarToolPath() string {
	if p := os.Getenv("CEDAR_FUZZ_TOOL"); p != "" {
		return p
	}
	return filepath.Join("..", "..", "test", "cedar-fuzz-tool", "target", "release", "cedar-fuzz-tool")
}

func startRustCedar(tb testing.TB) *rustCedar {
	tb.Helper()
	path := rustCedarToolPath()
	if _, err := os.Stat(path); err != nil {
		tb.Skipf("Rust Cedar fuzz tool not built at %s; run mise run build:fuzz-tool", path)
	}
	rustCedarOnce.Do(func() {
		cmd := exec.Command(path)
		cmd.Stderr = os.Stderr
		in, err := cmd.StdinPipe()
		if err != nil {
			errRustCedar = err
			return
		}
		out, err := cmd.StdoutPipe()
		if err != nil {
			errRustCedar = err
			return
		}
		if err := cmd.Start(); err != nil {
			errRustCedar = err
			return
		}
		rustCedarInst = &rustCedar{in: in, out: bufio.NewReader(out)}
	})
	if errRustCedar != nil {
		tb.Fatalf("starting Rust Cedar fuzz tool: %v", errRustCedar)
	}
	return rustCedarInst
}

func (r *rustCedar) call(op byte, input []byte) (text string, accepted bool, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	req := make([]byte, 5, 5+len(input))
	req[0] = op
	binary.BigEndian.PutUint32(req[1:], uint32(len(input)))
	req = append(req, input...)
	if _, err := r.in.Write(req); err != nil {
		return "", false, err
	}
	var head [5]byte
	if _, err := io.ReadFull(r.out, head[:]); err != nil {
		return "", false, err
	}
	body := make([]byte, binary.BigEndian.Uint32(head[1:]))
	if _, err := io.ReadFull(r.out, body); err != nil {
		return "", false, err
	}
	switch head[0] {
	case 0:
		return string(body), true, nil
	case 1:
		return string(body), false, nil
	default:
		return "", false, fmt.Errorf("unknown status %d", head[0])
	}
}

func (r *rustCedar) fromText(tb testing.TB, text []byte) (string, bool) {
	tb.Helper()
	out, ok, err := r.call('t', text)
	if err != nil {
		tb.Fatalf("Rust Cedar fuzz tool: %v", err)
	}
	return out, ok
}

func (r *rustCedar) fromJSON(tb testing.TB, b []byte) (string, bool) {
	tb.Helper()
	out, ok, err := r.call('j', b)
	if err != nil {
		tb.Fatalf("Rust Cedar fuzz tool: %v", err)
	}
	return out, ok
}

func FuzzRustCedarText(f *testing.F) {
	for _, text := range fuzzSeedPolicies {
		f.Add(text)
	}
	for _, b := range fuzzRoundtripJSONSeeds(f) {
		var p json.Policy
		if err := p.UnmarshalJSON(b); err != nil {
			f.Fatalf("seed %q: UnmarshalJSON: %v", b, err)
		}
		f.Add(string(marshalCedarText((*Policy)(&p))))
	}
	f.Fuzz(func(t *testing.T, in string) {
		var p Policy
		if err := p.UnmarshalCedar([]byte(in)); err != nil {
			return
		}
		goText := rustMarshalCedar(&p)
		rustText, ok := startRustCedar(t).fromText(t, goText)
		if !ok {
			t.Fatalf("Rust Cedar rejected Go's text:\n%s\nerror: %s", goText, rustText)
		}
		if string(goText) != rustText {
			t.Fatalf("text mismatch:\ngo:   %s\nrust: %s", goText, rustText)
		}
	})
}

func FuzzRustCedarJSON(f *testing.F) {
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
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, in []byte) {
		var p json.Policy
		goErr := p.UnmarshalJSONMode(in, json.ModeStrict)
		var pGo json.Policy
		if err := pGo.UnmarshalJSONMode(in, json.ModeGo); (err == nil) != (goErr == nil) {
			t.Fatalf("Go and strict modes disagree:\n%s\ngo error: %v\nstrict error: %v", in, err, goErr)
		}
		rustText, ok := startRustCedar(t).fromJSON(t, in)
		switch {
		case goErr != nil && ok:
			t.Fatalf("Go rejected JSON that Rust accepted:\n%s\ngo error: %v\nrust text: %s", in, goErr, rustText)
		case goErr == nil && !ok:
			t.Fatalf("Go accepted JSON that Rust rejected:\n%s\nrust error: %s", in, rustText)
		case goErr != nil:
			return
		}
		goText := rustMarshalCedar((*Policy)(&p))
		if string(goText) != rustText {
			t.Fatalf("text mismatch:\ngo:   %s\nrust: %s", goText, rustText)
		}
	})
}
