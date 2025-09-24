package parser_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/x/exp/parser"
)

func TestNewEncoder(t *testing.T) {
	_ = parser.NewEncoder(nil)
}
