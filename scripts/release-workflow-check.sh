#!/usr/bin/env bash
# Checks .github/workflows/release.yml against docs/release.md (Acceptance).
set -euo pipefail

f=.github/workflows/release.yml
fail() { echo "release-workflow-check: $1"; exit 1; }

[ -f "$f" ] || fail "$f is missing"
actionlint "$f"

# Starts on v* tags only: no branch, PR or manual trigger.
grep -q 'tags:' "$f" || fail "no tag trigger"
grep -Eq '^\s*-\s*"?v\*"?' "$f" || fail "tag trigger is not v*"
grep -Eq '^\s*(branches|pull_request|workflow_dispatch):' "$f" && fail "must not run on branches, PRs or manual dispatch"

# Token guard runs before goreleaser.
grep -q 'HOMEBREW_TAP_TOKEN is not set' "$f" || fail "no token guard message"
guard=$(grep -n 'HOMEBREW_TAP_TOKEN is not set' "$f" | head -1 | cut -d: -f1)
run=$(grep -n 'goreleaser-action' "$f" | head -1 | cut -d: -f1)
[ "$guard" -lt "$run" ] || fail "token guard must come before goreleaser"

# Tests, then goreleaser; release needs write access.
grep -q 'go test ./...' "$f" || fail "no go test step"
grep -q 'contents: write' "$f" || fail "needs contents: write"
grep -q 'release --clean' "$f" || fail "goreleaser must run: release --clean"
echo "release-workflow-check: ok"
