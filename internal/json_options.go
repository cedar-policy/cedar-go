package internal

import (
	jsonv1 "encoding/json"
	"encoding/json/v2"
)

// JSONOptions are the encoding/json/v2 options for decoding Cedar's JSON formats. They keep
// encoding/json's (v1) semantics, except that object member names are matched case-sensitively,
// as the Cedar JSON formats and the Rust implementation require. A member whose name differs from
// a field's only by case is treated as an unknown member.
var JSONOptions = json.JoinOptions(jsonv1.DefaultOptionsV1(), json.MatchCaseInsensitiveNames(false))
