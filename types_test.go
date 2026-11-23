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
		testutil.Must(cedar.NewPath("Foo::Bar")),
		testutil.Must(types.NewPath("Foo::Bar")),
	)
	testutil.Equals(t,
		testutil.Must(cedar.NewEntityType("Foo::Bar")),
		testutil.Must(types.NewEntityType("Foo::Bar")),
	)
	testutil.Equals(t,
		testutil.Must(cedar.NewAnnotationKey("if")),
		testutil.Must(types.NewAnnotationKey("if")),
	)
	testutil.Equals(t,
		testutil.Must(cedar.ParseDatetime("1970-01-01")),
		testutil.Must(types.ParseDatetime("1970-01-01")),
	)
	testutil.Equals(t,
		testutil.Must(cedar.ParseDecimal("1.5")),
		testutil.Must(types.ParseDecimal("1.5")),
	)
	testutil.Equals(t,
		testutil.Must(cedar.ParseDuration("1h30m")),
		testutil.Must(types.ParseDuration("1h30m")),
	)
	testutil.Equals(t,
		testutil.Must(cedar.ParseIPAddr("10.0.0.0/8")),
		testutil.Must(types.ParseIPAddr("10.0.0.0/8")),
	)
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
