package json

import (
	"bytes"
	"errors"
	"fmt"

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
	return p.UnmarshalJSONMode(b, ModeGo)
}

func (p *PolicySetJSON) UnmarshalJSONMode(b []byte, m Mode) error {
	var j struct {
		StaticPolicies jsontext.Value    `json:"staticPolicies"`
		Templates      templatesJSON     `json:"templates"`
		TemplateLinks  templateLinksJSON `json:"templateLinks"`
	}
	if err := json.Unmarshal(b, &j, json.RejectUnknownMembers(true), jsontext.AllowDuplicateNames(m != ModeGoLegacy)); err != nil {
		return err
	}
	if j.StaticPolicies == nil || j.StaticPolicies.Kind() == jsontext.KindNull {
		return errors.New(`"staticPolicies" is missing or null`)
	}
	if len(j.Templates) > 0 || len(j.TemplateLinks) > 0 {
		return errors.New("templates and template links are not supported")
	}
	policies, err := unmarshalStaticPolicies(j.StaticPolicies, m)
	if err != nil {
		return err
	}
	*p = PolicySetJSON{StaticPolicies: &policies, Templates: j.Templates, TemplateLinks: j.TemplateLinks}
	return nil
}

func unmarshalStaticPolicies(b jsontext.Value, m Mode) (PolicySet, error) {
	dec := jsontext.NewDecoder(bytes.NewReader(b), jsontext.AllowDuplicateNames(true))
	if tok, err := dec.ReadToken(); err != nil {
		return nil, err
	} else if tok.Kind() != jsontext.KindBeginObject {
		return nil, fmt.Errorf(`"staticPolicies" must be an object, got %v`, tok.Kind())
	}
	policies := PolicySet{}
	for dec.PeekKind() != jsontext.KindEndObject {
		tok, err := dec.ReadToken()
		if err != nil {
			return nil, err
		}
		id := tok.String()
		if _, ok := policies[id]; ok {
			return nil, fmt.Errorf("duplicate policy id %q", id)
		}
		v, err := dec.ReadValue()
		if err != nil {
			return nil, err
		}
		var policy Policy
		if err := policy.UnmarshalJSONMode(v, m); err != nil {
			return nil, fmt.Errorf("policy %q: %w", id, err)
		}
		policies[id] = policy
	}
	return policies, nil
}
