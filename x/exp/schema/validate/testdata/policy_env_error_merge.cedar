permit (
  principal,
  action == Action::"act",
  resource
) when {
  (principal is A && decimal("bad") == decimal("1.0")) || decimal("bad") == decimal("1.0")
};
