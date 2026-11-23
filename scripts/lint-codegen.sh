#!/bin/bash

# Verifies that `mise run generate` reproduces the committed generated code.
#
# Generation runs in a fresh checkout so untracked and ignored files in the
# developer's tree can't influence the result. Supports both jj and git.

set -euo pipefail

git_diff() {
	git add -A
	git diff --cached --exit-code
}

jj_diff() {
	if [ -n "$(jj diff --summary)" ]
	then
		jj diff --git
		return 1
	fi
}

source "$(dirname "$0")/lib/fresh-checkout.sh"

MISE_TRUSTED_CONFIG_PATHS="$dir" mise run generate
"${vcs}_diff"
