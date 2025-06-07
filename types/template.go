package types

type SlotID string

const (
	PrincipalSlot SlotID = "?principal"
	ResourceSlot  SlotID = "?resource"
)

func (s SlotID) String() string {
	return string(s)
}

func (s SlotID) isEntityReference() {}
