package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var ErrDecimal = fmt.Errorf("error parsing decimal value")
var ErrIP = fmt.Errorf("error parsing ip value")
var ErrType = fmt.Errorf("type error")

type Value interface {
	// String produces a string representation of the Value.
	String() string
	// Cedar produces a valid Cedar language representation of the Value.
	Cedar() string
	// ExplicitMarshalJSON marshals the Value into JSON using the explicit (if
	// applicable) JSON form, which is necessary for marshalling values within
	// Sets or Records where the type is not defined.
	ExplicitMarshalJSON() ([]byte, error)
	Equal(Value) bool
	TypeName() string
	deepClone() Value
}

func ZeroValue() Value {
	return nil
}

// A Boolean is a value that is either true or false.
type Boolean bool

const (
	True  = Boolean(true)
	False = Boolean(false)
)

func (a Boolean) Equal(bi Value) bool {
	b, ok := bi.(Boolean)
	return ok && a == b
}
func (v Boolean) TypeName() string { return "bool" }

// String produces a string representation of the Boolean, e.g. `true`.
func (v Boolean) String() string { return v.Cedar() }

// Cedar produces a valid Cedar language representation of the Boolean, e.g. `true`.
func (v Boolean) Cedar() string {
	return fmt.Sprint(bool(v))
}

// ExplicitMarshalJSON marshals the Boolean into JSON.
func (v Boolean) ExplicitMarshalJSON() ([]byte, error) { return json.Marshal(v) }
func (v Boolean) deepClone() Value                     { return v }

func ValueToBool(v Value) (Boolean, error) {
	bv, ok := v.(Boolean)
	if !ok {
		return false, fmt.Errorf("%w: expected bool, got %v", ErrType, v.TypeName())
	}
	return bv, nil
}

// A Long is a whole number without decimals that can range from -9223372036854775808 to 9223372036854775807.
type Long int64

func (a Long) Equal(bi Value) bool {
	b, ok := bi.(Long)
	return ok && a == b
}

// ExplicitMarshalJSON marshals the Long into JSON.
func (v Long) ExplicitMarshalJSON() ([]byte, error) { return json.Marshal(v) }
func (v Long) TypeName() string                     { return "long" }

// String produces a string representation of the Long, e.g. `42`.
func (v Long) String() string { return v.Cedar() }

// Cedar produces a valid Cedar language representation of the Long, e.g. `42`.
func (v Long) Cedar() string {
	return fmt.Sprint(int64(v))
}
func (v Long) deepClone() Value { return v }

func ValueToLong(v Value) (Long, error) {
	lv, ok := v.(Long)
	if !ok {
		return 0, fmt.Errorf("%w: expected long, got %v", ErrType, v.TypeName())
	}
	return lv, nil
}

// A String is a sequence of characters consisting of letters, numbers, or symbols.
type String string

func (a String) Equal(bi Value) bool {
	b, ok := bi.(String)
	return ok && a == b
}

// ExplicitMarshalJSON marshals the String into JSON.
func (v String) ExplicitMarshalJSON() ([]byte, error) { return json.Marshal(v) }
func (v String) TypeName() string                     { return "string" }

// String produces an unquoted string representation of the String, e.g. `hello`.
func (v String) String() string {
	return string(v)
}

// Cedar produces a valid Cedar language representation of the String, e.g. `"hello"`.
func (v String) Cedar() string {
	return strconv.Quote(string(v))
}
func (v String) deepClone() Value { return v }

func ValueToString(v Value) (String, error) {
	sv, ok := v.(String)
	if !ok {
		return "", fmt.Errorf("%w: expected string, got %v", ErrType, v.TypeName())
	}
	return sv, nil
}

// A Set is a collection of elements that can be of the same or different types.
type Set []Value

func (s Set) Contains(v Value) bool {
	for _, e := range s {
		if e.Equal(v) {
			return true
		}
	}
	return false
}

// Equals returns true if the sets are Equal.
func (s Set) Equals(b Set) bool { return s.Equal(b) }

