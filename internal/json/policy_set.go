package json

type PolicySet map[string]*Policy

type TemplateSet map[string]*Policy

type PolicySetJSON struct {
	StaticPolicies PolicySet   `json:"staticPolicies"`
	Templates      TemplateSet `json:"templates"`
}
