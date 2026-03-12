// Policy 0: is...in with unknown entity type on left
// Covers typeOfIsIn leftErr path (lines 628-630, 640-642)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    Unknown::"x" is User in resource
};

// Policy 1: is...in with unknown entity type on right
// Covers typeOfIsIn rightErr path (line 648-650)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal is User in Unknown::"x"
};

// Policy 2: equality with joined error on right side
// Inner equality Foo::"a" == Bar::"b" produces joined error
// Outer == right-side error flattening (line 341-343)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true == (Foo::"a" == Bar::"b")
};

// Policy 3: hasTag with right operand error
// Covers typeOfHasTag right error path (line 784-786)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.hasTag(principal.nonexistent)
};

// Policy 4: getTag with right operand error
// Covers typeOfGetTag right error path (lines 829-831)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.hasTag("x") && principal.getTag(principal.nonexistent) == "y"
};

// Policy 5: getTag with non-string key (Long)
// Covers typeOfGetTag non-string key (lines 833-835) and tagCapabilityKey non-string (line 1194-1196)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.hasTag("x") && principal.getTag(1) == "y"
};

// Policy 6: record literal with joined error in element
// Inner equality Foo::"a" == Bar::"b" produces joined error
// Covers typeOfRecord error flattening (lines 895-897)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    {"k": (Foo::"a" == Bar::"b")} == {"k": true}
};
