package cedar

import (
	"encoding/json"
	"fmt"
	"testing"
)

func mustDecimalValue(v string) Decimal {
	r, _ := newDecimalValue(v)
	return r
}

func mustIPValue(v string) IPAddr {
	r, _ := newIPValue(v)
	return r
}

func TestJSON_Value(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want Value
		err  error
	}{
		{"impliedEntity", `{ "type": "User", "id": "alice" }`, EntityUID{Type: "User", ID: "alice"}, nil},
		{"explicitEntity", `{ "__entity": { "type": "User", "id": "alice" } }`, EntityUID{Type: "User", ID: "alice"}, nil},
		{"impliedLongEntity", `{ "type": "User::External", "id": "alice" }`, EntityUID{Type: "User::External", ID: "alice"}, nil},
		{"explicitLongEntity", `{ "__entity": { "type": "User::External", "id": "alice" } }`, EntityUID{Type: "User::External", ID: "alice"}, nil},
		{"invalidJSON", `!@#$`, zeroValue(), errJSONDecode},
		{"numericOverflow", "12341234123412341234", zeroValue(), errJSONLongOutOfRange},
		{"unsupportedNull", "null", zeroValue(), errJSONUnsupportedType},
		{"explicitIP", `{ "__extn": { "fn": "ip", "arg": "222.222.222.7" } }`, mustIPValue("222.222.222.7"), nil},
		{"explicitSubnet", `{ "__extn": { "fn": "ip", "arg": "192.168.0.0/16" } }`, mustIPValue("192.168.0.0/16"), nil},
		{"explicitDecimal", `{ "__extn": { "fn": "decimal", "arg": "33.57" } }`, mustDecimalValue("33.57"), nil},
		{"invalidExtension", `{ "__extn": { "fn": "asdf", "arg": "blah" } }`, zeroValue(), errJSONInvalidExtn},
		{"badIP", `{ "__extn": { "fn": "ip", "arg": "bad" } }`, zeroValue(), errIP},
		{"badDecimal", `{ "__extn": { "fn": "decimal", "arg": "bad" } }`, zeroValue(), errDecimal},
		{"set", `[42]`, Set{Long(42)}, nil},
		{"record", `{"a":"b"}`, Record{"a": String("b")}, nil},
		{"bool", `false`, Boolean(false), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got Value
			ptr := &got
			err := unmarshalJSON([]byte(tt.in), ptr)
			assertError(t, err, tt.err)
			assertValue(t, got, tt.want)
			if tt.err != nil {
				return
			}

			// Now assert that when we Marshal/Unmarshal that value, we still
			// have what we started with
			gotJSON, err := (*ptr).ExplicitMarshalJSON()
			testutilOK(t, err)
			var gotRetry Value
			ptr = &gotRetry
			err = unmarshalJSON(gotJSON, ptr)
			testutilOK(t, err)
			testutilEquals(t, gotRetry, tt.want)
		})
	}
}

