// Policy 0: policy that applies to multiple actions, producing same error in each env
permit (
    principal is User,
    action,
    resource is Document
) when {
    1 + "bad"
};
