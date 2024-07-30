package cedar

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/parser"
)

func safeDoErr(f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return f()
}

func TestToEval(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		in    any
		out   evaler
		panic string
	}{
		{"happy", parser.Entity{
			Path: []string{"Action", "test"},
		},
			newLiteralEval(types.EntityValueFromSlice([]string{"Action", "test"})), ""},
		{"missingRelOp", parser.Relation{
			Add: parser.Add{
				Mults: []parser.Mult{
					{
						Unaries: []parser.Unary{
							{
								Member: parser.Member{
									Primary: parser.Primary{
										Type: parser.PrimaryEntity,
										Entity: parser.Entity{
											Path: []string{"Action", "test"},
										},
									},
								},
							},
						},
					},
				},
			},
			RelOpRhs: parser.Add{
				Mults: []parser.Mult{
					{
						Unaries: []parser.Unary{
							{
								Member: parser.Member{
									Primary: parser.Primary{
										Type: parser.PrimaryEntity,
										Entity: parser.Entity{
											Path: []string{"Action", "test"},
										},
									},
								},
							},
						},
					},
				},
			},
			Type:  parser.RelationRelOp,
			RelOp: "invalid",
		},
			nil, "missing RelOp case"},

		{"missingRelationType", parser.Relation{
			Add: parser.Add{
				Mults: []parser.Mult{
					{
						Unaries: []parser.Unary{
							{
								Member: parser.Member{
									Primary: parser.Primary{
										Type: parser.PrimaryEntity,
										Entity: parser.Entity{
											Path: []string{"Action", "test"},
										},
									},
								},
							},
						},
					},
				},
			},
			Type: "invalid",
		},
			nil, "missing RelationType case"},

		{"unknownAddOp", parser.Add{
			Mults: []parser.Mult{
				{
					Unaries: []parser.Unary{
						{
							Member: parser.Member{
								Primary: parser.Primary{
									Type: parser.PrimaryEntity,
									Entity: parser.Entity{
										Path: []string{"Action", "test"},
									},
								},
							},
						},
					},
				},
			},
			AddOps: []parser.AddOp{"invalid"},
		},
			nil, "unknown AddOp"},

		{"missingPrimaryType", parser.Primary{
			Type: parser.PrimaryType(-42),
		},
			nil, "missing PrimaryType case"},

		{"missingLiteralType", parser.Literal{
			Type: parser.LiteralType(-42),
		},
			nil, "missing LiteralType case"},

		{"missingVarType", parser.Var{
			Type: "invalid",
		},
			nil, "missing VarType case"},

		{"unknownNodeType", true,
			nil, "unknown node type bool"},

		{"missingAccessType", parser.Member{
			Primary: parser.Primary{
				Type: parser.PrimaryEntity,
				Entity: parser.Entity{
					Path: []string{"Action", "test"},
				},
			},
			Accesses: []parser.Access{
				{
					Type: parser.AccessType(-42),
				},
			},
		},
			nil, "missing AccessType case"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var out evaler
			err := safeDoErr(func() error {
				out = toEval(tt.in)
				return nil
			})
			testutil.Equals(t, out, tt.out)
			testutil.Equals(t, err != nil, tt.panic != "")
			if tt.panic != "" {
				testutil.FatalIf(t, !strings.Contains(err.Error(), tt.panic), "panic got %v want %v", err.Error(), tt.panic)
			}
		})
	}
}
