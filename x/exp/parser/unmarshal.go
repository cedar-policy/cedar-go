package parser

import (
	"io"

	"github.com/cedar-policy/cedar-go/internal/parser"
)

type Decoder = parser.Decoder

func NewDecoder(r io.Reader) *Decoder {
	return parser.NewDecoder(r)
}
