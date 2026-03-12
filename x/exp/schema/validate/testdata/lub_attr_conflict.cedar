// Policy 0: attribute access on entity LUB with incompatible attribute types
// TypeA has shared:Long, TypeB has shared:String - LUB fails
// Covers lookupEntityAttr LUB error path (cedar_type.go line 279-281)
permit (
    principal is TypeA,
    action == Action::"test",
    resource is Resource
) when {
    (if context has flag then TypeA::"a" else TypeB::"b").shared == "test"
};

// Policy 1: hasTag on entity LUB with incompatible tag types
// Separate policy to avoid && evaluation order issues
permit (
    principal is TypeA,
    action == Action::"test",
    resource is Resource
) when {
    (if context has flag then TypeA::"a" else TypeB::"b").hasTag("x")
};

// Policy 2: getTag on entity LUB with incompatible tag types
// TypeA has tags Long, TypeB has tags String - tag LUB fails
// Covers entityTagType LUB error path (cedar_type.go line 312-314)
permit (
    principal is TypeA,
    action == Action::"test",
    resource is Resource
) when {
    (if context has flag then TypeA::"a" else TypeB::"b").hasTag("x") &&
    (if context has flag then TypeA::"a" else TypeB::"b").getTag("x") == "y"
};
