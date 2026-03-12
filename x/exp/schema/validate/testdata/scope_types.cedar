// Policy 0: principal == entity (valid)
permit (
    principal == User::"alice",
    action == Action::"view",
    resource == Document::"doc1"
);

// Policy 1: principal in entity
permit (
    principal in Group::"admins",
    action == Action::"view",
    resource in Folder::"root"
);

// Policy 2: principal is type (valid)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
);

// Policy 3: principal is type in entity (type can be in entity)
permit (
    principal is User in Group::"staff",
    action == Action::"view",
    resource is Document
);

// Policy 4: action in with descendants
permit (
    principal is User,
    action in Action::"view",
    resource is Document
);

// Policy 5: action in set
permit (
    principal is User,
    action in [Action::"view", Action::"admin_view"],
    resource is Document
);

// Policy 6: resource is type in entity
permit (
    principal is User,
    action == Action::"view",
    resource is Document in Document::"doc1"
);
