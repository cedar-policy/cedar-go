// Policy 0: extension function with wrong number of args
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip("1.2.3.4", "5.6.7.8").isIpv4()
};

// Policy 1: extension constructor with non-literal in strict mode
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    decimal(principal.name) > decimal("1.0")
};
