#!/bin/bash

set -euo pipefail

max_input_bytes=20000
max_log_bytes=10000

new_inputs() {
	git ls-files --others -- ':(glob)**/testdata/fuzz/**'
}

collect() {
	local dest=$1
	mkdir -p "$dest"
	new_inputs | while IFS= read -r f; do
		mkdir -p "$dest/$(dirname "$f")"
		cp "$f" "$dest/$f"
	done
}

hashes() {
	new_inputs | while IFS= read -r f; do
		basename "$f"
	done
}

reported() {
	local seen=$1
	local h found=
	while IFS= read -r h; do
		found=1
		grep -qF "$h" "$seen" || return 1
	done < <(hashes)
	[ -n "$found" ]
}

body() {
	local log=$1
	local inputs
	inputs=$(new_inputs)

	echo "The nightly fuzz run failed."
	echo
	echo "- Run: $RUN_URL"
	echo "- Commit: $SHA"
	if [ -n "${ARTIFACT_URL:-}" ]; then
		echo "- Failing inputs artifact: $ARTIFACT_URL"
	fi
	echo

	if [ -z "$inputs" ]; then
		echo "No new fuzz input was written. See the log tail below and the run for the cause."
		echo
	else
		while IFS= read -r f; do
			local pkg=${f%%/testdata/fuzz/*}
			local target hash
			target=$(basename "$(dirname "$f")")
			hash=$(basename "$f")

			echo "## \`$target\` in \`./$pkg\`"
			echo
			if [ "$(wc -c <"$f")" -gt "$max_input_bytes" ]; then
				echo "The input is too large to include here. Download it from the artifact and save it as \`$f\`, then run:"
				echo
				echo '```sh'
			else
				echo "Recreate the input and run it:"
				echo
				echo '```sh'
				echo "mkdir -p $(dirname "$f")"
				echo "cat >$f <<'FUZZ_INPUT_EOF'"
				cat "$f"
				[ -z "$(tail -c 1 "$f")" ] || echo
				echo "FUZZ_INPUT_EOF"
			fi
			echo "go test -run=$target/$hash ./$pkg"
			echo '```'
			echo
			echo "Once fixed, commit \`$f\` so it runs as a regression test."
			echo
		done <<<"$inputs"
	fi

	if [ -s "$log" ]; then
		echo "## Log tail"
		echo
		echo '~~~'
		tail -c "$max_log_bytes" "$log"
		[ -z "$(tail -c 1 "$log")" ] || echo
		echo '~~~'
	fi
}

issue() {
	local log=$1
	local number
	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT

	body "$log" >"$tmp/body.md"
	number=$(gh issue list --repo "$GITHUB_REPOSITORY" --state open \
		--search "\"$TITLE\" in:title" --json number,title \
		--jq 'map(select(.title == env.TITLE)) | .[0].number // empty')
	if [ -z "$number" ]; then
		gh issue create --repo "$GITHUB_REPOSITORY" --title "$TITLE" --body-file "$tmp/body.md"
		return
	fi
	gh issue view "$number" --repo "$GITHUB_REPOSITORY" --json body,comments \
		--jq '.body, .comments[].body' >"$tmp/seen.txt"
	if reported "$tmp/seen.txt"; then
		echo "Issue #$number already reports these inputs"
		return
	fi
	gh issue comment "$number" --repo "$GITHUB_REPOSITORY" --body-file "$tmp/body.md"
}

case ${1:-} in
collect) collect "$2" ;;
issue) issue "$2" ;;
*)
	echo "usage: $0 collect <dir> | issue <log>" >&2
	exit 2
	;;
esac
