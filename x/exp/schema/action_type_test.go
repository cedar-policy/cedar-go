package schema_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

func TestResolveActionTypeReferences(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name        string
		cedar       string
		json        string
		entityType  types.EntityType
		wantType    resolved.IsType
		wantActions []types.EntityUID
	}{
		{
			name:        "root set",
			cedar:       `entity Staff { grants: Set<Action> }; action "read";`,
			json:        `{"":{"entityTypes":{"Staff":{"shape":{"type":"Record","attributes":{"grants":{"type":"Set","element":{"type":"Entity","name":"Action"}}}}}},"actions":{"read":{}}}}`,
			entityType:  "Staff",
			wantType:    resolved.SetType{Element: resolved.EntityType("Action")},
			wantActions: []types.EntityUID{types.NewEntityUID("Action", "read")},
		},
		{
			name:        "current namespace before root",
			cedar:       `action "root"; namespace App { entity Staff { grants: Action }; action "read"; }`,
			json:        `{"":{"actions":{"root":{}},"entityTypes":{}},"App":{"entityTypes":{"Staff":{"shape":{"type":"Record","attributes":{"grants":{"type":"Entity","name":"Action"}}}}},"actions":{"read":{}}}}`,
			entityType:  "App::Staff",
			wantType:    resolved.EntityType("App::Action"),
			wantActions: []types.EntityUID{types.NewEntityUID("Action", "root"), types.NewEntityUID("App::Action", "read")},
		},
		{
			name:        "root fallback",
			cedar:       `action "read"; namespace App { entity Staff { grants: Action }; }`,
			json:        `{"":{"actions":{"read":{}},"entityTypes":{}},"App":{"entityTypes":{"Staff":{"shape":{"type":"Record","attributes":{"grants":{"type":"Entity","name":"Action"}}}}},"actions":{}}}`,
			entityType:  "App::Staff",
			wantType:    resolved.EntityType("Action"),
			wantActions: []types.EntityUID{types.NewEntityUID("Action", "read")},
		},
		{
			name:        "qualified foreign namespace",
			cedar:       `namespace App { entity Staff { grants: Other::Action }; action "local"; } namespace Other { action "read"; }`,
			json:        `{"App":{"entityTypes":{"Staff":{"shape":{"type":"Record","attributes":{"grants":{"type":"Entity","name":"Other::Action"}}}}},"actions":{"local":{}}},"Other":{"actions":{"read":{}},"entityTypes":{}}}`,
			entityType:  "App::Staff",
			wantType:    resolved.EntityType("Other::Action"),
			wantActions: []types.EntityUID{types.NewEntityUID("App::Action", "local"), types.NewEntityUID("Other::Action", "read")},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, format := range []struct {
				name   string
				source string
				parse  func(*schema.Schema, []byte) error
			}{
				{"Cedar", tt.cedar, (*schema.Schema).UnmarshalCedar},
				{"JSON", tt.json, (*schema.Schema).UnmarshalJSON},
			} {
				t.Run(format.name, func(t *testing.T) {
					t.Parallel()
					var s schema.Schema
					testutil.OK(t, format.parse(&s, []byte(format.source)))
					r, err := s.Resolve()
					testutil.OK(t, err)
					testutil.Equals(t, r.Entities[tt.entityType].Shape["grants"].Type, tt.wantType)
					testutil.Equals(t, len(r.Entities), 1)
					testutil.Equals(t, len(r.Actions), len(tt.wantActions))
					for _, uid := range tt.wantActions {
						testutil.Equals(t, r.Actions[uid].Entity.UID, uid)
					}
				})
			}
		})
	}
}

