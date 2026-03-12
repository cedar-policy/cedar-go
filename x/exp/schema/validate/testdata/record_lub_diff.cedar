@id("strict_record_lub")
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (if principal.active then {"a": 1} else {"b": 2}) == {"a": 1} };

@id("permissive_record_lub")
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (if principal.active then {"a": 1, "c": 3} else {"a": 2, "d": 4}).a > 0 };
