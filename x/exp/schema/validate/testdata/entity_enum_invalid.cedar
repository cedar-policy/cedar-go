permit (
    principal in Status::"active",
    action == Action::"view",
    resource
) when { principal.status == Status::"inactive" };

permit (
    principal in Status::"bogus",
    action == Action::"view",
    resource
);

permit (
    principal,
    action == Action::"view",
    resource
) when { principal.status == Status::"bogus" };

permit (
    principal == Status::"bogus",
    action == Action::"check",
    resource
);

permit (
    principal,
    action == Action::"view",
    resource
) when { principal.status == Status::"o'k" };
