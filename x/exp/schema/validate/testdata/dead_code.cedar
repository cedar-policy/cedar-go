// Policy 0: false && expr - dead RHS with valid entity ref
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && User::"alice" == User::"bob"
};

// Policy 1: false && expr - dead RHS with INVALID entity ref (error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && Unknown::"x" == Unknown::"y"
};

// Policy 2: true || expr - dead RHS with valid entity ref
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true || User::"alice" == User::"bob"
};

// Policy 3: true || expr - dead RHS with INVALID entity ref (error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true || Unknown::"x" == Unknown::"y"
};

// Policy 4: if true then expr else dead - dead else with invalid entity ref (error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if true then true else Unknown::"x" == Unknown::"y"
};

// Policy 5: if false then dead else expr - dead then with invalid entity ref (error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if false then Unknown::"x" == Unknown::"y" else true
};

// Policy 6: == between disjoint entity types -> typeFalse
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal == resource
};

// Policy 7: is returns typeFalse when type doesn't match
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal is Document
};

// Policy 8: in returns typeFalse when hierarchy doesn't allow it (User can never be in Document)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal in resource
};

// Policy 9: != literal values (static evaluation)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 != 2
};

// Policy 10: == literal values true (static eval)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    "hello" == "hello"
};

// Policy 11: false || expr
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false || true
};

// Policy 12: if-then-else with runtime condition (not constant)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if principal is User then true else true
};
