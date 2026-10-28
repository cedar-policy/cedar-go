package scanner

import "slices"

var reservedKeywords = []string{"true", "false", "if", "then", "else", "in", "like", "has", "is", "__cedar"}

// IsReservedKeyword reports whether s is a reserved Cedar keyword.
func IsReservedKeyword(s string) bool {
	return slices.Contains(reservedKeywords, s)
}
