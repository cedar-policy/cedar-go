// Policy 0: containsAll with sub-error AND incompatible set types in strict
// Left set has error element, right set has different element type
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [1, 1 + "bad"].containsAll(["foo"])
};
