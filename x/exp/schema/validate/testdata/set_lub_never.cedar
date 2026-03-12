// Policy 0: if-then-else with typed set and empty set
// LUB(Set<Long>, Set<Never>) exercises leastUpperBound b=typeNever path
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if principal.active then [1, 2] else []).contains(1)
};

// Policy 1: if-then-else where both branches return typeTrue
// LUB(typeTrue, typeTrue) exercises exact-match LUB path
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if principal.active then true else true
};

// Policy 2: if-then-else where branches return typeFalse and typeBool
// LUB(typeFalse, typeBool) exercises bool folding
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if principal.active then (1 > 2) else principal.active
};

// Policy 3: has on context record with required attribute (returns typeTrue)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context has flag && context.flag
};
