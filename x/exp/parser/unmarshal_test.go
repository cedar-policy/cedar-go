package parser_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/x/exp/parser"
)

func TestNewDecoder(t *testing.T) {
	_ = parser.NewDecoder(nil)
}
