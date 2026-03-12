// Policy 0: arith error + type error in && left operand
// (principal.nonexistent + 1) returns typeLong + error, then && sees non-Bool
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (principal.nonexistent + 1) && true
};

// Policy 1: arith error + type error in || left operand
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (principal.nonexistent + 1) || true
};

// Policy 2: negate error + type error in && left operand
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (-(principal.nonexistent)) && true
};

// Policy 3: arith right side error
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 + principal.nonexistent > 0
};

// Policy 4: in right side error
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal in principal.nonexistent
};

// Policy 5: containsAll right side error
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [1].containsAll(principal.nonexistent)
};

// Policy 6: containsAny right side error
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [1].containsAny(principal.nonexistent)
};

// Policy 7: if-then-else with error in then branch
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if principal.active then principal.nonexistent else 1
};

// Policy 8: if-then-else with error in else branch
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if principal.active then 1 else principal.nonexistent
};

// Policy 9: comparison with right-side error
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 < principal.nonexistent
};

// Policy 10: equality with nested arith error (both sides type-checked)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (principal.nonexistent + 1) == 1
};

// Policy 11: || right side error
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.active || (principal.nonexistent == 1)
};

// Policy 12: || right side non-bool type from arith
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.active || principal.age
};

// Policy 13: && right side non-bool type from arith
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.active && principal.age
};

// Policy 14: not with non-bool type + error (negate returns Long + error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    !(-(principal.nonexistent))
};
