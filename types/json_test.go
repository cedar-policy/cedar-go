package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func zeroValue() Value {
	return nil
}

func AssertValue(t *testing.T, got, want Value) {
	t.Helper()
	testutil.FatalIf(
		t,
		(got != zeroValue() || want != zeroValue()) && (got == zeroValue() || want == zeroValue() || !got.Equal(want)),
		"got %v want %v", got, want)
}

func TestJSON_Value(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		in           string
		want         Value
		unmarshalErr error
		marshalErr   error
	}{
		{"entity-likeRecord", `{ "type": "User", "id": "alice" }`, NewRecord(RecordMap{"type": String("User"), "id": String("alice")}), nil, nil},
		{"explicitEntity", `{ "__entity": { "type": "User", "id": "alice" } }`, EntityUID{Type: "User", ID: "alice"}, nil, nil},
		{"explicitEntityExtraField", `{ "__entity": { "type": "User", "id": "alice" }, "extra": "stuff" }`, NewRecord(RecordMap{"__entity": NewRecord(RecordMap{"type": String("User"), "id": String("alice")}), "extra": String("stuff")}), nil, errJSONReservedKey},
		{"explicitLongEntity", `{ "__entity": { "type": "User::External", "id": "alice" } }`, EntityUID{Type: "User::External", ID: "alice"}, nil, nil},
		{"explicitEntityInvalidType", `{ "__entity": { "type": "User ) permit ( principal", "id": "alice" } }`, zeroValue(), errJSONInvalidType, nil},
		{"explicitEntityReservedType", `{ "__entity": { "type": "User::if", "id": "alice" } }`, zeroValue(), errJSONInvalidType, nil},
		{"explicitEntityInvalidTypeInRecord", `{ "a": { "__entity": { "type": "A B", "id": "alice" } } }`, zeroValue(), errJSONInvalidType, nil},
		{"invalidJSON", `!@#$`, zeroValue(), errJSONDecode, nil},
		{"numericOverflow", "12341234123412341234", zeroValue(), errJSONLongOutOfRange, nil},
		{"unsupportedNull", "null", zeroValue(), errJSONUnsupportedType, nil},
		{"unsupportedNullInSet", "[1, null]", zeroValue(), errJSONUnsupportedType, nil},
		{"unsupportedNullInRecord", `{"a": null}`, zeroValue(), errJSONUnsupportedType, nil},
		{"explicitIP", `{ "__extn": { "fn": "ip", "arg": "222.222.222.7" } }`, testutil.Must(ParseIPAddr("222.222.222.7")), nil, nil},
		{"explicitSubnet", `{ "__extn": { "fn": "ip", "arg": "192.168.0.0/16" } }`, testutil.Must(ParseIPAddr("192.168.0.0/16")), nil, nil},
		{"explicitDecimal", `{ "__extn": { "fn": "decimal", "arg": "33.57" } }`, testutil.Must(ParseDecimal("33.57")), nil, nil},
		{"explicitDatetime", `{ "__extn": { "fn": "datetime", "arg": "1970-01-01T00:00:01Z" } }`, testutil.Must(ParseDatetime("1970-01-01T00:00:01Z")), nil, nil},
		{"explicitDuration", `{ "__extn": { "fn": "duration", "arg": "1d12h30m30s500ms" } }`, testutil.Must(ParseDuration("1d12h30m30s500ms")), nil, nil},
		{"invalidExtension", `{ "__extn": { "fn": "asdf", "arg": "blah" } }`, zeroValue(), errJSONInvalidExtn, nil},
		{"badIP", `{ "__extn": { "fn": "ip", "arg": "bad" } }`, zeroValue(), errIP, nil},
		{"badDecimal", `{ "__extn": { "fn": "decimal", "arg": "bad" } }`, zeroValue(), errDecimal, nil},
		{"badDatetime", `{ "__extn": { "fn": "datetime", "arg": "bad" } }`, zeroValue(), errDatetime, nil},
		{"badDuration", `{ "__extn": { "fn": "duration", "arg": "bad" } }`, zeroValue(), errDuration, nil},
		{"extnArgNotString", `{ "__extn": { "fn": "ip", "arg": 1 } }`, zeroValue(), errJSONInvalidExtn, nil},
		{"extnArgExtn", `{ "__extn": { "fn": "ip", "arg": { "__extn": { "fn": "ip", "arg": "1.2.3.4" } } } }`, zeroValue(), errJSONInvalidExtn, nil},
		{"invalidExtensionArgNotString", `{ "__extn": { "fn": "asdf", "arg": 1 } }`, zeroValue(), errJSONInvalidExtn, nil},
		{"extnFnNotString", `{ "__extn": { "fn": 1, "arg": "1.2.3.4" } }`, NewRecord(RecordMap{"__extn": NewRecord(RecordMap{"fn": Long(1), "arg": String("1.2.3.4")})}), nil, errJSONReservedKey},
		{"entityNotRecord", `{ "__entity": 1 }`, NewRecord(RecordMap{"__entity": Long(1)}), nil, errJSONReservedKey},
		{"entityMissingType", `{ "__entity": { "id": "alice" } }`, NewRecord(RecordMap{"__entity": NewRecord(RecordMap{"id": String("alice")})}), nil, errJSONReservedKey},
		{"entityMissingID", `{ "__entity": { "type": "User" } }`, NewRecord(RecordMap{"__entity": NewRecord(RecordMap{"type": String("User")})}), nil, errJSONReservedKey},
		{"entityTypeNotString", `{ "__entity": { "type": 1, "id": "alice" } }`, NewRecord(RecordMap{"__entity": NewRecord(RecordMap{"type": Long(1), "id": String("alice")})}), nil, errJSONReservedKey},
		{"entityIDNotString", `{ "__entity": { "type": "User", "id": 1 } }`, NewRecord(RecordMap{"__entity": NewRecord(RecordMap{"type": String("User"), "id": Long(1)})}), nil, errJSONReservedKey},
		{"entityExtraInnerKey", `{ "__entity": { "type": "User", "id": "alice", "x": 1 } }`, EntityUID{Type: "User", ID: "alice"}, nil, nil},
		{"extnNotRecord", `{ "__extn": 1 }`, NewRecord(RecordMap{"__extn": Long(1)}), nil, errJSONReservedKey},
		{"extnMissingFn", `{ "__extn": { "arg": "1.2.3.4" } }`, NewRecord(RecordMap{"__extn": NewRecord(RecordMap{"arg": String("1.2.3.4")})}), nil, errJSONReservedKey},
		{"extnMissingArg", `{ "__extn": { "fn": "ip" } }`, NewRecord(RecordMap{"__extn": NewRecord(RecordMap{"fn": String("ip")})}), nil, errJSONReservedKey},
		{"set", `[42]`, NewSet(Long(42)), nil, nil},
		{"record", `{"a":"b"}`, NewRecord(RecordMap{"a": String("b")}), nil, nil},
		{"bool", `false`, Boolean(false), nil, nil},
		{"surroundingWhitespace", " [42] ", NewSet(Long(42)), nil, nil},
		{"trailingData", `[42] 43`, zeroValue(), errJSONDecode, nil},
		{"duplicateKey", `{"a":"b","a":"c"}`, zeroValue(), errJSONDecode, nil},
		{"nestedDuplicateKey", `{"a":{"b":1,"b":2}}`, zeroValue(), errJSONDecode, nil},
		{"invalidUTF8", "\"\xff\"", zeroValue(), errJSONDecode, nil},
		{"invalidUTF8Key", "{\"\xff\":1}", zeroValue(), errJSONDecode, nil},
		{"nonIntegerNumber", "1.5", zeroValue(), errJSONLongOutOfRange, nil},
		{"negativeZero", "-0", zeroValue(), errJSONLongOutOfRange, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got Value
			ptr := &got
			err := UnmarshalJSON([]byte(tt.in), ptr)
			testutil.ErrorIs(t, err, tt.unmarshalErr)
			AssertValue(t, got, tt.want)
			if tt.unmarshalErr != nil {
				return
			}

			// Now assert that when we Marshal/Unmarshal that value, we still
			// have what we started with, unless marshaling it is an expected
			// error (e.g. marshaling reserved keys like __entity in a Record)
			gotJSON, err := json.Marshal(ptr)
			testutil.ErrorIs(t, err, tt.marshalErr)
			if tt.marshalErr != nil {
				return
			}
			var gotRetry Value
			ptr = &gotRetry
			err = UnmarshalJSON(gotJSON, ptr)
			testutil.OK(t, err)
			testutil.Equals(t, gotRetry, tt.want)
		})
	}
}

