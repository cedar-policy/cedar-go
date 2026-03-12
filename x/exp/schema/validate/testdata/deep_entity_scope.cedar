// Policy 0: principal in Org scope - tests transitive closure in getEntityTypesIn
// Org has child Group, Group has child User, so all three are returned
permit (
    principal in Org::"acme",
    action == Action::"view",
    resource is Document
);
