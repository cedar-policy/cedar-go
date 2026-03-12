// Policy 0: is...in with non-entity right operand
// Exercises typeOfIsIn right-operand non-entity error path
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal is User in "hello"
};
