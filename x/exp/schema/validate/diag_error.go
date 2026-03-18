package validate

import "errors"

// mergeKeyedError carries a stable identity key used only for cross-environment
// diagnostic deduplication. The rendered Error() text remains unchanged.
type mergeKeyedError interface {
	MergeKey() string
}

type keyedDiagError struct {
	err error
	key string
}

func (e keyedDiagError) Error() string    { return e.err.Error() }
func (e keyedDiagError) Unwrap() error    { return e.err }
func (e keyedDiagError) MergeKey() string { return e.key }

func withMergeKey(err error, key string) error {
	return keyedDiagError{err: err, key: key}
}

func mergeIdentityKey(err error) (string, bool) {
	var mk mergeKeyedError
	if errors.As(err, &mk) {
		return mk.MergeKey(), true
	}
	return "", false
}
