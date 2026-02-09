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
	_, err := marshalNamespace("", ast.Namespace{
		CommonTypes: ast.CommonTypes{
			"Bad": ast.CommonType{Type: nil},
		},
	})
	testutil.Error(t, err)
}

func TestMarshalNamespaceEntityShapeError(t *testing.T) {
	_, err := marshalNamespace("", ast.Namespace{
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
	_, err := marshalNamespace("", ast.Namespace{
		Entities: ast.Entities{
			"Foo": ast.Entity{Tags: nil},
		},
	})
	testutil.OK(t, err)
}

func TestMarshalNamespaceEntityTagsError2(t *testing.T) {
	_, err := marshalNamespace("", ast.Namespace{
		Entities: ast.Entities{
			"Foo": ast.Entity{Tags: ast.SetType{Element: nil}},
		},
	})
	testutil.Error(t, err)
}

func TestMarshalNamespaceActionAnnotations(t *testing.T) {
	ns, err := marshalNamespace("", ast.Namespace{
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
	_, err := marshalNamespace("", ast.Namespace{
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

func TestUnmarshalCommonTypeError(t *testing.T) {
	_, err := unmarshalNamespace(jsonNamespace{
		EntityTypes: map[string]jsonEntityType{},
		Actions:     map[string]jsonAction{},
		CommonTypes: map[string]jsonCommonType{
			"Bad": {jsonType: jsonType{Type: "Unknown"}},
		},
	})
	testutil.Error(t, err)
}

func TestUnmarshalEntityShapeError(t *testing.T) {
	_, err := unmarshalNamespace(jsonNamespace{
		EntityTypes: map[string]jsonEntityType{
			"Foo": {Shape: &jsonType{
				Type: "Record",
				Attributes: map[string]jsonAttr{
					"bad": {jsonType: jsonType{Type: "Unknown"}},
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

func TestUnmarshalContextTypeError(t *testing.T) {
	_, err := unmarshalNamespace(jsonNamespace{
		EntityTypes: map[string]jsonEntityType{},
		Actions: map[string]jsonAction{
			"view": {AppliesTo: &jsonAppliesTo{
				Context: &jsonType{Type: "Unknown"},
			}},
		},
	})
	testutil.Error(t, err)
}

func TestUnmarshalSetElementError(t *testing.T) {
	_, err := unmarshalType(&jsonType{
		Type:    "Set",
		Element: &jsonType{Type: "Unknown"},
	})
	testutil.Error(t, err)
}

func TestUnmarshalRecordAttrError(t *testing.T) {
	_, err := unmarshalRecordType(&jsonType{
		Type: "Record",
		Attributes: map[string]jsonAttr{
			"bad": {jsonType: jsonType{Type: "Unknown"}},
		},
	})
	testutil.Error(t, err)
}
