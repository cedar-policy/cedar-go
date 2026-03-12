// Policy 0: && with left error, exercises typeOfAnd error paths
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (1 + "bad") && principal.name == "x"
};

// Policy 1: isEmpty on expression that errors with set recovery type
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if (1 + "bad") then principal.tags else principal.tags).isEmpty()
};

// Policy 2: like on expression that errors with string recovery type
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if (1 + "bad") then principal.name else principal.name) like "x*"
};

// Policy 3: is on expression that errors with entity recovery type
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if (1 + "bad") then principal else principal) is User
};

// Policy 4: attribute not found on multi-entity LUB
// User and Admin both have "name" but neither has "missing"
permit (
    principal,
    action == Action::"view",
    resource is Document
) when {
    principal.missing == "x"
};

// Policy 5: || with both sides erroring
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (1 + "bad") || (2 + "bad")
};