func (as Set) Equal(bi Value) bool {
	bs, ok := bi.(Set)
	if !ok {
		return false
	}
	for _, a := range as {
		if !bs.Contains(a) {
			return false
		}
	}
	for _, b := range bs {
		if !as.Contains(b) {
			return false
		}
	}
	return true
}

func (v *explicitValue) UnmarshalJSON(b []byte) error {
	return UnmarshalJSON(b, &v.Value)
}

func (v *Set) UnmarshalJSON(b []byte) error {
	var res []explicitValue
	err := json.Unmarshal(b, &res)
	if err != nil {
		return err
	}
	for _, vv := range res {
		*v = append(*v, vv.Value)
	}
	return nil
}

// MarshalJSON marshals the Set into JSON, the marshaller uses the explicit JSON
// form for all the values in the Set.
func (v Set) MarshalJSON() ([]byte, error) {
	w := &bytes.Buffer{}
	w.WriteByte('[')
	for i, vv := range v {
		if i > 0 {
			w.WriteByte(',')
		}
		b, err := vv.ExplicitMarshalJSON()
		if err != nil {
			return nil, err
		}
		w.Write(b)
	}
	w.WriteByte(']')
	return w.Bytes(), nil
}

// ExplicitMarshalJSON marshals the Set into JSON, the marshaller uses the
// explicit JSON form for all the values in the Set.
func (v Set) ExplicitMarshalJSON() ([]byte, error) { return v.MarshalJSON() }

func (v Set) TypeName() string { return "set" }

// String produces a string representation of the Set, e.g. `[1,2,3]`.
func (v Set) String() string { return v.Cedar() }

// Cedar produces a valid Cedar language representation of the Set, e.g. `[1,2,3]`.
func (v Set) Cedar() string {
	var sb strings.Builder
	sb.WriteRune('[')
	for i, elem := range v {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(elem.Cedar())
	}
	sb.WriteRune(']')
	return sb.String()
}
func (v Set) deepClone() Value { return v.DeepClone() }

// DeepClone returns a deep clone of the Set.
func (v Set) DeepClone() Set {
	if v == nil {
		return v
	}
	res := make(Set, len(v))
	for i, vv := range v {
		res[i] = vv.deepClone()
	}
	return res
}

func ValueToSet(v Value) (Set, error) {
	sv, ok := v.(Set)
	if !ok {
		return nil, fmt.Errorf("%w: expected set, got %v", ErrType, v.TypeName())
	}
	return sv, nil
}

// A Record is a collection of attributes. Each attribute consists of a name and
// an associated value. Names are simple strings. Values can be of any type.
type Record map[string]Value

// Equals returns true if the records are Equal.
func (r Record) Equals(b Record) bool { return r.Equal(b) }

func (a Record) Equal(bi Value) bool {
	b, ok := bi.(Record)
	if !ok || len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok || !av.Equal(bv) {
			return false
		}
	}
	return true
}

func (v *Record) UnmarshalJSON(b []byte) error {
	var res map[string]explicitValue
	err := json.Unmarshal(b, &res)
	if err != nil {
		return err
	}
	*v = Record{}
	for kk, vv := range res {
		(*v)[kk] = vv.Value
	}
	return nil
}

// MarshalJSON marshals the Record into JSON, the marshaller uses the explicit
// JSON form for all the values in the Record.
func (v Record) MarshalJSON() ([]byte, error) {
	w := &bytes.Buffer{}
	w.WriteByte('{')
	keys := maps.Keys(v)
	slices.Sort(keys)
	for i, kk := range keys {
		if i > 0 {
			w.WriteByte(',')
		}
		kb, _ := json.Marshal(kk) // json.Marshal cannot error on strings
		w.Write(kb)
		w.WriteByte(':')
		vv := v[kk]
		vb, err := vv.ExplicitMarshalJSON()
		if err != nil {
			return nil, err
		}
		w.Write(vb)
	}
	w.WriteByte('}')
	return w.Bytes(), nil
}

