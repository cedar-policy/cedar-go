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
		name string
		in   string
		want Value
		err  error
	}{
		{"entity-likeRecord", `{ "type": "User", "id": "alice" }`, NewRecord(RecordMap{"type": String("User"), "id": String("alice")}), nil},
		{"explicitEntity", `{ "__entity": { "type": "User", "id": "alice" } }`, EntityUID{Type: "User", ID: "alice"}, nil},
		{"explicitEntityExtraField", `{ "__entity": { "type": "User", "id": "alice" }, "extra": "stuff" }`, NewRecord(RecordMap{"__entity": NewRecord(RecordMap{"type": String("User"), "id": String("alice")}), "extra": String("stuff")}), nil},
		{"explicitLongEntity", `{ "__entity": { "type": "User::External", "id": "alice" } }`, EntityUID{Type: "User::External", ID: "alice"}, nil},
		{"invalidJSON", `!@#$`, zeroValue(), errJSONDecode},
		{"numericOverflow", "12341234123412341234", zeroValue(), errJSONLongOutOfRange},
		{"unsupportedNull", "null", zeroValue(), errJSONUnsupportedType},
		{"unsupportedNullInSet", "[1, null]", zeroValue(), errJSONUnsupportedType},
		{"unsupportedNullInRecord", `{"a": null}`, zeroValue(), errJSONUnsupportedType},
		{"explicitIP", `{ "__extn": { "fn": "ip", "arg": "222.222.222.7" } }`, testutil.Must(ParseIPAddr("222.222.222.7")), nil},
		{"explicitSubnet", `{ "__extn": { "fn": "ip", "arg": "192.168.0.0/16" } }`, testutil.Must(ParseIPAddr("192.168.0.0/16")), nil},
		{"explicitDecimal", `{ "__extn": { "fn": "decimal", "arg": "33.57" } }`, testutil.Must(ParseDecimal("33.57")), nil},
		{"explicitDatetime", `{ "__extn": { "fn": "datetime", "arg": "1970-01-01T00:00:01Z" } }`, testutil.Must(ParseDatetime("1970-01-01T00:00:01Z")), nil},
		{"explicitDuration", `{ "__extn": { "fn": "duration", "arg": "1d12h30m30s500ms" } }`, testutil.Must(ParseDuration("1d12h30m30s500ms")), nil},
		{"invalidExtension", `{ "__extn": { "fn": "asdf", "arg": "blah" } }`, zeroValue(), errJSONInvalidExtn},
		{"badIP", `{ "__extn": { "fn": "ip", "arg": "bad" } }`, zeroValue(), errIP},
		{"badDecimal", `{ "__extn": { "fn": "decimal", "arg": "bad" } }`, zeroValue(), errDecimal},
		{"badDatetime", `{ "__extn": { "fn": "datetime", "arg": "bad" } }`, zeroValue(), errDatetime},
		{"badDuration", `{ "__extn": { "fn": "duration", "arg": "bad" } }`, zeroValue(), errDuration},
		{"extnArgNotString", `{ "__extn": { "fn": "ip", "arg": 1 } }`, zeroValue(), errJSONInvalidExtn},
		{"extnArgExtn", `{ "__extn": { "fn": "ip", "arg": { "__extn": { "fn": "ip", "arg": "1.2.3.4" } } } }`, zeroValue(), errJSONInvalidExtn},
		{"invalidExtensionArgNotString", `{ "__extn": { "fn": "asdf", "arg": 1 } }`, zeroValue(), errJSONInvalidExtn},
		{"extnFnNotString", `{ "__extn": { "fn": 1, "arg": "1.2.3.4" } }`, NewRecord(RecordMap{"__extn": NewRecord(RecordMap{"fn": Long(1), "arg": String("1.2.3.4")})}), nil},
		{"entityNotRecord", `{ "__entity": 1 }`, NewRecord(RecordMap{"__entity": Long(1)}), nil},
		{"entityMissingType", `{ "__entity": { "id": "alice" } }`, NewRecord(RecordMap{"__entity": NewRecord(RecordMap{"id": String("alice")})}), nil},
		{"entityMissingID", `{ "__entity": { "type": "User" } }`, NewRecord(RecordMap{"__entity": NewRecord(RecordMap{"type": String("User")})}), nil},
		{"entityTypeNotString", `{ "__entity": { "type": 1, "id": "alice" } }`, NewRecord(RecordMap{"__entity": NewRecord(RecordMap{"type": Long(1), "id": String("alice")})}), nil},
		{"entityIDNotString", `{ "__entity": { "type": "User", "id": 1 } }`, NewRecord(RecordMap{"__entity": NewRecord(RecordMap{"type": String("User"), "id": Long(1)})}), nil},
		{"entityExtraInnerKey", `{ "__entity": { "type": "User", "id": "alice", "x": 1 } }`, EntityUID{Type: "User", ID: "alice"}, nil},
		{"extnNotRecord", `{ "__extn": 1 }`, NewRecord(RecordMap{"__extn": Long(1)}), nil},
		{"extnMissingFn", `{ "__extn": { "arg": "1.2.3.4" } }`, NewRecord(RecordMap{"__extn": NewRecord(RecordMap{"arg": String("1.2.3.4")})}), nil},
		{"extnMissingArg", `{ "__extn": { "fn": "ip" } }`, NewRecord(RecordMap{"__extn": NewRecord(RecordMap{"fn": String("ip")})}), nil},
		{"set", `[42]`, NewSet(Long(42)), nil},
		{"record", `{"a":"b"}`, NewRecord(RecordMap{"a": String("b")}), nil},
		{"bool", `false`, Boolean(false), nil},
		{"surroundingWhitespace", " [42] ", NewSet(Long(42)), nil},
		{"trailingData", `[42] 43`, zeroValue(), errJSONDecode},
		{"duplicateKey", `{"a":"b","a":"c"}`, zeroValue(), errJSONDecode},
		{"nestedDuplicateKey", `{"a":{"b":1,"b":2}}`, zeroValue(), errJSONDecode},
		{"invalidUTF8", "\"\xff\"", zeroValue(), errJSONDecode},
		{"invalidUTF8Key", "{\"\xff\":1}", zeroValue(), errJSONDecode},
		{"nonIntegerNumber", "1.5", zeroValue(), errJSONLongOutOfRange},
		{"negativeZero", "-0", zeroValue(), errJSONLongOutOfRange},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got Value
			ptr := &got
			err := UnmarshalJSON([]byte(tt.in), ptr)
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, got, tt.want)
			if tt.err != nil {
				return
			}

			// Now assert that when we Marshal/Unmarshal that value, we still
			// have what we started with
			gotJSON, err := json.Marshal(ptr)
			testutil.OK(t, err)
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
