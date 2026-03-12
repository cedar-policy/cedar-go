permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { context.session.token == "abc" };
