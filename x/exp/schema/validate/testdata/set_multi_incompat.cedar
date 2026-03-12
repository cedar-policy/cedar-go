// Policy 0: set literal with 3 incompatible types (Long, String, Bool)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [1, "hello", true].isEmpty()
};
