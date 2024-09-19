package types

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestIpAddr(t *testing.T) {
	t.Parallel()

	t.Run("hash", func(t *testing.T) {
		t.Parallel()

		ipaddr1, err := ParseIPAddr("0.0.0.0")
		testutil.OK(t, err)

		ipaddr2, err := ParseIPAddr("0.0.0.0")
		testutil.OK(t, err)

		ipaddr3, err := ParseIPAddr("0.0.0.1")
		testutil.OK(t, err)

		ipaddr4, err := ParseIPAddr("0.0.0.1")
		testutil.OK(t, err)

		testutil.Equals(t, ipaddr1.hash(), ipaddr2.hash())
		testutil.Equals(t, ipaddr3.hash(), ipaddr4.hash())

		// This isn't necessarily true for all IPAddrs, but we want to make sure we're not just returning the same hash
		// value for all IPAddrs
		testutil.FatalIf(t, ipaddr1.hash() == ipaddr3.hash(), "unexpected hash collision")
	})
}
