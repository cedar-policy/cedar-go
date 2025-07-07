package parser

import (
	"errors"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/schema/token"
)

func TestParserErrors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErrs []string // Substrings that should be present in error messages
	}{
		{
			name: "invalid token at schema level",
			input: `
				foo bar;
			`,
			wantErrs: []string{"unexpected token foo"},
		},
		{
			name: "missing comma in entity list",
			input: `
				entity Foo Bar Baz;
			`,
			wantErrs: []string{"expected ;, got Bar"},
		},
		{
			name: "reserved type name",
			input: `
				type String = {
					foo: String
				};
			`,
			wantErrs: []string{"reserved typename String"},
		},
		{
			name: "missing colon in applies to",
			input: `
				action DoSomething appliesTo {
					principal [User]
					resource: [Resource];
				};
			`,
			wantErrs: []string{"expected :"},
		},
		{
			name: "invalid applies to field",
			input: `
				action DoSomething appliesTo {
					foo: [User];
				};
			`,
			wantErrs: []string{"expected principal, resource, or context"},
		},
		{
			name: "missing comma in applies to",
			input: `
				action DoSomething appliesTo {
					principal: [User]
					resource: [Resource];
				};
			`,
			wantErrs: []string{"expected , or }"},
		},
		{
			name: "missing closing brace in record",
			input: `
				type Foo = {
					bar: String,
					baz: Bool
			`,
			wantErrs: []string{"expected , or }"},
		},
		{
			name: "missing comma in record",
			input: `
				type Foo = {
					bar: String
					baz: Bool
				};
			`,
			wantErrs: []string{"expected , or }"},
		},
		{
			name: "invalid Set type",
			input: `
				type Foo = Set<>;
			`,
			wantErrs: []string{"expected type"},
		},
		{
			name: "missing closing angle bracket",
			input: `
				type Foo = Set<String;
			`,
			wantErrs: []string{"expected >"},
		},
		{
			name: "missing type after equals",
			input: `
				type Foo = ;
			`,
			wantErrs: []string{"expected type"},
		},
		{
			name: "missing semicolon after declaration",
			input: `
				type Foo = String
				type Bar = Bool;
			`,
			wantErrs: []string{"expected ;"},
		},
		{
			name: "invalid path separator",
			input: `
				namespace Foo:Bar {
				}
			`,
			wantErrs: []string{"expected { after namespace path, got :"},
		},
		{
			name: "missing comma in type reference list",
			input: `
				action DoSomething in [User::Path::To;
			`,
			wantErrs: []string{"expected , or ]"},
		},
		{
			name: "missing closing bracket in principal list",
			input: `
				action DoSomething appliesTo {
					principal: [User::Path::To;
				}
			`,
			wantErrs: []string{"expected , or ]"},
		},
		{
			name: "invalid token after double colon in path",
			input: `
				entity User in [User::123::Entity];
			`,
			wantErrs: []string{"expected identifier after ::"},
		},
		{
			name: "invalid token where name expected",
			input: `
				action 123;
			`,
			wantErrs: []string{"expected name (identifier or string)"},
		},
		{
			name: "too many errors causing bailout",
			input: `
				type A ~ Other;
				type B ~ Other;
				type C ~ Other;
				type D ~ Other;
				type E ~ Other;
				type F ~ Other;
				type G ~ Other;
				type H ~ Other;
				type I ~ Other;
				type J ~ Other;
				type K ~ Other;
				type L ~ Other;
				type M ~ Other;
				type N ~ Other;
				type O ~ Other;
				type P ~ Other;
				type Q ~ Other;
			`,
			wantErrs: []string{"expected = after typename"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseFile("test.cedar", []byte(tt.input))
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var errs token.Errors
			if !errors.As(err, &errs) {
				t.Fatalf("expected []error, got %T", err)
			}

			// Convert errors to strings for easier matching
			var gotErrs []string
			for _, err := range errs {
				if terr, ok := err.(token.Error); ok {
					gotErrs = append(gotErrs, terr.Err.Error())
				} else {
					gotErrs = append(gotErrs, err.Error())
				}
			}

			// Check that each expected error substring is present
			for _, want := range tt.wantErrs {
				found := false
				for _, got := range gotErrs {
					if strings.Contains(got, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, got errors: %v", want, gotErrs)
				}
			}
		})
	}
}
