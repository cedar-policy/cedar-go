// Policy 0: comparison where right side has single error (attr not found)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 < principal.nonexistent
};

// Policy 1: comparison where right side has unwrapped single error
// context.level is Long, .subField access on Long returns single unexpectedTypeErr
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 < context.level.subField
};
