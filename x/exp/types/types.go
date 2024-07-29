package types

import "net"

type Value interface {
	isValue()
}

type Boolean bool

func (Boolean) isValue() {}

type String string

func (String) isValue() {}

type Long int64

func (Long) isValue() {}

type Set []Value

func (Set) isValue() {}

type Record map[string]Value

func (Record) isValue() {}

type EntityType string

type EntityUID struct {
	Type string
	ID   string
}

func (EntityUID) isValue() {}

type Decimal []float64

func (Decimal) isValue() {}

type IpAddr net.IPAddr

func (IpAddr) isValue() {}
