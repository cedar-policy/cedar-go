package types

import (
	"fmt"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/entities"
	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func TestBool(t *testing.T) {
	t.Parallel()
	t.Run("roundTrip", func(t *testing.T) {
		t.Parallel()
		v, err := ValueToBool(Boolean(true))
		testutil.OK(t, err)
		testutil.Equals(t, v, true)
	})

	t.Run("toBoolOnNonBool", func(t *testing.T) {
		t.Parallel()
		v, err := ValueToBool(Long(0))
		testutil.AssertError(t, err, ErrType)
		testutil.Equals(t, v, false)
	})

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		t1 := Boolean(true)
		t2 := Boolean(true)
		f := Boolean(false)
		zero := Long(0)
		testutil.FatalIf(t, !t1.Equal(t1), "%v not Equal to %v", t1, t1)
		testutil.FatalIf(t, !t1.Equal(t2), "%v not Equal to %v", t1, t2)
		testutil.FatalIf(t, t1.Equal(f), "%v Equal to %v", t1, f)
		testutil.FatalIf(t, f.Equal(t1), "%v Equal to %v", f, t1)
		testutil.FatalIf(t, f.Equal(zero), "%v Equal to %v", f, zero)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		AssertValueString(t, Boolean(true), "true")
	})

	t.Run("TypeName", func(t *testing.T) {
		t.Parallel()
		tn := Boolean(true).TypeName()
		testutil.Equals(t, tn, "bool")
	})
}

func TestLong(t *testing.T) {
	t.Parallel()
	t.Run("roundTrip", func(t *testing.T) {
		t.Parallel()
		v, err := ValueToLong(Long(42))
		testutil.OK(t, err)
		testutil.Equals(t, v, 42)
	})

	t.Run("toLongOnNonLong", func(t *testing.T) {
		t.Parallel()
		v, err := ValueToLong(Boolean(true))
		testutil.AssertError(t, err, ErrType)
		testutil.Equals(t, v, 0)
	})

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		one := Long(1)
		one2 := Long(1)
		zero := Long(0)
		f := Boolean(false)
		testutil.FatalIf(t, !one.Equal(one), "%v not Equal to %v", one, one)
		testutil.FatalIf(t, !one.Equal(one2), "%v not Equal to %v", one, one2)
		testutil.FatalIf(t, one.Equal(zero), "%v Equal to %v", one, zero)
		testutil.FatalIf(t, zero.Equal(one), "%v Equal to %v", zero, one)
		testutil.FatalIf(t, zero.Equal(f), "%v Equal to %v", zero, f)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		AssertValueString(t, Long(1), "1")
	})

	t.Run("TypeName", func(t *testing.T) {
		t.Parallel()
		tn := Long(1).TypeName()
		testutil.Equals(t, tn, "long")
	})
}

func TestString(t *testing.T) {
	t.Parallel()
	t.Run("roundTrip", func(t *testing.T) {
		t.Parallel()
		v, err := ValueToString(String("hello"))
		testutil.OK(t, err)
		testutil.Equals(t, v, "hello")
	})

	t.Run("toStringOnNonString", func(t *testing.T) {
		t.Parallel()
		v, err := ValueToString(Boolean(true))
		testutil.AssertError(t, err, ErrType)
		testutil.Equals(t, v, "")
	})

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		hello := String("hello")
		hello2 := String("hello")
		goodbye := String("goodbye")
		testutil.FatalIf(t, !hello.Equal(hello), "%v not Equal to %v", hello, hello)
		testutil.FatalIf(t, !hello.Equal(hello2), "%v not Equal to %v", hello, hello2)
		testutil.FatalIf(t, hello.Equal(goodbye), "%v Equal to %v", hello, goodbye)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		AssertValueString(t, String("hello"), `hello`)
		AssertValueString(t, String("hello\ngoodbye"), "hello\ngoodbye")
	})

	t.Run("TypeName", func(t *testing.T) {
		t.Parallel()
		tn := String("hello").TypeName()
		testutil.Equals(t, tn, "string")
	})
}

