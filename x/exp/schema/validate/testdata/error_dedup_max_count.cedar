// Policy 0: expression that produces different error counts across environments.
// For action "a" (principal User, x:Long): principal.x + "bad" → 1 "expected Long saw String" error
// For action "b" (principal Admin, x:String): principal.x + "bad" → 2 "expected Long saw String" errors
// This exercises the max-count merge path in typecheckConditions.
permit (
    principal,
    action,
    resource
) when {
    principal.x + "bad"
};
