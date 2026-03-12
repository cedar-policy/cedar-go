// Policy 0: action UID in condition - known action
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    action == Action::"view"
};

// Policy 1: action UID in condition - unknown action (type exists but ID not declared)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    action == Action::"nonexistent"
};

// Policy 2: action in set with action entities
permit (
    principal is User,
    action in [Action::"view", Action::"edit"],
    resource is Document
) when {
    true
};

// Policy 3: context variable access
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context == {}
};

// Policy 4: set literal with strings
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ["a", "b"].contains("a")
};

// Policy 5: record literal with error value
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    {x: principal.nonexistent} == {x: "test"}
};
