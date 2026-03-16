// Policy 0: if-then-else with false condition + condErr, entity ref error in dead then-branch
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if (if (1 + "bad") then false else false) then (Bogus::"x" == Bogus::"y") else true
};

// Policy 1: if-then-else with false condition + condErr, error in live else-branch
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if (if (1 + "bad") then false else false) then true else (1 + "bad2")
};
