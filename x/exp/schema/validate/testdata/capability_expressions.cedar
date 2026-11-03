// Policy 0: has on a getTag result guards accesses to its attributes
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.hasTag("t") &&
    principal.getTag("t") has foo &&
    principal.getTag("t").foo == "a"
};

// Policy 1: hasTag with a non-literal tag guards getTag with the same tag
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.hasTag(context.key) && principal.getTag(context.key).bar == 1
};

// Policy 2: hasTag doesn't guard getTag with a different tag
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.hasTag(context.key) && principal.getTag(context.other).bar == 1
};

// Policy 3: has guards accesses on any expression, not just variables and
// attribute accesses
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if context.flag then principal else resource.owner) has nick &&
    (if context.flag then principal else resource.owner).nick == "n"
};

// Policy 4: record literals with the same elements are the same expression,
// regardless of the order of the elements
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.hasTag(if context.r == {a: 1, b: 2} then "x" else "y") &&
    principal.getTag(if context.r == {b: 2, a: 1} then "x" else "y").bar == 1
};

// Policy 5: hasTag doesn't guard getTag on a different entity
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.hasTag(context.key) &&
    resource.owner.getTag(context.key).bar == 1
};

// Policy 6: has guards an optional attribute below a getTag result
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.hasTag("t") &&
    principal.getTag("t") has inner &&
    principal.getTag("t").inner has s &&
    principal.getTag("t").inner.s == "a"
};

// Policy 7: has on one tag doesn't guard another tag's attribute
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.hasTag("j") &&
    principal.hasTag("k") &&
    principal.getTag("j") has foo &&
    principal.getTag("k").foo == "a"
};
