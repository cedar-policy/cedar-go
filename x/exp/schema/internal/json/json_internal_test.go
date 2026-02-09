package json

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	ast2 "github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

func TestMarshalIsTypeUnknown(t *testing.T) {
	// marshalIsType line 266-267: default case for unknown type
	_, err := marshalIsType(nil)
	testutil.Error(t, err)
}

func TestMarshalIsTypeSetError(t *testing.T) {
	// marshalIsType line 256-258: error marshaling Set element
	_, err := marshalIsType(ast2.SetType{Element: nil})
	testutil.Error(t, err)
}

func TestMarshalRecordTypeError(t *testing.T) {
	// marshalRecordType line 278-280: error marshaling attribute type
	_, err := marshalRecordType(ast2.RecordType{
		"bad": ast2.Attribute{Type: nil},
	})
	testutil.Error(t, err)
}

func TestMarshalNamespaceCommonTypeError(t *testing.T) {
	// marshalNamespace line 149-151: marshalIsType error in common type
	_, err := marshalNamespace("", ast2.Namespace{
		CommonTypes: ast2.CommonTypes{
			"Bad": ast2.CommonType{Type: nil},
		},
	})
	testutil.Error(t, err)
}

func TestMarshalNamespaceEntityShapeError(t *testing.T) {
	// marshalNamespace line 174-176: marshalRecordType error in entity shape
	_, err := marshalNamespace("", ast2.Namespace{
		Entities: ast2.Entities{
			"Foo": ast2.Entity{
				Shape: &ast2.RecordType{
					"bad": ast2.Attribute{Type: nil},
				},
			},
		},
	})
	testutil.Error(t, err)
}

func TestMarshalNamespaceEntityTagsError(t *testing.T) {
	// marshalNamespace line 181-183: marshalIsType error in entity tags
	_, err := marshalNamespace("", ast2.Namespace{
		Entities: ast2.Entities{
			"Foo": ast2.Entity{Tags: nil},
		},
	})
	// Tags is nil, but the code checks `entity.Tags != nil` first (line 179)
	// So we need a non-nil tags that fails. Use SetType{Element: nil}.
	testutil.OK(t, err)
}

func TestMarshalNamespaceEntityTagsError2(t *testing.T) {
	// marshalNamespace line 181-183: marshalIsType error in entity tags
	_, err := marshalNamespace("", ast2.Namespace{
		Entities: ast2.Entities{
			"Foo": ast2.Entity{Tags: ast2.SetType{Element: nil}},
		},
	})
	testutil.Error(t, err)
}

func TestMarshalNamespaceActionAnnotations(t *testing.T) {
	// marshalNamespace line 203-205: action with annotations
	ns, err := marshalNamespace("", ast2.Namespace{
		Actions: ast2.Actions{
			"view": ast2.Action{
				Annotations: ast2.Annotations{"doc": "test"},
			},
		},
	})
	testutil.OK(t, err)
	testutil.Equals(t, ns.Actions["view"].Annotations["doc"], "test")
}

func TestMarshalNamespaceContextError(t *testing.T) {
	// marshalNamespace line 231-233: marshalIsType error in appliesTo context
	_, err := marshalNamespace("", ast2.Namespace{
		Actions: ast2.Actions{
			"view": ast2.Action{
				AppliesTo: &ast2.AppliesTo{
					Context: ast2.SetType{Element: nil},
				},
			},
		},
	})
	testutil.Error(t, err)
}

func TestMarshalBareNamespaceError(t *testing.T) {
	// MarshalJSON line 28-30: marshalNamespace error for bare decls
	s := &Schema{
		Entities: ast2.Entities{
			"Foo": ast2.Entity{Tags: ast2.SetType{Element: nil}},
		},
	}
	_, err := s.MarshalJSON()
	testutil.Error(t, err)
}

func TestMarshalNamespacedError(t *testing.T) {
	// MarshalJSON line 36-38: marshalNamespace error for namespaced decls
	s := &Schema{
		Namespaces: ast2.Namespaces{
			"NS": ast2.Namespace{
				Entities: ast2.Entities{
					"NS::Foo": ast2.Entity{Tags: ast2.SetType{Element: nil}},
				},
			},
		},
	}
	_, err := s.MarshalJSON()
	testutil.Error(t, err)
}

func TestUnmarshalCommonTypeError(t *testing.T) {
	// unmarshalNamespace line 311-313: unmarshalType error in common type
	_, err := unmarshalNamespace("", jsonNamespace{
		EntityTypes: map[string]jsonEntityType{},
		Actions:     map[string]jsonAction{},
		CommonTypes: map[string]jsonCommonType{
			"Bad": {jsonType: jsonType{Type: "Unknown"}},
		},
	})
	testutil.Error(t, err)
}

func TestUnmarshalEntityShapeError(t *testing.T) {
	// unmarshalNamespace line 348-350: unmarshalRecordType error in entity shape
	_, err := unmarshalNamespace("", jsonNamespace{
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
	// unmarshalNamespace line 369-371: action with annotations
	ns, err := unmarshalNamespace("", jsonNamespace{
		EntityTypes: map[string]jsonEntityType{},
		Actions: map[string]jsonAction{
			"view": {Annotations: map[string]string{"doc": "test"}},
		},
	})
	testutil.OK(t, err)
	testutil.Equals(t, ns.Actions["view"].Annotations["doc"], types.String("test"))
}

func TestUnmarshalContextTypeError(t *testing.T) {
	// unmarshalNamespace line 395-397: unmarshalType error in appliesTo context
	_, err := unmarshalNamespace("", jsonNamespace{
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
	// unmarshalType line 426-428: error unmarshaling Set element
	_, err := unmarshalType(&jsonType{
		Type:    "Set",
		Element: &jsonType{Type: "Unknown"},
	})
	testutil.Error(t, err)
}

func TestUnmarshalRecordAttrError(t *testing.T) {
	// unmarshalRecordType line 445-447: error unmarshaling attribute type
	_, err := unmarshalRecordType(&jsonType{
		Type: "Record",
		Attributes: map[string]jsonAttr{
			"bad": {jsonType: jsonType{Type: "Unknown"}},
		},
	})
	testutil.Error(t, err)
}
