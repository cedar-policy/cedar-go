// Policy 0: negate on non-Long expression
// Exercises typeOfNegate error path
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    -principal > 0
};

// Policy 1: dead code with NotEquals, comparison operators, is...in
// Exercises validateEntityRefs for various node types in dead code
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && (1 != 2 && 1 < 2 && 1 <= 2 && 1 > 0 && 1 >= 0 && (principal is User in Group::"staff"))
};

// Policy 2: dead code with record containing invalid entity ref
// Exercises validateEntityRefs NodeTypeRecord error path
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && {"a": Unknown::"x"} == {"a": 1}
};

// Policy 3: dead code with set containing invalid entity ref
// Exercises validateEntityRefs NodeTypeSet error path
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && [Unknown::"x"].contains(1)
};

// Policy 4: has capability upgrade - second has after first has in &&
// Exercises typeOfHas capability upgrade path
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal has name && principal has name && principal.name == "admin"
};
