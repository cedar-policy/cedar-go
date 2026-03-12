// Policy 0: action in top-level action, requiring recursive descendant search
// leaf -> mid -> top, so isActionDescendant(leaf, top) needs recursion
permit (
    principal is User,
    action in Action::"top",
    resource is Document
);
