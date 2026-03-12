// Policy 0: getTag with if-then-else key expression (unsafe access error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag(if true then "role" else "admin") == "manager"
};

// Policy 1: getTag with non-constant condition in if-then-else
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag(if principal == principal then "x" else "y") == "z"
};

// Policy 2: getTag with nested getTag key (formatTagExprRust getTag case)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag(principal.getTag("key")) == "value"
};

// Policy 3: getTag with Long value in if condition (formatCedarExpr Long)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag(if 42 then "x" else "y") == "z"
};

// Policy 4: getTag with context in if condition (formatCedarExpr variable)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag(if context then "x" else "y") == "z"
};

// Policy 5: getTag with variable in then branch (formatCedarExpr variable)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag(if true then context else "y") == "z"
};

// Policy 6: getTag with nested if-then-else in then branch
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag(if true then (if true then "a" else "b") else "c") == "z"
};

