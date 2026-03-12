// Policy 0: set with one erroring element and one non-erroring element of different types
// triggers error recovery LUB path where LUB fails (Long vs String)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [(1 + "bad"), "hello"].isEmpty()
};
