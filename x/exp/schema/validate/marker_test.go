package validate

import (
	"testing"
)

func TestMarkerMethods(t *testing.T) {
	typeNever{}.isCedarType()
	typeTrue{}.isCedarType()
	typeFalse{}.isCedarType()
	typeBool{}.isCedarType()
	typeLong{}.isCedarType()
	typeString{}.isCedarType()
	typeSet{}.isCedarType()
	typeRecord{}.isCedarType()
	typeEntity{}.isCedarType()
	typeExtension{}.isCedarType()
}

func TestCedarTypeKindRankDefault(t *testing.T) {
	// nil cannot occur in normal type flow, but we keep a default for defensive
	// behavior and exercise it explicitly here.
	var nilType cedarType
	if got := cedarTypeKindRank(nilType); got != -1 {
		t.Fatalf("cedarTypeKindRank(nil) = %d, want -1", got)
	}
}

func TestCedarTypeNameDefault(t *testing.T) {
	// nil cannot occur in normal type flow, but we keep a default for defensive
	// behavior and exercise it explicitly here.
	var nilType cedarType
	if got := cedarTypeName(nilType); got != "?" {
		t.Fatalf("cedarTypeName(nil) = %q, want %q", got, "?")
	}
}
