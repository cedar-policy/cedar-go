// Policy 0: access on entity with optional attribute to trigger error with optional field type
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.profile + 1
};