// ExplicitMarshalJSON marshals the Record into JSON, the marshaller uses the
// explicit JSON form for all the values in the Record.
func (v Record) ExplicitMarshalJSON() ([]byte, error) { return v.MarshalJSON() }
func (r Record) TypeName() string                     { return "record" }

// String produces a string representation of the Record, e.g. `{"a":1,"b":2,"c":3}`.
func (r Record) String() string { return r.Cedar() }

// Cedar produces a valid Cedar language representation of the Record, e.g. `{"a":1,"b":2,"c":3}`.
func (r Record) Cedar() string {
	var sb strings.Builder
	sb.WriteRune('{')
	first := true
	keys := maps.Keys(r)
	slices.Sort(keys)
	for _, k := range keys {
		v := r[k]
		if !first {
			sb.WriteString(",")
		}
		first = false
		sb.WriteString(strconv.Quote(k))
		sb.WriteString(":")
		sb.WriteString(v.Cedar())
	}
	sb.WriteRune('}')
	return sb.String()
}
func (v Record) deepClone() Value { return v.DeepClone() }

// DeepClone returns a deep clone of the Record.
func (v Record) DeepClone() Record {
	if v == nil {
		return v
	}
	res := make(Record, len(v))
	for k, vv := range v {
		res[k] = vv.deepClone()
	}
	return res
}

func ValueToRecord(v Value) (Record, error) {
	rv, ok := v.(Record)
	if !ok {
		return nil, fmt.Errorf("%w: expected record got %v", ErrType, v.TypeName())
	}
	return rv, nil
}

// An EntityUID is the identifier for a principal, action, or resource.
type EntityUID struct {
	Type string
	ID   string
}

func NewEntityUID(typ, id string) EntityUID {
	return EntityUID{
		Type: typ,
		ID:   id,
	}
}

// IsZero returns true if the EntityUID has an empty Type and ID.
func (a EntityUID) IsZero() bool {
	return a.Type == "" && a.ID == ""
}

func (a EntityUID) Equal(bi Value) bool {
	b, ok := bi.(EntityUID)
	return ok && a == b
}
func (v EntityUID) TypeName() string { return fmt.Sprintf("(entity of type `%s`)", v.Type) }

// String produces a string representation of the EntityUID, e.g. `Type::"id"`.
func (v EntityUID) String() string { return v.Cedar() }

// Cedar produces a valid Cedar language representation of the EntityUID, e.g. `Type::"id"`.
func (v EntityUID) Cedar() string {
	return v.Type + "::" + strconv.Quote(v.ID)
}

func (v *EntityUID) UnmarshalJSON(b []byte) error {
	// TODO: review after adding support for schemas
	var res entityValueJSON
	if err := json.Unmarshal(b, &res); err != nil {
		return err
	}
	if res.Entity != nil {
		v.Type = res.Entity.Type
		v.ID = res.Entity.ID
		return nil
	} else if res.Type != nil && res.ID != nil { // require both Type and ID to parse "implicit" JSON
		v.Type = *res.Type
		v.ID = *res.ID
		return nil
	}
	return errJSONEntityNotFound
}

// ExplicitMarshalJSON marshals the EntityUID into JSON using the implicit form.
func (v EntityUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(entityValueJSON{
		Type: &v.Type,
		ID:   &v.ID,
	})
}

// ExplicitMarshalJSON marshals the EntityUID into JSON using the explicit form.
func (v EntityUID) ExplicitMarshalJSON() ([]byte, error) {
	return json.Marshal(entityValueJSON{
		Entity: &extEntity{
			Type: v.Type,
			ID:   v.ID,
		},
	})
}
func (v EntityUID) deepClone() Value { return v }

func ValueToEntity(v Value) (EntityUID, error) {
	ev, ok := v.(EntityUID)
	if !ok {
		return EntityUID{}, fmt.Errorf("%w: expected (entity of type `any_entity_type`), got %v", ErrType, v.TypeName())
	}
	return ev, nil
}

