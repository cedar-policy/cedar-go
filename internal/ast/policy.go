package ast

import "github.com/cedar-policy/cedar-go/types"

type PolicySet map[string]PolicySetEntry

type PolicySetEntry struct {
	Policy   Policy
	Position Position
}

func (p PolicySetEntry) TmpGetAnnotations() map[string]string {
	res := make(map[string]string, len(p.Policy.Annotations))
	for _, e := range p.Policy.Annotations {
		res[string(e.Key)] = string(e.Value)
	}
	return res
}
func (p PolicySetEntry) TmpGetEffect() bool {
	return bool(p.Policy.Effect)
}

type AnnotationType struct {
	Key   types.String // TODO: review type
	Value types.String
}
type Condition bool

const (
	ConditionWhen   = true
	ConditionUnless = false
)

type ConditionType struct {
	Condition Condition
	Body      IsNode
}

type Effect bool

const (
	EffectPermit Effect = true
	EffectForbid Effect = false
)

type Policy struct {
	Effect      Effect
	Annotations []AnnotationType
	Principal   IsScopeNode
	Action      IsScopeNode
	Resource    IsScopeNode
	Conditions  []ConditionType
}

func newPolicy(effect Effect, annotations []AnnotationType) *Policy {
	return &Policy{
		Effect:      effect,
		Annotations: annotations,
		Principal:   Scope(NewPrincipalNode()).All(),
		Action:      Scope(NewActionNode()).All(),
		Resource:    Scope(NewResourceNode()).All(),
	}
}

func Permit() *Policy {
	return newPolicy(EffectPermit, nil)
}

func Forbid() *Policy {
	return newPolicy(EffectForbid, nil)
}

func (p *Policy) When(node Node) *Policy {
	p.Conditions = append(p.Conditions, ConditionType{Condition: ConditionWhen, Body: node.v})
	return p
}

func (p *Policy) Unless(node Node) *Policy {
	p.Conditions = append(p.Conditions, ConditionType{Condition: ConditionUnless, Body: node.v})
	return p
}
