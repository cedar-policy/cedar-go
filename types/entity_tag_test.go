package types_test

import (
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestEntityTagMarshalJSON(t *testing.T) {
	t.Parallel()
	e := types.Entity{
		UID:        types.NewEntityUID("FooType", "1"),
		Parents:    types.NewEntityUIDSet(),
		Attributes: types.Record{},
		Tags: cedar.NewRecord(types.RecordMap{
			"key":      types.String("value"),
			"entity":   types.NewEntityUID("FootType", "1"),
			"datetime": types.NewDatetime(time.Unix(0, 0)),
		}),
	}

	testutil.JSONMarshalsTo(t, e,
		`{
			"uid": {"type":"FooType","id":"1"},
			"parents": [],
			"attrs":{},
            "tags": {
        		"datetime": {
        			"__extn": {
        				"fn": "datetime",
        				"arg": "1970-01-01T00:00:00.000Z"
        			}
        		},
        		"entity": {
        			"__entity": {
        				"type": "FootType",
        				"id": "1"
        			}
        		},
        		"key": "value"
            }
		}`)
}
