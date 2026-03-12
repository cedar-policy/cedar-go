// Policy 0: ip() called with wrong number of args (0 instead of 1)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip().isIpv4()
};