func EntityValueFromSlice(v []string) EntityUID {
	return EntityUID{
		Type: strings.Join(v[:len(v)-1], "::"),
		ID:   v[len(v)-1],
	}
}

// Path is the type portion of an EntityUID
type Path string

func (a Path) Equal(bi Value) bool {
	b, ok := bi.(Path)
	return ok && a == b
}
func (v Path) TypeName() string { return fmt.Sprintf("(Path of type `%s`)", v) }

func (v Path) String() string                       { return string(v) }
func (v Path) Cedar() string                        { return string(v) }
func (v Path) ExplicitMarshalJSON() ([]byte, error) { return json.Marshal(string(v)) }
func (v Path) deepClone() Value                     { return v }

func ValueToPath(v Value) (Path, error) {
	ev, ok := v.(Path)
	if !ok {
		return "", fmt.Errorf("%w: expected (Path of type `any_entity_type`), got %v", ErrType, v.TypeName())
	}
	return ev, nil
}

func PathFromSlice(v []string) Path {
	return Path(strings.Join(v, "::"))
}

// A Decimal is a value with both a whole number part and a decimal part of no
// more than four digits. In Go this is stored as an int64, the precision is
// defined by the constant DecimalPrecision.
type Decimal int64

// DecimalPrecision is the precision of a Decimal.
const DecimalPrecision = 10000

// ParseDecimal takes a string representation of a decimal number and converts it into a Decimal type.
func ParseDecimal(s string) (Decimal, error) {
	// Check for empty string.
	if len(s) == 0 {
		return Decimal(0), fmt.Errorf("%w: string too short", ErrDecimal)
	}
	i := 0

	// Parse an optional '-'.
	negative := false
	if s[i] == '-' {
		negative = true
		i++
		if i == len(s) {
			return Decimal(0), fmt.Errorf("%w: string too short", ErrDecimal)
		}
	}

	// Parse the required first digit.
	c := rune(s[i])
	if !unicode.IsDigit(c) {
		return Decimal(0), fmt.Errorf("%w: unexpected character %s", ErrDecimal, strconv.QuoteRune(c))
	}
	integer := int64(c - '0')
	i++

	// Parse any other digits, ending with i pointing to '.'.
	for ; ; i++ {
		if i == len(s) {
			return Decimal(0), fmt.Errorf("%w: string missing decimal point", ErrDecimal)
		}
		c = rune(s[i])
		if c == '.' {
			break
		}
		if !unicode.IsDigit(c) {
			return Decimal(0), fmt.Errorf("%w: unexpected character %s", ErrDecimal, strconv.QuoteRune(c))
		}
		integer = 10*integer + int64(c-'0')
		if integer > 922337203685477 {
			return Decimal(0), fmt.Errorf("%w: overflow", ErrDecimal)
		}
	}

	// Advance past the '.'.
	i++

	// Parse the fraction part
	fraction := int64(0)
	fractionDigits := 0
	for ; i < len(s); i++ {
		c = rune(s[i])
		if !unicode.IsDigit(c) {
			return Decimal(0), fmt.Errorf("%w: unexpected character %s", ErrDecimal, strconv.QuoteRune(c))
		}
		fraction = 10*fraction + int64(c-'0')
		fractionDigits++
	}

	// Adjust the fraction part based on how many digits we parsed.
	switch fractionDigits {
	case 0:
		return Decimal(0), fmt.Errorf("%w: missing digits after decimal point", ErrDecimal)
	case 1:
		fraction *= 1000
	case 2:
		fraction *= 100
	case 3:
		fraction *= 10
	case 4:
	default:
		return Decimal(0), fmt.Errorf("%w: too many digits after decimal point", ErrDecimal)
	}

	// Check for overflow before we put the number together.
	if integer >= 922337203685477 && (fraction > 5808 || (!negative && fraction == 5808)) {
		return Decimal(0), fmt.Errorf("%w: overflow", ErrDecimal)
	}

	// Put the number together.
	if negative {
		// Doing things in this order keeps us from overflowing when parsing
		// -922337203685477.5808. This isn't technically necessary because the
		// go spec defines arithmetic to be well-defined when overflowing.
		// However, doing things this way doesn't hurt, so let's be pedantic.
		return Decimal(DecimalPrecision*-integer - fraction), nil
	} else {
		return Decimal(DecimalPrecision*integer + fraction), nil
	}
}

