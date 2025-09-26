package cedar

import (
	"io"

	"github.com/cedar-policy/cedar-go/ast"
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

// Decoder reads, parses and compiles [Policy] statements in the human-readable format specified by the [Cedar documentation]
// from an [io.Reader].
//
// [Cedar documentation]: https://docs.cedarpolicy.com/policies/syntax-grammar.html
type Decoder struct {
	dec *parser.Decoder
}

// NewDecoder returns a new [Decoder].
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{dec: parser.NewDecoder(r)}
}

// Decode parses and compiles a single [Policy] statement from the underlying [io.Reader].
func (e *Decoder) Decode(p *Policy) error {
	var parserPolicy parser.Policy

	err := e.dec.Decode(&parserPolicy)
	if err != nil {
		return err
	}

	*p = *NewPolicyFromAST((*ast.Policy)(&parserPolicy))

	return nil
}
