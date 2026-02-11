package schema_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

var wantCedar = `
@doc("Address information")
@personal_information
type Address = {
	@also("town")
	city: String,
	country: Country,
	street: String,
	zipcode?: String
};

type decimal = {
	decimal: Long,
	whole: Long
};

entity Admin;

entity Country;

entity System in Admin {
	version: String
};

entity Role enum ["superuser", "operator"];

action audit appliesTo {
	principal: Admin,
	resource: [MyApp::Document, System]
};

@doc("Doc manager")
namespace MyApp {
	type Metadata = {
		created: datetime,
		tags: Set<String>
	};

	entity Department {
		budget: decimal
	};

	entity Document {
		public: Bool,
		title: String
	};

	entity Group in Department {
		metadata: Metadata,
		name: String
	};

	@doc("User entity")
	entity User in Group {
		active: Bool,
		address: Address,
		email: String,
		level: Long
	};

	entity Status enum ["draft", "published", "archived"];

	@doc("View or edit document")
	action edit appliesTo {
		principal: User,
		resource: Document,
		context: {
			ip: ipaddr,
			timestamp: datetime
		}
	};

	action manage appliesTo {
		principal: User,
		resource: [Document, Group]
	};

	@doc("View or edit document")
	action view appliesTo {
		principal: User,
		resource: Document,
		context: {
			ip: ipaddr,
			timestamp: datetime
		}
	};
}
`

var wantJSON = `{
  "": {
    "entityTypes": {
      "Admin": {},
      "Country": {},
      "Role": {
        "enum": ["superuser", "operator"]
      },
      "System": {
        "memberOfTypes": ["Admin"],
        "shape": {
          "type": "Record",
          "attributes": {
            "version": {
              "type": "EntityOrCommon",
              "name": "String"
            }
          }
        }
      }
    },
    "actions": {
      "audit": {
        "appliesTo": {
          "principalTypes": ["Admin"],
          "resourceTypes": ["MyApp::Document", "System"]
        }
      }
    },
    "commonTypes": {
      "Address": {
        "type": "Record",
        "attributes": {
          "city": {
            "type": "EntityOrCommon",
            "name": "String",
            "annotations": {
              "also": "town"
            }
          },
          "country": {
            "type": "EntityOrCommon",
            "name": "Country"
          },
          "street": {
            "type": "EntityOrCommon",
            "name": "String"
          },
          "zipcode": {
            "type": "EntityOrCommon",
            "name": "String",
            "required": false
          }
        },
        "annotations": {
          "doc": "Address information",
          "personal_information": ""
        }
      },
      "decimal": {
        "type": "Record",
        "attributes": {
          "decimal": {
            "type": "EntityOrCommon",
            "name": "Long"
          },
          "whole": {
            "type": "EntityOrCommon",
            "name": "Long"
          }
        }
      }
    }
  },
  "MyApp": {
    "annotations": {
      "doc": "Doc manager"
    },
    "entityTypes": {
      "Department": {
        "shape": {
          "type": "Record",
          "attributes": {
            "budget": {
              "type": "EntityOrCommon",
              "name": "decimal"
            }
          }
        }
      },
      "Document": {
        "shape": {
          "type": "Record",
          "attributes": {
            "public": {
              "type": "EntityOrCommon",
              "name": "Bool"
            },
            "title": {
              "type": "EntityOrCommon",
              "name": "String"
            }
          }
        }
      },
      "Group": {
        "memberOfTypes": ["Department"],
        "shape": {
          "type": "Record",
          "attributes": {
            "metadata": {
              "type": "EntityOrCommon",
              "name": "Metadata"
            },
            "name": {
              "type": "EntityOrCommon",
              "name": "String"
            }
          }
        }
      },
      "Status": {
        "enum": ["draft", "published", "archived"]
      },
      "User": {
        "memberOfTypes": ["Group"],
        "shape": {
          "type": "Record",
          "attributes": {
            "active": {
              "type": "EntityOrCommon",
              "name": "Bool"
            },
            "address": {
              "type": "EntityOrCommon",
              "name": "Address"
            },
            "email": {
              "type": "EntityOrCommon",
              "name": "String"
            },
            "level": {
              "type": "EntityOrCommon",
              "name": "Long"
            }
          }
        },
        "annotations": {
          "doc": "User entity"
        }
      }
    },
    "actions": {
      "edit": {
        "appliesTo": {
          "principalTypes": ["User"],
          "resourceTypes": ["Document"],
          "context": {
            "type": "Record",
            "attributes": {
              "ip": {
                "type": "EntityOrCommon",
                "name": "ipaddr"
              },
              "timestamp": {
                "type": "EntityOrCommon",
                "name": "datetime"
              }
            }
          }
        },
        "annotations": {
          "doc": "View or edit document"
        }
      },
      "manage": {
        "appliesTo": {
          "principalTypes": ["User"],
          "resourceTypes": ["Document", "Group"]
        }
      },
      "view": {
        "appliesTo": {
          "principalTypes": ["User"],
          "resourceTypes": ["Document"],
          "context": {
            "type": "Record",
            "attributes": {
              "ip": {
                "type": "EntityOrCommon",
                "name": "ipaddr"
              },
              "timestamp": {
                "type": "EntityOrCommon",
                "name": "datetime"
              }
            }
          }
        },
        "annotations": {
          "doc": "View or edit document"
        }
      }
    },
    "commonTypes": {
      "Metadata": {
        "type": "Record",
        "attributes": {
          "created": {
            "type": "EntityOrCommon",
            "name": "datetime"
          },
          "tags": {
            "type": "Set",
            "element": {
              "type": "EntityOrCommon",
              "name": "String"
            }
          }
        }
      }
    }
  }
}`

