// Policy 0: negate on expression that returns non-Long type WITH an error
// This triggers the dual error path in typeOfNegate where both
// a sub-expression error AND a type mismatch are reported
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    -(principal == Unknown::"x") > 0
};
