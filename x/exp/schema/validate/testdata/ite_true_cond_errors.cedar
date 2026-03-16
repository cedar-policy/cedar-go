// Policy 0: ITE with typeTrue condition + condErr (lines 383-385)
// The inner ITE `if (1 + "bad") then true else true` has error but both branches are typeTrue,
// so it returns (typeTrue, err). This feeds the outer ITE as condition = (typeTrue, err).
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if (if (1 + "bad") then true else true) then true else true
};

// Policy 1: ITE with typeTrue condition + condErr + entity ref error in dead else branch (lines 377-379)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if (if (1 + "bad") then true else true) then true else Bogus::"x" == Bogus::"y"
};
