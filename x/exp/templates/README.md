# Cedar Templates for Go

Cedar Templates is a feature that extends the Cedar policy language in Go by allowing you to create policy templates with placeholder variables that can be filled in at runtime. This README explains the basics of Cedar Templates and provides examples of how to use them.

## Overview

Cedar policy language provides a way to define access control policies for your applications. Templates enhance this capability by allowing you to create policy patterns that can be instantiated with specific values at runtime. This is particularly useful when you need to create similar policies for different entities without duplicating policy code.

## Key Concepts

- **Template**: A Cedar policy with placeholders (slots) that can be filled in at runtime.
- **Slots**: Placeholders in a template denoted by a question mark followed by an identifier (e.g., `?principal`).
- **Linking**: The process of binding concrete values to slots in a template to create a usable policy.
- **PolicySet**: A collection of policies and templates that can be used for authorization decisions.

## Basic Usage

### Creating a Template

A template looks like a regular Cedar policy but includes slots (marked with `?`) for values to be filled in later:

```go
templateStr := `permit (
    principal == ?principal,
    action,
    resource == ?resource
)
when { resource.owner == principal };`

var template templates.Template
err := template.UnmarshalCedar([]byte(templateStr))
if err != nil {
    // handle error
}
```

### Creating a PolicySet and Adding Templates

```go
// Create a new empty PolicySet
policySet := templates.NewPolicySet()

// Add a template to the PolicySet
templateID := cedar.PolicyID("access_template")
policySet.AddTemplate(templateID, &template)
```

### Linking a Template to Create a Policy

Once you have a template, you can link it with specific entity values to create a concrete policy:

```go
// Define the slot values
slotValues := map[types.SlotID]types.EntityUID{
    "?principal": types.NewEntityUID("User", "alice"),
    "?resource": types.NewEntityUID("Document", "report"),
}

// Link the template to create a policy
linkID := cedar.PolicyID("alice_report_access")
err = policySet.LinkTemplate(templateID, linkID, slotValues)
if err != nil {
    // handle error
}
```

### Using Templates for Authorization

```go
// Create a request
request := cedar.Request{
    Principal: cedar.NewEntityUID("User", "alice"),
    Action:    cedar.NewEntityUID("Action", "read"),
    Resource:  cedar.NewEntityUID("Document", "report"),
    Context:   types.NewRecord(nil),
}

// Create an entity store with relevant entities
entities := types.NewEntityMap()
// Add entities to the store...

// Make an authorization decision
decision, diagnostic := templates.Authorize(policySet, entities, request)

// Check the decision
if decision == cedar.Allow {
    // Access granted
} else {
    // Access denied
}
```

## Advanced Examples

### Example 1: Role-Based Access Control Template

```go
// Template that grants access based on role
roleBasedTemplate := `permit (
    principal,
    action,
    resource
)
when { principal.roles.contains(?role) };`

// Link with a specific role
roleSlots := map[types.SlotID]types.EntityUID{
    "?role": types.NewEntityUID("Role", "admin"),
}
policySet.LinkTemplate(cedar.PolicyID("role_template"), cedar.PolicyID("admin_access"), roleSlots)
```

### Example 2: Resource Ownership Template

```go
// Template for resource ownership
ownershipTemplate := `permit (
    principal == ?owner,
    action in [Action::"read", Action::"write", Action::"delete"],
    resource == ?resource
);`

// Link with specific owner and resource
ownershipSlots := map[types.SlotID]types.EntityUID{
    "?owner": types.NewEntityUID("User", "bob"),
    "?resource": types.NewEntityUID("Photo", "vacation"),
}
policySet.LinkTemplate(cedar.PolicyID("ownership_template"), cedar.PolicyID("bob_photo_ownership"), ownershipSlots)
```

### Example 3: Handling Multiple Templates

```go
// Load templates from Cedar language text
policySetStr := `
// Resource ownership template
template ownership_tpl(principal, resource) {
    permit(
        principal == ?principal,
        action in [Action::"read", Action::"write"],
        resource == ?resource
    );
}

// Role-based access template
template role_tpl(role) {
    permit(
        principal,
        action,
        resource
    )
    when { principal.roles.contains(?role) };
}
`

policySet, err := templates.NewPolicySetFromBytes("policies.cedar", []byte(policySetStr))
if err != nil {
    // handle error
}

// Link templates
policySet.LinkTemplate("ownership_tpl", "alice_doc1_ownership", map[types.SlotID]types.EntityUID{
    "?principal": types.NewEntityUID("User", "alice"),
    "?resource": types.NewEntityUID("Document", "doc1"),
})

policySet.LinkTemplate("role_tpl", "admin_access", map[types.SlotID]types.EntityUID{
    "?role": types.NewEntityUID("Role", "admin"),
})
```

## Working with Template Outputs

After linking a template, the resulting policy can be:

1. Used for authorization via the `templates.Authorize()` function
2. Serialized to Cedar language format with `MarshalCedar()`
3. Serialized to JSON format with `MarshalJSON()`

## Notes and Best Practices

1. **Template Management**: Keep track of template IDs and linked policy IDs to manage them effectively.
2. **Error Handling**: Always check for errors when parsing templates, linking them, or making authorization decisions.
3. **Entity Management**: Ensure your entity store contains all entities referenced in your policies and templates.
4. **Slot Validation**: Verify that all required slots are provided when linking a template.
5. **Experimental Status**: Note that the templates package is in the experimental (`x/exp`) namespace and may undergo changes.

## Additional Resources

- [Cedar Policy Documentation](https://docs.cedarpolicy.com/)
- [Cedar Templates Documentation](https://docs.cedarpolicy.com/policies/templates.html)
- [Go API Reference](https://pkg.go.dev/github.com/cedar-policy/cedar-go)

## License

Cedar is licensed under the Apache License, Version 2.0