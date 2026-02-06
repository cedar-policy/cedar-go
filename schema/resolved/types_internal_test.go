package resolved

import "testing"

func TestIsTypeMarkers(t *testing.T) {
	StringType{}.isType()
	LongType{}.isType()
	BoolType{}.isType()
	ExtensionType("ipaddr").isType()
	SetType{}.isType()
	RecordType{}.isType()
	EntityType("User").isType()
}
