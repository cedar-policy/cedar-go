package ast

import "github.com/cedar-policy/cedar-go/types"

type extInfo struct {
	Args     int
	IsMethod bool
}

var ExtMap = map[types.String]extInfo{
	"ip":      {Args: 1, IsMethod: false},
	"decimal": {Args: 1, IsMethod: false},

	"lessThan":           {Args: 2, IsMethod: true},
	"lessThanOrEqual":    {Args: 2, IsMethod: true},
	"greaterThan":        {Args: 2, IsMethod: true},
	"greaterThanOrEqual": {Args: 2, IsMethod: true},

	"isIpv4":      {Args: 1, IsMethod: true},
	"isIpv6":      {Args: 1, IsMethod: true},
	"isLoopback":  {Args: 1, IsMethod: true},
	"isMulticast": {Args: 1, IsMethod: true},
	"isInRange":   {Args: 2, IsMethod: true},
}
