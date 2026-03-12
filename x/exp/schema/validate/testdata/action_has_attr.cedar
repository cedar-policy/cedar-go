// Policy 0: action scope all with conditions (covers matchesActionConstraint empty constraints)
permit (
    principal is User,
    action,
    resource is Document
)
when { principal has name };

// Policy 1: has on action entity (covers hasResultTypeEntity action entity path)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { action has someAttr };