func TestSet(t *testing.T) {
	t.Parallel()
	t.Run("roundTrip", func(t *testing.T) {
		t.Parallel()
		v := Set{Boolean(true), Long(1)}
		slice, err := ValueToSet(v)
		testutil.OK(t, err)
		v2 := slice
		testutil.FatalIf(t, !v.Equal(v2), "got %v want %v", v, v2)
	})

	t.Run("ToSetOnNonSet", func(t *testing.T) {
		t.Parallel()
		v, err := ValueToSet(Boolean(true))
		testutil.AssertError(t, err, ErrType)
		testutil.Equals(t, v, nil)
	})

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		empty := Set{}
		empty2 := Set{}
		oneTrue := Set{Boolean(true)}
		oneTrue2 := Set{Boolean(true)}
		oneFalse := Set{Boolean(false)}
		nestedOnce := Set{empty, oneTrue, oneFalse}
		nestedOnce2 := Set{empty, oneTrue, oneFalse}
		nestedTwice := Set{empty, oneTrue, oneFalse, nestedOnce}
		nestedTwice2 := Set{empty, oneTrue, oneFalse, nestedOnce}
		oneTwoThree := Set{
			Long(1), Long(2), Long(3),
		}
		threeTwoTwoOne := Set{
			Long(3), Long(2), Long(2), Long(1),
		}

		testutil.FatalIf(t, !empty.Equals(empty), "%v not Equal to %v", empty, empty)
		testutil.FatalIf(t, !empty.Equals(empty2), "%v not Equal to %v", empty, empty2)
		testutil.FatalIf(t, !oneTrue.Equals(oneTrue), "%v not Equal to %v", oneTrue, oneTrue)
		testutil.FatalIf(t, !oneTrue.Equals(oneTrue2), "%v not Equal to %v", oneTrue, oneTrue2)
		testutil.FatalIf(t, !nestedOnce.Equals(nestedOnce), "%v not Equal to %v", nestedOnce, nestedOnce)
		testutil.FatalIf(t, !nestedOnce.Equals(nestedOnce2), "%v not Equal to %v", nestedOnce, nestedOnce2)
		testutil.FatalIf(t, !nestedTwice.Equals(nestedTwice), "%v not Equal to %v", nestedTwice, nestedTwice)
		testutil.FatalIf(t, !nestedTwice.Equals(nestedTwice2), "%v not Equal to %v", nestedTwice, nestedTwice2)
		testutil.FatalIf(t, !oneTwoThree.Equals(threeTwoTwoOne), "%v not Equal to %v", oneTwoThree, threeTwoTwoOne)

		testutil.FatalIf(t, empty.Equals(oneFalse), "%v Equal to %v", empty, oneFalse)
		testutil.FatalIf(t, oneTrue.Equals(oneFalse), "%v Equal to %v", oneTrue, oneFalse)
		testutil.FatalIf(t, nestedOnce.Equals(nestedTwice), "%v Equal to %v", nestedOnce, nestedTwice)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		AssertValueString(t, Set{}, "[]")
		AssertValueString(
			t,
			Set{Boolean(true), Long(1)},
			"[true,1]")
	})

	t.Run("TypeName", func(t *testing.T) {
		t.Parallel()
		tn := Set{}.TypeName()
		testutil.Equals(t, tn, "set")
	})
}

