// Package names implements Cedar's grammar for identifiers and names.
package names

import (
	"slices"
	"strings"
)

var reservedKeywords = []string{"true", "false", "if", "then", "else", "in", "like", "has", "is", "__cedar"}

// IsReservedKeyword reports whether s is a reserved Cedar keyword.
func IsReservedKeyword(s string) bool {
	return slices.Contains(reservedKeywords, s)
}

// IsIdentRune reports whether ch may appear in an identifier, either as its
// first rune or as a later one.
func IsIdentRune(ch rune, first bool) bool {
	return ch == '_' || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') && !first
}

// IsAnyIdent reports whether s is an identifier, including reserved keywords.
func IsAnyIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if !IsIdentRune(r, i == 0) {
			return false
		}
	}
	return true
}

// IsIdent reports whether s is an identifier that isn't a reserved keyword.
func IsIdent(s string) bool {
	return IsAnyIdent(s) && !IsReservedKeyword(s)
}

// IsName reports whether s is a sequence of one or more identifiers, none of
// them reserved keywords, separated by "::".
func IsName(s string) bool {
	for part := range strings.SplitSeq(s, "::") {
		if !IsIdent(part) {
			return false
		}
	}
	return true
}
