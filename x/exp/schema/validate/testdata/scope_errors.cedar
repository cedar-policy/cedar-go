// Policy 0: principal == unknown entity type
permit (
    principal == Unknown::"x",
    action == Action::"view",
    resource is Document
);

// Policy 1: principal in unknown entity type
permit (
    principal in Unknown::"x",
    action == Action::"view",
    resource is Document
);

// Policy 2: principal is unknown type
permit (
    principal is Unknown,
    action == Action::"view",
    resource is Document
);

// Policy 3: action == unknown action
permit (
    principal is User,
    action == Action::"nonexistent",
    resource is Document
);

// Policy 4: action in unknown action
permit (
    principal is User,
    action in Action::"nonexistent",
    resource is Document
);

// Policy 5: resource == unknown entity type
permit (
    principal is User,
    action == Action::"view",
    resource == Unknown::"x"
);

// Policy 6: principal is type in entity - type cannot be in entity (mismatch)
permit (
    principal is Document in Document::"doc1",
    action == Action::"view",
    resource is Document
);

// Policy 7: resource is type in entity
permit (
    principal is User,
    action == Action::"view",
    resource is Document in Document::"doc1"
);

// Policy 8: principal is type in entity - unknown type
permit (
    principal is Unknown in Group::"g",
    action == Action::"view",
    resource is Document
);

// Policy 9: principal is type in entity - unknown entity
permit (
    principal is User in Unknown::"x",
    action == Action::"view",
    resource is Document
);

// Policy 10: resource in entity
permit (
    principal is User,
    action == Action::"view",
    resource in Document::"doc1"
);

// Policy 11: resource is type
permit (
    principal is User,
    action == Action::"view",
    resource is Document
);
