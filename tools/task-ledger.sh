#!/usr/bin/env bash
# task-ledger.sh — the one canonical progress ledger per task.
#
# Canonical path: <worktree>/docs/tasks/<id>/progress.md
# Legacy path:    <worktree>/.superpowers/sdd/plan/progress.md
#
# Usage:
#   task-ledger.sh path <task>      Print the canonical ledger path
#   task-ledger.sh init <task>      Create the ledger with its header if absent; print path
#   task-ledger.sh migrate <task>   Move a legacy ledger to the canonical path
#                                   (prints: no ledger | migrated | canonical;
#                                   exit 1 and both paths on stderr on conflict)
#   task-ledger.sh strays <task>    Print non-canonical progress.md files naming the task
#
# Resolver exit codes (3 none, 4 ambiguous) propagate. TASK_RESOLVE overrides
# the resolver path.

set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESOLVE="${TASK_RESOLVE:-$HERE/task-resolve.sh}"

usage() { sed -n '2,16p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//' >&2; exit 2; }

[ $# -eq 2 ] || usage
cmd="$1"; query="$2"

rc=0
resolved="$("$RESOLVE" "$query")" || rc=$?
[ "$rc" -eq 0 ] || exit "$rc"
IFS=$'\t' read -r id _ worktree <<<"$resolved"

canonical="$worktree/docs/tasks/$id/progress.md"
legacy="$worktree/.superpowers/sdd/plan/progress.md"

# names_task <file>: the first 3 lines name the id as a whole token.
names_task() {
  head -3 "$1" | grep -Eq "(^|[^A-Za-z0-9_-])${id}(/|[[:space:]]|$)"
}

case "$cmd" in
  path)
    echo "$canonical"
    ;;
  init)
    if [ ! -e "$canonical" ]; then
      mkdir -p "$(dirname "$canonical")"
      printf '# SDD ledger — plan: docs/tasks/%s/plan.md\n' "$id" > "$canonical"
    fi
    echo "$canonical"
    ;;
  migrate)
    have_legacy=0
    if [ -f "$legacy" ] && names_task "$legacy"; then have_legacy=1; fi
    if [ -f "$canonical" ] && [ "$have_legacy" -eq 1 ]; then
      echo "$canonical" >&2
      echo "$legacy" >&2
      exit 1
    elif [ -f "$canonical" ]; then
      echo canonical
    elif [ "$have_legacy" -eq 1 ]; then
      mkdir -p "$(dirname "$canonical")"
      cp "$legacy" "$canonical"
      rm "$legacy"
      git -C "$worktree" add "$canonical"
      git -C "$worktree" commit -q -m "chore($id): migrate progress ledger" -- "$canonical"
      echo migrated
    else
      echo "no ledger"
    fi
    ;;
  strays)
    dirs=()
    for d in "$worktree/docs/tasks" "$worktree/.superpowers"; do
      [ -d "$d" ] && dirs+=("$d")
    done
    [ ${#dirs[@]} -gt 0 ] || exit 0
    while IFS= read -r f; do
      [ "$f" = "$canonical" ] && continue
      names_task "$f" && echo "$f"
    done < <(find "${dirs[@]}" -name progress.md -type f | sort)
    ;;
  *)
    usage
    ;;
esac
