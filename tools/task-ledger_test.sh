#!/usr/bin/env bash
# task-ledger_test.sh — hermetic tests for tools/task-ledger.sh.

set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
[ -x "$HERE/task-ledger.sh" ] || { echo "FATAL: $HERE/task-ledger.sh not executable" >&2; exit 2; }

fails=0
assert_eq()  { if [ "$2" = "$3" ]; then echo "ok   - $1"; else echo "FAIL - $1 (want '$2', got '$3')" >&2; fails=$((fails+1)); fi; }
assert_has() { case "$3" in *"$2"*) echo "ok   - $1";; *) echo "FAIL - $1 (missing '$2')" >&2; fails=$((fails+1));; esac; }

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
repo="$tmp/repo"
mkdir -p "$repo/docs/tasks/task-300-ledger"
git -C "$repo" init -q -b main
git -C "$repo" config user.email t@t.t
git -C "$repo" config user.name t
git -C "$repo" config commit.gpgsign false
echo prd > "$repo/docs/tasks/task-300-ledger/prd.md"
git -C "$repo" add -A
git -C "$repo" commit -qm base
git -C "$repo" tag base

stub="$tmp/resolve.sh"
cat > "$stub" <<STUB
#!/usr/bin/env bash
case "\$1" in
  300|task-300-ledger) printf 'task-300-ledger\tdocs/tasks/task-300-ledger\t$repo\n' ;;
  *) exit 3 ;;
esac
STUB
chmod +x "$stub"
export TASK_RESOLVE="$stub"

L="$HERE/task-ledger.sh"
canon="$repo/docs/tasks/task-300-ledger/progress.md"
legacy="$repo/.superpowers/sdd/plan/progress.md"
reset() { git -C "$repo" reset -q --hard base; rm -rf "$repo/.superpowers" "$canon" "$legacy"; }
mklegacy() { mkdir -p "$(dirname "$legacy")"; printf '%s\nTask 1: complete (x)\n' "$1" > "$legacy"; }
hdr='# SDD ledger — plan: docs/tasks/task-300-ledger/plan.md'

out="$("$L" path 300)"; rc=$?
assert_eq "path" "$canon" "$out"; assert_eq "path rc" 0 "$rc"
"$L" path 999 >/dev/null 2>&1; assert_eq "path unresolved" 3 "$?"

"$L" init 300 >/dev/null
assert_eq "init creates header" "$hdr" "$(head -1 "$canon")"
echo "Task 1: complete (x)" >> "$canon"; before="$(cat "$canon")"
"$L" init 300 >/dev/null
assert_eq "init idempotent" "$before" "$(cat "$canon")"

reset
assert_eq "migrate none" "no ledger" "$("$L" migrate 300)"

reset; mklegacy "$hdr"; orig="$(cat "$legacy"; echo x)"
assert_eq "migrate legacy only" "migrated" "$("$L" migrate 300)"
assert_eq "migrate bytes equal" "$orig" "$(cat "$canon"; echo x)"
gone=0; [ -e "$legacy" ] && gone=1; assert_eq "migrate legacy gone" 0 "$gone"
assert_eq "migrate commit" "chore(task-300-ledger): migrate progress ledger" "$(git -C "$repo" log -1 --format=%s)"

reset; "$L" init 300 >/dev/null
assert_eq "migrate canonical only" "canonical" "$("$L" migrate 300)"

reset; "$L" init 300 >/dev/null; mklegacy "$hdr"
c1="$(cat "$canon")"; l1="$(cat "$legacy")"
err="$("$L" migrate 300 2>&1 >/dev/null)"; rc=$?
assert_eq "migrate both rc" 1 "$rc"
assert_has "migrate both names canonical" "$canon" "$err"
assert_has "migrate both names legacy" "$legacy" "$err"
assert_eq "migrate both canonical untouched" "$c1" "$(cat "$canon")"
assert_eq "migrate both legacy untouched" "$l1" "$(cat "$legacy")"

reset; mklegacy '# SDD ledger — plan: docs/tasks/task-301-other/plan.md'; l1="$(cat "$legacy")"
assert_eq "migrate foreign legacy" "no ledger" "$("$L" migrate 300)"
assert_eq "migrate foreign legacy untouched" "$l1" "$(cat "$legacy")"

reset; "$L" init 300 >/dev/null; mklegacy "$hdr"
assert_eq "strays lists legacy" "$legacy" "$("$L" strays 300)"

reset; mklegacy '# SDD ledger — plan: docs/tasks/task-301-other/plan.md'
assert_eq "strays ignores other task" "" "$("$L" strays 300)"

reset; mklegacy '# SDD ledger — plan: docs/tasks/task-300-ledgerx/plan.md'
assert_eq "strays ignores prefix-sharing task" "" "$("$L" strays 300)"

reset; "$L" init 300 >/dev/null
assert_eq "strays ignores canonical" "" "$("$L" strays 300)"

[ "$fails" -eq 0 ] || { echo "$fails failure(s)" >&2; exit 1; }
