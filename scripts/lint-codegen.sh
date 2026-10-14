#!/bin/bash

# Verifies that `mise run generate` reproduces the committed generated code.
#
# Generation runs in a fresh checkout so untracked and ignored files in the
# developer's tree can't influence the result. Supports both jj and git.

set -euo pipefail

#        _ _
#   __ _(_) |_
#  / _` | | __|
# | (_| | | |_
#  \__, |_|\__|
#  |___/
git_create() {
	git worktree add --detach --force "$dir" HEAD >/dev/null
}

git_diff() {
	git add -A
	git diff --cached --exit-code
}

git_cleanup() {
	git -C "$repo" worktree remove --force "$dir" >/dev/null 2>&1 || true
}

#    _  _
#   (_)(_)
#   | || |
#   | || |
#  _/ |/ |
# |__/__/
jj_workspace="lint-codegen-$$"

jj_create() {
	jj workspace add --quiet --name "$jj_workspace" --revision @ "$dir"
}

jj_diff() {
	if [ -n "$(jj diff --summary)" ]
	then
		jj diff --git
		return 1
	fi
}

jj_cleanup() {
	# Forgetting a workspace only discards its working-copy commit when it's
	# empty, so abandon it to avoid leaking the regenerated diff.
	jj -R "$repo" abandon --quiet "$jj_workspace@" >/dev/null 2>&1 || true
	jj -R "$repo" workspace forget "$jj_workspace" >/dev/null 2>&1 || true
}

#   ___ ___  _ __ ___  _ __ ___   ___  _ __
#  / __/ _ \| '_ ` _ \| '_ ` _ \ / _ \| '_ \
# | (_| (_) | | | | | | | | | | | (_) | | | |
#  \___\___/|_| |_| |_|_| |_| |_|\___/|_| |_|
if repo="$(jj root 2>/dev/null)"
then
	vcs=jj
else
	vcs=git
	repo="$(git rev-parse --show-toplevel)"
fi

dir="$(mktemp -d)"
trap '"${vcs}_cleanup"; rm -rf "$dir"' EXIT

"${vcs}_create"
cd "$dir"
MISE_TRUSTED_CONFIG_PATHS="$dir" mise run generate
"${vcs}_diff"
