// Policy 0: getTag with if-then-else where then branch is an Access node
// Exercises formatCedarExpr NodeTypeAccess case
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag(if true then context.foo else "y") == "z"
};

// Policy 1: getTag with if-then-else where then branch is a getTag node
// Exercises formatCedarExpr NodeTypeGetTag case
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag(if true then principal.getTag("x") else "y") == "z"
};

// Policy 2: getTag with if-then-else where condition is an arithmetic expression
// Exercises formatCedarExpr default <expr> case
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag(if (1 + 2) then "x" else "y") == "z"
};

// Policy 3: getTag with if-then-else where then branch is an entity literal
// Exercises formatCedarExpr NodeValue EntityUID case
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag(if true then User::"alice" else "y") == "z"
};
