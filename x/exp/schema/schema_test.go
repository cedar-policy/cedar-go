package schema_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
	ast2 "github.com/cedar-policy/cedar-go/x/exp/schema/ast"
	resolved2 "github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
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
var wantAST = &ast2.Schema{
	CommonTypes: ast2.CommonTypes{
		"Address": ast2.CommonType{
			Annotations: ast2.Annotations{
				"doc":                  "Address information",
				"personal_information": "",
			},
			Type: ast2.RecordType{
				"city": ast2.Attribute{
					Type: ast2.TypeRef("String"),
					Annotations: ast2.Annotations{
						"also": "town",
					},
				},
				"country": ast2.Attribute{Type: ast2.TypeRef("Country")},
				"street":  ast2.Attribute{Type: ast2.TypeRef("String")},
				"zipcode": ast2.Attribute{Type: ast2.TypeRef("String"), Optional: true},
			},
		},
		"decimal": ast2.CommonType{
			Type: ast2.RecordType{
				"decimal": ast2.Attribute{Type: ast2.TypeRef("Long")},
				"whole":   ast2.Attribute{Type: ast2.TypeRef("Long")},
			},
		},
	},
	Entities: ast2.Entities{
		"Admin":   ast2.Entity{},
		"Country": ast2.Entity{},
		"System": ast2.Entity{
			ParentTypes: []ast2.EntityTypeRef{"Admin"},
			Shape: &ast2.RecordType{
				"version": ast2.Attribute{Type: ast2.TypeRef("String")},
			},
		},
	},
	Enums: ast2.Enums{
		"Role": ast2.Enum{
			Values: []types.String{"superuser", "operator"},
		},
	},
	Actions: ast2.Actions{
		"audit": ast2.Action{
			AppliesTo: &ast2.AppliesTo{
				Principals: []ast2.EntityTypeRef{"Admin"},
				Resources:  []ast2.EntityTypeRef{"MyApp::Document", "System"},
			},
		},
	},
	Namespaces: ast2.Namespaces{
		"MyApp": ast2.Namespace{
			Annotations: ast2.Annotations{
				"doc": "Doc manager",
			},
			CommonTypes: ast2.CommonTypes{
				"Metadata": ast2.CommonType{
					Type: ast2.RecordType{
						"created": ast2.Attribute{Type: ast2.TypeRef("datetime")},
						"tags":    ast2.Attribute{Type: ast2.SetType{Element: ast2.TypeRef("String")}},
					},
				},
			},
			Entities: ast2.Entities{
				"Department": ast2.Entity{
					Shape: &ast2.RecordType{
						"budget": ast2.Attribute{Type: ast2.TypeRef("decimal")},
					},
				},
				"Document": ast2.Entity{
					Shape: &ast2.RecordType{
						"public": ast2.Attribute{Type: ast2.TypeRef("Bool")},
						"title":  ast2.Attribute{Type: ast2.TypeRef("String")},
					},
				},
				"Group": ast2.Entity{
					ParentTypes: []ast2.EntityTypeRef{"Department"},
					Shape: &ast2.RecordType{
						"metadata": ast2.Attribute{Type: ast2.TypeRef("Metadata")},
						"name":     ast2.Attribute{Type: ast2.TypeRef("String")},
					},
				},
				"User": ast2.Entity{
					ParentTypes: []ast2.EntityTypeRef{"Group"},
					Annotations: ast2.Annotations{
						"doc": "User entity",
					},
					Shape: &ast2.RecordType{
						"active":  ast2.Attribute{Type: ast2.TypeRef("Bool")},
						"address": ast2.Attribute{Type: ast2.TypeRef("Address")},
						"email":   ast2.Attribute{Type: ast2.TypeRef("String")},
						"level":   ast2.Attribute{Type: ast2.TypeRef("Long")},
					},
				},
			},
			Enums: ast2.Enums{
				"Status": ast2.Enum{
					Values: []types.String{"draft", "published", "archived"},
				},
			},
			Actions: ast2.Actions{
				"edit": ast2.Action{
					Annotations: ast2.Annotations{
						"doc": "View or edit document",
					},
					AppliesTo: &ast2.AppliesTo{
						Principals: []ast2.EntityTypeRef{"User"},
						Resources:  []ast2.EntityTypeRef{"Document"},
						Context: ast2.RecordType{
							"ip":        ast2.Attribute{Type: ast2.TypeRef("ipaddr")},
							"timestamp": ast2.Attribute{Type: ast2.TypeRef("datetime")},
						},
					},
				},
				"manage": ast2.Action{
					AppliesTo: &ast2.AppliesTo{
						Principals: []ast2.EntityTypeRef{"User"},
						Resources:  []ast2.EntityTypeRef{"Document", "Group"},
					},
				},
				"view": ast2.Action{
					Annotations: ast2.Annotations{
						"doc": "View or edit document",
					},
					AppliesTo: &ast2.AppliesTo{
						Principals: []ast2.EntityTypeRef{"User"},
						Resources:  []ast2.EntityTypeRef{"Document"},
						Context: ast2.RecordType{
							"ip":        ast2.Attribute{Type: ast2.TypeRef("ipaddr")},
							"timestamp": ast2.Attribute{Type: ast2.TypeRef("datetime")},
						},
					},
				},
			},
		},
	},
}

