package types_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestIP(t *testing.T) {
	t.Parallel()
	t.Run("ParseAndString", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			in     string
			parses bool
			out    string
		}{
			{"0.0.0.0", true, "0.0.0.0"},
			{"0.0.0.1", true, "0.0.0.1"},
			{"127.0.0.1", true, "127.0.0.1"},
			{"127.0.0.1/32", true, "127.0.0.1"},
			{"127.0.0.1/24", true, "127.0.0.1/24"},
			{"127.1.2.3/8", true, "127.1.2.3/8"},
			{"::/128", true, "::"},
			{"::1/128", true, "::1"},
			{"2001:db8::1", true, "2001:db8::1"},
			{"2001:db8::1:0:0:1", true, "2001:db8::1:0:0:1"},
			{"::ffff:192.0.2.128", false, ""},
			{"::ffff:c000:0280", true, "::ffff:192.0.2.128"},
			{"2001:db8::1/32", true, "2001:db8::1/32"},
			{"2001:db8::1:0:0:1/96", true, "2001:db8::1:0:0:1/96"},
			{"::ffff:192.0.2.128/24", false, ""},
			{"::ffff:192.0.2.128/120", false, ""},
			{"::ffff:c000:0280/24", true, "::ffff:192.0.2.128/24"},
			{"::ffff:c000:0280/120", true, "::ffff:192.0.2.128/120"},
			{"6b6b:f00::32ff:ffff:6368/00", false, ""}, // leading zero(s)
			{"garbage", false, ""},
			{"c5c5:c5c5:c5c5:c5c5:c5c5:c5c5:c5c5:c5c5/68", true, "c5c5:c5c5:c5c5:c5c5:c5c5:c5c5:c5c5:c5c5/68"},
		}
		for _, tt := range tests {
			tt := tt
			var testName string
			if tt.parses {
				testName = fmt.Sprintf("%s-parses-and-prints-as-%s", tt.in, tt.out)
			} else {
				testName = fmt.Sprintf("%s-does-not-parse", tt.in)
			}
			t.Run(testName, func(t *testing.T) {
				t.Parallel()
				i, err := types.ParseIPAddr(tt.in)
				if tt.parses {
					testutil.OK(t, err)
					testutil.Equals(t, i.String(), tt.out)
				} else {
					testutil.Error(t, err)
				}
			})
		}
	})

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			lhs, rhs string
			equal    bool
		}{
			{"0.0.0.0", "0.0.0.0", true},
			{"0.0.0.0", "0.0.0.0/32", true},
			{"127.0.0.1", "127.0.0.1", true},
			{"127.0.0.1", "127.0.0.1/32", true},
			{"::", "::", true},
			{"::", "::/128", true},
			{"::1", "::1", true},
			{"::1", "::1/128", true},
			{"::", "0.0.0.0", false},
			{"::1", "127.0.0.1", false},
			{"::ffff:c000:0280", "192.0.2.128", false},
			{"1.2.3.4", "1.2.3.4", true},
			{"1.2.3.4", "1.2.3.4/32", true},
			{"1.2.3.4/32", "1.2.3.4/32", true},
			{"1.2.3.4/24", "1.2.3.4/24", true},
			{"1.2.3.0/24", "1.2.3.255/24", false},
			{"1.2.3.0/24", "1.2.3.0/25", false},
			{"::ffff:c000:0280/24", "::/24", false},
			{"::ffff:c000:0280/120", "192.0.2.0/24", false},
			{"2001:db8::1/32", "2001:db8::/32", false},
			{"2001:db8::1:0:0:1/96", "2001:db8:0:0:1::/96", false},
			{"c5c5:c5c5:c5c5:c5c5:c5c5:c5c5:c5c5:c5c5/68", "c5c5:c5c5:c5c5:c5c5:c5c5:5cc5:c5c5:c5c5/68", false},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("ip(%v).Equal(ip(%v))", tt.lhs, tt.rhs), func(t *testing.T) {
				t.Parallel()
				lhs, err := types.ParseIPAddr(tt.lhs)
				testutil.OK(t, err)
				rhs, err := types.ParseIPAddr(tt.rhs)
				testutil.OK(t, err)
				equal := lhs.Equal(rhs)
				if equal != tt.equal {
					t.Fatalf("expected ip(%v).Equal(ip(%v)) to be %v instead of %v", tt.lhs, tt.rhs, tt.equal, equal)
				}
				if equal {
					testutil.FatalIf(
						t,
						!lhs.Contains(rhs),
						"ip(%v) and ip(%v) compare Equal but !ip(%v).contains(ip(%v))", tt.lhs, tt.rhs, tt.lhs, tt.rhs)
					testutil.FatalIf(
						t,
						!rhs.Contains(lhs),
						"ip(%v) and ip(%v) compare Equal but !ip(%v).contains(ip(%v))", tt.rhs, tt.lhs, tt.rhs, tt.lhs)
				}
			})
		}
	})

	t.Run("isIPv4", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			val            string
			isIPv4, isIPv6 bool
		}{
			{"0.0.0.0", true, false},
			{"0.0.0.0/32", true, false},
			{"127.0.0.1", true, false},
			{"127.0.0.1/32", true, false},
			{"::", false, true},
			{"::1", false, true},
			{"::/128", false, true},
			{"::1/128", false, true},
			{"::ffff:c000:0280", false, true},
			{"::ffff:c000:0280/128", false, true},
			{"::ffff:c000:0280/24", false, true},
			{"2001:db8::1", false, true},
			{"2001:db8::1:0:0:1", false, true},
			{"2001:db8::1/32", false, true},
			{"2001:db8::1:0:0:1/96", false, true},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("ip(%v).isIPv{4,6}()", tt.val), func(t *testing.T) {
				t.Parallel()
				val, err := types.ParseIPAddr(tt.val)
				testutil.OK(t, err)
				isIPv4 := val.IsIPv4()
				if isIPv4 != tt.isIPv4 {
					t.Fatalf("expected ip(%v).isIPv4() to be %v instead of %v", tt.val, tt.isIPv4, isIPv4)
				}
				isIPv6 := val.IsIPv6()
				if isIPv6 != tt.isIPv6 {
					t.Fatalf("expected ip(%v).isIPv6() to be %v instead of %v", tt.val, tt.isIPv6, isIPv6)
				}
			})
		}
	})

	t.Run("isLoopback", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			val        string
			isLoopback bool
		}{
			{"0.0.0.0", false},
			{"127.0.0.1", true},
			{"127.0.0.2", true},
			{"127.0.0.1/32", true},
			{"127.0.0.1/24", true},
			{"127.0.0.1/8", true},
			{"127.0.0.1/7", false},
			{"::", false},
			{"::1", true},
			{"::/128", false},
			{"::1/128", true},
			{"::1/127", false},
			{"::ffff:8000:0001", false},
			{"::ffff:8000:0002", false},
			{"::ffff:8000:0001/128", false},
			{"::ffff:8000:0002/128", false},
			{"::ffff:8000:0001/104", false},
			{"::ffff:8000:0002/104", false},
			{"::ffff:8000:0001/100", false},
			{"::ffff:8000:0002/100", false},
			{"2001:db8::1", false},
			{"2001:db8::1:0:0:1", false},
			{"2001:db8::1/32", false},
			{"2001:db8::1:0:0:1/96", false},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("ip(%v).isLoopback()", tt.val), func(t *testing.T) {
				t.Parallel()
				val, err := types.ParseIPAddr(tt.val)
				testutil.OK(t, err)
				isLoopback := val.IsLoopback()
				if isLoopback != tt.isLoopback {
					t.Fatalf("expected ip(%v).isLoopback() to be %v instead of %v", tt.val, tt.isLoopback, isLoopback)
				}
			})
		}
	})

	t.Run("isMulticast", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			val         string
			isMulticast bool
		}{
			{"0.0.0.0", false},
			{"127.0.0.1", false},
			{"223.255.255.255", false},
			{"224.0.0.0", true},
			{"239.255.255.255", true},
			{"240.0.0.0", false},
			{"feff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", false},
			{"ff00::", true},
			{"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", true},
			{"ff00::/8", true},
			{"ff00::/7", false},
			{"224.0.0.0/4", true},
			{"224.0.0.0/3", false},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("ip(%v).isMulticast()", tt.val), func(t *testing.T) {
				t.Parallel()
				val, err := types.ParseIPAddr(tt.val)
				testutil.OK(t, err)
				isMulticast := val.IsMulticast()
				if isMulticast != tt.isMulticast {
					t.Fatalf("expected ip(%v).isMulticast() to be %v instead of %v", tt.val, tt.isMulticast, isMulticast)
				}
			})
		}
	})

	t.Run("contains", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			lhs, rhs string
			contains bool
		}{
			{"0.0.0.0/31", "0.0.0.0", true},
			{"0.0.0.0", "0.0.0.0/31", false},
			{"255.255.0.0/16", "255.255.255.255", true},
			{"255.255.0.0/16", "255.255.255.248/28", true},
			{"255.255.0.0/16", "255.255.255.0/24", true},
			{"255.255.0.0/16", "255.255.248.0/20", true},
			{"255.255.0.0/16", "255.255.0.0/16", true},
			{"255.255.0.0/16", "255.254.0.0/15", false},
			{"255.255.0.0/16", "255.254.255.0/24", false},
			{"::ffff:c000:0280", "192.0.2.128", false},
			{"2001:db8::/120", "2001:db8::2", true},
			{"2001:db8::/64", "2001:db8:0:0:dead:f00d::/96", true},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("ip(%v).contains(ip(%v))", tt.lhs, tt.rhs), func(t *testing.T) {
				t.Parallel()
				lhs, err := types.ParseIPAddr(tt.lhs)
				testutil.OK(t, err)
				rhs, err := types.ParseIPAddr(tt.rhs)
				testutil.OK(t, err)
				contains := lhs.Contains(rhs)
				if contains != tt.contains {
					t.Fatalf("expected ip(%v).contains(ip(%v)) to be %v instead of %v", tt.lhs, tt.rhs, tt.contains, contains)
				}
			})
		}
	})
	t.Run("MarshalCedar", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(
			t,
			string(testutil.Must(types.ParseIPAddr("10.0.0.42")).MarshalCedar()),
			`ip("10.0.0.42")`)
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		t.Parallel()
		expected := `{
			"__extn": {
				"fn": "ip",
				"arg": "12.34.56.78"
			}
		}`
		i1 := testutil.Must(types.ParseIPAddr("12.34.56.78"))
		testutil.JSONMarshalsTo(t, i1, expected)

		var i2 types.IPAddr
		err := json.Unmarshal([]byte(expected), &i2)
		testutil.OK(t, err)
		testutil.Equals(t, i1, i2)
	})

	t.Run("UnmarshalJSON/error", func(t *testing.T) {
		t.Parallel()
		var dt2 types.IPAddr
		err := json.Unmarshal([]byte("{}"), &dt2)
		testutil.Error(t, err)
	})
}
