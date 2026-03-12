// Policy 0: empty set in if condition with extension errors in branches
permit (
    principal is a,
    action == Action::"action",
    resource is a
) when {
    (if ([].isEmpty()) then ip("") else ip("")).isIpv4()
};

// Policy 1: entity record access - attribute not found via dot notation
permit (
    principal is a,
    action == Action::"action",
    resource is a
) when {
    (resource.r).nonexistent == "test"
};

// Policy 2: tag expression with getTag and if-then-else (const true - dead branch)
permit (
    principal is a,
    action == Action::"action",
    resource is a
) when {
    resource.getTag(if true then "k" else "k2") == "v"
};

// Policy 3: unicode in extension arg
permit (
    principal is a,
    action == Action::"action",
    resource is a
) when {
    ip("\u{1}\u{2}") == ip("0.0.0.0")
};

// Policy 4: if-then-else with const false (dead branch replacement in tag)
permit (
    principal is a,
    action == Action::"action",
    resource is a
) when {
    resource.getTag(if false then "a" else "b") == "v"
};

// Policy 5: entity record access with non-identifier key (bracket notation)
permit (
    principal is a,
    action == Action::"action",
    resource is a
) when {
    (resource.r)[""] == "test"
};
