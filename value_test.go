package cedar

import (
	"fmt"
	"testing"
)

func TestBool(t *testing.T) {
	t.Parallel()
	t.Run("roundTrip", func(t *testing.T) {
		t.Parallel()
		v, err := valueToBool(Boolean(true))
		testutilOK(t, err)
		testutilEquals(t, v, true)
	})

	t.Run("toBoolOnNonBool", func(t *testing.T) {
		t.Parallel()
		v, err := valueToBool(Long(0))
		assertError(t, err, errType)
		testutilEquals(t, v, false)
	})

	t.Run("equal", func(t *testing.T) {
		t.Parallel()
		t1 := Boolean(true)
		t2 := Boolean(true)
		f := Boolean(false)
		zero := Long(0)
		testutilFatalIf(t, !t1.equal(t1), "%v not equal to %v", t1, t1)
		testutilFatalIf(t, !t1.equal(t2), "%v not equal to %v", t1, t2)
		testutilFatalIf(t, t1.equal(f), "%v equal to %v", t1, f)
		testutilFatalIf(t, f.equal(t1), "%v equal to %v", f, t1)
		testutilFatalIf(t, f.equal(zero), "%v equal to %v", f, zero)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		assertValueString(t, Boolean(true), "true")
	})

	t.Run("typeName", func(t *testing.T) {
		t.Parallel()
		tn := Boolean(true).typeName()
		testutilEquals(t, tn, "bool")
	})
}

func TestLong(t *testing.T) {
	t.Parallel()
	t.Run("roundTrip", func(t *testing.T) {
		t.Parallel()
		v, err := valueToLong(Long(42))
		testutilOK(t, err)
		testutilEquals(t, v, 42)
	})

	t.Run("toLongOnNonLong", func(t *testing.T) {
		t.Parallel()
		v, err := valueToLong(Boolean(true))
		assertError(t, err, errType)
		testutilEquals(t, v, 0)
	})

	t.Run("equal", func(t *testing.T) {
		t.Parallel()
		one := Long(1)
		one2 := Long(1)
		zero := Long(0)
		f := Boolean(false)
		testutilFatalIf(t, !one.equal(one), "%v not equal to %v", one, one)
		testutilFatalIf(t, !one.equal(one2), "%v not equal to %v", one, one2)
		testutilFatalIf(t, one.equal(zero), "%v equal to %v", one, zero)
		testutilFatalIf(t, zero.equal(one), "%v equal to %v", zero, one)
		testutilFatalIf(t, zero.equal(f), "%v equal to %v", zero, f)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		assertValueString(t, Long(1), "1")
	})

	t.Run("typeName", func(t *testing.T) {
		t.Parallel()
		tn := Long(1).typeName()
		testutilEquals(t, tn, "long")
	})
}

func TestString(t *testing.T) {
	t.Parallel()
	t.Run("roundTrip", func(t *testing.T) {
		t.Parallel()
		v, err := valueToString(String("hello"))
		testutilOK(t, err)
		testutilEquals(t, v, "hello")
	})

	t.Run("toStringOnNonString", func(t *testing.T) {
		t.Parallel()
		v, err := valueToString(Boolean(true))
		assertError(t, err, errType)
		testutilEquals(t, v, "")
	})

	t.Run("equal", func(t *testing.T) {
		t.Parallel()
		hello := String("hello")
		hello2 := String("hello")
		goodbye := String("goodbye")
		testutilFatalIf(t, !hello.equal(hello), "%v not equal to %v", hello, hello)
		testutilFatalIf(t, !hello.equal(hello2), "%v not equal to %v", hello, hello2)
		testutilFatalIf(t, hello.equal(goodbye), "%v equal to %v", hello, goodbye)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		assertValueString(t, String("hello"), `hello`)
		assertValueString(t, String("hello\ngoodbye"), "hello\ngoodbye")
	})

	t.Run("typeName", func(t *testing.T) {
		t.Parallel()
		tn := String("hello").typeName()
		testutilEquals(t, tn, "string")
	})
}

