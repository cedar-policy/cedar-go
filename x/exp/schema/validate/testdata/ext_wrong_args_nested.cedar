// Policy 0: extension method with wrong arg count and nested error in args
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip("1.2.3.4").isIpv4(1 + "bad")
};
