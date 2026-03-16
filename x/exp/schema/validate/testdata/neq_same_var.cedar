// Policy 0: principal != principal (negated variable equality)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal != principal
};
