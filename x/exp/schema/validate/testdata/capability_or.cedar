// Policy 0: capability intersection through || - both branches have same has guard
// principal.name access should work because has("name") is true on both sides
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (principal has name || principal has name) && principal.name == "x" };

// Policy 1: capability NOT propagated through || - different attrs on each side
// principal.email access should fail because only RHS has the guard
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (principal has name || principal has email) && principal.email == "y" };
