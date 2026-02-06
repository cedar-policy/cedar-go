package schema_test

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/schema"
)

const exampleCedar = `entity User in [Group] {
	name: String,
	age?: Long
};

entity Group;

entity Photo {
	owner: User,
	tags: Set<String>
};

action viewPhoto appliesTo {
	principal: User,
	resource: Photo,
	context: {}
};
`

func ExampleSchema() {
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(exampleCedar)); err != nil {
		fmt.Println("schema parse error:", err)
		return
	}

	resolved, err := s.Resolve()
	if err != nil {
		fmt.Println("schema resolve error:", err)
		return
	}

	for entityType := range resolved.Entities {
		fmt.Println("entity:", entityType)
	}
	for actionUID := range resolved.Actions {
		fmt.Println("action:", actionUID)
	}
	// Unordered output:
	// entity: User
	// entity: Group
	// entity: Photo
	// action: Action::"viewPhoto"
}
