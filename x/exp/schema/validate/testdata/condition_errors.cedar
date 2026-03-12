// Policy 0: access non-existent context attribute (record attr not found)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context.nonexistent == "x"
};

// Policy 1: condition with non-boolean result type
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name
};