func TestRecord(t *testing.T) {
	t.Parallel()
	t.Run("roundTrip", func(t *testing.T) {
		t.Parallel()
		v := Record{
			"foo": Boolean(true),
			"bar": Long(1),
		}
		map_, err := ValueToRecord(v)
		testutil.OK(t, err)
		v2 := map_
		testutil.FatalIf(t, !v.Equal(v2), "got %v want %v", v, v2)
	})

	t.Run("toRecordOnNonRecord", func(t *testing.T) {
		t.Parallel()
		v, err := ValueToRecord(String("hello"))
		testutil.AssertError(t, err, ErrType)
		testutil.Equals(t, v, nil)
	})

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		empty := Record{}
		empty2 := Record{}
		twoElems := Record{
			"foo": Boolean(true),
			"bar": String("blah"),
		}
		twoElems2 := Record{
			"foo": Boolean(true),
			"bar": String("blah"),
		}
		differentValues := Record{
			"foo": Boolean(false),
			"bar": String("blaz"),
		}
		differentKeys := Record{
			"foo": Boolean(false),
			"bar": Long(1),
		}
		nested := Record{
			"one":  Long(1),
			"two":  Long(2),
			"nest": twoElems,
		}
		nested2 := Record{
			"one":  Long(1),
			"two":  Long(2),
			"nest": twoElems,
		}

		testutil.FatalIf(t, !empty.Equals(empty), "%v not Equal to %v", empty, empty)
		testutil.FatalIf(t, !empty.Equals(empty2), "%v not Equal to %v", empty, empty2)

		testutil.FatalIf(t, !twoElems.Equals(twoElems), "%v not Equal to %v", twoElems, twoElems)
		testutil.FatalIf(t, !twoElems.Equals(twoElems2), "%v not Equal to %v", twoElems, twoElems2)

		testutil.FatalIf(t, !nested.Equals(nested), "%v not Equal to %v", nested, nested)
		testutil.FatalIf(t, !nested.Equals(nested2), "%v not Equal to %v", nested, nested2)

		testutil.FatalIf(t, nested.Equals(twoElems), "%v Equal to %v", nested, twoElems)
		testutil.FatalIf(t, twoElems.Equals(differentValues), "%v Equal to %v", twoElems, differentValues)
		testutil.FatalIf(t, twoElems.Equals(differentKeys), "%v Equal to %v", twoElems, differentKeys)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		AssertValueString(t, Record{}, "{}")
		AssertValueString(
			t,
			Record{"foo": Boolean(true)},
			`{"foo":true}`)
		AssertValueString(
			t,
			Record{
				"foo": Boolean(true),
				"bar": String("blah"),
			},
			`{"bar":"blah","foo":true}`)
	})

	t.Run("TypeName", func(t *testing.T) {
		t.Parallel()
		tn := Record{}.TypeName()
		testutil.Equals(t, tn, "record")
	})
}

func TestEntity(t *testing.T) {
	t.Parallel()
	t.Run("roundTrip", func(t *testing.T) {
		t.Parallel()
		want := EntityUID{Type: "User", ID: "bananas"}
		v, err := ValueToEntity(want)
		testutil.OK(t, err)
		testutil.Equals(t, v, want)
	})
	t.Run("ToEntityOnNonEntity", func(t *testing.T) {
		t.Parallel()
		v, err := ValueToEntity(String("hello"))
		testutil.AssertError(t, err, ErrType)
		testutil.Equals(t, v, EntityUID{})
	})

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		twoElems := EntityUID{"type", "id"}
		twoElems2 := EntityUID{"type", "id"}
		differentValues := EntityUID{"asdf", "vfds"}
		testutil.FatalIf(t, !twoElems.Equal(twoElems), "%v not Equal to %v", twoElems, twoElems)
		testutil.FatalIf(t, !twoElems.Equal(twoElems2), "%v not Equal to %v", twoElems, twoElems2)
		testutil.FatalIf(t, twoElems.Equal(differentValues), "%v Equal to %v", twoElems, differentValues)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		AssertValueString(t, EntityUID{Type: "type", ID: "id"}, `type::"id"`)
		AssertValueString(t, EntityUID{Type: "namespace::type", ID: "id"}, `namespace::type::"id"`)
	})

	t.Run("TypeName", func(t *testing.T) {
		t.Parallel()
		tn := EntityUID{"T", "id"}.TypeName()
		testutil.Equals(t, tn, "(entity of type `T`)")
	})
}

