// Policy 0: action scope "all" with principal constraint
// Iterates over all actions including one with no AppliesTo
// Covers validateActionApplication action.AppliesTo == nil continue path
permit (
    principal is User,
    action,
    resource
) when { true };
