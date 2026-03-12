// Policy 0: true && expr - constant fold LHS true
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true && (principal.name == "alice")
};

// Policy 1: expr && false - constant fold RHS false
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (principal.name == "alice") && false
};

// Policy 2: false || expr - constant fold LHS false
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false || (principal.name == "alice")
};

// Policy 3: expr || true - constant fold RHS true
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (principal.name == "alice") || true
};

// Policy 4: expr || false - constant fold RHS false
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (principal.name == "alice") || false
};

// Policy 5: !true - constant fold to false, whole when is false
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    !(1 == 1) && (principal.name == "alice")
};

// Policy 6: !false - constant fold to true
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    !(1 == 2) && (principal.name == "alice")
};

// Policy 7: if false then X else Y - takes else branch
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if (1 == 2) then false else (principal.name == "alice")
};

// Policy 8: != with disjoint entity types
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (principal != resource) && true
};
