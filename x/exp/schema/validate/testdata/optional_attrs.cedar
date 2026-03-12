// Policy 0: has guard on optional entity attr (ok)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal has nickname && principal.nickname == "nick"
};

// Policy 1: access optional without has guard (error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.nickname == "nick"
};

// Policy 2: has on record (context) - optional attr
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context has reason && context.reason == "test"
};

// Policy 3: has on record - required attr (typeTrue)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal has name && principal.name == "alice"
};

// Policy 4: has on entity - attr doesn't exist (typeFalse)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal has nonexistent
};

// Policy 5: has on record - attr doesn't exist (typeFalse)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context has nonexistent
};
