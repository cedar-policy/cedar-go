// Policy 0: && with non-bool LHS that also has sub-expression error
// principal.nonexistent returns error AND principal.name returns String (non-bool)
// This tests the dual error path in typeOfAnd
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name && true
};

// Policy 1: || with non-bool LHS that also has sub-expression error
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name || true
};

// Policy 2: ! on non-bool expression
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    !principal.name
};

// Policy 3: && non-bool RHS
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true && principal.name
};

// Policy 4: || non-bool RHS
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false || principal.name
};
