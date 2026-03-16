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
