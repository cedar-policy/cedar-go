package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"net/netip"
	"strings"
)

// An IPAddr is value that represents an IP address. It can be either IPv4 or IPv6.
// The value can represent an individual address or a range of addresses.
type IPAddr netip.Prefix

// ParseIPAddr takes a string representation of an IP address and converts it into an IPAddr type.
func ParseIPAddr(s string) (IPAddr, error) {
	// We disallow IPv4-mapped IPv6 addresses in dotted notation because Cedar does.
	if strings.Count(s, ":") >= 2 && strings.Count(s, ".") >= 2 {
		return IPAddr{}, fmt.Errorf("%w: cannot parse IPv4 addresses embedded in IPv6 addresses", ErrIP)
	} else if net, err := netip.ParsePrefix(s); err == nil {
		return IPAddr(net), nil
	} else if addr, err := netip.ParseAddr(s); err == nil {
		return IPAddr(netip.PrefixFrom(addr, addr.BitLen())), nil
	} else {
		return IPAddr{}, fmt.Errorf("%w: error parsing IP address %s", ErrIP, s)
	}
}

func (a IPAddr) Equal(bi Value) bool {
	b, ok := bi.(IPAddr)
	return ok && a == b
}

// MarshalCedar produces a valid MarshalCedar language representation of the IPAddr, e.g. `ip("127.0.0.1")`.
func (v IPAddr) MarshalCedar() []byte { return []byte(`ip("` + v.String() + `")`) }

// String produces a string representation of the IPAddr, e.g. `127.0.0.1`.
func (v IPAddr) String() string {
	if v.Prefix().Bits() == v.Addr().BitLen() {
		return v.Addr().String()
	}
	return v.Prefix().String()
}

func (v IPAddr) Prefix() netip.Prefix {
	return netip.Prefix(v)
}

func (v IPAddr) IsIPv4() bool {
	return v.Addr().Is4()
}

func (v IPAddr) IsIPv6() bool {
	return v.Addr().Is6()
}

func (v IPAddr) IsLoopback() bool {
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
	return v.Prefix().Masked().Addr().IsLoopback()
}

func (v IPAddr) Addr() netip.Addr {
	return netip.Prefix(v).Addr()
}

func (v IPAddr) IsMulticast() bool {
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
	var min_prefix_len int
	if v.IsIPv4() {
		min_prefix_len = 4
	} else {
		min_prefix_len = 8
	}
	return v.Addr().IsMulticast() && v.Prefix().Bits() >= min_prefix_len
}

func (c IPAddr) Contains(o IPAddr) bool {
	return c.Prefix().Contains(o.Addr()) && c.Prefix().Bits() <= o.Prefix().Bits()
}

func (v *IPAddr) UnmarshalJSON(b []byte) error {
	var arg string
	if len(b) > 0 && b[0] == '"' {
		if err := json.Unmarshal(b, &arg); err != nil {
			return errors.Join(errJSONDecode, err)
		}
	} else {
		// NOTE: cedar supports two other forms, for now we're only supporting the smallest implicit explicit form.
		// The following are not supported:
		// "ip(\"192.168.0.42\")"
		// {"fn":"ip","arg":"192.168.0.42"}
		var res extValueJSON
		if err := json.Unmarshal(b, &res); err != nil {
			return errors.Join(errJSONDecode, err)
		}
		if res.Extn == nil {
			return errJSONExtNotFound
		}
		if res.Extn.Fn != "ip" {
			return errJSONExtFnMatch
		}
		arg = res.Extn.Arg
	}
	vv, err := ParseIPAddr(arg)
	if err != nil {
		return err
	}
	*v = vv
	return nil
}

// ExplicitMarshalJSON marshals the IPAddr into JSON using the implicit form.
func (v IPAddr) MarshalJSON() ([]byte, error) { return []byte(`"` + v.String() + `"`), nil }

// ExplicitMarshalJSON marshals the IPAddr into JSON using the explicit form.
func (v IPAddr) ExplicitMarshalJSON() ([]byte, error) {
	if v.Prefix().Bits() == v.Prefix().Addr().BitLen() {
		return json.Marshal(extValueJSON{
			Extn: &extn{
				Fn:  "ip",
				Arg: v.Addr().String(),
			},
		})
	}
	return json.Marshal(extValueJSON{
		Extn: &extn{
			Fn:  "ip",
			Arg: v.String(),
		},
	})
}

func (v IPAddr) hash() uint64 {
	// MarshalBinary() cannot actually fail
	bytes, _ := netip.Prefix(v).MarshalBinary()
	h := fnv.New64()
	_, _ = h.Write(bytes)
	return h.Sum64()
}
