permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if principal == User::"a" then false else 1) == true
};
