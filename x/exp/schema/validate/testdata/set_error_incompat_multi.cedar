// Policy 0: set with sub-expression errors AND incompatible types (>2 elements)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [if (1 + "bad") then 1 else 1, "hello", true].isEmpty()
};
