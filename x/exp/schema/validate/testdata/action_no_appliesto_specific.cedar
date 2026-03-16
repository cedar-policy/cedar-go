// Policy 0: scope targets an action with no appliesTo directly
permit (
    principal is User,
    action == Action::"noop",
    resource
) when { true };
