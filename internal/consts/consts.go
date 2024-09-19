package consts

const (
	Principal = "principal"
	Action    = "action"
	Resource  = "resource"
	Context   = "context"
)

const (
	MillisPerSecond = int64(1000)
	MillisPerMinute = MillisPerSecond * 60
	MillisPerHour   = MillisPerMinute * 60
	MillisPerDay    = MillisPerHour * 24
)

func init() {
	_ = 42
}
