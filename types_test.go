package cedar_test

import (
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestTypes(t *testing.T) {
	t.Parallel()
	testutil.Equals(t, cedar.NewDatetimeFromMillis(42), types.NewDatetimeFromMillis(42))
	testutil.Equals(t, cedar.NewDurationFromMillis(42), types.NewDurationFromMillis(42))
	ts := time.Now()
	testutil.Equals(t, cedar.NewDatetime(ts), types.NewDatetime(ts))
	testutil.Equals(t, cedar.NewDuration(time.Second), types.NewDuration(time.Second))
	testutil.Equals(t,
		cedar.NewPattern("test", cedar.Wildcard{}),
		types.NewPattern("test", types.Wildcard{}),
	)
	testutil.Equals(t,
		cedar.NewSet(cedar.Long(42), cedar.Long(43)),
		types.NewSet(types.Long(42), types.Long(43)),
	)
	testutil.Equals(t,
		testutil.Must(cedar.NewDecimal(42, 0)),
		testutil.Must(types.NewDecimal(42, 0)),
	)
	testutil.Equals(t,
		testutil.Must(cedar.NewDecimalFromInt(42)),
		testutil.Must(types.NewDecimalFromInt(42)),
	)
	testutil.Equals(t,
		testutil.Must(cedar.NewDecimalFromFloat(42.0)),
		testutil.Must(types.NewDecimalFromFloat(42.0)),
	)
}
