// Policy 0: boolean operators, variables, attribute access, has, like, comparison, arithmetic
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    // && with booleans, attribute access, has guard for optional-less attrs
    principal.active && resource.public &&
    // like on string
    principal.name like "*alice*" &&
    // comparison <
    principal.age < 100 &&
    // arithmetic + - *
    (principal.age + 1 - 2) * 3 > 0 &&
    // set contains
    principal.scores.contains(42) &&
    // set containsAll, containsAny
    principal.scores.containsAll([1, 2]) &&
    principal.scores.containsAny([3, 4]) &&
    // set isEmpty
    ![5, 6].isEmpty() &&
    // equality between same types
    principal.name == "alice" &&
    // negation
    -(principal.age) < 0 &&
    // not
    !resource.public &&
    // if-then-else with bool condition and compatible branches
    (if resource.public then "yes" else "no") == "yes" &&
    // is on entity
    principal is User &&
    // in between entities
    principal in principal &&
    // record literal
    {name: "test"} == {name: "test"} &&
    // extension constructors
    ip("192.168.1.1").isIpv4() &&
    ip("::1").isIpv6() &&
    ip("127.0.0.1").isLoopback() &&
    ip("224.0.0.1").isMulticast() &&
    ip("192.168.1.1").isInRange(ip("192.168.0.0/16")) &&
    context.amount.lessThan(decimal("10.0")) &&
    context.amount.lessThanOrEqual(decimal("10.0")) &&
    context.amount.greaterThan(decimal("0.0")) &&
    context.amount.greaterThanOrEqual(decimal("0.0")) &&
    context.time.toDate() == context.time.toDate() &&
    context.time.toTime() == context.time.toTime() &&
    context.dur.toMilliseconds() > 0 &&
    context.dur.toSeconds() > 0 &&
    context.dur.toMinutes() >= 0 &&
    context.dur.toHours() >= 0 &&
    context.dur.toDays() >= 0 &&
    context.time > datetime("2024-01-01") &&
    context.dur < duration("1h") &&
    // hasTag / getTag
    principal.hasTag("role") &&
    principal.getTag("role").containsAll(["admin"])
};

// Policy 1: || with booleans
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource.public || principal.active
};

// Policy 2: != and equality between different types (permissive OK)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name != "bob" &&
    1 == 1 &&
    principal.age <= 100 &&
    principal.age >= 0
};

// Policy 3: entity in set of entities
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal in [principal, resource.owner]
};
