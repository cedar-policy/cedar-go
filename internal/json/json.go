package json

import (
	"github.com/go-json-experiment/json"

	"github.com/cedar-policy/cedar-go/types"
)

type policyJSON struct {
	Annotations annotationsJSON `json:"annotations,omitempty"`
	Effect      string          `json:"effect"`
	Principal   scopeJSON       `json:"principal"`
	Action      scopeJSON       `json:"action"`
	Resource    scopeJSON       `json:"resource"`
	Conditions  conditionsJSON  `json:"conditions"`
}

// annotationsJSON, conditionsJSON, and scopeEntitiesJSON may be missing but,
// as in Rust Cedar, not null.
type annotationsJSON map[string]string

type conditionsJSON []conditionJSON

type scopeEntitiesJSON []types.ImplicitlyMarshaledEntityUID

// scopeInJSON uses the implicit form of EntityUID JSON serialization to match the Rust SDK
type scopeInJSON struct {
	Entity *types.ImplicitlyMarshaledEntityUID `json:"entity"`
}

// scopeJSON uses the implicit form of EntityUID JSON serialization to match the Rust SDK
type scopeJSON struct {
	Op         string                              `json:"op"`
	Entity     *types.ImplicitlyMarshaledEntityUID `json:"entity,omitzero"`
	Entities   scopeEntitiesJSON                   `json:"entities,omitzero"`
	EntityType string                              `json:"entity_type,omitempty"`
	In         *scopeInJSON                        `json:"in,omitzero"`
}

type conditionJSON struct {
	Kind string   `json:"kind"`
	Body nodeJSON `json:"body"`
}

type binaryJSON struct {
	Left  nodeJSON `json:"left"`
	Right nodeJSON `json:"right"`
}

type unaryJSON struct {
	Arg nodeJSON `json:"arg"`
}

type strJSON struct {
	Left nodeJSON `json:"left"`
	Attr *string  `json:"attr"`
}

type likeJSON struct {
	Left    nodeJSON       `json:"left"`
	Pattern *types.Pattern `json:"pattern"`
}

type isJSON struct {
	Left       nodeJSON  `json:"left"`
	EntityType string    `json:"entity_type"`
	In         *nodeJSON `json:"in,omitzero"`
}

type ifThenElseJSON struct {
	If   nodeJSON `json:"if"`
	Then nodeJSON `json:"then"`
	Else nodeJSON `json:"else"`
}

type arrayJSON []nodeJSON

type recordJSON map[string]nodeJSON

type extensionJSON struct {
	Name string
	Args arrayJSON
}

// MarshalJSON encodes the extension call as a single-member object whose name
// is the extension function or method name.
func (e *extensionJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]arrayJSON{e.Name: e.Args})
}

type valueJSON struct {
	v types.Value
}

func (e *valueJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.v)
}

func (e *valueJSON) UnmarshalJSON(b []byte) error {
	return types.UnmarshalJSON(b, &e.v)
}

type nodeJSON struct {
	// Value
	Value *valueJSON `json:"Value,omitzero"` // could be any

	// Var
	Var *string `json:"Var,omitzero"`

	// Slot
	// Unknown

	// !, neg, and isEmpty operators
	Not     *unaryJSON `json:"!,omitzero"`
	Negate  *unaryJSON `json:"neg,omitzero"`
	IsEmpty *unaryJSON `json:"isEmpty,omitzero"`

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny, getTag, hasTag
	Equals             *binaryJSON `json:"==,omitzero"`
	NotEquals          *binaryJSON `json:"!=,omitzero"`
	In                 *binaryJSON `json:"in,omitzero"`
	LessThan           *binaryJSON `json:"<,omitzero"`
	LessThanOrEqual    *binaryJSON `json:"<=,omitzero"`
	GreaterThan        *binaryJSON `json:">,omitzero"`
	GreaterThanOrEqual *binaryJSON `json:">=,omitzero"`
	And                *binaryJSON `json:"&&,omitzero"`
	Or                 *binaryJSON `json:"||,omitzero"`
	Add                *binaryJSON `json:"+,omitzero"`
	Subtract           *binaryJSON `json:"-,omitzero"` //nolint:staticcheck // "-,omitempty" means field name is "-" with omitempty, not skip
	Multiply           *binaryJSON `json:"*,omitzero"`
	Contains           *binaryJSON `json:"contains,omitzero"`
	ContainsAll        *binaryJSON `json:"containsAll,omitzero"`
	ContainsAny        *binaryJSON `json:"containsAny,omitzero"`
	GetTag             *binaryJSON `json:"getTag,omitzero"`
	HasTag             *binaryJSON `json:"hasTag,omitzero"`

	// ., has
	Access *strJSON `json:".,omitzero"`
	Has    *strJSON `json:"has,omitzero"`

	// is
	Is *isJSON `json:"is,omitzero"`

	// like
	Like *likeJSON `json:"like,omitzero"`

	// if-then-else
	IfThenElse *ifThenElseJSON `json:"if-then-else,omitzero"`

	// Set
	Set *arrayJSON `json:"Set,omitzero"`

	// Record
	Record *recordJSON `json:"Record,omitzero"`

	// Any other method: decimal, datetime, duration, ip, lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual, isIpv4, isIpv6, isLoopback, isMulticast, isInRange, toDate, toTime, toDays, toHours, toMinutes, toSeconds, toMilliseconds, offset, durationSince
	ExtensionCall *extensionJSON `json:"-"`
}