func TestDecimal(t *testing.T) {
	t.Parallel()
	{
		tests := []struct{ in, out string }{
			{"1.2345", "1.2345"},
			{"1.2340", "1.234"},
			{"1.2300", "1.23"},
			{"1.2000", "1.2"},
			{"1.0000", "1.0"},
			{"1.234", "1.234"},
			{"1.230", "1.23"},
			{"1.200", "1.2"},
			{"1.000", "1.0"},
			{"1.23", "1.23"},
			{"1.20", "1.2"},
			{"1.00", "1.0"},
			{"1.2", "1.2"},
			{"1.0", "1.0"},
			{"01.2345", "1.2345"},
			{"01.2340", "1.234"},
			{"01.2300", "1.23"},
			{"01.2000", "1.2"},
			{"01.0000", "1.0"},
			{"01.234", "1.234"},
			{"01.230", "1.23"},
			{"01.200", "1.2"},
			{"01.000", "1.0"},
			{"01.23", "1.23"},
			{"01.20", "1.2"},
			{"01.00", "1.0"},
			{"01.2", "1.2"},
			{"01.0", "1.0"},
			{"1234.5678", "1234.5678"},
			{"1234.5670", "1234.567"},
			{"1234.5600", "1234.56"},
			{"1234.5000", "1234.5"},
			{"1234.0000", "1234.0"},
			{"1234.567", "1234.567"},
			{"1234.560", "1234.56"},
			{"1234.500", "1234.5"},
			{"1234.000", "1234.0"},
			{"1234.56", "1234.56"},
			{"1234.50", "1234.5"},
			{"1234.00", "1234.0"},
			{"1234.5", "1234.5"},
			{"1234.0", "1234.0"},
			{"0.0", "0.0"},
			{"00.0", "0.0"},
			{"000000000000000000000000000000000000000000000000000000000000000000.0", "0.0"},
			{"922337203685477.5807", "922337203685477.5807"},
			{"-922337203685477.5808", "-922337203685477.5808"},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%s->%s", tt.in, tt.out), func(t *testing.T) {
				t.Parallel()
				d, err := ParseDecimal(tt.in)
				testutil.OK(t, err)
				testutil.Equals(t, d.String(), tt.out)
			})
		}
	}

	{
		tests := []struct{ in, errStr string }{
			{"", "error parsing decimal value: string too short"},
			{"-", "error parsing decimal value: string too short"},
			{"a", "error parsing decimal value: unexpected character 'a'"},
			{"-a", "error parsing decimal value: unexpected character 'a'"},
			{"'", `error parsing decimal value: unexpected character '\''`},
			{`-\\`, `error parsing decimal value: unexpected character '\\'`},
			{"0", "error parsing decimal value: string missing decimal point"},
			{"1a", "error parsing decimal value: unexpected character 'a'"},
			{"1a", "error parsing decimal value: unexpected character 'a'"},
			{"1.", "error parsing decimal value: missing digits after decimal point"},
			{"1.00000", "error parsing decimal value: too many digits after decimal point"},
			{"1.a", "error parsing decimal value: unexpected character 'a'"},
			{"1.0a", "error parsing decimal value: unexpected character 'a'"},
			{"1.0000a", "error parsing decimal value: unexpected character 'a'"},
			{"1.0000a", "error parsing decimal value: unexpected character 'a'"},

			{"1000000000000000.0", "error parsing decimal value: overflow"},
			{"-1000000000000000.0", "error parsing decimal value: overflow"},
			{"922337203685477.5808", "error parsing decimal value: overflow"},
			{"922337203685478.0", "error parsing decimal value: overflow"},
			{"-922337203685477.5809", "error parsing decimal value: overflow"},
			{"-922337203685478.0", "error parsing decimal value: overflow"},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%s->%s", tt.in, tt.errStr), func(t *testing.T) {
				t.Parallel()
				_, err := ParseDecimal(tt.in)
				testutil.AssertError(t, err, ErrDecimal)
				testutil.Equals(t, err.Error(), tt.errStr)
			})
		}
	}

	t.Run("roundTrip", func(t *testing.T) {
		t.Parallel()
		dv, err := ParseDecimal("1.20")
		testutil.OK(t, err)
		v, err := ValueToDecimal(dv)
		testutil.OK(t, err)
		testutil.FatalIf(t, !v.Equal(dv), "got %v want %v", v, dv)
	})

	t.Run("toDecimalOnNonDecimal", func(t *testing.T) {
		t.Parallel()
		v, err := ValueToDecimal(Boolean(true))
		testutil.AssertError(t, err, ErrType)
		testutil.Equals(t, v, 0)
	})

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		one := Decimal(10000)
		one2 := Decimal(10000)
		zero := Decimal(0)
		f := Boolean(false)
		testutil.FatalIf(t, !one.Equal(one), "%v not Equal to %v", one, one)
		testutil.FatalIf(t, !one.Equal(one2), "%v not Equal to %v", one, one2)
		testutil.FatalIf(t, one.Equal(zero), "%v Equal to %v", one, zero)
		testutil.FatalIf(t, zero.Equal(one), "%v Equal to %v", zero, one)
		testutil.FatalIf(t, zero.Equal(f), "%v Equal to %v", zero, f)
	})

	t.Run("TypeName", func(t *testing.T) {
		t.Parallel()
		tn := Decimal(0).TypeName()
		testutil.Equals(t, tn, "decimal")
	})
}

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
				i, err := ParseIPAddr(tt.in)
				if tt.parses {
					testutil.OK(t, err)
					testutil.Equals(t, i.String(), tt.out)
				} else {
					testutil.Error(t, err)
				}
			})
		}
	})

	t.Run("toIPOnNonIP", func(t *testing.T) {
		t.Parallel()
		v, err := ValueToIP(Boolean(true))
		testutil.AssertError(t, err, ErrType)
		testutil.Equals(t, v, IPAddr{})
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
				lhs, err := ParseIPAddr(tt.lhs)
				testutil.OK(t, err)
				rhs, err := ParseIPAddr(tt.rhs)
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
				val, err := ParseIPAddr(tt.val)
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
				val, err := ParseIPAddr(tt.val)
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
				val, err := ParseIPAddr(tt.val)
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
				lhs, err := ParseIPAddr(tt.lhs)
				testutil.OK(t, err)
				rhs, err := ParseIPAddr(tt.rhs)
				testutil.OK(t, err)
				contains := lhs.Contains(rhs)
				if contains != tt.contains {
					t.Fatalf("expected ip(%v).contains(ip(%v)) to be %v instead of %v", tt.lhs, tt.rhs, tt.contains, contains)
				}
			})
		}
	})

	t.Run("TypeName", func(t *testing.T) {
		t.Parallel()
		tn := IPAddr{}.TypeName()
		testutil.Equals(t, tn, "IP")
	})
}

