permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip("1.2.3.4") == 1
};
