package cedar

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	EntityOpEquals       = "=="
	EntityOpIs           = "is"
	EntityOpIn           = "in"
	EntityOpAll          = "All"
	EffectPermit         = "permit"
	EffectForbid         = "forbid"
	InValidEffectMessage = "not a valid Effect"
	InValidPrincipal     = "principal can't be null"
	InValidAction        = "action can't be null"
	InValidResource      = "resource can't be null"
	CrPrincipal          = "principal"
	CrAction             = "action"
	CrResource           = "resource"
)

type JsonPolicy struct {
	Effect    string            `json:"effect"`
	Principal *JsonEntityWithIn `json:"principal"`
	Action    *JsonEntity       `json:"action"`
	Resource  *JsonEntityWithIn `json:"resource"`
	Condition *Condition        `json:"conditions"`
	ID        string            `json:"id"`
}

type JsonEntityWithIn struct {
	Op         string          `json:"op"`
	Entity     *JsonEntityBody `json:"entity"`
	EntityType string          `json:"entityType"`
	In         *JsonEntity     `json:"in"`
}

type JsonEntity struct {
	Op       string           `json:"op"`
	Entity   *JsonEntityBody  `json:"entity"`
	Entities []JsonEntityBody `json:"entities"`
}

type JsonEntityBody struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type Condition struct {
	Kind    string `json:"kind"`
	Content string `json:"content"` //expecting to capture condition string as-is and validate using cedar libs
}

type PolicyParser struct {
}

// JsonToPolicy converts a JSON representation of the Policy to its textual counterpart.
// It doesn't check the syntax; please use cedar.NewPolicySet to validate output.
func (p *PolicyParser) JsonToPolicy(b []byte) (string, error) {
	var jp []JsonPolicy
	if err := json.Unmarshal(b, &jp); err != nil {
		return "", err
	}
	var policyStrings []string
	for _, policy := range jp {
		ps, err := convertSingleJsonPolicyToText(policy)
		if err != nil {
			return "", err
		}
		policyStrings = append(policyStrings, ps)
	}
	return strings.Join(policyStrings, "\n"), nil
}

func convertSingleJsonPolicyToText(jp JsonPolicy) (string, error) {
	var b strings.Builder

	// Handle the permit or forbid keyword
	if jp.Effect == EffectPermit {
		b.WriteString(EffectPermit)
	} else if jp.Effect == EffectForbid {
		b.WriteString(EffectForbid)
	} else {
		return "", errors.New(InValidEffectMessage)
	}

	b.WriteString("(")

	// Handle principal
	if jp.Principal == nil {
		return "", errors.New(InValidPrincipal)
	}
	principalStr := handleEntityWithIn(jp.Principal)
	if principalStr != "" {
		b.WriteString(fmt.Sprintf("%s %s", CrPrincipal, principalStr))
	}

	// Handle action
	if jp.Action == nil {
		return "", errors.New(InValidAction)
	}
	actionStr := handleEntity(jp.Action)
	if actionStr != "" {
		if principalStr != "" {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%s %s", CrAction, actionStr))
	}

	// Handle resource
	if jp.Resource == nil {
		return "", errors.New(InValidResource)
	}
	resourceStr := handleEntityWithIn(jp.Resource)
	if resourceStr != "" {
		if principalStr != "" || actionStr != "" {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%s %s", CrResource, resourceStr))
	}

	b.WriteString(")")

	// Handle conditions
	if jp.Condition != nil {
		b.WriteString(fmt.Sprintf(" %s { ", jp.Condition.Kind))
		b.WriteString(jp.Condition.Content)
		b.WriteString(" }")
	}

	return b.String() + ";", nil
}

// handleEntity converts an entity involved in a policy (action) to its textual representation.
func handleEntity(entity *JsonEntity) string {
	if entity.Op == EntityOpIn && len(entity.Entities) > 0 {
		var entities []string
		for _, e := range entity.Entities {
			entities = append(entities, fmt.Sprintf("%s::\"%s\"", e.Type, e.ID))
		}
		return fmt.Sprintf("%s [%s]", EntityOpIn, strings.Join(entities, ", "))
	} else if entity.Entity != nil {
		return fmt.Sprintf("%s %s::\"%s\"", entity.Op, entity.Entity.Type, entity.Entity.ID)
	}
	return ""
}

// handleEntityWithIn converts an entity involved in a policy (resource or principal) to its textual representation.
func handleEntityWithIn(jsonEntityWithIn *JsonEntityWithIn) string {
	var b strings.Builder
	if jsonEntityWithIn.Op == EntityOpAll {
		return ""
	}
	if jsonEntityWithIn.Op == EntityOpIs {
		b.WriteString(" is " + jsonEntityWithIn.EntityType)
	}
	if jsonEntityWithIn.Op == EntityOpEquals {
		b.WriteString(fmt.Sprintf("%s %s::\"%s\"", jsonEntityWithIn.Op, jsonEntityWithIn.Entity.Type, jsonEntityWithIn.Entity.ID))
	}
	if jsonEntityWithIn.In != nil {
		b.WriteString(fmt.Sprintf(" %s %s::\"%s\"", EntityOpIn, jsonEntityWithIn.In.Entity.Type, jsonEntityWithIn.In.Entity.ID))
	}
	return b.String()
}
