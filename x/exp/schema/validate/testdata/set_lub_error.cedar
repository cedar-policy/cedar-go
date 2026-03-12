// Policy 0: if-then-else with incompatible set element types
// Set<Long> vs Set<String> → LUB of Long and String fails
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (if principal.active then [1, 2] else ["a", "b"]) == [1] };
