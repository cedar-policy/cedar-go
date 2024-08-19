package types

import (
	"encoding/json"
	"strings"
)

// EntityType is the type portion of an EntityUID
type EntityType string

func (a EntityType) Equal(bi Value) bool {
	b, ok := bi.(EntityType)
	return ok && a == b
}

func (v EntityType) String() string                       { return string(v) }
func (v EntityType) Cedar() string                        { return string(v) }
func (v EntityType) ExplicitMarshalJSON() ([]byte, error) { return json.Marshal(string(v)) }
func (v EntityType) deepClone() Value                     { return v }

func EntityTypeFromSlice(v []string) EntityType {
	return EntityType(strings.Join(v, "::"))
}
