package json

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

func TestMarshalIsTypeUnknown(t *testing.T) {
	_, err := marshalIsType(nil)
	testutil.Error(t, err)
}

func TestMarshalIsTypeSetError(t *testing.T) {
	_, err := marshalIsType(ast.SetType{Element: nil})
	testutil.Error(t, err)
}

func TestMarshalRecordTypeError(t *testing.T) {
	_, err := marshalRecordType(ast.RecordType{
		"bad": ast.Attribute{Type: nil},
	})
	testutil.Error(t, err)
}

func TestMarshalNamespaceCommonTypeError(t *testing.T) {
	_, err := marshalNamespace(ast.Namespace{
		CommonTypes: ast.CommonTypes{
			"Bad": ast.CommonType{Type: nil},
		},
	})
	testutil.Error(t, err)
}

func TestMarshalNamespaceEntityShapeError(t *testing.T) {
	_, err := marshalNamespace(ast.Namespace{
		Entities: ast.Entities{
			"Foo": ast.Entity{
				Shape: ast.RecordType{
					"bad": ast.Attribute{Type: nil},
				},
			},
		},
	})
	testutil.Error(t, err)
}

func TestMarshalNamespaceEntityTagsError(t *testing.T) {
	// Tags is nil, but the code checks `entity.Tags != nil` first
	// So we need a non-nil tags that fails. Use SetType{Element: nil}.
	_, err := marshalNamespace(ast.Namespace{
		Entities: ast.Entities{
			"Foo": ast.Entity{Tags: nil},
		},
	})
	testutil.OK(t, err)
}

func TestMarshalNamespaceEntityTagsError2(t *testing.T) {
	_, err := marshalNamespace(ast.Namespace{
		Entities: ast.Entities{
			"Foo": ast.Entity{Tags: ast.SetType{Element: nil}},
		},
	})
	testutil.Error(t, err)
}

func TestMarshalNamespaceActionAnnotations(t *testing.T) {
	ns, err := marshalNamespace(ast.Namespace{
		Actions: ast.Actions{
			"view": ast.Action{
				Annotations: ast.Annotations{"doc": "test"},
			},
		},
	})
	testutil.OK(t, err)
	testutil.Equals(t, ns.Actions["view"].Annotations["doc"], "test")
}

func TestMarshalNamespaceContextError(t *testing.T) {
	_, err := marshalNamespace(ast.Namespace{
		Actions: ast.Actions{
			"view": ast.Action{
				AppliesTo: &ast.AppliesTo{
					Context: ast.SetType{Element: nil},
				},
			},
		},
	})
	testutil.Error(t, err)
}

func TestMarshalBareNamespaceError(t *testing.T) {
	s := &Schema{
		Entities: ast.Entities{
			"Foo": ast.Entity{Tags: ast.SetType{Element: nil}},
		},
	}
	_, err := s.MarshalJSON()
	testutil.Error(t, err)
}

func TestMarshalNamespacedError(t *testing.T) {
	s := &Schema{
		Namespaces: ast.Namespaces{
			"NS": ast.Namespace{
				Entities: ast.Entities{
					"Foo": ast.Entity{Tags: ast.SetType{Element: nil}},
				},
			},
		},
	}
	_, err := s.MarshalJSON()
	testutil.Error(t, err)
}

func TestUnmarshalCommonTypeShorthand(t *testing.T) {
	// A non-builtin "type" string is parsed as an EntityOrCommon reference;
	// resolution of the referenced name happens later.
	ns, err := unmarshalNamespace(jsonNamespace{
		EntityTypes: map[string]jsonEntityType{},
		Actions:     map[string]jsonAction{},
		CommonTypes: map[string]jsonCommonType{
			"Alias": {jsonType: jsonType{Type: "PersonType"}},
		},
	})
	testutil.OK(t, err)
	testutil.Equals(t, ns.CommonTypes["Alias"].Type, ast.IsType(ast.TypeRef("PersonType")))
}

// A Set type without an "element" field is the only remaining parse-time
// error from unmarshalType. The next few tests exercise that error
// propagating up through each nesting level of unmarshalNamespace and its
// helpers.