func TestSet(t *testing.T) {
	t.Parallel()
	t.Run("roundTrip", func(t *testing.T) {
		t.Parallel()
		v := Set{Boolean(true), Long(1)}
		slice, err := valueToSet(v)
		testutilOK(t, err)
		v2 := slice
		testutilFatalIf(t, !v.equal(v2), "got %v want %v", v, v2)
	})

	t.Run("ToSetOnNonSet", func(t *testing.T) {
		t.Parallel()
		v, err := valueToSet(Boolean(true))
		assertError(t, err, errType)
		testutilEquals(t, v, nil)
	})

	t.Run("equal", func(t *testing.T) {
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

		testutilFatalIf(t, !empty.Equals(empty), "%v not equal to %v", empty, empty)
		testutilFatalIf(t, !empty.Equals(empty2), "%v not equal to %v", empty, empty2)
		testutilFatalIf(t, !oneTrue.Equals(oneTrue), "%v not equal to %v", oneTrue, oneTrue)
		testutilFatalIf(t, !oneTrue.Equals(oneTrue2), "%v not equal to %v", oneTrue, oneTrue2)
		testutilFatalIf(t, !nestedOnce.Equals(nestedOnce), "%v not equal to %v", nestedOnce, nestedOnce)
		testutilFatalIf(t, !nestedOnce.Equals(nestedOnce2), "%v not equal to %v", nestedOnce, nestedOnce2)
		testutilFatalIf(t, !nestedTwice.Equals(nestedTwice), "%v not equal to %v", nestedTwice, nestedTwice)
		testutilFatalIf(t, !nestedTwice.Equals(nestedTwice2), "%v not equal to %v", nestedTwice, nestedTwice2)
		testutilFatalIf(t, !oneTwoThree.Equals(threeTwoTwoOne), "%v not equal to %v", oneTwoThree, threeTwoTwoOne)

		testutilFatalIf(t, empty.Equals(oneFalse), "%v equal to %v", empty, oneFalse)
		testutilFatalIf(t, oneTrue.Equals(oneFalse), "%v equal to %v", oneTrue, oneFalse)
		testutilFatalIf(t, nestedOnce.Equals(nestedTwice), "%v equal to %v", nestedOnce, nestedTwice)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		assertValueString(t, Set{}, "[]")
		assertValueString(
			t,
			Set{Boolean(true), Long(1)},
			"[true,1]")
	})

	t.Run("typeName", func(t *testing.T) {
		t.Parallel()
		tn := Set{}.typeName()
		testutilEquals(t, tn, "set")
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
		map_, err := valueToRecord(v)
		testutilOK(t, err)
		v2 := map_
		testutilFatalIf(t, !v.equal(v2), "got %v want %v", v, v2)
	})

	t.Run("toRecordOnNonRecord", func(t *testing.T) {
		t.Parallel()
		v, err := valueToRecord(String("hello"))
		assertError(t, err, errType)
		testutilEquals(t, v, nil)
	})

	t.Run("equal", func(t *testing.T) {
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

		testutilFatalIf(t, !empty.Equals(empty), "%v not equal to %v", empty, empty)
		testutilFatalIf(t, !empty.Equals(empty2), "%v not equal to %v", empty, empty2)

		testutilFatalIf(t, !twoElems.Equals(twoElems), "%v not equal to %v", twoElems, twoElems)
		testutilFatalIf(t, !twoElems.Equals(twoElems2), "%v not equal to %v", twoElems, twoElems2)

		testutilFatalIf(t, !nested.Equals(nested), "%v not equal to %v", nested, nested)
		testutilFatalIf(t, !nested.Equals(nested2), "%v not equal to %v", nested, nested2)

		testutilFatalIf(t, nested.Equals(twoElems), "%v equal to %v", nested, twoElems)
		testutilFatalIf(t, twoElems.Equals(differentValues), "%v equal to %v", twoElems, differentValues)
		testutilFatalIf(t, twoElems.Equals(differentKeys), "%v equal to %v", twoElems, differentKeys)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		assertValueString(t, Record{}, "{}")
		assertValueString(
			t,
			Record{"foo": Boolean(true)},
			`{"foo":true}`)
		assertValueString(
			t,
			Record{
				"foo": Boolean(true),
				"bar": String("blah"),
			},
			`{"bar":"blah","foo":true}`)
	})

	t.Run("typeName", func(t *testing.T) {
		t.Parallel()
		tn := Record{}.typeName()
		testutilEquals(t, tn, "record")
	})
}

