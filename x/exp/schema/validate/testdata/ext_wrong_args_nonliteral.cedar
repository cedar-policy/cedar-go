// Policy 0: extension constructor with wrong arg count and non-literal arg
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    decimal(if true then "1.0" else "2.0", "3.0").greaterThan(decimal("0.0"))
};
