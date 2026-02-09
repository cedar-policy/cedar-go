// Package resolved transforms an AST schema into a resolved schema
// where all type references are fully qualified and common types are inlined.
package resolved

import (
	"fmt"
	"iter"
	"strings"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

// Schema is a Cedar schema with resolved types and indexed declarations.
type Schema struct {
	Namespaces map[types.Path]Namespace
	Entities   map[types.EntityType]Entity
	Enums      map[types.EntityType]Enum
	Actions    map[types.EntityUID]Action
}

// Namespace represents a resolved namespace.
type Namespace struct {
	Name        types.Path
	Annotations Annotations
}

// Entity is a resolved entity type definition.
type Entity struct {
	Name        types.EntityType
	Annotations Annotations
	ParentTypes []types.EntityType
	Shape       RecordType
	Tags        IsType
}

// Enum is a resolved enum entity type definition.
type Enum struct {
	Name        types.EntityType
	Annotations Annotations
	Values      []types.String
}

// EntityUIDs iterates over valid EntityUIDs for this enum.
func (e Enum) EntityUIDs() iter.Seq[types.EntityUID] {
	return func(yield func(types.EntityUID) bool) {
		for _, v := range e.Values {
			if !yield(types.NewEntityUID(e.Name, v)) {
				return
			}
		}
	}
}

// AppliesTo defines the resolved principal, resource, and context types for an action.
type AppliesTo struct {
	Principals []types.EntityType
	Resources  []types.EntityType
	Context    RecordType
}

// Action is a resolved action definition.
type Action struct {
	Name        types.String
	Annotations Annotations
	Parents     []types.EntityUID
	AppliesTo   *AppliesTo
}

// Resolve transforms an AST schema into a fully resolved schema.
func Resolve(s *ast.Schema) (*Schema, error) {
	r := &resolverState{
		entityTypes: make(map[types.EntityType]bool),
		enumTypes:   make(map[types.EntityType]bool),
		commonTypes: make(map[types.Path]ast.IsType),
	}

	// Phase 1: Register all declarations
	if err := r.registerDecls("", s.Entities, s.Enums, s.CommonTypes); err != nil {
		return nil, err
	}
	for nsName, ns := range s.Namespaces {
		if err := r.registerDecls(nsName, ns.Entities, ns.Enums, ns.CommonTypes); err != nil {
			return nil, err
		}
	}

	// Phase 2: Check for illegal shadowing (RFC 70)
	if err := checkShadowing(s); err != nil {
		return nil, err
	}

	// Phase 3: Detect cycles in common types
	if err := r.detectCommonTypeCycles(); err != nil {
		return nil, err
	}

	// Phase 4: Resolve everything
	result := &Schema{
		Namespaces: make(map[types.Path]Namespace),
		Entities:   make(map[types.EntityType]Entity),
		Enums:      make(map[types.EntityType]Enum),
		Actions:    make(map[types.EntityUID]Action),
	}

	// Resolve bare declarations
	if err := r.resolveEntities("", s.Entities, result); err != nil {
		return nil, err
	}
	r.resolveEnums("", s.Enums, result)
	if err := r.resolveActions("", s.Actions, result); err != nil {
		return nil, err
	}

	// Resolve namespaced declarations
	for nsName, ns := range s.Namespaces {
		result.Namespaces[nsName] = Namespace{
			Name:        nsName,
			Annotations: Annotations(ns.Annotations),
		}
		if err := r.resolveEntities(nsName, ns.Entities, result); err != nil {
			return nil, err
		}
		r.resolveEnums(nsName, ns.Enums, result)
		if err := r.resolveActions(nsName, ns.Actions, result); err != nil {
			return nil, err
		}
	}

	// Phase 5: Validate and resolve action membership
	if err := r.validateActionMembership(result); err != nil {
		return nil, err
	}

	return result, nil
}

type resolverState struct {
	entityTypes map[types.EntityType]bool
	enumTypes   map[types.EntityType]bool
	commonTypes map[types.Path]ast.IsType
}

func (r *resolverState) registerDecls(nsName types.Path, entities ast.Entities, enums ast.Enums, commonTypes ast.CommonTypes) error {
	for name := range entities {
		if _, ok := enums[name]; ok {
			return fmt.Errorf("%q is declared twice", qualifyEntityType(nsName, name))
		}
		r.entityTypes[qualifyEntityType(nsName, name)] = true
	}
	for name := range enums {
		r.enumTypes[qualifyEntityType(nsName, name)] = true
	}
	for name, ct := range commonTypes {
		r.commonTypes[qualifyPath(nsName, name)] = ct.Type
	}
	return nil
}

// checkShadowing returns an error if any namespaced entity type, common type,
// or action shadows a declaration with the same basename in the empty namespace.
// See https://github.com/cedar-policy/rfcs/blob/main/text/0070-disallow-empty-namespace-shadowing.md
func checkShadowing(s *ast.Schema) error {
	// Collect bare (empty namespace) entity and common type basenames
	bareTypes := make(map[types.Ident]bool)
	for name := range s.Entities {
		bareTypes[name] = true
	}
	for name := range s.Enums {
		bareTypes[name] = true
	}
	for name := range s.CommonTypes {
		bareTypes[name] = true
	}

	// Check each namespace for conflicts
	for nsName, ns := range s.Namespaces {
		for name := range ns.Entities {
			if bareTypes[name] {
				return fmt.Errorf("definition of %q illegally shadows the existing definition of %q", string(nsName)+"::"+string(name), name)
			}
		}
		for name := range ns.Enums {
			if bareTypes[name] {
				return fmt.Errorf("definition of %q illegally shadows the existing definition of %q", string(nsName)+"::"+string(name), name)
			}
		}
		for name := range ns.CommonTypes {
			if bareTypes[name] {
				return fmt.Errorf("definition of %q illegally shadows the existing definition of %q", string(nsName)+"::"+string(name), name)
			}
		}
	}

	// Check bare action names against namespaced actions
	bareActions := make(map[types.String]bool)
	for name := range s.Actions {
		bareActions[name] = true
	}
	for nsName, ns := range s.Namespaces {
		for name := range ns.Actions {
			if bareActions[name] {
				return fmt.Errorf("definition of %q illegally shadows the existing definition of %q",
					string(nsName)+"::Action::\""+string(name)+"\"",
					"Action::\""+string(name)+"\"")
			}
		}
	}

	return nil
}

func (r *resolverState) detectCommonTypeCycles() error {
	// Build dependency graph
	deps := make(map[types.Path][]types.Path)
	for name, typ := range r.commonTypes {
		ns := extractNamespace(name)
		refs := collectTypeRefs(typ)
		for _, ref := range refs {
			resolved := r.resolveTypeRefPath(ns, ref)
			if _, ok := r.commonTypes[resolved]; ok {
				deps[name] = append(deps[name], resolved)
			}
		}
	}

	// Kahn's algorithm for topological sort / cycle detection
	inDegree := make(map[types.Path]int)
	for name := range r.commonTypes {
		inDegree[name] = 0
	}
	for _, neighbors := range deps {
		for _, n := range neighbors {
			inDegree[n]++
		}
	}

	var queue []types.Path
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	visited := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		visited++
		for _, neighbor := range deps[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if visited != len(r.commonTypes) {
		// Find a cycle for the error message
		for name := range inDegree {
			if inDegree[name] > 0 {
				return fmt.Errorf("cycle detected in common type definitions involving %q", name)
			}
		}
	}

	return nil
}

func (r *resolverState) resolveEntities(nsName types.Path, entities ast.Entities, result *Schema) error {
	for name, entity := range entities {
		qualName := qualifyEntityType(nsName, name)
		resolved := Entity{
			Name:        qualName,
			Annotations: Annotations(entity.Annotations),
		}
		for _, ref := range entity.ParentTypes {
			et, err := r.resolveEntityTypeRef(nsName, ref)
			if err != nil {
				return fmt.Errorf("entity %q: %w", qualName, err)
			}
			resolved.ParentTypes = append(resolved.ParentTypes, et)
		}
		if entity.Shape != nil {
			rec, err := r.resolveRecordType(nsName, *entity.Shape)
			if err != nil {
				return fmt.Errorf("entity %q shape: %w", qualName, err)
			}
			resolved.Shape = rec
		}
		if entity.Tags != nil {
			tags, err := r.resolveType(nsName, entity.Tags)
			if err != nil {
				return fmt.Errorf("entity %q tags: %w", qualName, err)
			}
			resolved.Tags = tags
		}
		result.Entities[qualName] = resolved
	}
	return nil
}

func (r *resolverState) resolveEnums(nsName types.Path, enums ast.Enums, result *Schema) {
	for name, enum := range enums {
		qualName := qualifyEntityType(nsName, name)
		result.Enums[qualName] = Enum{
			Name:        qualName,
			Annotations: Annotations(enum.Annotations),
			Values:      enum.Values,
		}
	}
}

func (r *resolverState) resolveActions(nsName types.Path, actions ast.Actions, result *Schema) error {
	for name, action := range actions {
		actionTypeName := qualifyActionType(nsName)
		uid := types.NewEntityUID(actionTypeName, types.String(name))
		resolved := Action{
			Name:        name,
			Annotations: Annotations(action.Annotations),
		}
		for _, ref := range action.Parents {
			resolved.Parents = append(resolved.Parents, resolveActionParentRef(nsName, ref))
		}
		if action.AppliesTo != nil {
			at := &AppliesTo{}
			for _, p := range action.AppliesTo.Principals {
				et, err := r.resolveEntityTypeRef(nsName, p)
				if err != nil {
					return fmt.Errorf("action %q principal: %w", name, err)
				}
				at.Principals = append(at.Principals, et)
			}
			if at.Principals == nil {
				at.Principals = []types.EntityType{}
			}
			for _, res := range action.AppliesTo.Resources {
				et, err := r.resolveEntityTypeRef(nsName, res)
				if err != nil {
					return fmt.Errorf("action %q resource: %w", name, err)
				}
				at.Resources = append(at.Resources, et)
			}
			if at.Resources == nil {
				at.Resources = []types.EntityType{}
			}
			if action.AppliesTo.Context != nil {
				ctx, err := r.resolveType(nsName, action.AppliesTo.Context)
				if err != nil {
					return fmt.Errorf("action %q context: %w", name, err)
				}
				rec, ok := ctx.(RecordType)
				if !ok {
					return fmt.Errorf("action %q context must resolve to a record type", name)
				}
				at.Context = rec
			} else {
				at.Context = RecordType{}
			}
			resolved.AppliesTo = at
		}
		result.Actions[uid] = resolved
	}
	return nil
}

func (r *resolverState) resolveType(ns types.Path, t ast.IsType) (IsType, error) {
	switch t := t.(type) {
	case ast.StringType:
		return StringType{}, nil
	case ast.LongType:
		return LongType{}, nil
	case ast.BoolType:
		return BoolType{}, nil
	case ast.ExtensionType:
		return ExtensionType(t), nil
	case ast.SetType:
		elem, err := r.resolveType(ns, t.Element)
		if err != nil {
			return nil, err
		}
		return SetType{Element: elem}, nil
	case ast.RecordType:
		return r.resolveRecordType(ns, t)
	case ast.EntityTypeRef:
		et, err := r.resolveEntityTypeRef(ns, t)
		if err != nil {
			return nil, err
		}
		return EntityType(et), nil
	case ast.TypeRef:
		return r.resolveTypeRef(ns, t)
	default:
		return nil, fmt.Errorf("unknown AST type: %T", t)
	}
}

func (r *resolverState) resolveRecordType(ns types.Path, rec ast.RecordType) (RecordType, error) {
	result := make(RecordType, len(rec))
	for name, attr := range rec {
		t, err := r.resolveType(ns, attr.Type)
		if err != nil {
			return nil, fmt.Errorf("attribute %q: %w", name, err)
		}
		result[name] = Attribute{
			Type:        t,
			Optional:    attr.Optional,
			Annotations: Annotations(attr.Annotations),
		}
	}
	return result, nil
}

func (r *resolverState) resolveEntityTypeRef(ns types.Path, ref ast.EntityTypeRef) (types.EntityType, error) {
	path := types.Path(ref)
	// If it's already a qualified path (contains ::), resolve directly
	if strings.Contains(string(path), "::") {
		et := types.EntityType(path)
		if r.entityTypes[et] || r.enumTypes[et] {
			return et, nil
		}
		return "", fmt.Errorf("undefined entity type %q", path)
	}
	// Unqualified: try NS::Name first, then bare Name
	if ns != "" {
		qualified := types.EntityType(string(ns) + "::" + string(path))
		if r.entityTypes[qualified] || r.enumTypes[qualified] {
			return qualified, nil
		}
	}
	bare := types.EntityType(path)
	if r.entityTypes[bare] || r.enumTypes[bare] {
		return bare, nil
	}
	return "", fmt.Errorf("undefined entity type %q", path)
}

// resolveTypeRef resolves a type reference (TypeRef) following the Cedar disambiguation rules:
// 1. Check if NS::N is declared as a common type
// 2. Check if NS::N is declared as an entity type
// 3. Check if N (empty namespace) is declared as a common type
// 4. Check if N (empty namespace) is declared as an entity type
// 5. Check if N is a built-in type
// 6. Error
func (r *resolverState) resolveTypeRef(ns types.Path, ref ast.TypeRef) (IsType, error) {
	// Qualified: resolve directly
	if strings.Contains(string(ref), "::") {
		return r.resolveQualifiedTypeRef(ref)
	}

	// Unqualified: follow disambiguation rules
	if ns != "" {
		qualifiedPath := types.Path(string(ns) + "::" + string(ref))
		// 1. Check NS::N as common type
		if ct, ok := r.commonTypes[qualifiedPath]; ok {
			return r.resolveType(ns, ct)
		}
		// 2. Check NS::N as entity type
		qualifiedET := types.EntityType(qualifiedPath)
		if r.entityTypes[qualifiedET] || r.enumTypes[qualifiedET] {
			return EntityType(qualifiedET), nil
		}
	}

	// 3. Check N as common type in empty namespace
	path := types.Path(ref)
	if ct, ok := r.commonTypes[path]; ok {
		return r.resolveType("", ct)
	}

	// 4. Check N as entity type in empty namespace
	bareET := types.EntityType(ref)
	if r.entityTypes[bareET] || r.enumTypes[bareET] {
		return EntityType(bareET), nil
	}

	// 5. Check built-in types
	if t := lookupBuiltin(path); t != nil {
		return t, nil
	}

	return nil, fmt.Errorf("undefined type %q", ref)
}

func (r *resolverState) resolveQualifiedTypeRef(ref ast.TypeRef) (IsType, error) {
	// Check for __cedar:: prefix first
	if strings.HasPrefix(string(ref), "__cedar::") {
		builtinName := ref[len("__cedar::"):]
		if t := lookupBuiltin(types.Path(builtinName)); t != nil {
			return t, nil
		}
		return nil, fmt.Errorf("undefined built-in type %q", ref)
	}

	// Try as common type first
	path := types.Path(ref)
	if ct, ok := r.commonTypes[path]; ok {
		ns := extractNamespace(path)
		return r.resolveType(ns, ct)
	}
	// Try as entity type
	et := types.EntityType(ref)
	if r.entityTypes[et] || r.enumTypes[et] {
		return EntityType(et), nil
	}
	return nil, fmt.Errorf("undefined type %q", ref)
}

func (r *resolverState) resolveTypeRefPath(ns types.Path, ref ast.TypeRef) types.Path {
	if strings.Contains(string(ref), "::") {
		return types.Path(ref)
	}
	if ns != "" {
		qualifiedPath := types.Path(string(ns) + "::" + string(ref))
		if _, ok := r.commonTypes[qualifiedPath]; ok {
			return qualifiedPath
		}
	}
	return types.Path(ref)
}

func resolveActionParentRef(ns types.Path, ref ast.ParentRef) types.EntityUID {
	if types.EntityType(ref.Type) == "" {
		// Bare reference: action in same namespace
		actionType := qualifyActionType(ns)
		return types.NewEntityUID(actionType, ref.ID)
	}
	return types.NewEntityUID(types.EntityType(ref.Type), ref.ID)
}

func (r *resolverState) validateActionMembership(result *Schema) error {
	// Build action UID set
	actionUIDs := make(map[types.EntityUID]bool)
	for uid := range result.Actions {
		actionUIDs[uid] = true
	}

	// Validate references and detect cycles
	for uid, action := range result.Actions {
		for _, parent := range action.Parents {
			if !actionUIDs[parent] {
				return fmt.Errorf("action %s: undefined parent action %s", uid, parent)
			}
		}
	}

	// Detect cycles using DFS
	visited := make(map[types.EntityUID]int) // 0=unvisited, 1=visiting, 2=done
	var visit func(types.EntityUID) error
	visit = func(uid types.EntityUID) error {
		switch visited[uid] {
		case 1:
			return fmt.Errorf("cycle detected in action hierarchy involving %s", uid)
		case 2:
			return nil
		}
		visited[uid] = 1
		action := result.Actions[uid]
		for _, parent := range action.Parents {
			if err := visit(parent); err != nil {
				return err
			}
		}
		visited[uid] = 2
		return nil
	}

	for uid := range result.Actions {
		if err := visit(uid); err != nil {
			return err
		}
	}

	return nil
}

func lookupBuiltin(path types.Path) IsType {
	switch path {
	case "String":
		return StringType{}
	case "Long":
		return LongType{}
	case "Bool", "Boolean":
		return BoolType{}
	case "ipaddr":
		return ExtensionType("ipaddr")
	case "decimal":
		return ExtensionType("decimal")
	case "datetime":
		return ExtensionType("datetime")
	case "duration":
		return ExtensionType("duration")
	default:
		return nil
	}
}

func collectTypeRefs(t ast.IsType) []ast.TypeRef {
	switch t := t.(type) {
	case ast.TypeRef:
		return []ast.TypeRef{t}
	case ast.SetType:
		return collectTypeRefs(t.Element)
	case ast.RecordType:
		var refs []ast.TypeRef
		for _, attr := range t {
			refs = append(refs, collectTypeRefs(attr.Type)...)
		}
		return refs
	default:
		return nil
	}
}

func qualifyEntityType(ns types.Path, name types.Ident) types.EntityType {
	if ns != "" {
		return types.EntityType(string(ns) + "::" + string(name))
	}
	return types.EntityType(name)
}

func qualifyPath(ns types.Path, name types.Ident) types.Path {
	if ns != "" {
		return types.Path(string(ns) + "::" + string(name))
	}
	return types.Path(name)
}

func qualifyActionType(ns types.Path) types.EntityType {
	if ns != "" {
		return types.EntityType(string(ns) + "::Action")
	}
	return types.EntityType("Action")
}

func extractNamespace(path types.Path) types.Path {
	s := string(path)
	if idx := strings.LastIndex(s, "::"); idx >= 0 {
		return types.Path(s[:idx])
	}
	return ""
}
