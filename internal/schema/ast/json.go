// Package ast defines the structure for a Cedar schema file.
//
// The schema is defined by the JSON format: https://docs.cedarpolicy.com/schema/json-schema.html
//
// The ast can be serialized 1-1 with the JSON format.
package ast

// JSONSchema represents the top-level Cedar schema structure
type JSONSchema map[string]*JSONNamespace

// JSONNamespace contains entity types, actions, and optional common types
type JSONNamespace struct {
	EntityTypes map[string]*JSONEntity     `json:"entityTypes"`
	Actions     map[string]*JSONAction     `json:"actions"`
	CommonTypes map[string]*JSONCommonType `json:"commonTypes,omitempty"`
	Annotations map[string]string          `json:"annotations,omitempty"`
}

// JSONEntity defines the structure and relationships of an entity
type JSONEntity struct {
	MemberOfTypes []string          `json:"memberOfTypes,omitempty"`
	Shape         *JSONType         `json:"shape,omitempty"`
	Tags          *JSONType         `json:"tags,omitempty"`
	Enum          []string          `json:"enum,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

// JSONAction defines what can perform an action and what it applies to
type JSONAction struct {
	MemberOf    []*JSONMember     `json:"memberOf,omitempty"`
	AppliesTo   *JSONAppliesTo    `json:"appliesTo"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type JSONMember struct {
	ID   string `json:"id"`
	Type string `json:"type,omitempty"`
}

// JSONAppliesTo defines what types can perform an action and what it applies to
type JSONAppliesTo struct {
	PrincipalTypes []string  `json:"principalTypes"`
	ResourceTypes  []string  `json:"resourceTypes"`
	Context        *JSONType `json:"context,omitempty"`
}

// JSONType represents the various type definitions possible in Cedar
type JSONType struct {
	Type        string                    `json:"type"`
	Element     *JSONType                 `json:"element,omitempty"`    // For Set types
	Name        string                    `json:"name,omitempty"`       // For Entity types
	Attributes  map[string]*JSONAttribute `json:"attributes,omitempty"` // For Record types
	Annotations map[string]string         `json:"annotations,omitempty"`
}

// JSONAttribute represents a field in a Record type
type JSONAttribute struct {
	Type        string                    `json:"type"`
	Required    bool                      `json:"required"`
	Element     *JSONType                 `json:"element,omitempty"`    // For Set types
	Name        string                    `json:"name,omitempty"`       // For Entity types
	Attributes  map[string]*JSONAttribute `json:"attributes,omitempty"` // For nested Record types
	Annotations map[string]string         `json:"annotations,omitempty"`
}

// JSONCommonType represents a reusable type definition
type JSONCommonType struct {
	*JSONType
}
