package ast

import "github.com/cedar-policy/cedar-go/x/exp/types"

func Boolean(b types.Boolean) Node {
	return newValueNode(nodeTypeBoolean, b)
}

func True() Node {
	return Boolean(true)
}

func False() Node {
	return Boolean(false)
}

func String(s types.String) Node {
	return newValueNode(nodeTypeString, s)
}

func Long(l types.Long) Node {
	return newValueNode(nodeTypeLong, l)
}

func Set(s types.Set) Node {
	return newValueNode(nodeTypeSet, s)
}

func Record(r types.Record) Node {
	return newValueNode(nodeTypeMap, r)
}

func EntityType(e types.EntityType) Node {
	return newValueNode(nodeTypeEntityType, e)
}

func Entity(e types.EntityUID) Node {
	return newValueNode(nodeTypeEntity, e)
}

func Decimal(d types.Decimal) Node {
	return newValueNode(nodeTypeEntity, d)
}

func IpAddr(i types.IpAddr) Node {
	return newValueNode(nodeTypeIpAddr, i)
}

func newValueNode(op opType, v any) Node {
	return Node{op: op, value: v}
}