func TestUnmarshalCommonTypeError(t *testing.T) {
	_, err := unmarshalNamespace(jsonNamespace{
		EntityTypes: map[string]jsonEntityType{},
		Actions:     map[string]jsonAction{},
		CommonTypes: map[string]jsonCommonType{
			"Bad": {jsonType: jsonType{Type: "Set"}},
		},
	})
	testutil.Error(t, err)
}

func TestUnmarshalEntityShapeShorthand(t *testing.T) {
	ns, err := unmarshalNamespace(jsonNamespace{
		EntityTypes: map[string]jsonEntityType{
			"Foo": {Shape: &jsonType{
				Type: "Record",
				Attributes: map[string]jsonAttr{
					"bar": {jsonType: jsonType{Type: "PersonType"}},
				},
			}},
		},
		Actions: map[string]jsonAction{},
	})
	testutil.OK(t, err)
	shape := ns.Entities["Foo"].Shape
	testutil.Equals(t, shape["bar"].Type, ast.IsType(ast.TypeRef("PersonType")))
}

func TestUnmarshalEntityShapeError(t *testing.T) {
	_, err := unmarshalNamespace(jsonNamespace{
		EntityTypes: map[string]jsonEntityType{
			"Foo": {Shape: &jsonType{
				Type: "Record",
				Attributes: map[string]jsonAttr{
					"bad": {jsonType: jsonType{Type: "Set"}},
				},
			}},
		},
		Actions: map[string]jsonAction{},
	})
	testutil.Error(t, err)
}

func TestUnmarshalActionAnnotations(t *testing.T) {
	ns, err := unmarshalNamespace(jsonNamespace{
		EntityTypes: map[string]jsonEntityType{},
		Actions: map[string]jsonAction{
			"view": {Annotations: map[string]string{"doc": "test"}},
		},
	})
	testutil.OK(t, err)
	testutil.Equals(t, ns.Actions["view"].Annotations["doc"], types.String("test"))
}

func TestUnmarshalContextTypeShorthand(t *testing.T) {
	ns, err := unmarshalNamespace(jsonNamespace{
		EntityTypes: map[string]jsonEntityType{},
		Actions: map[string]jsonAction{
			"view": {AppliesTo: &jsonAppliesTo{
				Context: &jsonType{Type: "ContextType"},
			}},
		},
	})
	testutil.OK(t, err)
	testutil.Equals(t, ns.Actions["view"].AppliesTo.Context, ast.IsType(ast.TypeRef("ContextType")))
}

func TestUnmarshalContextTypeError(t *testing.T) {
	_, err := unmarshalNamespace(jsonNamespace{
		EntityTypes: map[string]jsonEntityType{},
		Actions: map[string]jsonAction{
			"view": {AppliesTo: &jsonAppliesTo{
				Context: &jsonType{Type: "Set"},
			}},
		},
	})
	testutil.Error(t, err)
}

func TestUnmarshalSetElementShorthand(t *testing.T) {
	got, err := unmarshalType(&jsonType{
		Type:    "Set",
		Element: &jsonType{Type: "PersonType"},
	})
	testutil.OK(t, err)
	testutil.Equals(t, got, ast.IsType(ast.SetType{Element: ast.TypeRef("PersonType")}))
}

func TestUnmarshalSetElementError(t *testing.T) {
	_, err := unmarshalType(&jsonType{
		Type:    "Set",
		Element: &jsonType{Type: "Set"},
	})
	testutil.Error(t, err)
}

func TestUnmarshalRecordAttrShorthand(t *testing.T) {
	got, err := unmarshalRecordType(&jsonType{
		Type: "Record",
		Attributes: map[string]jsonAttr{
			"bar": {jsonType: jsonType{Type: "PersonType"}},
		},
	})
	testutil.OK(t, err)
	testutil.Equals(t, got["bar"].Type, ast.IsType(ast.TypeRef("PersonType")))
}

func TestUnmarshalRecordAttrError(t *testing.T) {
	_, err := unmarshalRecordType(&jsonType{
		Type: "Record",
		Attributes: map[string]jsonAttr{
			"bad": {jsonType: jsonType{Type: "Set"}},
		},
	})
	testutil.Error(t, err)
}
