// Policy 0: extension constructor non-literal where arg itself has error
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip(if (1 + "bad") then "1.2.3.4" else "5.6.7.8").isIpv4()
};
