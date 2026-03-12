// Policy 0: LUB of typeTrue + typeFalse → typeBool
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { if principal has name then true else false };

// Policy 1: LUB of typeBool + typeTrue → typeBool (via has optional)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { if principal has name then (principal has email) else true };

// Policy 2: LUB of typeFalse + typeFalse → typeFalse
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { if principal has name then false else false };

// Policy 3: LUB of typeLong + typeLong → typeLong
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (if principal has name then 1 else 2) > 0 };

// Policy 4: LUB of typeString + typeString → typeString
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (if principal has name then "a" else "b") like "*" };

// Policy 5: LUB of typeSet + typeSet → typeSet
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (if principal has name then [1, 2] else [3, 4]).contains(1) };

// Policy 6: LUB of typeExtension + typeExtension → typeExtension
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (if principal has name then ip("10.0.0.1") else ip("10.0.0.2")).isLoopback() };

// Policy 7: LUB of typeEntity + typeEntity → typeEntity
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (if principal has name then User::"a" else User::"b") == principal };
