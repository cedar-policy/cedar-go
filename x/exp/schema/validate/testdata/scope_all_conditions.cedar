// Policy 0: all scopes are ScopeTypeAll with conditions
permit (
    principal,
    action,
    resource
)
when { principal.active };
