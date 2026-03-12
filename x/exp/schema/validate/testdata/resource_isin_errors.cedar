// Policy 0: resource is Unknown in Document - unknown type in is...in
permit (
    principal is User,
    action == Action::"view",
    resource is Unknown in Document::"doc1"
);

// Policy 1: resource is Document in Unknown - unknown entity in is...in
permit (
    principal is User,
    action == Action::"view",
    resource is Document in Unknown::"x"
);
