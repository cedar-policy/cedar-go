package json

import (
	"encoding/json"

	"github.com/cedar-policy/cedar-go/types"
)

type policyJSON struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Effect      string            `json:"effect"`
	Principal   scopeJSON         `json:"principal"`
	Action      scopeJSON         `json:"action"`
	Resource    scopeJSON         `json:"resource"`
	Conditions  []conditionJSON   `json:"conditions,omitempty"`
}

// scopeInJSON uses the implicit form of EntityUID JSON serialization to match the Rust SDK
type scopeInJSON struct {
	Entity types.ImplicitlyMarshaledEntityUID `json:"entity"`
}

// scopeJSON uses the implicit form of EntityUID JSON serialization to match the Rust SDK
type scopeJSON struct {
	Op         string                               `json:"op"`
	Entity     *types.ImplicitlyMarshaledEntityUID  `json:"entity,omitempty"`
	Entities   []types.ImplicitlyMarshaledEntityUID `json:"entities,omitempty"`
	EntityType string                               `json:"entity_type,omitempty"`
	In         *scopeInJSON                         `json:"in,omitempty"`
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
	Attr string   `json:"attr"`
}

type likeJSON struct {
	Left    nodeJSON      `json:"left"`
	Pattern types.Pattern `json:"pattern"`
}

type isJSON struct {
	Left       nodeJSON  `json:"left"`
	EntityType string    `json:"entity_type"`
	In         *nodeJSON `json:"in,omitempty"`
}

type ifThenElseJSON struct {
	If   nodeJSON `json:"if"`
	Then nodeJSON `json:"then"`
	Else nodeJSON `json:"else"`
}

type arrayJSON []nodeJSON

type recordJSON map[string]nodeJSON

type extensionJSON map[string]arrayJSON

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
	Value *valueJSON `json:"Value,omitempty"` // could be any

	// Var
	Var *string `json:"Var,omitempty"`

	// Slot
	// Unknown

	// !, neg, and isEmpty operators
	Not     *unaryJSON `json:"!,omitempty"`
	Negate  *unaryJSON `json:"neg,omitempty"`
	IsEmpty *unaryJSON `json:"isEmpty,omitempty"`

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny, getTag, hasTag
	Equals             *binaryJSON `json:"==,omitempty"`
	NotEquals          *binaryJSON `json:"!=,omitempty"`
	In                 *binaryJSON `json:"in,omitempty"`
	LessThan           *binaryJSON `json:"<,omitempty"`
	LessThanOrEqual    *binaryJSON `json:"<=,omitempty"`
	GreaterThan        *binaryJSON `json:">,omitempty"`
	GreaterThanOrEqual *binaryJSON `json:">=,omitempty"`
	And                *binaryJSON `json:"&&,omitempty"`
	Or                 *binaryJSON `json:"||,omitempty"`
	Add                *binaryJSON `json:"+,omitempty"`
	Subtract           *binaryJSON `json:"-,omitempty"`
	Multiply           *binaryJSON `json:"*,omitempty"`
	Contains           *binaryJSON `json:"contains,omitempty"`
	ContainsAll        *binaryJSON `json:"containsAll,omitempty"`
	ContainsAny        *binaryJSON `json:"containsAny,omitempty"`
	GetTag             *binaryJSON `json:"getTag,omitempty"`
	HasTag             *binaryJSON `json:"hasTag,omitempty"`

	// ., has
	Access *strJSON `json:".,omitempty"`
	Has    *strJSON `json:"has,omitempty"`

	// is
	Is *isJSON `json:"is,omitempty"`

	// like
	Like *likeJSON `json:"like,omitempty"`

	// if-then-else
	IfThenElse *ifThenElseJSON `json:"if-then-else,omitempty"`

	// Set
	Set arrayJSON `json:"Set,omitempty"`

	// Record
	Record recordJSON `json:"Record,omitempty"`

	// Any other method: decimal, datetime, duration, ip, lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual, isIpv4, isIpv6, isLoopback, isMulticast, isInRange, toDate, toTime, toDays, toHours, toMinutes, toSeconds, toMilliseconds, offset, durationSince
	ExtensionCall extensionJSON `json:"-"`
}
