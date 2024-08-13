package parser

import (
	"testing"
)

// https://go.dev/doc/tutorial/fuzz
// mkdir testdata
// go test -fuzz=FuzzTokenize -fuzztime 60s
// go test -fuzz=FuzzParse -fuzztime 60s

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

func FuzzParse(f *testing.F) {
	tests := []string{
		`permit(principal,action,resource);`,
		`forbid(principal,action,resource);`,
		`permit(principal,action,resource in asdf::"1234");`,
		`permit(principal,action,resource) when { resource in "foo" };`,
		`permit(principal,action,resource) when { context.x == 42 };`,
		`permit(principal,action,resource) when { context.x == 42 };`,
		`permit(principal,action,resource) when { principal.x == 42 };`,
		`permit(principal,action,resource) when { principal.x == 42 };`,
		`permit(principal,action,resource) when { principal in parent::"bob" };`,
		`permit(principal == coder::"cuzco",action,resource);`,
		`permit(principal in team::"osiris",action,resource);`,
		`permit(principal,action == table::"drop",resource);`,
		`permit(principal,action in scary::"stuff",resource);`,
		`permit(principal,action in [scary::"stuff"],resource);`,
		`permit(principal,action,resource == table::"whatever");`,
		`permit(principal,action,resource) unless { false };`,
		`permit(principal,action,resource) when { (if true then true else true) };`,
		`permit(principal,action,resource) when { (true || false) };`,
		`permit(principal,action,resource) when { (true && true) };`,
		`permit(principal,action,resource) when { (1<2) && (1<=1) && (2>1) && (1>=1) && (1!=2) && (1==1)};`,
		`permit(principal,action,resource) when { principal in principal };`,
		`permit(principal,action,resource) when { principal has name };`,
		`permit(principal,action,resource) when { 40+3-1==42 };`,
		`permit(principal,action,resource) when { 6*7==42 };`,
		`permit(principal,action,resource) when { -42==-42 };`,
		`permit(principal,action,resource) when { !(1+1==42) };`,
		`permit(principal,action,resource) when { [1,2,3].contains(2) };`,
		`permit(principal,action,resource) when { {name:"bob"} has name };`,
		`permit(principal,action,resource) when { action in action };`,
		`permit(principal,action,resource) when { [1,2,3].contains(2) };`,
		`permit(principal,action,resource) when { [1,2,3].contains(2,3) };`,
		`permit(principal,action,resource) when { [1,2,3].containsAll([2,3]) };`,
		`permit(principal,action,resource) when { [1,2,3].containsAll(2,3) };`,
		`permit(principal,action,resource) when { [1,2,3].containsAny([2,5]) };`,
		`permit(principal,action,resource) when { [1,2,3].containsAny(2,5) };`,
		`permit(principal,action,resource) when { {name:"bob"}["name"] == "bob" };`,
		`permit(principal,action,resource) when { [1,2,3].shuffle() };`,
		`permit(principal,action,resource) when { "bananas" like "*nan*" };`,
		`permit(principal,action,resource) when { fooBar("10") };`,
		`permit(principal,action,resource) when { decimal(1, 2) };`,
		`permit(principal,action,resource) when { ip() };`,
		`permit(principal,action,resource) when { ip("1.2.3.4").isIpv4(true) };`,
		`permit(principal,action,resource) when { ip("1.2.3.4").isIpv6(true) };`,
		`permit(principal,action,resource) when { ip("1.2.3.4").isLoopback(true) };`,
		`permit(principal,action,resource) when { ip("1.2.3.4").isMulticast(true) };`,
		`permit(principal,action,resource) when { ip("1.2.3.4").isInRange() };`,
	}
	for _, tt := range tests {
		f.Add(tt)
	}
	f.Fuzz(func(_ *testing.T, orig string) {
		// intentionally ignore parse errors
		var policy Policy
		_ = policy.UnmarshalCedar([]byte(orig))
	})
}
