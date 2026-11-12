#!/usr/bin/env bash
# merge-main_test.sh — hermetic tests for tools/merge-main.sh. Every case runs
# in a throwaway origin/clone pair; the real repo is never touched.

set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REAL="$(cd "$HERE/.." && pwd)"
[ -x "$HERE/merge-main.sh" ] || { echo "FATAL: merge-main.sh not executable" >&2; exit 2; }

fails=0
assert_eq() {
  if [ "$2" = "$3" ]; then echo "ok   - $1"
  else echo "FAIL - $1 (want '$2', got '$3')" >&2; fails=$((fails + 1)); fi
}
assert_contains() {
  case "$3" in
    *"$2"*) echo "ok   - $1" ;;
    *) echo "FAIL - $1 (want substring '$2', got '$3')" >&2; fails=$((fails + 1)) ;;
  esac
}

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_SYSTEM=/dev/null
export GIT_AUTHOR_NAME=t GIT_AUTHOR_EMAIL=t@t.t GIT_COMMITTER_NAME=t GIT_COMMITTER_EMAIL=t@t.t
n=0
W=""
# mkfix: fresh origin + work clone on branch main; sets W.
mkfix() {
  n=$((n + 1))
  local d="$tmp/fx$n"
  mkdir -p "$d"
  git init -q --bare -b main "$d/origin.git"
  git clone -q "$d/origin.git" "$d/work" 2>/dev/null
  W="$d/work"
  cd "$W"
  git checkout -q -b main 2>/dev/null || true
  mkdir -p tools docs
  cp "$REAL/tools/merge-main.sh" tools/
  printf 'base' > docs/other.txt
  printf 'sum-base\n' > go.work.sum
  git add -A
  git commit -qm base
  git push -q origin main
}

# diverge [other|sum]: task-1 and main each add a distinct file; main pushed.
#   other: both sides also edit docs/other.txt (real conflict)
#   sum:   main edits go.work.sum (branch copy must win)
diverge() {
  git checkout -q -b task-1
  printf 'A' > docs/A.txt; git add docs/A.txt
  [ "${1:-}" != other ] || printf 'a' > docs/other.txt
  git commit -qam a
  git checkout -q main
  printf 'B' > docs/B.txt; git add docs/B.txt
  [ "${1:-}" != other ] || printf 'b' > docs/other.txt
  [ "${1:-}" != sum ] || printf 'sum-main\n' > go.work.sum
  git commit -qam b
  git push -q origin main
  git checkout -q task-1
}

run() { # sets rc, out, err
  rc=0
  tools/merge-main.sh "$@" > "$tmp/out" 2> "$tmp/err" || rc=$?
  out="$(cat "$tmp/out")"; err="$(cat "$tmp/err")"
}
has_merge_head() { git rev-parse -q --verify MERGE_HEAD >/dev/null 2>&1 && echo yes || echo no; }

# --- clean merge, then up to date
mkfix; diverge
run
assert_eq "clean: rc" 0 "$rc"
assert_eq "clean: two parents" 2 "$(git log -1 --format=%P | wc -w)"
assert_eq "clean: has A" A "$(cat docs/A.txt)"
assert_eq "clean: has B" B "$(cat docs/B.txt)"
assert_eq "clean: tree clean" "" "$(git status --porcelain)"
head1="$(git rev-parse HEAD)"
run
assert_eq "uptodate: rc" 0 "$rc"
assert_eq "uptodate: stdout" "merge-main: already up to date" "$out"
assert_eq "uptodate: no new commit" "$head1" "$(git rev-parse HEAD)"

# --- unstaged go.work.sum churn is discarded; staged merge result wins
mkfix; diverge sum
git checkout -q main; printf 'sum-main2\n' > go.work.sum; git commit -qam sum2; git push -q origin main; git checkout -q task-1
git fetch -q origin main
git merge -q --no-ff --no-commit origin/main
printf 'churn\n' > go.work.sum
run --continue
assert_eq "sum: rc" 0 "$rc"
assert_eq "sum: merged copy committed" sum-main2 "$(git show HEAD:go.work.sum)"
assert_eq "sum: tree clean" "" "$(git status --porcelain)"

# --- go.work.sum churn before the merge does not block it
mkfix; diverge sum
printf 'churn\n' > go.work.sum
run
assert_eq "pre-churn: rc" 0 "$rc"
assert_eq "pre-churn: main's copy merged" sum-main "$(git show HEAD:go.work.sum)"
assert_eq "pre-churn: tree clean" "" "$(git status --porcelain)"

# --- real conflict, then continue
mkfix; diverge other
run
assert_eq "conflict: rc" 1 "$rc"
assert_contains "conflict: lists other.txt" docs/other.txt "$out"
assert_eq "conflict: MERGE_HEAD" yes "$(has_merge_head)"
mh="$(git rev-parse MERGE_HEAD)"
run
assert_eq "mid-merge: rc" 1 "$rc"
assert_contains "mid-merge: stderr" "merge in progress; resolve and run tools/merge-main.sh --continue" "$err"
assert_eq "mid-merge: MERGE_HEAD unchanged" "$mh" "$(git rev-parse MERGE_HEAD)"
printf 'ab' > docs/other.txt
git add docs/other.txt
run --continue
assert_eq "continue: rc" 0 "$rc"
assert_eq "continue: two parents" 2 "$(git log -1 --format=%P | wc -w)"
assert_eq "continue: other.txt" ab "$(cat docs/other.txt)"

# --- continue with nothing
mkfix
git checkout -q -b task-1
run --continue
assert_eq "continue none: rc" 3 "$rc"
assert_contains "continue none: stderr" "no merge in progress" "$err"

# --- usage
run --bogus
assert_eq "usage: rc" 3 "$rc"

# --- no-fetch merges the local origin/main ref
mkfix; diverge
bare_main="$(git rev-parse origin/main)"
git checkout -q --detach main
git commit -q --allow-empty -m local-only
local_only="$(git rev-parse HEAD)"
git checkout -q task-1
git update-ref refs/remotes/origin/main "$local_only"
MERGE_MAIN_NO_FETCH=1 run
assert_eq "no-fetch: rc" 0 "$rc"
assert_eq "no-fetch: ref unchanged" "$local_only" "$(git rev-parse origin/main)"
assert_eq "no-fetch: merged local ref" "$local_only" "$(git rev-parse HEAD^2)"
[ "$bare_main" != "$local_only" ] || { echo "FAIL - no-fetch fixture" >&2; fails=$((fails + 1)); }

# --- dirty tree: merge refuses to start
mkfix; diverge
printf 'dirty' > docs/B.txt
pre="$(git rev-parse HEAD)"
run
[ "$rc" -ne 0 ] && r=nonzero || r=zero
assert_eq "dirty: nonzero rc" nonzero "$r"
assert_eq "dirty: no MERGE_HEAD" no "$(has_merge_head)"
assert_eq "dirty: nothing staged" "" "$(git diff --cached --name-only)"
assert_eq "dirty: no new commit" "$pre" "$(git rev-parse HEAD)"

cd "$REAL"
[ "$fails" -eq 0 ] || { echo "$fails failure(s)" >&2; exit 1; }
echo "all ok"
