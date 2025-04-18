package types

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/netip"
	"strings"

	"github.com/cedar-policy/cedar-go/internal"
)

var errIP = internal.ErrIP

// An IPAddr is value that represents an IP address. It can be either IPv4 or IPv6.
// The value can represent an individual address or a range of addresses.
type IPAddr netip.Prefix

// ParseIPAddr takes a string representation of an IP address and converts it into an IPAddr type.
func ParseIPAddr(s string) (IPAddr, error) {
	// We disallow IPv4-mapped IPv6 addresses in dotted notation because Cedar does.
	if strings.Count(s, ":") >= 2 && strings.Count(s, ".") >= 2 {
		return IPAddr{}, fmt.Errorf("%w: cannot parse IPv4 addresses embedded in IPv6 addresses", errIP)
	} else if net, err := netip.ParsePrefix(s); err == nil {
		return IPAddr(net), nil
	} else if addr, err := netip.ParseAddr(s); err == nil {
		return IPAddr(netip.PrefixFrom(addr, addr.BitLen())), nil
	}
	return IPAddr{}, fmt.Errorf("%w: error parsing IP address %s", errIP, s)
}

// Equal returns true if the argument is equal to a
func (i IPAddr) Equal(bi Value) bool {
	b, ok := bi.(IPAddr)
	return ok && i == b
}

// MarshalCedar produces a valid MarshalCedar language representation of the IPAddr, e.g. `ip("127.0.0.1")`.
func (i IPAddr) MarshalCedar() []byte { return []byte(`ip("` + i.String() + `")`) }

// String produces a string representation of the IPAddr, e.g. `127.0.0.1`.
func (i IPAddr) String() string {
	if i.Prefix().Bits() == i.Addr().BitLen() {
		return i.Addr().String()
	}
	return i.Prefix().String()
}

func (i IPAddr) Prefix() netip.Prefix {
	return netip.Prefix(i)
}

func (i IPAddr) IsIPv4() bool {
	return i.Addr().Is4()
}

func (i IPAddr) IsIPv6() bool {
	return i.Addr().Is6()
}

func (i IPAddr) IsLoopback() bool {
	// This comment is in the Cedar Rust implementation:
	//
	// 		Loopback addresses are "127.0.0.0/8" for IpV4 and "::1" for IpV6
	//
	// 		Unlike the implementation of `is_multicast`, we don't need to test prefix
	//
	// 		The reason for IpV6 is obvious: There's only one loopback address
	//
	// 		The reason for IpV4 is that provided the truncated ip address is a
	// 		loopback address, its prefix cannot be less than 8 because
	// 		otherwise its more significant byte cannot be 127
	return i.Prefix().Masked().Addr().IsLoopback()
}

func (i IPAddr) Addr() netip.Addr {
	return netip.Prefix(i).Addr()
}

func (i IPAddr) IsMulticast() bool {
	// This comment is in the Cedar Rust implementation:
	//
	// 		Multicast addresses are "224.0.0.0/4" for IpV4 and "ff00::/8" for
	// 		IpV6
	//
	// 		If an IpNet's addresses are multicast addresses, calling
	// 		`is_in_range()` over it and its associated net above should
	// 		evaluate to true
	//
	// 		The implementation uses the property that if `ip1/prefix1` is in
	// 		range `ip2/prefix2`, then `ip1` is in `ip2/prefix2` and `prefix1 >=
	// 		prefix2`
	var minPrefixLen int
	if i.IsIPv4() {
		minPrefixLen = 4
	} else {
		minPrefixLen = 8
	}
	return i.Addr().IsMulticast() && i.Prefix().Bits() >= minPrefixLen
}

func (i IPAddr) Contains(o IPAddr) bool {
	return i.Prefix().Contains(o.Addr()) && i.Prefix().Bits() <= o.Prefix().Bits()
}

// UnmarshalJSON implements encoding/json.Unmarshaler for IPAddr
//
// It is capable of unmarshaling 3 different representations supported by Cedar
//   - { "__extn": { "fn": "ip", "arg": "12.34.56.78" }}
//   - { "fn": "ip", "arg": "12.34.56.78" }
//   - "12.34.56.78"
func (i *IPAddr) UnmarshalJSON(b []byte) error {
	vv, err := unmarshalExtensionValue(b, "ip", ParseIPAddr)
	if err != nil {
		return err
	}

	*i = vv
	return nil
}

// MarshalJSON marshals the IPAddr into JSON using the explicit form.
func (i IPAddr) MarshalJSON() ([]byte, error) {
	if i.Prefix().Bits() == i.Prefix().Addr().BitLen() {
		return json.Marshal(extValueJSON{
			Extn: &extn{
				Fn:  "ip",
				Arg: i.Addr().String(),
			},
		})
	}
	return json.Marshal(extValueJSON{
		Extn: &extn{
			Fn:  "ip",
			Arg: i.String(),
		},
	})
}

func (i IPAddr) hash() uint64 {
	// MarshalBinary() cannot actually fail
	bytes, _ := netip.Prefix(i).MarshalBinary()
	h := fnv.New64()
	_, _ = h.Write(bytes)
	return h.Sum64()
}