// wantAST is the expected AST structure for the test schema.
// The Cedar parser produces ast.TypeRef for all type names (including
// builtins like String, Long, Bool). Resolution happens later.
var wantAST = &ast.Schema{
	CommonTypes: ast.CommonTypes{
		"Address": ast.CommonType{
			Annotations: ast.Annotations{
				"doc":                  "Address information",
				"personal_information": "",
			},
			Type: ast.RecordType{
				"city": ast.Attribute{
					Type: ast.TypeRef("String"),
					Annotations: ast.Annotations{
						"also": "town",
					},
				},
				"country": ast.Attribute{Type: ast.TypeRef("Country")},
				"street":  ast.Attribute{Type: ast.TypeRef("String")},
				"zipcode": ast.Attribute{Type: ast.TypeRef("String"), Optional: true},
			},
		},
		"decimal": ast.CommonType{
			Type: ast.RecordType{
				"decimal": ast.Attribute{Type: ast.TypeRef("Long")},
				"whole":   ast.Attribute{Type: ast.TypeRef("Long")},
			},
		},
	},
	Entities: ast.Entities{
		"Admin":   ast.Entity{},
		"Country": ast.Entity{},
		"System": ast.Entity{
			ParentTypes: []ast.EntityTypeRef{"Admin"},
			Shape: ast.RecordType{
				"version": ast.Attribute{Type: ast.TypeRef("String")},
			},
		},
	},
	Enums: ast.Enums{
		"Role": ast.Enum{
			Values: []types.String{"superuser", "operator"},
		},
	},
	Actions: ast.Actions{
		"audit": ast.Action{
			AppliesTo: &ast.AppliesTo{
				Principals: []ast.EntityTypeRef{"Admin"},
				Resources:  []ast.EntityTypeRef{"MyApp::Document", "System"},
			},
		},
	},
	Namespaces: ast.Namespaces{
		"MyApp": ast.Namespace{
			Annotations: ast.Annotations{
				"doc": "Doc manager",
			},
			CommonTypes: ast.CommonTypes{
				"Metadata": ast.CommonType{
					Type: ast.RecordType{
						"created": ast.Attribute{Type: ast.TypeRef("datetime")},
						"tags":    ast.Attribute{Type: ast.SetType{Element: ast.TypeRef("String")}},
					},
				},
			},
			Entities: ast.Entities{
				"Department": ast.Entity{
					Shape: ast.RecordType{
						"budget": ast.Attribute{Type: ast.TypeRef("decimal")},
					},
				},
				"Document": ast.Entity{
					Shape: ast.RecordType{
						"public": ast.Attribute{Type: ast.TypeRef("Bool")},
						"title":  ast.Attribute{Type: ast.TypeRef("String")},
					},
				},
				"Group": ast.Entity{
					ParentTypes: []ast.EntityTypeRef{"Department"},
					Shape: ast.RecordType{
						"metadata": ast.Attribute{Type: ast.TypeRef("Metadata")},
						"name":     ast.Attribute{Type: ast.TypeRef("String")},
					},
				},
				"User": ast.Entity{
					ParentTypes: []ast.EntityTypeRef{"Group"},
					Annotations: ast.Annotations{
						"doc": "User entity",
					},
					Shape: ast.RecordType{
						"active":  ast.Attribute{Type: ast.TypeRef("Bool")},
						"address": ast.Attribute{Type: ast.TypeRef("Address")},
						"email":   ast.Attribute{Type: ast.TypeRef("String")},
						"level":   ast.Attribute{Type: ast.TypeRef("Long")},
					},
				},
			},
			Enums: ast.Enums{
				"Status": ast.Enum{
					Values: []types.String{"draft", "published", "archived"},
				},
			},
			Actions: ast.Actions{
				"edit": ast.Action{
					Annotations: ast.Annotations{
						"doc": "View or edit document",
					},
					AppliesTo: &ast.AppliesTo{
						Principals: []ast.EntityTypeRef{"User"},
						Resources:  []ast.EntityTypeRef{"Document"},
						Context: ast.RecordType{
							"ip":        ast.Attribute{Type: ast.TypeRef("ipaddr")},
							"timestamp": ast.Attribute{Type: ast.TypeRef("datetime")},
						},
					},
				},
				"manage": ast.Action{
					AppliesTo: &ast.AppliesTo{
						Principals: []ast.EntityTypeRef{"User"},
						Resources:  []ast.EntityTypeRef{"Document", "Group"},
					},
				},
				"view": ast.Action{
					Annotations: ast.Annotations{
						"doc": "View or edit document",
					},
					AppliesTo: &ast.AppliesTo{
						Principals: []ast.EntityTypeRef{"User"},
						Resources:  []ast.EntityTypeRef{"Document"},
						Context: ast.RecordType{
							"ip":        ast.Attribute{Type: ast.TypeRef("ipaddr")},
							"timestamp": ast.Attribute{Type: ast.TypeRef("datetime")},
						},
					},
				},
			},
		},
	},
}

