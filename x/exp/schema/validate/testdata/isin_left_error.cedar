// Policy 0: is...in where the left operand has a sub-expression error
// AND returns a non-entity type. (1 + "hello") returns typeLong with error,
// triggering the error propagation path in typeOfIsIn.
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (1 + "hello") is User in User::"alice"
};
