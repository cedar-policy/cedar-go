permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [[], [1], true, false, principal has opt, 1, "s", principal, resource, {"a": 1}, {"a": 1, "b": "x"}, ip("1.1.1.1"), datetime("2020-01-01T00:00:00Z")].isEmpty()
};