// wantResolved is the expected resolved schema structure.
// All type references have been fully qualified and common types inlined.
var wantResolved = &resolved.Schema{
	Namespaces: map[types.Namespace]resolved.Namespace{
		"MyApp": {
			Name: "MyApp",
			Annotations: resolved.Annotations{
				"doc": "Doc manager",
			},
		},
	},
	Entities: map[types.EntityType]resolved.Entity{
		"Admin": {
			Name: "Admin",
		},
		"Country": {
			Name: "Country",
		},
		"System": {
			Name:        "System",
			ParentTypes: []types.EntityType{"Admin"},
			Shape: resolved.RecordType{
				"version": resolved.Attribute{Type: resolved.StringType{}},
			},
		},
		"MyApp::Department": {
			Name: "MyApp::Department",
			Shape: resolved.RecordType{
				"budget": resolved.Attribute{
					Type: resolved.RecordType{
						"decimal": resolved.Attribute{Type: resolved.LongType{}},
						"whole":   resolved.Attribute{Type: resolved.LongType{}},
					},
				},
			},
		},
		"MyApp::Document": {
			Name: "MyApp::Document",
			Shape: resolved.RecordType{
				"public": resolved.Attribute{Type: resolved.BoolType{}},
				"title":  resolved.Attribute{Type: resolved.StringType{}},
			},
		},
		"MyApp::Group": {
			Name:        "MyApp::Group",
			ParentTypes: []types.EntityType{"MyApp::Department"},
			Shape: resolved.RecordType{
				"metadata": resolved.Attribute{
					Type: resolved.RecordType{
						"created": resolved.Attribute{Type: resolved.ExtensionType("datetime")},
						"tags":    resolved.Attribute{Type: resolved.SetType{Element: resolved.StringType{}}},
					},
				},
				"name": resolved.Attribute{Type: resolved.StringType{}},
			},
		},
		"MyApp::User": {
			Name:        "MyApp::User",
			Annotations: resolved.Annotations{"doc": "User entity"},
			ParentTypes: []types.EntityType{"MyApp::Group"},
			Shape: resolved.RecordType{
				"active": resolved.Attribute{Type: resolved.BoolType{}},
				"address": resolved.Attribute{
					Type: resolved.RecordType{
						"city": resolved.Attribute{
							Type:        resolved.StringType{},
							Annotations: resolved.Annotations{"also": "town"},
						},
						"country": resolved.Attribute{Type: resolved.EntityType("Country")},
						"street":  resolved.Attribute{Type: resolved.StringType{}},
						"zipcode": resolved.Attribute{Type: resolved.StringType{}, Optional: true},
					},
				},
				"email": resolved.Attribute{Type: resolved.StringType{}},
				"level": resolved.Attribute{Type: resolved.LongType{}},
			},
		},
	},
	Enums: map[types.EntityType]resolved.Enum{
		"Role": {
			Name:   "Role",
			Values: []types.EntityUID{types.NewEntityUID("Role", "superuser"), types.NewEntityUID("Role", "operator")},
		},
		"MyApp::Status": {
			Name:   "MyApp::Status",
			Values: []types.EntityUID{types.NewEntityUID("MyApp::Status", "draft"), types.NewEntityUID("MyApp::Status", "published"), types.NewEntityUID("MyApp::Status", "archived")},
		},
	},
	Actions: map[types.EntityUID]resolved.Action{
		types.NewEntityUID("Action", "audit"): {
			Name: "audit",
			AppliesTo: &resolved.AppliesTo{
				Principals: []types.EntityType{"Admin"},
				Resources:  []types.EntityType{"MyApp::Document", "System"},
				Context:    resolved.RecordType{},
			},
		},
		types.NewEntityUID("MyApp::Action", "edit"): {
			Name:        "edit",
			Annotations: resolved.Annotations{"doc": "View or edit document"},
			AppliesTo: &resolved.AppliesTo{
				Principals: []types.EntityType{"MyApp::User"},
				Resources:  []types.EntityType{"MyApp::Document"},
				Context: resolved.RecordType{
					"ip":        resolved.Attribute{Type: resolved.ExtensionType("ipaddr")},
					"timestamp": resolved.Attribute{Type: resolved.ExtensionType("datetime")},
				},
			},
		},
		types.NewEntityUID("MyApp::Action", "manage"): {
			Name: "manage",
			AppliesTo: &resolved.AppliesTo{
				Principals: []types.EntityType{"MyApp::User"},
				Resources:  []types.EntityType{"MyApp::Document", "MyApp::Group"},
				Context:    resolved.RecordType{},
			},
		},
		types.NewEntityUID("MyApp::Action", "view"): {
			Name:        "view",
			Annotations: resolved.Annotations{"doc": "View or edit document"},
			AppliesTo: &resolved.AppliesTo{
				Principals: []types.EntityType{"MyApp::User"},
				Resources:  []types.EntityType{"MyApp::Document"},
				Context: resolved.RecordType{
					"ip":        resolved.Attribute{Type: resolved.ExtensionType("ipaddr")},
					"timestamp": resolved.Attribute{Type: resolved.ExtensionType("datetime")},
				},
			},
		},
	},
}

