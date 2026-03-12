// Policy 0: in with entity types in hierarchy (User in Group is valid)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal in Group::"admins"
};

// Policy 1: in with entity in its own type (always valid)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal in User::"alice"
};

// Policy 2: entity hierarchy - User can be in Org transitively (User -> Group -> Org)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal in Org::"acme"
};

// Policy 3: action scope in with descendant (edit in view)
permit (
    principal is User,
    action in Action::"view",
    resource is Document
) when {
    true
};

// Policy 4: scope with action == entity
permit (
    principal == User::"alice",
    action == Action::"view",
    resource == Document::"doc1"
);

// Policy 5: principal scope == action entity
permit (
    principal is User,
    action == Action::"edit",
    resource is Document
);
