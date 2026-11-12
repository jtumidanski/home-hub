#!/usr/bin/env bash
# merge-main.sh — merge origin/main into the current branch.
#
#   tools/merge-main.sh              start the merge
#   tools/merge-main.sh --continue   finish after resolving real conflicts
#
# Exit codes: 0 merged or up to date, 1 real conflicts / merge in progress,
# 3 nothing to continue / usage.
#
# Env: MERGE_MAIN_NO_FETCH=1 to skip the fetch.
#
# Used by the /execute-task branch end (claude-drive protocol): exit 1 or 3
# there is a NEEDS OPERATOR stop.

set -euo pipefail

continue_mode=0
case "${1:-}" in
  "") ;;
  --continue) continue_mode=1 ;;
  *) echo "usage: merge-main.sh [--continue]" >&2; exit 3 ;;
esac

merge_in_progress() { git rev-parse -q --verify MERGE_HEAD >/dev/null 2>&1; }

report_unmerged() {
  local unmerged
  unmerged="$(git diff --name-only --diff-filter=U)"
  [ -n "$unmerged" ] || return 0
  echo "merge-main: unmerged paths:"
  echo "$unmerged"
  echo "merge-main: resolve them, git add, then run tools/merge-main.sh --continue"
  exit 1
}

if [ "$continue_mode" -eq 1 ]; then
  if ! merge_in_progress; then
    echo "merge-main: no merge in progress" >&2
    exit 3
  fi
  report_unmerged
else
  if merge_in_progress; then
    echo "merge-main: merge in progress; resolve and run tools/merge-main.sh --continue" >&2
    exit 1
  fi
  if [ "${MERGE_MAIN_NO_FETCH:-}" != "1" ]; then
    git fetch origin main
  fi
  if git merge-base --is-ancestor origin/main HEAD; then
    echo "merge-main: already up to date"
    exit 0
  fi
  merge_rc=0
  git merge --no-ff --no-commit origin/main || merge_rc=$?
  if [ "$merge_rc" -ne 0 ] && ! merge_in_progress; then
    echo "merge-main: git merge failed to start (exit $merge_rc); nothing staged or committed" >&2
    exit 1
  fi
  report_unmerged
fi

# Discard unstaged go.work.sum churn (go tooling rewrites it) so the merge
# commit leaves a clean tree. The staged merge result is untouched.
if [ -n "$(git ls-files -m -- go.work.sum)" ]; then
  git checkout -- go.work.sum
fi
git commit --no-edit
exit 0
