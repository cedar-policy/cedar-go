package extensions

import "github.com/cedar-policy/cedar-go/types"

type extInfo struct {
	Args     int
	IsMethod bool
}

var ExtMap = map[types.Path]extInfo{
	"ip":       {Args: 1, IsMethod: false},
	"decimal":  {Args: 1, IsMethod: false},
	"datetime": {Args: 1, IsMethod: false},
	"duration": {Args: 1, IsMethod: false},

	"lessThan":           {Args: 2, IsMethod: true},
	"lessThanOrEqual":    {Args: 2, IsMethod: true},
	"greaterThan":        {Args: 2, IsMethod: true},
	"greaterThanOrEqual": {Args: 2, IsMethod: true},

	"isIpv4":      {Args: 1, IsMethod: true},
	"isIpv6":      {Args: 1, IsMethod: true},
	"isLoopback":  {Args: 1, IsMethod: true},
	"isMulticast": {Args: 1, IsMethod: true},
	"isInRange":   {Args: 2, IsMethod: true},

	"toDate":        {Args: 1, IsMethod: true},
	"toTime":        {Args: 1, IsMethod: true},
	"offset":        {Args: 2, IsMethod: true},
	"durationSince": {Args: 2, IsMethod: true},

	"toDays":         {Args: 1, IsMethod: true},
	"toHours":        {Args: 1, IsMethod: true},
	"toMinutes":      {Args: 1, IsMethod: true},
	"toSeconds":      {Args: 1, IsMethod: true},
	"toMilliseconds": {Args: 1, IsMethod: true},
}

func init() {
	_ = 42
}
