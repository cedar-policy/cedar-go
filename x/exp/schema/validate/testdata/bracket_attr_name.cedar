// Policy 0: nested attr access where parent has digit-start name, child not found
// triggers isValidCedarIdent failure for non-alpha start → bracket notation in error
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal["123start"].nonexistent == "x"
};

// Policy 1: nested attr access where parent has dash in name, child not found
// triggers isValidCedarIdent failure for non-alphanumeric continue → bracket notation
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal["has-dash"].nonexistent == "x"
};
