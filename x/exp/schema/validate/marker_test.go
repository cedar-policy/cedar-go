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

	// Defensive type name/sort key paths for types that rarely appear in error messages
	cedarTypeSortKey(typeNever{})
	cedarTypeSortKey(typeBool{})
	cedarTypeName(typeNever{})
	cedarEntityTypeName(entityLUB{})

}
