#!/bin/bash

# Verifies that the public API has no incompatible changes since the latest
# release, using gorelease.
#
# Packages under x/exp have no compatibility guarantee, so changes to them are
# ignored.
#
# Should the day come where we want to make a v2 with breaking changes, we'll
# have to revisit how this linter works. For now, we don't intend to do so.

set -euo pipefail

# gorelease refuses to run with uncommitted changes.
source "$(dirname "$0")/lib/fresh-checkout.sh"

# gorelease exits non-zero when it finds incompatible changes, which may all
# be in x/exp, so its exit status alone isn't enough.
report="$(gorelease 2>&1)" || true
if ! grep -qx '# summary' <<<"$report"
then
	echo "$report" >&2
	exit 1
fi

# Print each problem as "<package>: <message>": incompatible changes and
# errors in the release, outside x/exp, and any diagnostics.
problems="$(awk -v xexp="$(go list -m)/x/exp" '
	/^# / { pkg = substr($0, 3); section = ""; next }
	/^## / { section = substr($0, 4); next }
	/^$/ { next }
	pkg == "summary" { next }
	pkg == "diagnostics" { print "diagnostics: " $0; next }
	pkg == xexp || index(pkg, xexp "/") == 1 { next }
	section == "incompatible changes" || section == "errors in release version:" { print pkg ": " $0 }
' <<<"$report")"

sed -n '/^# summary$/,$p' <<<"$report"
if [ -n "$problems" ]
then
	printf '\nIncompatible changes since the latest release:\n%s\n' "$problems" >&2
	exit 1
fi
echo
echo "No incompatible changes."
