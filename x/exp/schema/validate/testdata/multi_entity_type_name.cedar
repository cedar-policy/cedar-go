permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if principal == User::"a" then
        (if principal == User::"b" then User::"c" else Admin::"d")
    else 1) == 1
};