func Test_unmarshalExtensionValue(t *testing.T) {
	t.Parallel()

	type extType struct {
		extValue string
	}

	parseExtValue := func(s string) (extType, error) {
		return extType{s}, nil
	}
	parseErr := fmt.Errorf("failed to parse")

	tests := []struct {
		name      string
		in        string
		extName   string
		parse     func(string) (extType, error)
		wantValue extType
		wantErr   error
	}{
		{
			name:      "explicit",
			in:        `{ "__extn": { "fn": "extType", "arg": "value" } }`,
			extName:   "extType",
			parse:     parseExtValue,
			wantValue: extType{extValue: "value"},
			wantErr:   nil,
		},
		{
			name:      "implicit/string",
			in:        `"value"`,
			extName:   "extType",
			parse:     parseExtValue,
			wantValue: extType{extValue: "value"},
			wantErr:   nil,
		},
		{
			name:      "implicit/JSON",
			in:        `{ "fn": "extType", "arg": "value" }`,
			extName:   "extType",
			parse:     parseExtValue,
			wantValue: extType{extValue: "value"},
			wantErr:   nil,
		},
		{
			name:      "implicit/badString",
			in:        `"bad`,
			extName:   "extType",
			parse:     parseExtValue,
			wantValue: extType{},
			wantErr:   errJSONDecode,
		},
		{
			name:      "ip/implicit/badJSON",
			in:        `{ "fn": "extType", "arg": 1 }`,
			extName:   "extType",
			parse:     parseExtValue,
			wantValue: extType{},
			wantErr:   errJSONDecode,
		},
		{
			name:      "implicit/badFn",
			in:        `{"fn": "bad", "arg": "value"}`,
			extName:   "extType",
			parse:     parseExtValue,
			wantValue: extType{},
			wantErr:   errJSONExtFnMatch,
		},
		{
			name:      "badParse",
			in:        `{ "__extn": { "fn": "extType", "arg": "someBadString" } }`,
			extName:   "extType",
			parse:     func(string) (extType, error) { return extType{}, parseErr },
			wantValue: extType{},
			wantErr:   parseErr,
		},
		{
			name:      "badJSON",
			in:        "bad",
			extName:   "extType",
			parse:     parseExtValue,
			wantValue: extType{},
			wantErr:   errJSONDecode,
		},
		{
			name:      "badFn",
			in:        `{ "__extn": { "fn": "bad", "arg": "value" } }`,
			extName:   "extType",
			parse:     parseExtValue,
			wantValue: extType{},
			wantErr:   errJSONExtFnMatch,
		},
		{
			name:      "explicit/miscased",
			in:        `{ "__EXTN": { "fn": "extType", "arg": "value" } }`,
			extName:   "extType",
			parse:     parseExtValue,
			wantValue: extType{},
			wantErr:   errJSONExtNotFound,
		},
		{
			name:      "implicit/miscased",
			in:        `{ "FN": "extType", "ARG": "value" }`,
			extName:   "extType",
			parse:     parseExtValue,
			wantValue: extType{},
			wantErr:   errJSONExtNotFound,
		},
		{
			name:      "duplicateKey",
			in:        `{ "__extn": { "fn": "extType", "fn": "extType", "arg": "value" } }`,
			extName:   "extType",
			parse:     parseExtValue,
			wantValue: extType{},
			wantErr:   errJSONDecode,
		},
		{
			name:      "implicit/invalidUTF8",
			in:        "\"\xff\"",
			extName:   "extType",
			parse:     parseExtValue,
			wantValue: extType{},
			wantErr:   errJSONDecode,
		},
		{
			name:      "extNotFound",
			in:        "{ }",
			extName:   "extType",
			parse:     parseExtValue,
			wantValue: extType{},
			wantErr:   errJSONExtNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotValue, gotErr := unmarshalExtensionValue([]byte(tt.in), tt.extName, tt.parse)
			testutil.Equals(t, gotValue, tt.wantValue)
			testutil.ErrorIs(t, gotErr, tt.wantErr)
		})
	}
}