func TestDeepClone(t *testing.T) {
	t.Parallel()
	t.Run("Boolean", func(t *testing.T) {
		t.Parallel()
		a := Boolean(true)
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a = Boolean(false)
		testutil.Equals(t, a, Boolean(false))
		testutil.Equals(t, b, Value(Boolean(true)))
	})
	t.Run("Long", func(t *testing.T) {
		t.Parallel()
		a := Long(42)
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a = Long(43)
		testutil.Equals(t, a, Long(43))
		testutil.Equals(t, b, Value(Long(42)))
	})
	t.Run("String", func(t *testing.T) {
		t.Parallel()
		a := String("cedar")
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a = String("policy")
		testutil.Equals(t, a, String("policy"))
		testutil.Equals(t, b, Value(String("cedar")))
	})
	t.Run("EntityUID", func(t *testing.T) {
		t.Parallel()
		a := NewEntityUID("Action", "test")
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a.ID = "bananas"
		testutil.Equals(t, a, NewEntityUID("Action", "bananas"))
		testutil.Equals(t, b, Value(NewEntityUID("Action", "test")))
	})

	t.Run("Set", func(t *testing.T) {
		t.Parallel()
		a := Set{Long(42)}
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a[0] = String("bananas")
		testutil.Equals(t, a, Set{String("bananas")})
		testutil.Equals(t, b, Value(Set{Long(42)}))
	})
	t.Run("NilSet", func(t *testing.T) {
		t.Parallel()
		var a Set
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
	})

	t.Run("Record", func(t *testing.T) {
		t.Parallel()
		a := Record{"key": Long(42)}
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a["key"] = String("bananas")
		testutil.Equals(t, a, Record{"key": String("bananas")})
		testutil.Equals(t, b, Value(Record{"key": Long(42)}))
	})

	t.Run("NilRecord", func(t *testing.T) {
		t.Parallel()
		var a Record
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
	})

	t.Run("Decimal", func(t *testing.T) {
		t.Parallel()
		a := Decimal(42)
		b := a.deepClone()
		testutil.Equals(t, Value(a), b)
		a = Decimal(43)
		testutil.Equals(t, a, Decimal(43))
		testutil.Equals(t, b, Value(Decimal(42)))
	})

	t.Run("IPAddr", func(t *testing.T) {
		t.Parallel()
		a := mustIPValue("127.0.0.42")
		b := a.deepClone()
		testutil.Equals(t, a.Cedar(), b.Cedar())
		a = mustIPValue("127.0.0.43")
		testutil.Equals(t, a.Cedar(), mustIPValue("127.0.0.43").Cedar())
		testutil.Equals(t, b.Cedar(), mustIPValue("127.0.0.42").Cedar())
	})
}

