package scanner

import (
	"testing"
)

// https://go.dev/doc/tutorial/fuzz
// mkdir testdata
// go test -fuzz=FuzzTokenize -fuzztime 60s

func FuzzTokenize(f *testing.F) {
	tests := []string{
		`These are some identifiers`,
		`0 1 1234`,
		`-1 9223372036854775807 -9223372036854775808`,
		`"" "string" "\"\'\n\r\t\\\0" "\x123" "\u{0}\u{10fFfF}"`,
		`"*" "\*" "*\**"`,
		`@.,;(){}[]+-*`,
		`:::`,
		`!!=<<=>>=`,
		`||&&`,
		`// single line comment`,
		`/*`,
		`multiline comment`,
		`// embedded comment does nothing`,
		`*/`,
		`'/%|&=`,
	}
	for _, tt := range tests {
		f.Add(tt)
	}
	f.Fuzz(func(t *testing.T, orig string) {
		toks, err := Tokenize([]byte(orig))
		if err != nil {
			if toks != nil {
				t.Errorf("toks != nil on err")
			}
		}
	})
}