func TestTypedJSONUnmarshal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		f         func(b []byte) (Value, error)
		in        string
		wantValue Value
		wantErr   error
	}{
		{
			name: "string",
			f: func(b []byte) (Value, error) {
				var res String
				err := json.Unmarshal(b, &res)
				return res, err
			},
			in:        `"hello"`,
			wantValue: String("hello"),
			wantErr:   nil,
		},
		{
			name: "ip",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "ip", "arg": "222.222.222.7" } }`,
			wantValue: mustIPValue("222.222.222.7"),
			wantErr:   nil,
		},
		{
			name: "ip/implicit",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"222.222.222.7"`,
			wantValue: mustIPValue("222.222.222.7"),
			wantErr:   nil,
		},
		{
			name: "ip/implicit/badJSON",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"bad`,
			wantValue: IPAddr{},
			wantErr:   errJSONDecode,
		},
		{
			name: "ip/badArg",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "ip", "arg": "bad" } }`,
			wantValue: IPAddr{},
			wantErr:   errIP,
		},
		{
			name: "ip/badJSON",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `bad`,
			wantValue: IPAddr{},
			wantErr:   errJSONDecode,
		},
		{
			name: "ip/badFn",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "bad", "arg": "222.222.222.7" } }`,
			wantValue: IPAddr{},
			wantErr:   errJSONExtFnMatch,
		},
		{
			name: "ip/ExtNotFound",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ }`,
			wantValue: IPAddr{},
			wantErr:   errJSONExtNotFound,
		},

		{
			name: "decimal",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "decimal", "arg": "1234.5678" } }`,
			wantValue: mustDecimalValue("1234.5678"),
			wantErr:   nil,
		},
		{
			name: "decimal/implicit",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"1234.5678"`,
			wantValue: mustDecimalValue("1234.5678"),
			wantErr:   nil,
		},
		{
			name: "decimal/implicit/badJSON",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"bad`,
			wantValue: Decimal(0),
			wantErr:   errJSONDecode,
		},
		{
			name: "decimal/badArg",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "decimal", "arg": "bad" } }`,
			wantValue: Decimal(0),
			wantErr:   errDecimal,
		},
		{
			name: "decimal/badJSON",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `bad`,
			wantValue: Decimal(0),
			wantErr:   errJSONDecode,
		},
		{
			name: "decimal/badFn",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "bad", "arg": "1234.5678" } }`,
			wantValue: Decimal(0),
			wantErr:   errJSONExtFnMatch,
		},
		{
			name: "decimal/ExtNotFound",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ }`,
			wantValue: Decimal(0),
			wantErr:   errJSONExtNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotValue, gotErr := tt.f([]byte(tt.in))
			testutilEquals(t, gotValue, tt.wantValue)
			assertError(t, gotErr, tt.wantErr)
		})
	}
}

func TestJSONMarshal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		in          Value
		outExplicit string
		outImplicit string
	}{
		{"record", Record{
			"ak": String("av"),
			"ck": String("cv"),
			"bk": String("bv"),
		}, `{"ak":"av","bk":"bv","ck":"cv"}`, `{"ak":"av","bk":"bv","ck":"cv"}`},
		{"recordWithExt", Record{
			"ip": mustIPValue("222.222.222.7"),
		}, `{"ip":{"__extn":{"fn":"ip","arg":"222.222.222.7"}}}`, `{"ip":{"__extn":{"fn":"ip","arg":"222.222.222.7"}}}`},
		{"set", Set{
			String("av"),
			String("cv"),
			String("bv"),
		}, `["av","cv","bv"]`, `["av","cv","bv"]`},
		{"setWithExt", Set{mustIPValue("222.222.222.7")},
			`[{"__extn":{"fn":"ip","arg":"222.222.222.7"}}]`, `[{"__extn":{"fn":"ip","arg":"222.222.222.7"}}]`},
		{"entity", EntityUID{"User", "alice"}, `{"__entity":{"type":"User","id":"alice"}}`, `{"type":"User","id":"alice"}`},
		{"ip", mustIPValue("222.222.222.7"), `{"__extn":{"fn":"ip","arg":"222.222.222.7"}}`, `"222.222.222.7"`},
		{"decimal", mustDecimalValue("33.57"), `{"__extn":{"fn":"decimal","arg":"33.57"}}`, `"33.57"`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			outExplicit, err := tt.in.ExplicitMarshalJSON()
			testutilOK(t, err)
			testutilEquals(t, string(outExplicit), tt.outExplicit)
			outImplicit, err := json.Marshal(tt.in)
			testutilOK(t, err)
			testutilEquals(t, string(outImplicit), tt.outImplicit)
		})
	}
}

type jsonErr struct{}

func (j *jsonErr) String() string                       { return "" }
func (j *jsonErr) Cedar() string                        { return "" }
func (j *jsonErr) equal(Value) bool                     { return false }
func (j *jsonErr) ExplicitMarshalJSON() ([]byte, error) { return nil, fmt.Errorf("jsonErr") }
func (j *jsonErr) typeName() string                     { return "jsonErr" }
func (j *jsonErr) deepClone() Value                     { return nil }

func TestJSONSet(t *testing.T) {
	t.Parallel()
	t.Run("UnmarshalErr", func(t *testing.T) {
		t.Parallel()
		var s Set
		err := json.Unmarshal([]byte(`[{"__extn":{"fn":"err"}}]`), &s)
		testutilError(t, err)
	})
	t.Run("MarshalErr", func(t *testing.T) {
		t.Parallel()
		s := Set{&jsonErr{}}
		_, err := json.Marshal(s)
		testutilError(t, err)
	})
}

func TestJSONRecord(t *testing.T) {
	t.Parallel()
	t.Run("UnmarshalErr", func(t *testing.T) {
		t.Parallel()
		var r Record
		err := json.Unmarshal([]byte(`{"key":{"__extn":{"fn":"err"}}}`), &r)
		testutilError(t, err)
	})
	t.Run("MarshalKeyErrImpossible", func(t *testing.T) {
		t.Parallel()
		r := Record{}
		k := []byte{0xde, 0x01}
		r[string(k)] = Boolean(false)
		v, err := json.Marshal(r)
		// this demonstrates that invalid keys will still result in json
		testutilEquals(t, string(v), `{"\ufffd\u0001":false}`)
		testutilOK(t, err)
	})
	t.Run("MarshalValueErr", func(t *testing.T) {
		t.Parallel()
		r := Record{"key": &jsonErr{}}
		_, err := json.Marshal(r)
		testutilError(t, err)
	})
}

func TestEntitiesJSON(t *testing.T) {
	t.Parallel()
	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()
		e := Entities{}
		ent := Entity{
			UID:        NewEntityUID("Type", "id"),
			Parents:    []EntityUID{},
			Attributes: Record{"key": Long(42)},
		}
		e[ent.UID] = ent
		b, err := e.MarshalJSON()
		testutilOK(t, err)
		testutilEquals(t, string(b), `[{"uid":{"type":"Type","id":"id"},"attrs":{"key":42}}]`)
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()
		b := []byte(`[{"uid":{"type":"Type","id":"id"},"parents":[],"attrs":{"key":42}}]`)
		var e Entities
		err := json.Unmarshal(b, &e)
		testutilOK(t, err)
		want := Entities{}
		ent := Entity{
			UID:        NewEntityUID("Type", "id"),
			Parents:    []EntityUID{},
			Attributes: Record{"key": Long(42)},
		}
		want[ent.UID] = ent
		testutilEquals(t, e, want)
	})

	t.Run("UnmarshalErr", func(t *testing.T) {
		t.Parallel()
		var e Entities
		err := e.UnmarshalJSON([]byte(`!@#$`))
		testutilError(t, err)
	})
}

