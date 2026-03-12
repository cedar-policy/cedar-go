// Policy 0: error propagation through arithmetic (left operand)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nonexistent + 1 > 0
};

// Policy 1: error propagation through negate
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    -(principal.nonexistent) > 0
};

// Policy 2: error propagation through in (left operand)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nonexistent in principal
};

// Policy 3: error propagation through contains (receiver)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nonexistent.contains(1)
};

// Policy 4: error propagation through && (left side has sub-expr error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (principal.nonexistent == 1) && true
};

// Policy 5: error propagation through || (left side has sub-expr error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (principal.nonexistent == 1) || true
};

// Policy 6: error propagation through ! (sub-expr error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    !(principal.nonexistent == 1)
};

// Policy 7: error propagation through like
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nonexistent like "*"
};

// Policy 8: error propagation through is
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nonexistent is User
};

// Policy 9: error propagation through if-then-else condition
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if (principal.nonexistent == 1) then true else false
};

// Policy 10: error propagation through has
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nonexistent has foo
};

// Policy 11: error propagation through containsAll (receiver)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nonexistent.containsAll([1])
};

// Policy 12: error propagation through isEmpty
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nonexistent.isEmpty()
};

// Policy 13: error propagation through comparison
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nonexistent < 1
};

// Policy 14: error propagation through hasTag
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nonexistent.hasTag("x")
};

// Policy 15: error propagation through getTag
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nonexistent.getTag("x") == "y"
};

// Policy 16: error propagation through set literal
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [principal.nonexistent].contains(1)
};

// Policy 17: error propagation through record literal
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    {"a": principal.nonexistent} == {"a": 1}
};

// Policy 18: error propagation through && (right side has sub-expr error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true && (principal.nonexistent == 1)
};

// Policy 19: error propagation through || (right side has sub-expr error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false || (principal.nonexistent == 1)
};
