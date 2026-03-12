// Policy 0: && with non-boolean left operand
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name && true
};

// Policy 1: || with non-boolean operand
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name || true
};

// Policy 2: ! on non-boolean
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    !principal.name
};

// Policy 3: if-then-else with non-boolean condition
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if principal.name then true else false
};

// Policy 4: < with non-comparable type
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name < "z"
};

// Policy 5: arithmetic with non-Long operand
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name + 1 > 0
};

// Policy 6: negate on non-Long
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    -(principal.name) > 0
};

// Policy 7: in with non-entity left
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name in principal
};

// Policy 8: in with non-entity right
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal in principal.name
};

// Policy 9: contains on non-set
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name.contains("a")
};

// Policy 10: containsAll with non-set
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name.containsAll(["a"])
};

// Policy 11: isEmpty on non-set
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name.isEmpty()
};

// Policy 12: like on non-string
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.age like "*"
};

// Policy 13: is on non-entity
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name is User
};

// Policy 14: has on non-entity/record
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name has foo
};

// Policy 15: attribute access on non-entity/record
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name.foo == "bar"
};

// Policy 16: attribute not found
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nonexistent == "test"
};

// Policy 17: hasTag on non-entity
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name.hasTag("x")
};

// Policy 18: getTag on non-entity
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name.getTag("x") == "y"
};

// Policy 19: extension function wrong arg type
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip(principal.age).isIpv4()
};

// Policy 20: invalid ip value
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip("not-an-ip").isIpv4()
};

// Policy 21: invalid decimal value
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    decimal("not-a-decimal").lessThan(decimal("1.0"))
};

// Policy 22: getTag without hasTag guard (unsafe access)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag("x") == "y"
};

// Policy 23: hasTag with non-string key
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.hasTag(1)
};

// Policy 24: && with non-bool right operand
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true && principal.name
};

// Policy 25: || with non-bool right operand
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false || principal.name
};

// Policy 26: attribute access without has guard on optional (entity attr)
// Not applicable since our schema doesn't have optional entity attrs
// Instead: if-then-else branches with incompatible types
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if principal.active then 1 else "no"
};

// Policy 27: is...in with non-entity operands
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name is User in principal
};

// Policy 28: containsAny on non-set right operand
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ["a"].containsAny(principal.name)
};

// Policy 29: invalid datetime
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    datetime("not-a-date") == datetime("2024-01-01")
};

// Policy 30: invalid duration
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    duration("not-a-dur") == duration("1h")
};
