package validate

import "github.com/cedar-policy/cedar-go/types"

type extFuncSig struct {
	isConstructor bool
	argTypes      []cedarType
	returnType    cedarType
}

var extFuncTypes = map[types.Path]extFuncSig{
	// Constructors
	"ip":       {isConstructor: true, argTypes: []cedarType{typeString{}}, returnType: typeExtension{"ipaddr"}},
	"decimal":  {isConstructor: true, argTypes: []cedarType{typeString{}}, returnType: typeExtension{"decimal"}},
	"datetime": {isConstructor: true, argTypes: []cedarType{typeString{}}, returnType: typeExtension{"datetime"}},
	"duration": {isConstructor: true, argTypes: []cedarType{typeString{}}, returnType: typeExtension{"duration"}},

	// Decimal methods
	"lessThan":           {argTypes: []cedarType{typeExtension{"decimal"}, typeExtension{"decimal"}}, returnType: typeBool{}},
	"lessThanOrEqual":    {argTypes: []cedarType{typeExtension{"decimal"}, typeExtension{"decimal"}}, returnType: typeBool{}},
	"greaterThan":        {argTypes: []cedarType{typeExtension{"decimal"}, typeExtension{"decimal"}}, returnType: typeBool{}},
	"greaterThanOrEqual": {argTypes: []cedarType{typeExtension{"decimal"}, typeExtension{"decimal"}}, returnType: typeBool{}},

	// IPAddr methods
	"isIpv4":      {argTypes: []cedarType{typeExtension{"ipaddr"}}, returnType: typeBool{}},
	"isIpv6":      {argTypes: []cedarType{typeExtension{"ipaddr"}}, returnType: typeBool{}},
	"isLoopback":  {argTypes: []cedarType{typeExtension{"ipaddr"}}, returnType: typeBool{}},
	"isMulticast": {argTypes: []cedarType{typeExtension{"ipaddr"}}, returnType: typeBool{}},
	"isInRange":   {argTypes: []cedarType{typeExtension{"ipaddr"}, typeExtension{"ipaddr"}}, returnType: typeBool{}},

	// Datetime methods
	"toDate":        {argTypes: []cedarType{typeExtension{"datetime"}}, returnType: typeExtension{"datetime"}},
	"toTime":        {argTypes: []cedarType{typeExtension{"datetime"}}, returnType: typeExtension{"duration"}},
	"offset":        {argTypes: []cedarType{typeExtension{"datetime"}, typeExtension{"duration"}}, returnType: typeExtension{"datetime"}},
	"durationSince": {argTypes: []cedarType{typeExtension{"datetime"}, typeExtension{"datetime"}}, returnType: typeExtension{"duration"}},

	// Duration methods
	"toDays":         {argTypes: []cedarType{typeExtension{"duration"}}, returnType: typeLong{}},
	"toHours":        {argTypes: []cedarType{typeExtension{"duration"}}, returnType: typeLong{}},
	"toMinutes":      {argTypes: []cedarType{typeExtension{"duration"}}, returnType: typeLong{}},
	"toSeconds":      {argTypes: []cedarType{typeExtension{"duration"}}, returnType: typeLong{}},
	"toMilliseconds": {argTypes: []cedarType{typeExtension{"duration"}}, returnType: typeLong{}},
}
