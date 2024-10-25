package parser

import (
  "fmt"
  "github.com/cedar-policy/cedar-go/internal/ast"
  "github.com/cedar-policy/cedar-go/types"
  "strings"
)

//todo: how to clone policy body to avoid modifying the original template?
//todo: add invariant len(slotEnv) == len(template.Slots)
func LinkTemplateToPolicy(template ast.Template, linkID string, slotEnv map[string]string) ast.LinkedPolicy {
  p := template.Body
  templateID, _ := findAnnotation(p, "id")

  p.Principal = linkScope(p.Principal, slotEnv)
  p.Resource = linkScope(p.Resource, slotEnv)

  p.Annotate("id", types.String(linkID))

  return ast.LinkedPolicy{
    TemplateID: templateID.Value.String(),
    LinkID:     linkID,
    Policy:     &p,
  }
}

func findAnnotation(p ast.Policy, key string) (ast.AnnotationType, bool) {
  identKey := types.Ident(key)

  for _, annotation := range p.Annotations {
    if annotation.Key == identKey {
      return annotation, true
    }
  }

  return ast.AnnotationType{}, false
}

func linkScope[T ast.IsScopeNode](scope T, slotEnv map[string]string) T {
  var linkedScope any

  switch t := any(scope).(type) {
  case ast.ScopeTypeAll:
    linkedScope = t
  case ast.ScopeTypeEq:
    t.Entity = resolveSlot(t.Entity, slotEnv)

    linkedScope = t
  case ast.ScopeTypeIn:
  case ast.ScopeTypeInSet:

  case ast.ScopeTypeIs:

  case ast.ScopeTypeIsIn:
  default:
    panic(fmt.Sprintf("unknown scope type %T", t))
  }

  return linkedScope.(T)
}

//todo: should we panic on this? or just trust that the interface is correct?
func resolveSlot(ef types.EntityReference, slotEnv map[string]string) types.EntityReference {
  switch e := ef.(type) {
  case types.EntityUID:
    return e
  case types.VariableSlot:
    return parseEntityUID(slotEnv[e.Name.String()])
  default:
    panic(fmt.Sprintf("unknown entity reference type %T", e))
  }
}

func parseEntityUID(euid string) types.EntityUID {
  typeIDseparator := strings.LastIndex(euid, "::")

  etype := euid[:typeIDseparator]
  eID := strings.Trim(euid[typeIDseparator+2:], `"`)

  return types.NewEntityUID(types.EntityType(etype), types.String(eID))
}
