// Policy 0: optional attribute of an entity's record attribute
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource.meta.s == "a"
};

// Policy 1: optional attribute two records below an entity
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource.meta has inner && resource.meta.inner.s == "a"
};

// Policy 2: the path stops at the nearest entity, not the variable
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource.owner.meta.s == "a"
};

// Policy 3: optional attribute of a record in context
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context.session.s == "a"
};

// Policy 4: optional attribute of a record-typed tag
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource.hasTag("k") && resource.getTag("k").s == "a"
};