func TestJSONMarshal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   Value
		out  string
	}{
		{
			"record",
			NewRecord(RecordMap{
				"ak": String("av"),
				"ck": String("cv"),
				"bk": String("bv"),
			}),
			`{"ak":"av","bk":"bv","ck":"cv"}`,
		},
		{
			"recordWithExt",
			NewRecord(RecordMap{
				"ip": testutil.Must(ParseIPAddr("222.222.222.7")),
			}),
			`{"ip":{"__extn":{"fn":"ip","arg":"222.222.222.7"}}}`,
		},
		{
			"set",
			NewSet(
				String("av"),
				String("cv"),
				String("bv"),
			),
			`["cv","bv","av"]`,
		},
		{
			"entity",
			EntityUID{"User", "alice"},
			`{"__entity":{"type":"User","id":"alice"}}`,
		},
		{
			"ip",
			testutil.Must(ParseIPAddr("222.222.222.7")),
			`{"__extn":{"fn":"ip","arg":"222.222.222.7"}}`,
		},
		{
			"decimal",
			testutil.Must(ParseDecimal(("33.57"))),
			`{"__extn":{"fn":"decimal","arg":"33.57"}}`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := json.Marshal(tt.in)
			testutil.OK(t, err)
			testutil.Equals(t, string(out), tt.out)
		})
	}
}

type jsonErr struct{}

func (j *jsonErr) String() string               { return "" }
func (j *jsonErr) MarshalCedar() []byte         { return nil }
func (j *jsonErr) Equal(Value) bool             { return false }
func (j *jsonErr) MarshalJSON() ([]byte, error) { return nil, fmt.Errorf("jsonErr") }
func (j *jsonErr) hash() uint64                 { return 0 }

