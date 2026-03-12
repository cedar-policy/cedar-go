// Policy 0: resource is Type in Entity - type cannot be in entity
// Group cannot be a descendant of a Document
permit (
    principal is User,
    action == Action::"view",
    resource is Group in Document::"doc1"
);

// Policy 1: action entity UID in principal scope
permit (
    principal == Action::"view",
    action == Action::"view",
    resource is Document
);

// Policy 2: resource in with unknown entity type
permit (
    principal is User,
    action == Action::"view",
    resource in Unknown::"x"
);

// Policy 3: resource is unknown type
permit (
    principal is User,
    action == Action::"view",
    resource is Unknown
);
