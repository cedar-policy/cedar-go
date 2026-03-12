// if-then-else with non-constant condition and incompatible entity types in branches
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if principal.active then principal else resource) == principal
};
