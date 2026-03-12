// Policy 0: action in set literal in condition (true: leaf is in top)
permit (
    principal is User,
    action == Action::"leaf",
    resource is Document
) when {
    action in [Action::"top"]
};

// Policy 1: action in single entity in condition (false: top is NOT in leaf)
permit (
    principal is User,
    action == Action::"top",
    resource is Document
) when {
    action in Action::"leaf"
};

// Policy 2: action in set literal in condition (true: leaf is in mid)
permit (
    principal is User,
    action == Action::"leaf",
    resource is Document
) when {
    action in [Action::"mid", Action::"top"]
};
