permit (
    principal,
    action == Action::"view",
    resource
)
when { principal is User in Group::"admins" };
