package json

import (
	"errors"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

type PolicySet map[string]Policy

// PolicySetJSON is the JSON form of a policy set. cedar-go doesn't support
// templates, so "templates" and "templateLinks" are always written empty. Rust
// Cedar requires both.
type PolicySetJSON struct {
	StaticPolicies *PolicySet        `json:"staticPolicies"`
	Templates      templatesJSON     `json:"templates"`
	TemplateLinks  templateLinksJSON `json:"templateLinks"`
}

type templatesJSON map[string]jsontext.Value

type templateLinksJSON []jsontext.Value

func (t *templatesJSON) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return unmarshalNonNull(dec, (*map[string]jsontext.Value)(t))
}

func (t *templateLinksJSON) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return unmarshalNonNull(dec, (*[]jsontext.Value)(t))
}

func (p *PolicySetJSON) MarshalJSON() ([]byte, error) {
	type policySetJSON PolicySetJSON
	return json.Marshal((*policySetJSON)(p), json.Deterministic(true))
}

// UnmarshalJSON decodes a policy set. As in Rust Cedar, "staticPolicies" is
// required, and unknown members and null are rejected. Rust Cedar also
// requires "templates" and "templateLinks", but they may be missing here
// because cedar-go previously omitted them. Since cedar-go doesn't support
// templates, they must be empty: ignoring them would silently drop
// template-linked policies.
func (p *PolicySetJSON) UnmarshalJSON(b []byte) error {
	type policySetJSON PolicySetJSON
	var j policySetJSON
	if err := json.Unmarshal(b, &j, json.RejectUnknownMembers(true)); err != nil {
		return err
	}
	if j.StaticPolicies == nil {
		return errors.New(`"staticPolicies" is missing or null`)
	}
	if len(j.Templates) > 0 || len(j.TemplateLinks) > 0 {
		return errors.New("templates and template links are not supported")
	}
	*p = PolicySetJSON(j)
	return nil
}
