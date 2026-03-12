// Policy 0: access on record from multi-entity LUB, nested attr not found
// principal.profile returns record with source LUB{User, Admin}
// then .nonexistent triggers attrNotFoundError with multi-entity record source
permit (
    principal,
    action == Action::"view",
    resource is Document
) when {
    principal.profile.nonexistent == "x"
};

// Policy 1: if-then-else with non-constant condition producing multi-entity LUB
// principal is User, resource is Document, both have profile record attr
// .profile returns record from multi-entity source, .nonexistent fails
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if principal.profile.name == "a" then principal else resource).profile.nonexistent == "x"
};
