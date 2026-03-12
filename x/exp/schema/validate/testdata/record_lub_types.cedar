// Policy 0: records with different number of attributes (strict error)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (if principal.active then {"a": 1} else {"a": 1, "b": 2}) == {"a": 1} };

// Policy 1: records with same key but incompatible types (strict error, permissive drops attr)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (if principal.active then {"x": 1} else {"x": "hello"}) == {"x": 1} };
