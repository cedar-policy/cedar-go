permit (principal, action, resource)
when { context.rec == { req: 1, opt: 1 } };

permit (principal, action, resource)
when { context.rec != { req: 1, opt: 1 } };

permit (principal, action, resource)
when { (if principal == resource then context.rec else { req: 1, opt: 1 }).req == 1 };

permit (principal, action, resource)
when { [context.rec, { req: 1, opt: 1 }].contains(context.rec) };

permit (principal, action, resource)
when { [context.rec].contains({ req: 1, opt: 1 }) };

permit (principal, action, resource)
when { context.nest == { inner: { a: 1 } } };

permit (principal, action, resource)
when { context.rec == context.rec2 };

permit (principal, action, resource)
when { context.rec == { req: 1 } };

permit (principal, action, resource)
when { context.rec == context.rec3 };

permit (principal, action, resource)
when { [context.rec, context.rec3, context.rec].isEmpty() };
