package cedar

import "github.com/cedar-policy/cedar-go/x/exp/parser"

// ported from Go's stdlib and reduced to our scope.
// https://golang.org/src/path/filepath/match.go?s=1226:1284#L34

// Match reports whether name matches the shell file name pattern.
// The pattern syntax is:
//
//	pattern:
//		{ term }
//	term:
//		'*'         matches any sequence of non-Separator characters
//		c           matches character c (c != '*')
func match(p parser.Pattern, name string) (matched bool) {
Pattern:
	for i, comp := range p.Comps {
		lastChunk := i == len(p.Comps)-1
		if comp.Star && comp.Chunk == "" {
			return true
		}
		// Look for Match at current position.
		t, ok := matchChunk(comp.Chunk, name)
		// if we're the last chunk, make sure we've exhausted the name
		// otherwise we'll give a false result even if we could still Match
		// using the star
		if ok && (len(t) == 0 || !lastChunk) {
			name = t
			continue
		}
		if comp.Star {
			// Look for Match skipping i+1 bytes.
			for i := 0; i < len(name); i++ {
				t, ok := matchChunk(comp.Chunk, name[i+1:])
				if ok {
					// if we're the last chunk, make sure we exhausted the name
					if lastChunk && len(t) > 0 {
						continue
					}
					name = t
					continue Pattern
				}
			}
		}
		return false
	}
	return len(name) == 0
}

// matchChunk checks whether chunk matches the beginning of s.
// If so, it returns the remainder of s (after the Match).
// Chunk is all single-character operators: literals, char classes, and ?.
func matchChunk(chunk, s string) (rest string, ok bool) {
	for len(chunk) > 0 {
		if len(s) == 0 {
			return
		}
		if chunk[0] != s[0] {
			return
		}
		s = s[1:]
		chunk = chunk[1:]
	}
	return s, true
}
