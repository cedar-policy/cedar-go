package ast

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

type inJSON struct {
	Entity types.EntityUID `json:"entity"`
}

type scopeJSON struct {
	Op         string            `json:"op"`
	Entity     *types.EntityUID  `json:"entity,omitempty"`
	Entities   []types.EntityUID `json:"entities,omitempty"`
	EntityType string            `json:"entity_type,omitempty"`
	In         *inJSON           `json:"in,omitempty"`
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

type ifThenElseJSON struct {
	If   nodeJSON `json:"if"`
	Then nodeJSON `json:"then"`
	Else nodeJSON `json:"else"`
}

type arrayJSON []nodeJSON

type recordJSON map[string]nodeJSON

type nodeJSON struct {

	// Value
	Value *json.RawMessage `json:"Value"` // could be any

	// Var
	Var *string `json:"Var"`

	// Slot
	// Unknown

	// ! or neg operators
	Not    *unaryJSON `json:"!"`
	Negate *unaryJSON `json:"neg"`

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny
	Equals             *binaryJSON `json:"=="`
	NotEquals          *binaryJSON `json:"!="`
	In                 *binaryJSON `json:"in"`
	LessThan           *binaryJSON `json:"<"`
	LessThanOrEqual    *binaryJSON `json:"<="`
	GreaterThan        *binaryJSON `json:">"`
	GreaterThanOrEqual *binaryJSON `json:">="`
	And                *binaryJSON `json:"&&"`
	Or                 *binaryJSON `json:"||"`
	Plus               *binaryJSON `json:"+"`
	Minus              *binaryJSON `json:"-"`
	Times              *binaryJSON `json:"*"`
	Contains           *binaryJSON `json:"contains"`
	ContainsAll        *binaryJSON `json:"containsAll"`
	ContainsAny        *binaryJSON `json:"containsAny"`

	// ., has
	Access *strJSON `json:"."`
	Has    *strJSON `json:"has"`

	// like
	Like *strJSON `json:"like"`

	// if-then-else
	IfThenElse *ifThenElseJSON `json:"if-then-else"`

	// Set
	Set arrayJSON `json:"Set"`

	// Record
	Record recordJSON `json:"Record"`

	// Any other function: decimal, ip
	Decimal arrayJSON `json:"decimal"`
	IP      arrayJSON `json:"ip"`

	// Any other method: lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual, isIpv4, isIpv6, isLoopback, isMulticast, isInRange
	LessThanExt           arrayJSON `json:"lessThan"`
	LessThanOrEqualExt    arrayJSON `json:"lessThanOrEqual"`
	GreaterThanExt        arrayJSON `json:"greaterThan"`
	GreaterThanOrEqualExt arrayJSON `json:"greaterThanOrEqual"`
	IsIpv4Ext             arrayJSON `json:"isIpv4"`
	IsIpv6Ext             arrayJSON `json:"isIpv6"`
	IsLoopbackExt         arrayJSON `json:"isLoopback"`
	IsMulticastExt        arrayJSON `json:"isMulticast"`
	IsInRangeExt          arrayJSON `json:"isInRange"`
}
