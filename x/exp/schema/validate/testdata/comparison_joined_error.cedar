// Policy 0: comparison where the RIGHT operand has a joined error
// -(principal == Unknown::"x") produces a joined error (unknown entity + type mismatch)
// Using it as the right operand of < exercises the right-side error flattening path
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    0 < -(principal == Unknown::"x")
};
