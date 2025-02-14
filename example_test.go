package cedar_test

import (
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go"
)

const policyCedar = `permit (
	principal == User::"alice",
	action == Action::"view",
	resource in Album::"jane_vacation"
  );
`

const entitiesJSON = `[
  {
    "uid": { "type": "User", "id": "alice" },
    "attrs": { "age": 18 },
    "parents": []
  },
  {
    "uid": { "type": "Photo", "id": "VacationPhoto94.jpg" },
    "attrs": {},
    "parents": [{ "type": "Album", "id": "jane_vacation" }]
  }
]`

func Example() {
	var policy cedar.Policy
	if err := policy.UnmarshalCedar([]byte(policyCedar)); err != nil {
		fmt.Println("policy unmarshal error:", err)
		return
	}

	ps := cedar.NewPolicySet()
	ps.Add("policy0", &policy)

	var entities cedar.EntityMap
	if err := json.Unmarshal([]byte(entitiesJSON), &entities); err != nil {
		fmt.Println("entity unmarshal error:", err)
		return
	}

	req := cedar.Request{
		Principal: cedar.NewEntityUID("User", "alice"),
		Action:    cedar.NewEntityUID("Action", "view"),
		Resource:  cedar.NewEntityUID("Photo", "VacationPhoto94.jpg"),
		Context: cedar.NewRecord(cedar.RecordMap{
			"demoRequest": cedar.True,
		}),
	}

	ok, _ := ps.IsAuthorized(entities, req)
	fmt.Println(ok)
	// Output:
	// allow
}
