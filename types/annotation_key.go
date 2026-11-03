package types

import (
	"errors"
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/scanner"
)

// AnnotationKey is the key of a policy annotation, e.g. in `@id("p1")`, `id`
// is the key. Create an AnnotationKey with NewAnnotationKey. Converting an
// arbitrary string directly to an AnnotationKey is unsupported: the result may
// marshal to Cedar text that means something other than intended.
type AnnotationKey string

// Ident is the former name of AnnotationKey.
//
// Deprecated: Use AnnotationKey.
type Ident = AnnotationKey

var errInvalidAnnotationKey = errors.New("invalid annotation key")

// NewAnnotationKey returns the AnnotationKey named by s, or an error if s isn't
// a single Cedar identifier. Unlike most identifiers in Cedar, annotation keys
// may be reserved keywords, such as `if` or `true`.
func NewAnnotationKey(s string) (AnnotationKey, error) {
	if !scanner.IsAnyIdent(s) {
		return "", fmt.Errorf("%w: %q", errInvalidAnnotationKey, s)
	}
	return AnnotationKey(s), nil
}
