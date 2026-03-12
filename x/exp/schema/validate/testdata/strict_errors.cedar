// Policy 0: empty set literal (strict error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [].isEmpty()
};

// Policy 1: incompatible equality types in strict mode
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.age == principal.name
};

// Policy 2: extension constructor with non-literal arg in strict mode
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip(principal.name).isIpv4()
};

// Policy 3: set with incompatible entity types in strict mode
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [principal, resource].isEmpty()
};

// Policy 4: if-then-else with incompatible entity types in strict mode
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if true then principal else resource) == principal
};

// Policy 5: contains with incompatible element type in strict mode
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [1, 2, 3].contains("foo")
};

// Policy 6: containsAll with incompatible element types in strict mode
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [1, 2].containsAll(["a", "b"])
};
