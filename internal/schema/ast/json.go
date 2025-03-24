// Package ast defines the structure for a Cedar schema file.
//
// The schema is defined by the JSON format: https://docs.cedarpolicy.com/schema/json-schema.html
//
// The ast can be serialized 1-1 with the JSON format.
package ast

// JsonSchema represents the top-level Cedar schema structure
type JsonSchema map[string]*JsonNamespace

// JsonNamespace contains entity types, actions, and optional common types
type JsonNamespace struct {
	EntityTypes map[string]*JsonEntity     `json:"entityTypes"`
	Actions     map[string]*JsonAction     `json:"actions"`
	CommonTypes map[string]*JsonCommonType `json:"commonTypes,omitempty"`
}

// JsonEntity defines the structure and relationships of an entity
type JsonEntity struct {
	MemberOfTypes []string  `json:"memberOfTypes,omitempty"`
	Shape         *JsonType `json:"shape,omitempty"`
	Tags          *JsonType `json:"tags,omitempty"`
}

// JsonAction defines what can perform an action and what it applies to
type JsonAction struct {
	MemberOf  []*JsonMember  `json:"memberOf,omitempty"`
	AppliesTo *JsonAppliesTo `json:"appliesTo"`
}

type JsonMember struct {
	ID   string `json:"id"`
	Type string `json:"type,omitempty"`
}

// JsonAppliesTo defines what types can perform an action and what it applies to
type JsonAppliesTo struct {
	PrincipalTypes []string  `json:"principalTypes"`
	ResourceTypes  []string  `json:"resourceTypes"`
	Context        *JsonType `json:"context,omitempty"`
}

// JsonType represents the various type definitions possible in Cedar
type JsonType struct {
	Type       string                    `json:"type"`
	Element    *JsonType                 `json:"element,omitempty"`    // For Set types
	Name       string                    `json:"name,omitempty"`       // For Entity types
	Attributes map[string]*JsonAttribute `json:"attributes,omitempty"` // For Record types
}

// JsonAttribute represents a field in a Record type
type JsonAttribute struct {
	Type       string                    `json:"type"`
	Required   bool                      `json:"required"`
	Element    *JsonType                 `json:"element,omitempty"`    // For Set types
	Name       string                    `json:"name,omitempty"`       // For Entity types
	Attributes map[string]*JsonAttribute `json:"attributes,omitempty"` // For nested Record types
}

// JsonCommonType represents a reusable type definition
type JsonCommonType struct {
	*JsonType
}
