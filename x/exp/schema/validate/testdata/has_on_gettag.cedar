// Policy 0: has guards an optional attribute of a record-typed tag
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource.hasTag("k") && resource.getTag("k") has s && resource.getTag("k").s == "a"
};

// Policy 1: has guards an optional attribute below a record-typed tag
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource.hasTag("k") &&
    resource.getTag("k") has inner &&
    resource.getTag("k").inner has s &&
    resource.getTag("k").inner.s == "a"
};

// Policy 2: has on one tag does not guard another tag's attribute
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource.hasTag("j") &&
    resource.hasTag("k") &&
    resource.getTag("j") has s &&
    resource.getTag("k").s == "a"
};
