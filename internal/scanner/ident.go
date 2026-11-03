package scanner

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