func TestEntity(t *testing.T) {
	t.Parallel()
	t.Run("roundTrip", func(t *testing.T) {
		t.Parallel()
		want := EntityUID{Type: "User", ID: "bananas"}
		v, err := valueToEntity(want)
		testutilOK(t, err)
		testutilEquals(t, v, want)
	})
	t.Run("ToEntityOnNonEntity", func(t *testing.T) {
		t.Parallel()
		v, err := valueToEntity(String("hello"))
		assertError(t, err, errType)
		testutilEquals(t, v, EntityUID{})
	})

	t.Run("equal", func(t *testing.T) {
		t.Parallel()
		twoElems := EntityUID{"type", "id"}
		twoElems2 := EntityUID{"type", "id"}
		differentValues := EntityUID{"asdf", "vfds"}
		testutilFatalIf(t, !twoElems.equal(twoElems), "%v not equal to %v", twoElems, twoElems)
		testutilFatalIf(t, !twoElems.equal(twoElems2), "%v not equal to %v", twoElems, twoElems2)
		testutilFatalIf(t, twoElems.equal(differentValues), "%v equal to %v", twoElems, differentValues)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		assertValueString(t, EntityUID{Type: "type", ID: "id"}, `type::"id"`)
		assertValueString(t, EntityUID{Type: "namespace::type", ID: "id"}, `namespace::type::"id"`)
	})

	t.Run("typeName", func(t *testing.T) {
		t.Parallel()
		tn := EntityUID{"T", "id"}.typeName()
		testutilEquals(t, tn, "(entity of type `T`)")
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
				d, err := newDecimalValue(tt.in)
				testutilOK(t, err)
				testutilEquals(t, d.String(), tt.out)
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
				_, err := newDecimalValue(tt.in)
				assertError(t, err, errDecimal)
				testutilEquals(t, err.Error(), tt.errStr)
			})
		}
	}

	t.Run("roundTrip", func(t *testing.T) {
		t.Parallel()
		dv, err := newDecimalValue("1.20")
		testutilOK(t, err)
		v, err := valueToDecimal(dv)
		testutilOK(t, err)
		testutilFatalIf(t, !v.equal(dv), "got %v want %v", v, dv)
	})

	t.Run("toDecimalOnNonDecimal", func(t *testing.T) {
		t.Parallel()
		v, err := valueToDecimal(Boolean(true))
		assertError(t, err, errType)
		testutilEquals(t, v, 0)
	})

	t.Run("equal", func(t *testing.T) {
		t.Parallel()
		one := Decimal(10000)
		one2 := Decimal(10000)
		zero := Decimal(0)
		f := Boolean(false)
		testutilFatalIf(t, !one.equal(one), "%v not equal to %v", one, one)
		testutilFatalIf(t, !one.equal(one2), "%v not equal to %v", one, one2)
		testutilFatalIf(t, one.equal(zero), "%v equal to %v", one, zero)
		testutilFatalIf(t, zero.equal(one), "%v equal to %v", zero, one)
		testutilFatalIf(t, zero.equal(f), "%v equal to %v", zero, f)
	})

	t.Run("typeName", func(t *testing.T) {
		t.Parallel()
		tn := Decimal(0).typeName()
		testutilEquals(t, tn, "decimal")
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
				i, err := newIPValue(tt.in)
				if tt.parses {
					testutilOK(t, err)
					testutilEquals(t, i.String(), tt.out)
				} else {
					testutilError(t, err)
				}
			})
		}
	})

	t.Run("toIPOnNonIP", func(t *testing.T) {
		t.Parallel()
		v, err := valueToIP(Boolean(true))
		assertError(t, err, errType)
		testutilEquals(t, v, IPAddr{})
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
			t.Run(fmt.Sprintf("ip(%v).equal(ip(%v))", tt.lhs, tt.rhs), func(t *testing.T) {
				t.Parallel()
				lhs, err := newIPValue(tt.lhs)
				testutilOK(t, err)
				rhs, err := newIPValue(tt.rhs)
				testutilOK(t, err)
				equal := lhs.equal(rhs)
				if equal != tt.equal {
					t.Fatalf("expected ip(%v).equal(ip(%v)) to be %v instead of %v", tt.lhs, tt.rhs, tt.equal, equal)
				}
				if equal {
					testutilFatalIf(
						t,
						!lhs.contains(rhs),
						"ip(%v) and ip(%v) compare equal but !ip(%v).contains(ip(%v))", tt.lhs, tt.rhs, tt.lhs, tt.rhs)
					testutilFatalIf(
						t,
						!rhs.contains(lhs),
						"ip(%v) and ip(%v) compare equal but !ip(%v).contains(ip(%v))", tt.rhs, tt.lhs, tt.rhs, tt.lhs)
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
				val, err := newIPValue(tt.val)
				testutilOK(t, err)
				isIPv4 := val.isIPv4()
				if isIPv4 != tt.isIPv4 {
					t.Fatalf("expected ip(%v).isIPv4() to be %v instead of %v", tt.val, tt.isIPv4, isIPv4)
				}
				isIPv6 := val.isIPv6()
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
				val, err := newIPValue(tt.val)
				testutilOK(t, err)
				isLoopback := val.isLoopback()
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
				val, err := newIPValue(tt.val)
				testutilOK(t, err)
				isMulticast := val.isMulticast()
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
				lhs, err := newIPValue(tt.lhs)
				testutilOK(t, err)
				rhs, err := newIPValue(tt.rhs)
				testutilOK(t, err)
				contains := lhs.contains(rhs)
				if contains != tt.contains {
					t.Fatalf("expected ip(%v).contains(ip(%v)) to be %v instead of %v", tt.lhs, tt.rhs, tt.contains, contains)
				}
			})
		}
	})

	t.Run("typeName", func(t *testing.T) {
		t.Parallel()
		tn := IPAddr{}.typeName()
		testutilEquals(t, tn, "IP")
	})
}

