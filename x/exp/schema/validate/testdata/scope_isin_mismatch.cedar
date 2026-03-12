// Policy 0: principal is User in Document - User can't be in Document hierarchy
permit (
    principal is User in Document::"doc1",
    action == Action::"view",
    resource is Document
);

// Policy 1: resource is Document in Group - Document can't be in Group hierarchy
permit (
    principal is User,
    action == Action::"view",
    resource is Document in Group::"g1"
);
