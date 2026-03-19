permit(
  principal,
  action,
  resource
) when {
  principal != principal
};

permit(
  principal,
  action,
  resource
) when {
  if (if ([].containsAll([])) then (action == action) else (action == action)) then true else false
};

permit(
  principal,
  action,
  resource
) when {
  if (if ([].containsAll([])) then (action != action) else (action != action)) then true else ((1 + "x") == 1)
};

permit(
  principal,
  action,
  resource
) when {
  [1 + "a", true, false].contains(true)
};

permit(
  principal,
  action,
  resource
) when {
  [if ([].containsAll([])) then principal else principal, resource].contains(principal)
};

permit(
  principal,
  action,
  resource
) when {
  ip(context.rec.a, "x") == ip(context.rec.a, "x")
};

permit(
  principal,
  action,
  resource
) when {
  context.rec == { a: 1 }
};

permit(
  principal,
  action,
  resource
) when {
  principal.getTag(resource.r)
};

permit(
  principal,
  action,
  resource
) when {
  [].containsAll([])
};

permit(
  principal,
  action,
  resource
) when {
  if true then context.flag else 1
};

permit(
  principal,
  action,
  resource
) when {
  [if context.flag then action else (if context.flag then principal else resource), Q::"q1"].contains(principal)
};

permit(
  principal,
  action,
  resource
) when {
  [if 1 then principal else principal, principal, Q::"q1"].contains(principal)
};

permit(
  principal,
  action,
  resource
) when {
  if [] then true else false
};

permit(
  principal,
  action,
  resource
) when {
  [context.flag, 1, 2].contains(1)
};

permit(
  principal,
  action,
  resource
) when {
  [[], [1], ["x"]].contains([])
};
