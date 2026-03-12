permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if principal == User::"a" then User::"a" else Admin::"b").email == "test"
};