func TestResolveUndeclaredActionTypeReferences(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name  string
		cedar string
		json  string
	}{
		{
			name:  "no actions",
			cedar: `entity Staff { grants: Action };`,
			json:  `{"":{"entityTypes":{"Staff":{"shape":{"type":"Record","attributes":{"grants":{"type":"Entity","name":"Action"}}}}},"actions":{}}}`,
		},
		{
			name:  "qualified namespace without actions",
			cedar: `action "read"; namespace Empty {} entity Staff { grants: Empty::Action };`,
			json:  `{"":{"entityTypes":{"Staff":{"shape":{"type":"Record","attributes":{"grants":{"type":"Entity","name":"Empty::Action"}}}}},"actions":{"read":{}}},"Empty":{"actions":{},"entityTypes":{}}}`,
		},
		{
			name:  "foreign actions do not declare root type",
			cedar: `namespace Other { action "read"; } namespace App { entity Staff { grants: Action }; }`,
			json:  `{"Other":{"actions":{"read":{}},"entityTypes":{}},"App":{"entityTypes":{"Staff":{"shape":{"type":"Record","attributes":{"grants":{"type":"Entity","name":"Action"}}}}},"actions":{}}}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, format := range []struct {
				name   string
				source string
				parse  func(*schema.Schema, []byte) error
			}{
				{"Cedar", tt.cedar, (*schema.Schema).UnmarshalCedar},
				{"JSON", tt.json, (*schema.Schema).UnmarshalJSON},
			} {
				t.Run(format.name, func(t *testing.T) {
					t.Parallel()
					var s schema.Schema
					testutil.OK(t, format.parse(&s, []byte(format.source)))
					_, err := s.Resolve()
					testutil.Error(t, err)
				})
			}
		})
	}
}

func TestResolveActionTypeCommonTypePrecedence(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		cedar    string
		wantType resolved.IsType
	}{
		{
			name:     "local common type before local action type",
			cedar:    `namespace App { type Action = Long; entity Staff { grants: Action }; action "read"; }`,
			wantType: resolved.LongType{},
		},
		{
			name:     "local action type before root common type",
			cedar:    `type Action = Long; namespace App { entity Staff { grants: Action }; action "read"; }`,
			wantType: resolved.EntityType("App::Action"),
		},
		{
			name:     "local action type prevents false common type cycle",
			cedar:    `type Action = App::T; namespace App { action "read"; type T = Action; entity Staff { grants: T }; }`,
			wantType: resolved.EntityType("App::Action"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var s schema.Schema
			testutil.OK(t, s.UnmarshalCedar([]byte(tt.cedar)))
			r, err := s.Resolve()
			testutil.OK(t, err)
			testutil.Equals(t, r.Entities["App::Staff"].Shape["grants"].Type, tt.wantType)
		})
	}
}

func TestResolveActionTypeForbiddenEntityRoles(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name  string
		cedar string
		json  string
	}{
		{
			name:  "entity parent",
			cedar: `entity Staff in Action; action "read";`,
			json:  `{"":{"entityTypes":{"Staff":{"memberOfTypes":["Action"]}},"actions":{"read":{}}}}`,
		},
		{
			name:  "action principal",
			cedar: `entity Endpoint; action "read" appliesTo { principal: Action, resource: Endpoint };`,
			json:  `{"":{"entityTypes":{"Endpoint":{}},"actions":{"read":{"appliesTo":{"principalTypes":["Action"],"resourceTypes":["Endpoint"]}}}}}`,
		},
		{
			name:  "action resource",
			cedar: `entity Staff; action "read" appliesTo { principal: Staff, resource: Action };`,
			json:  `{"":{"entityTypes":{"Staff":{}},"actions":{"read":{"appliesTo":{"principalTypes":["Staff"],"resourceTypes":["Action"]}}}}}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, format := range []struct {
				name   string
				source string
				parse  func(*schema.Schema, []byte) error
			}{
				{"Cedar", tt.cedar, (*schema.Schema).UnmarshalCedar},
				{"JSON", tt.json, (*schema.Schema).UnmarshalJSON},
			} {
				t.Run(format.name, func(t *testing.T) {
					t.Parallel()
					var s schema.Schema
					testutil.OK(t, format.parse(&s, []byte(format.source)))
					_, err := s.Resolve()
					testutil.Error(t, err)
				})
			}
		})
	}
}
