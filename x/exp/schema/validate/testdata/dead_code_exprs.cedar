// Policy 0: dead branch with arithmetic, comparison, negation
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && (1 + 2 - 3 * 1 > 0 && -(1) < 2 && 1 <= 2 && 2 >= 1)
};

// Policy 1: dead branch with set ops, contains, containsAll, containsAny, isEmpty
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && ([1, 2, 3].contains(1) && [1, 2].containsAll([1]) && [1, 2].containsAny([2]) && ![3, 4].isEmpty())
};

// Policy 2: dead branch with in, is, like, has, access
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && (principal in Group::"staff" && principal is User && principal.name like "*" && principal has name && principal.name == "x")
};

// Policy 3: dead branch with record, extension call, or, not
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && ({"a": 1} == {"a": 1} || !false || ip("10.0.0.1").isLoopback())
};

// Policy 4: dead branch with hasTag, getTag
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && (principal.hasTag("role") && principal.getTag("role") == "admin")
};

// Policy 5: dead branch with is...in
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && (principal is User in Group::"staff")
};

// Policy 6: dead branch with nested if-then-else
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && (if true then true else false)
};
