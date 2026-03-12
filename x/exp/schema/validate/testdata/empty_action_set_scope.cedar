// Policy 0: empty action set in scope (strict error)
permit (
    principal is User,
    action in [],
    resource is Document
);
