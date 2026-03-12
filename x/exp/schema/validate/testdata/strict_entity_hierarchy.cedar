// Policy 0: if-then-else with entity types related via hierarchy
// Admin is a descendant of User, but strict mode considers them
// incompatible (only same-type equality counts for entity LUB compatibility)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if principal.name == "admin" then User::"alice" else Admin::"bob").name == "test"
};
