// Policy 0: LUB of typeFalse + typeTrue via if-then-else (covers LUB bool path)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if principal has name then false else true
};

// Policy 1: Dead code with extension call containing invalid entity ref
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && ip(BadType::"x")
};

// Policy 2: principal in empty set (permissive ok, strict error for empty set)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal in []
};

// Policy 3: action ScopeTypeAll with noop action (no AppliesTo) in schema
permit (
    principal is User,
    action,
    resource is Document
) when {
    true
};

// Policy 4: `is` on multi-entity LUB returns typeBool (not typeTrue/typeFalse)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if principal has name then principal else resource) is User
};

// Policy 5: `has` on non-variable expression (record literal) covers exprVarName return ""
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    {a: 1, b: "hello"} has a
};

// Policy 6: ip() with non-string literal in strict mode (strict error only)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip(42).isLoopback()
};

// Policy 7: hasTag with non-literal key (covers tagCapabilityKey non-NodeValue branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource.hasTag(context.tagKey)
};

// Policy 8: getTag with non-literal key without hasTag guard (covers formatTagExpr non-literal)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource.getTag(context.tagKey) == "value"
};
