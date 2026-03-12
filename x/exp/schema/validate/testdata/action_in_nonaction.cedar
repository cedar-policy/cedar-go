// Policy 0: action in set of non-action entity in condition
// exercises exprToActionEUIDs non-action fallback and no-actions RHS path
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    action in [User::"alice"]
};

// Policy 1: action in set with variable element (not a literal)
// exercises exprToActionEUIDs nil return for non-literal element
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    action in [principal]
};

// Policy 2: action in non-set entity in condition
// exercises exprToActionEUIDs final nil return for non-set RHS
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    action in principal
};
