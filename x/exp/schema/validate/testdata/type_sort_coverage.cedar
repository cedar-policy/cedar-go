// Policy 0: set with Bool + Bool + Long — exercises cedarTypeCompare simple-types case
// and cedarTypeOrdinal for typeBool
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [principal.a, principal.b, 1].isEmpty()
};

// Policy 1: set with two different extension types — exercises cedarTypeCompare extension case
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [ip("1.2.3.4"), decimal("1.0")].isEmpty()
};

// Policy 2: set with entities + non-entity — exercises cedarTypeCompare entity case
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [User::"a", Document::"b", 1].isEmpty()
};

// Policy 3: if-then-else with empty set vs Long in permissive — exercises cedarTypeName(typeNever)
// In permissive mode, [] has type Set<Never>, producing incompatibility with Long.
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if principal.a then [] else 1
};
