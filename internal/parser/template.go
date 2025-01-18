package parser

import (
    "fmt"
    "github.com/cedar-policy/cedar-go/internal/ast"
    "github.com/cedar-policy/cedar-go/types"
    "strings"
)

func LinkTemplateToPolicy(template ast.Template, linkID string, slotEnv map[string]string) (ast.LinkedPolicy, error) {
    body := template.ClonePolicy()
    if len(body.Slots()) != len(slotEnv) {
        return ast.LinkedPolicy{}, fmt.Errorf("slot env length %d does not match template slot length %d", len(slotEnv), len(body.Slots()))
    }

    templateID, _ := findAnnotation(body, "id")

    for _, slot := range body.Slots() {
        switch slot {
        case types.PrincipalSlot:
            body.Principal = linkScope(template.Principal, slotEnv)
        case types.ResourceSlot:
            body.Resource = linkScope(template.Resource, slotEnv)
        }
    }

    body.Annotate("id", types.String(linkID))

    return ast.LinkedPolicy{
        TemplateID: templateID.Value.String(),
        LinkID:     linkID,
        Policy:     &body,
    }, nil
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
    var linkedScope any = scope

    switch t := any(scope).(type) {
    case ast.ScopeTypeEq:
        t.Entity = resolveSlot(t.Entity, slotEnv)

        linkedScope = t
    case ast.ScopeTypeIn:
        t.Entity = resolveSlot(t.Entity, slotEnv)

        linkedScope = t
    case ast.ScopeTypeIsIn:
        t.Entity = resolveSlot(t.Entity, slotEnv)

        linkedScope = t
    default:
        panic(fmt.Sprintf("unknown scope type %T", t))
    }

    return linkedScope.(T)
}

// todo: should we panic on this? or just trust that the interface is correct?
func resolveSlot(ef types.EntityReference, slotEnv map[string]string) types.EntityReference {
    switch e := ef.(type) {
    case types.EntityUID:
        return e
    case types.VariableSlot:
        return parseEntityUID(slotEnv[e.ID.String()])
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