func TestDeepClone(t *testing.T) {
	t.Parallel()
	t.Run("Boolean", func(t *testing.T) {
		t.Parallel()
		a := Boolean(true)
		b := a.deepClone()
		testutilEquals(t, Value(a), b)
		a = Boolean(false)
		testutilEquals(t, a, Boolean(false))
		testutilEquals(t, b, Value(Boolean(true)))
	})
	t.Run("Long", func(t *testing.T) {
		t.Parallel()
		a := Long(42)
		b := a.deepClone()
		testutilEquals(t, Value(a), b)
		a = Long(43)
		testutilEquals(t, a, Long(43))
		testutilEquals(t, b, Value(Long(42)))
	})
	t.Run("String", func(t *testing.T) {
		t.Parallel()
		a := String("cedar")
		b := a.deepClone()
		testutilEquals(t, Value(a), b)
		a = String("policy")
		testutilEquals(t, a, String("policy"))
		testutilEquals(t, b, Value(String("cedar")))
	})
	t.Run("EntityUID", func(t *testing.T) {
		t.Parallel()
		a := NewEntityUID("Action", "test")
		b := a.deepClone()
		testutilEquals(t, Value(a), b)
		a.ID = "bananas"
		testutilEquals(t, a, NewEntityUID("Action", "bananas"))
		testutilEquals(t, b, Value(NewEntityUID("Action", "test")))
	})

	t.Run("Set", func(t *testing.T) {
		t.Parallel()
		a := Set{Long(42)}
		b := a.deepClone()
		testutilEquals(t, Value(a), b)
		a[0] = String("bananas")
		testutilEquals(t, a, Set{String("bananas")})
		testutilEquals(t, b, Value(Set{Long(42)}))
	})
	t.Run("NilSet", func(t *testing.T) {
		t.Parallel()
		var a Set
		b := a.deepClone()
		testutilEquals(t, Value(a), b)
	})

	t.Run("Record", func(t *testing.T) {
		t.Parallel()
		a := Record{"key": Long(42)}
		b := a.deepClone()
		testutilEquals(t, Value(a), b)
		a["key"] = String("bananas")
		testutilEquals(t, a, Record{"key": String("bananas")})
		testutilEquals(t, b, Value(Record{"key": Long(42)}))
	})

	t.Run("NilRecord", func(t *testing.T) {
		t.Parallel()
		var a Record
		b := a.deepClone()
		testutilEquals(t, Value(a), b)
	})

	t.Run("Decimal", func(t *testing.T) {
		t.Parallel()
		a := Decimal(42)
		b := a.deepClone()
		testutilEquals(t, Value(a), b)
		a = Decimal(43)
		testutilEquals(t, a, Decimal(43))
		testutilEquals(t, b, Value(Decimal(42)))
	})

	t.Run("IPAddr", func(t *testing.T) {
		t.Parallel()
		a := mustIPValue("127.0.0.42")
		b := a.deepClone()
		testutilEquals(t, a.Cedar(), b.Cedar())
		a = mustIPValue("127.0.0.43")
		testutilEquals(t, a.Cedar(), mustIPValue("127.0.0.43").Cedar())
		testutilEquals(t, b.Cedar(), mustIPValue("127.0.0.42").Cedar())
	})
}

func TestPath(t *testing.T) {
	t.Parallel()
	t.Run("equal", func(t *testing.T) {
		t.Parallel()
		a := path("X")
		b := path("X")
		c := path("Y")
		testutilEquals(t, a.equal(b), true)
		testutilEquals(t, b.equal(a), true)
		testutilEquals(t, a.equal(c), false)
		testutilEquals(t, c.equal(a), false)
	})
	t.Run("typeName", func(t *testing.T) {
		t.Parallel()
		a := path("X")
		testutilEquals(t, a.typeName(), "(path of type `X`)")
	})
	t.Run("String", func(t *testing.T) {
		t.Parallel()
		a := path("X")
		testutilEquals(t, a.String(), "X")
	})
	t.Run("Cedar", func(t *testing.T) {
		t.Parallel()
		a := path("X")
		testutilEquals(t, a.Cedar(), "X")
	})
	t.Run("ExplicitMarshalJSON", func(t *testing.T) {
		t.Parallel()
		a := path("X")
		v, err := a.ExplicitMarshalJSON()
		testutilOK(t, err)
		testutilEquals(t, string(v), `"X"`)
	})
	t.Run("deepClone", func(t *testing.T) {
		t.Parallel()
		a := path("X")
		b := a.deepClone()
		c, ok := b.(path)
		testutilEquals(t, ok, true)
		testutilEquals(t, c, a)
	})

	t.Run("pathFromSlice", func(t *testing.T) {
		t.Parallel()
		a := pathFromSlice([]string{"X", "Y"})
		testutilEquals(t, a, path("X::Y"))
	})

}
