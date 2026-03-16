// Policy 0: getTag with tag name derived from resource variable (covers resource lubKey path)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag(resource.name) == "public"
};