func (a Decimal) Equal(bi Value) bool {
	b, ok := bi.(Decimal)
	return ok && a == b
}

func (v Decimal) TypeName() string { return "decimal" }

// Cedar produces a valid Cedar language representation of the Decimal, e.g. `decimal("12.34")`.
func (v Decimal) Cedar() string { return `decimal("` + v.String() + `")` }

// String produces a string representation of the Decimal, e.g. `12.34`.
func (v Decimal) String() string {
	var res string
	if v < 0 {
		// Make sure we don't overflow here. Also, go truncates towards zero.
		integer := v / DecimalPrecision
		decimal := integer*DecimalPrecision - v
		res = fmt.Sprintf("-%d.%04d", -integer, decimal)
	} else {
		res = fmt.Sprintf("%d.%04d", v/DecimalPrecision, v%DecimalPrecision)
	}

	// Trim off up to three trailing zeros.
	right := len(res)
	for trimmed := 0; right-1 >= 0 && trimmed < 3; right, trimmed = right-1, trimmed+1 {
		if res[right-1] != '0' {
			break
		}
	}
	return res[:right]
}

func (v *Decimal) UnmarshalJSON(b []byte) error {
	var arg string
	if len(b) > 0 && b[0] == '"' {
		if err := json.Unmarshal(b, &arg); err != nil {
			return errors.Join(errJSONDecode, err)
		}
	} else {
		// NOTE: cedar supports two other forms, for now we're only supporting the smallest implicit and explicit form.
		// The following are not supported:
		// "decimal(\"1234.5678\")"
		// {"fn":"decimal","arg":"1234.5678"}
		var res extValueJSON
		if err := json.Unmarshal(b, &res); err != nil {
			return errors.Join(errJSONDecode, err)
		}
		if res.Extn == nil {
			return errJSONExtNotFound
		}
		if res.Extn.Fn != "decimal" {
			return errJSONExtFnMatch
		}
		arg = res.Extn.Arg
	}
	vv, err := ParseDecimal(arg)
	if err != nil {
		return err
	}
	*v = vv
	return nil
}

// ExplicitMarshalJSON marshals the Decimal into JSON using the implicit form.
func (v Decimal) MarshalJSON() ([]byte, error) { return []byte(`"` + v.String() + `"`), nil }

// ExplicitMarshalJSON marshals the Decimal into JSON using the explicit form.
func (v Decimal) ExplicitMarshalJSON() ([]byte, error) {
	return json.Marshal(extValueJSON{
		Extn: &extn{
			Fn:  "decimal",
			Arg: v.String(),
		},
	})
}
func (v Decimal) deepClone() Value { return v }

func ValueToDecimal(v Value) (Decimal, error) {
	d, ok := v.(Decimal)
	if !ok {
		return 0, fmt.Errorf("%w: expected decimal, got %v", ErrType, v.TypeName())
	}
	return d, nil
}

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

func (v IPAddr) TypeName() string { return "IP" }

// Cedar produces a valid Cedar language representation of the IPAddr, e.g. `ip("127.0.0.1")`.
func (v IPAddr) Cedar() string { return `ip("` + v.String() + `")` }

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

// in this case, netip.Prefix does contain a pointer, but
// the interface given is immutable, so it is safe to return
func (v IPAddr) deepClone() Value { return v }

func ValueToIP(v Value) (IPAddr, error) {
	i, ok := v.(IPAddr)
	if !ok {
		return IPAddr{}, fmt.Errorf("%w: expected ipaddr, got %v", ErrType, v.TypeName())
	}
	return i, nil
}
