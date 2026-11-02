permit(principal, action, resource) when { 1 < datetime("2020-01-01") };
permit(principal, action, resource) when { datetime("2020-01-01") < 1 };
permit(principal, action, resource) when { datetime("2020-01-01") < duration("1h") };