// wantResolved is the expected resolved schema structure.
// All type references have been fully qualified and common types inlined.
var wantResolved = &resolved2.Schema{
	Namespaces: map[types.Path]resolved2.Namespace{
		"MyApp": {
			Name: "MyApp",
			Annotations: resolved2.Annotations{
				"doc": "Doc manager",
			},
		},
	},
	Entities: map[types.EntityType]resolved2.Entity{
		"Admin": {
			Name: "Admin",
		},
		"Country": {
			Name: "Country",
		},
		"System": {
			Name:        "System",
			ParentTypes: []types.EntityType{"Admin"},
			Shape: resolved2.RecordType{
				"version": resolved2.Attribute{Type: resolved2.StringType{}},
			},
		},
		"MyApp::Department": {
			Name: "MyApp::Department",
			Shape: resolved2.RecordType{
				"budget": resolved2.Attribute{
					Type: resolved2.RecordType{
						"decimal": resolved2.Attribute{Type: resolved2.LongType{}},
						"whole":   resolved2.Attribute{Type: resolved2.LongType{}},
					},
				},
			},
		},
		"MyApp::Document": {
			Name: "MyApp::Document",
			Shape: resolved2.RecordType{
				"public": resolved2.Attribute{Type: resolved2.BoolType{}},
				"title":  resolved2.Attribute{Type: resolved2.StringType{}},
			},
		},
		"MyApp::Group": {
			Name:        "MyApp::Group",
			ParentTypes: []types.EntityType{"MyApp::Department"},
			Shape: resolved2.RecordType{
				"metadata": resolved2.Attribute{
					Type: resolved2.RecordType{
						"created": resolved2.Attribute{Type: resolved2.ExtensionType("datetime")},
						"tags":    resolved2.Attribute{Type: resolved2.SetType{Element: resolved2.StringType{}}},
					},
				},
				"name": resolved2.Attribute{Type: resolved2.StringType{}},
			},
		},
		"MyApp::User": {
			Name:        "MyApp::User",
			Annotations: resolved2.Annotations{"doc": "User entity"},
			ParentTypes: []types.EntityType{"MyApp::Group"},
			Shape: resolved2.RecordType{
				"active": resolved2.Attribute{Type: resolved2.BoolType{}},
				"address": resolved2.Attribute{
					Type: resolved2.RecordType{
						"city": resolved2.Attribute{
							Type:        resolved2.StringType{},
							Annotations: resolved2.Annotations{"also": "town"},
						},
						"country": resolved2.Attribute{Type: resolved2.EntityType("Country")},
						"street":  resolved2.Attribute{Type: resolved2.StringType{}},
						"zipcode": resolved2.Attribute{Type: resolved2.StringType{}, Optional: true},
					},
				},
				"email": resolved2.Attribute{Type: resolved2.StringType{}},
				"level": resolved2.Attribute{Type: resolved2.LongType{}},
			},
		},
	},
	Enums: map[types.EntityType]resolved2.Enum{
		"Role": {
			Name:   "Role",
			Values: []types.String{"superuser", "operator"},
		},
		"MyApp::Status": {
			Name:   "MyApp::Status",
			Values: []types.String{"draft", "published", "archived"},
		},
	},
	Actions: map[types.EntityUID]resolved2.Action{
		types.NewEntityUID("Action", "audit"): {
			Name: "audit",
			AppliesTo: &resolved2.AppliesTo{
				Principals: []types.EntityType{"Admin"},
				Resources:  []types.EntityType{"MyApp::Document", "System"},
				Context:    resolved2.RecordType{},
			},
		},
		types.NewEntityUID("MyApp::Action", "edit"): {
			Name:        "edit",
			Annotations: resolved2.Annotations{"doc": "View or edit document"},
			AppliesTo: &resolved2.AppliesTo{
				Principals: []types.EntityType{"MyApp::User"},
				Resources:  []types.EntityType{"MyApp::Document"},
				Context: resolved2.RecordType{
					"ip":        resolved2.Attribute{Type: resolved2.ExtensionType("ipaddr")},
					"timestamp": resolved2.Attribute{Type: resolved2.ExtensionType("datetime")},
				},
			},
		},
		types.NewEntityUID("MyApp::Action", "manage"): {
			Name: "manage",
			AppliesTo: &resolved2.AppliesTo{
				Principals: []types.EntityType{"MyApp::User"},
				Resources:  []types.EntityType{"MyApp::Document", "MyApp::Group"},
				Context:    resolved2.RecordType{},
			},
		},
		types.NewEntityUID("MyApp::Action", "view"): {
			Name:        "view",
			Annotations: resolved2.Annotations{"doc": "View or edit document"},
			AppliesTo: &resolved2.AppliesTo{
				Principals: []types.EntityType{"MyApp::User"},
				Resources:  []types.EntityType{"MyApp::Document"},
				Context: resolved2.RecordType{
					"ip":        resolved2.Attribute{Type: resolved2.ExtensionType("ipaddr")},
					"timestamp": resolved2.Attribute{Type: resolved2.ExtensionType("datetime")},
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

	t.Run("EmptySchema", func(t *testing.T) {
		t.Parallel()
		s := schema.NewSchemaFromAST(&ast2.Schema{})
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
