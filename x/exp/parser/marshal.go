package parser

import (
	"io"

	"github.com/cedar-policy/cedar-go/internal/parser"
)

type Encoder = parser.Encoder

func NewEncoder(w io.Writer) *Encoder {
	return parser.NewEncoder(w)
}