func TestJSONSet(t *testing.T) {
	t.Parallel()
	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()
		var s Set
		err := json.Unmarshal([]byte(`[1, "a", {"__entity": {"type": "User", "id": "alice"}}, 1]`), &s)
		testutil.OK(t, err)
		testutil.Equals(t, s, NewSet(Long(1), String("a"), EntityUID{Type: "User", ID: "alice"}))
	})
	t.Run("UnmarshalErr", func(t *testing.T) {
		t.Parallel()
		var s Set
		err := json.Unmarshal([]byte(`{"not":["a","set"]}`), &s)
		testutil.Error(t, err)
	})
	t.Run("UnmarshalDuplicateKeyErr", func(t *testing.T) {
		t.Parallel()
		var s Set
		err := json.Unmarshal([]byte(`[{"a":1,"a":2}]`), &s)
		testutil.Error(t, err)
	})
	t.Run("UnmarshalInvalidUTF8Err", func(t *testing.T) {
		t.Parallel()
		var s Set
		err := json.Unmarshal([]byte("[\"\xff\"]"), &s)
		testutil.Error(t, err)
	})
	t.Run("MarshalErr", func(t *testing.T) {
		t.Parallel()
		s := NewSet(&jsonErr{})
		_, err := json.Marshal(s)
		testutil.Error(t, err)
	})
}

func TestJSONRecord(t *testing.T) {
	t.Parallel()
	t.Run("UnmarshalErr", func(t *testing.T) {
		t.Parallel()
		var r Record
		err := json.Unmarshal([]byte(`["not", "a", "record"]`), &r)
		testutil.Error(t, err)
	})
	t.Run("UnmarshalDuplicateKeyErr", func(t *testing.T) {
		t.Parallel()
		var r Record
		err := json.Unmarshal([]byte(`{"a":1,"a":2}`), &r)
		testutil.Error(t, err)
	})
	t.Run("UnmarshalInvalidUTF8KeyErr", func(t *testing.T) {
		t.Parallel()
		var r Record
		err := json.Unmarshal([]byte("{\"\xff\":1}"), &r)
		testutil.Error(t, err)
	})
	// As in Rust Cedar, an object that decodes as an EntityUID or an
	// extension value isn't a Record.
	t.Run("UnmarshalEscapeErr", func(t *testing.T) {
		t.Parallel()
		for _, in := range []string{
			`{"__entity":{"type":"User","id":"alice"}}`,
			`{"__extn":{"fn":"ip","arg":"1.2.3.4"}}`,
			`{"__extn":{"fn":"unknown","arg":"x"}}`,
		} {
			var r Record
			testutil.Error(t, json.Unmarshal([]byte(in), &r))
		}
	})
	// An object with members other than "__entity" is an ordinary Record.
	t.Run("UnmarshalEscapeShaped", func(t *testing.T) {
		t.Parallel()
		var r Record
		testutil.OK(t, json.Unmarshal([]byte(`{"__entity":{"type":"User","id":"alice"},"x":1}`), &r))
		testutil.Equals(t, r, NewRecord(RecordMap{
			"__entity": NewRecord(RecordMap{"type": String("User"), "id": String("alice")}),
			"x":        Long(1),
		}))
	})
	// As in Rust Cedar, null isn't a Record, e.g. for a request's context.
	t.Run("UnmarshalNullErr", func(t *testing.T) {
		t.Parallel()
		var r Record
		testutil.Error(t, json.Unmarshal([]byte(`null`), &r))
		var req Request
		testutil.Error(t, json.Unmarshal([]byte(`{"principal":{"type":"User","id":"a"},"action":{"type":"Action","id":"a"},"resource":{"type":"User","id":"a"},"context":null}`), &req))
	})
	t.Run("UnmarshalEntityAttrsEscapeErr", func(t *testing.T) {
		t.Parallel()
		var e Entity
		testutil.Error(t, json.Unmarshal([]byte(`{"uid":{"type":"User","id":"alice"},"parents":[],"attrs":{"__entity":{"type":"User","id":"bob"}}}`), &e))
	})
	t.Run("MarshalKeyErr", func(t *testing.T) {
		t.Parallel()
		k := []byte{0xde, 0x01}
		m := RecordMap{String(k): Boolean(false)}
		r := NewRecord(m)
		_, err := json.Marshal(r)
		testutil.Error(t, err)
	})
	t.Run("MarshalValueErr", func(t *testing.T) {
		t.Parallel()
		r := NewRecord(RecordMap{"key": &jsonErr{}})
		_, err := json.Marshal(r)
		testutil.Error(t, err)
	})
}
