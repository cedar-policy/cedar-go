package json

import "github.com/cedar-policy/cedar-go/types"

type PolicySet map[string]*Policy

type TemplateSet map[string]*Policy

type LinkedPolicy struct {
	TemplateID string                                        `json:"templateId"`
	LinkID     string                                        `json:"newId"`
	Values     map[string]types.ImplicitlyMarshaledEntityUID `json:"values"`
}

type PolicySetJSON struct {
	StaticPolicies PolicySet      `json:"staticPolicies"`
	Templates      TemplateSet    `json:"templates"`
	TemplateLinks  []LinkedPolicy `json:"templateLinks,omitempty"`
}
