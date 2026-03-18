permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [context, {"a": 1, "r": 1}, 1].isEmpty()
};

permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [{"z": 1}, {"z": 1}, 1].isEmpty()
};

permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if false then true else NoSuchType::"x"
};