func TestPath(t *testing.T) {
	t.Parallel()
	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		a := Path("X")
		b := Path("X")
		c := Path("Y")
		testutil.Equals(t, a.Equal(b), true)
		testutil.Equals(t, b.Equal(a), true)
		testutil.Equals(t, a.Equal(c), false)
		testutil.Equals(t, c.Equal(a), false)
	})
	t.Run("TypeName", func(t *testing.T) {
		t.Parallel()
		a := Path("X")
		testutil.Equals(t, a.TypeName(), "(Path of type `X`)")
	})
	t.Run("String", func(t *testing.T) {
		t.Parallel()
		a := Path("X")
		testutil.Equals(t, a.String(), "X")
	})
	t.Run("Cedar", func(t *testing.T) {
		t.Parallel()
		a := Path("X")
		testutil.Equals(t, a.Cedar(), "X")
	})
	t.Run("ExplicitMarshalJSON", func(t *testing.T) {
		t.Parallel()
		a := Path("X")
		v, err := a.ExplicitMarshalJSON()
		testutil.OK(t, err)
		testutil.Equals(t, string(v), `"X"`)
	})
	t.Run("deepClone", func(t *testing.T) {
		t.Parallel()
		a := Path("X")
		b := a.deepClone()
		c, ok := b.(Path)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, c, a)
	})

	t.Run("pathFromSlice", func(t *testing.T) {
		t.Parallel()
		a := PathFromSlice([]string{"X", "Y"})
		testutil.Equals(t, a, Path("X::Y"))
	})

}

func TestEntities(t *testing.T) {
	t.Parallel()
	t.Run("Clone", func(t *testing.T) {
		t.Parallel()
		e := entities.Entities{
			EntityUID{Type: "A", ID: "A"}: {},
			EntityUID{Type: "A", ID: "B"}: {},
			EntityUID{Type: "B", ID: "A"}: {},
			EntityUID{Type: "B", ID: "B"}: {},
		}
		clone := e.Clone()
		testutil.Equals(t, clone, e)
	})

}
