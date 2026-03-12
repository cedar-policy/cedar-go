// Policy 0: set with 3 elements where 2 have same type (Long, Long, String)
// triggers typeIncompatErrMulti with dedup to 2 names
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [1, 2, "hello"].isEmpty()
};

// Policy 1: set with 3 entity elements from different hierarchies (strict)
// triggers checkStrictEntityLUB error with 3+ elements
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [principal, resource, principal].isEmpty()
};