func TestSchema(t *testing.T) {
	t.Parallel()

	t.Run("UnmarshalCedar", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema
		err := s.UnmarshalCedar([]byte(wantCedar))
		testutil.OK(t, err)
		testutil.Equals(t, s.AST(), wantAST)
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema
		err := s.UnmarshalJSON([]byte(wantJSON))
		testutil.OK(t, err)
		testutil.Equals(t, s.AST(), wantAST)
	})

	t.Run("MarshalCedar", func(t *testing.T) {
		t.Parallel()
		s := schema.NewSchemaFromAST(wantAST)
		b, err := s.MarshalCedar()
		testutil.OK(t, err)
		stringEquals(t, string(b), wantCedar)
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		t.Parallel()
		s := schema.NewSchemaFromAST(wantAST)
		b, err := s.MarshalJSON()
		testutil.OK(t, err)
		stringEquals(t, string(normalizeJSON(t, b)), string(normalizeJSON(t, []byte(wantJSON))))
	})

	t.Run("Resolve", func(t *testing.T) {
		t.Parallel()
		s := schema.NewSchemaFromAST(wantAST)
		r, err := s.Resolve()
		testutil.OK(t, err)
		testutil.Equals(t, r, wantResolved)
	})

	t.Run("CedarRoundTrip", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema
		testutil.OK(t, s.UnmarshalCedar([]byte(wantCedar)))
		b, err := s.MarshalCedar()
		testutil.OK(t, err)
		var s2 schema.Schema
		testutil.OK(t, s2.UnmarshalCedar(b))
		testutil.Equals(t, s2.AST(), wantAST)
	})

	t.Run("JSONRoundTrip", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema
		testutil.OK(t, s.UnmarshalJSON([]byte(wantJSON)))
		b, err := s.MarshalJSON()
		testutil.OK(t, err)
		var s2 schema.Schema
		testutil.OK(t, s2.UnmarshalJSON(b))
		testutil.Equals(t, s2.AST(), wantAST)
	})

	t.Run("CedarToJSONRoundTrip", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema
		testutil.OK(t, s.UnmarshalCedar([]byte(wantCedar)))
		jsonBytes, err := s.MarshalJSON()
		testutil.OK(t, err)
		var s2 schema.Schema
		testutil.OK(t, s2.UnmarshalJSON(jsonBytes))
		testutil.Equals(t, s2.AST(), wantAST)
	})

	t.Run("JSONToCedarRoundTrip", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema
		testutil.OK(t, s.UnmarshalJSON([]byte(wantJSON)))
		cedarBytes, err := s.MarshalCedar()
		testutil.OK(t, err)
		var s2 schema.Schema
		testutil.OK(t, s2.UnmarshalCedar(cedarBytes))
		testutil.Equals(t, s2.AST(), wantAST)
	})

	t.Run("JSONMarshalInterface", func(t *testing.T) {
		t.Parallel()
		s := schema.NewSchemaFromAST(wantAST)
		b, err := json.Marshal(s)
		testutil.OK(t, err)
		var s2 schema.Schema
		testutil.OK(t, json.Unmarshal(b, &s2))
		testutil.Equals(t, s2.AST(), wantAST)
	})

	t.Run("UnmarshalCedarErr", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema
		const filename = "path/to/my-file-name.cedarschema"
		s.SetFilename(filename)
		err := s.UnmarshalCedar([]byte("LSKJDFN"))
		testutil.Error(t, err)
		testutil.FatalIf(t, !strings.Contains(err.Error(), filename+":1:1"), "expected %q in error: %v", filename, err)
	})

	t.Run("UnmarshalJSONErr", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema
		err := s.UnmarshalJSON([]byte("LSKJDFN"))
		testutil.Error(t, err)
	})

	t.Run("ResolveErr", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema
		testutil.OK(t, s.UnmarshalCedar([]byte(`entity User in [NonExistent];`)))
		_, err := s.Resolve()
		testutil.Error(t, err)
	})

	t.Run("ZeroValueSchema", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema

		b, err := s.MarshalCedar()
		testutil.OK(t, err)
		testutil.Equals(t, string(b), "")

		jb, err := s.MarshalJSON()
		testutil.OK(t, err)
		testutil.Equals(t, string(jb), "{}")

		r, err := s.Resolve()
		testutil.OK(t, err)
		testutil.Equals(t, r != nil, true)

		testutil.Equals(t, s.AST() != nil, true)
	})

	t.Run("EmptySchema", func(t *testing.T) {
		t.Parallel()
		s := schema.NewSchemaFromAST(&ast.Schema{})
		b, err := s.MarshalCedar()
		testutil.OK(t, err)
		testutil.Equals(t, string(b), "")

		jb, err := s.MarshalJSON()
		testutil.OK(t, err)
		testutil.Equals(t, string(jb), "{}")
	})
}

func stringEquals(t *testing.T, got, want string) {
	t.Helper()
	testutil.Equals(t, strings.TrimSpace(got), strings.TrimSpace(want))
}

func normalizeJSON(t *testing.T, in []byte) []byte {
	t.Helper()
	var out any
	err := json.Unmarshal(in, &out)
	testutil.OK(t, err)
	b, err := json.MarshalIndent(out, "", "  ")
	testutil.OK(t, err)
	return b
}
