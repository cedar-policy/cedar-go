permit (
    principal,
    action == Action::"act",
    resource
) when {
    [[], [1], true, false, principal has opt, 1, "s", principal, resource, {}, {"a": 1}, {"a": 1, "b": "x"}, ip("1.1.1.1"), datetime("2020-01-01T00:00:00Z")].isEmpty()
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    {"a": 1} == {"b": 1}
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    {"b": 1} == {"a": 1}
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    {"a": "x"} == {"a": 1}
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    [principal, resource, if principal has opt then principal else resource, 1].isEmpty()
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    if decimal("bad") then true else true
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    if 1 then true else false
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    if false then NoSuchType::"x" else true
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    if true then true else NoSuchType::"x"
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    [NoSuchType::"x", 1].isEmpty()
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    [1, "x"].isEmpty()
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    [principal, resource].isEmpty()
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    [decimal("bad"), 1].isEmpty()
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    [decimal("bad"), 1, "x"].isEmpty()
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    ip(if true then "x" else "x", "y") == ip("1.1.1.1")
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    ip(if true then "x" else "x") == ip("1.1.1.1")
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    ip(1) == ip("1.1.1.1")
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    ip("bad") == ip("1.1.1.1")
};

permit (
    principal,
    action == Action::"act",
    resource
) when {
    ip(decimal("bad"), "x") == ip("1.1.1.1")
};
