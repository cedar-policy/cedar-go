// Policy 0: && where left has error with typeFalse recovery (line 175)
// (if (error) then false else false) produces (typeFalse, error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if (1 + "bad") then false else false) && true
};

// Policy 1: && where left has error (non-typeFalse), right has error with typeFalse recovery (line 186)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (1 + "bad") && (if (2 + "bad") then false else false)
};

// Policy 2: && where left is ok (typeBool), right has error with typeFalse recovery (line 204)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.active && (if (1 + "bad") then false else false)
};

// Policy 3: && where left has error with typeFalse recovery AND right has bad entity ref (line 176)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if (1 + "bad") then false else false) && (Bogus::"x" == Bogus::"y")
};