func TestJSONEffect(t *testing.T) {
	t.Parallel()
	t.Run("MarshalPermit", func(t *testing.T) {
		t.Parallel()
		e := Permit
		b, err := e.MarshalJSON()
		testutilOK(t, err)
		testutilEquals(t, string(b), `"permit"`)
	})
	t.Run("MarshalForbid", func(t *testing.T) {
		t.Parallel()
		e := Forbid
		b, err := e.MarshalJSON()
		testutilOK(t, err)
		testutilEquals(t, string(b), `"forbid"`)
	})
	t.Run("UnmarshalPermit", func(t *testing.T) {
		t.Parallel()
		var e Effect
		err := json.Unmarshal([]byte(`"permit"`), &e)
		testutilOK(t, err)
		testutilEquals(t, e, Permit)
	})
	t.Run("UnmarshalForbid", func(t *testing.T) {
		t.Parallel()
		var e Effect
		err := json.Unmarshal([]byte(`"forbid"`), &e)
		testutilOK(t, err)
		testutilEquals(t, e, Forbid)
	})
}

func TestJSONDecision(t *testing.T) {
	t.Parallel()
	t.Run("MarshalAllow", func(t *testing.T) {
		t.Parallel()
		d := Allow
		b, err := d.MarshalJSON()
		testutilOK(t, err)
		testutilEquals(t, string(b), `"allow"`)
	})
	t.Run("MarshalDeny", func(t *testing.T) {
		t.Parallel()
		d := Deny
		b, err := d.MarshalJSON()
		testutilOK(t, err)
		testutilEquals(t, string(b), `"deny"`)
	})
	t.Run("UnmarshalAllow", func(t *testing.T) {
		t.Parallel()
		var d Decision
		err := json.Unmarshal([]byte(`"allow"`), &d)
		testutilOK(t, err)
		testutilEquals(t, d, Allow)
	})
	t.Run("UnmarshalDeny", func(t *testing.T) {
		t.Parallel()
		var d Decision
		err := json.Unmarshal([]byte(`"deny"`), &d)
		testutilOK(t, err)
		testutilEquals(t, d, Deny)
	})
}
