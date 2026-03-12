// Policy 0: contains with RHS sub-expression error
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [1].contains(principal.nonexistent)
};

// Policy 1: contains with unrelated entity types (strict error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [principal].contains(resource)
};

// Policy 2: equality with both sides having errors (error flattening on RHS)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nonexistent == resource.nonexistent
};

// Policy 3: record literal with error element (error flattening)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    {"a": principal.nonexistent, "b": resource.nonexistent} == {"a": 1, "b": 2}
};
