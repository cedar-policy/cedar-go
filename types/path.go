package types

import "strings"

// Path is the type portion of an EntityUID
type Path string

func PathFromSlice(v []string) Path {
	return Path(strings.Join(v, "::"))
}
