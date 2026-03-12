// Policy 0: nested has guard with access on optional nested attr (ok)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context.session has token && context.session.token == "abc"
};

// Policy 1: access nested optional without has guard (error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context.session.token == "abc"
};
