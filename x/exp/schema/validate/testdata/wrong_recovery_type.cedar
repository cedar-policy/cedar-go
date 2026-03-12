// Policy 0: isEmpty on non-set recovery type (Long)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (1 + "bad").isEmpty()
};

// Policy 1: like on non-string recovery type (Long)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (1 + "bad") like "x*"
};

// Policy 2: is on non-entity recovery type (Long)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (1 + "bad") is User
};

// Policy 3: attribute access on non-entity/record with sub-error
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (1 + "bad").someAttr == "x"
};

// Policy 4: attribute access on entity with sub-error
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if (1 + "bad") then principal else principal).name == "x"
};
