// Test subtraction with valid Long operands
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    5 - 3 == 2
};

// Test multiplication with valid Long operands
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    3 * 4 == 12
};

// Test invalid subtraction: string operand
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    5 - "two" == 0
};

// Test invalid multiplication: string operand
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    "three" * 4 == 0
};
