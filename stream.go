package cedar

import (
	"io"

	"github.com/cedar-policy/cedar-go/internal/parser"
)

// Encoder encodes [Policy] statements in the human-readable format specified by the [Cedar documentation]
// and writes them to an [io.Writer].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/syntax-grammar.html
type Encoder struct {
	enc *parser.Encoder
}

// NewEncoder returns a new [Encoder].
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{enc: parser.NewEncoder(w)}
}

// Encode encodes and writes a single [Policy] statement to the underlying [io.Writer].
func (e *Encoder) Encode(p *Policy) error {
	return e.enc.Encode((*parser.Policy)(p.AST()))
}
