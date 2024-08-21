package json

type PolicySet map[string]*Policy

type PolicySetJSON struct {
	StaticPolicies PolicySet `json:"staticPolicies"`
}
