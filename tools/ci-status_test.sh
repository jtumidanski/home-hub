#!/usr/bin/env bash
# ci-status_test.sh — CLI suite for tools/ci-status.sh.
#
# A bash stub `tea` is placed first on PATH. It serves fixtures from $STUB_FIX:
#   pulls/<n>            -> pulls.<k>.json (k = pulls call count, last file reused);
#                           if pulls.always-moves exists, a fresh head SHA (dyn<k>) every call
#   commits/<sha>/status -> status.<sha>.json, else status.<k>.json (k = status call count)
#   $STUB_FIX/tea.rc     -> if present, tea exits with that code (stdout from tea.out)
set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="${CI_STATUS_SCRIPT:-$HERE/ci-status.sh}"
REAL="$HERE/testdata/ci-status"
[ -x "$SCRIPT" ] || { echo "FATAL: ci-status.sh not executable" >&2; exit 2; }

fails=0
assert_eq() { if [ "$2" = "$3" ]; then echo "ok   - $1"; else echo "FAIL - $1 (want '$2', got '$3')" >&2; fails=$((fails + 1)); fi; }
assert_has() { case "$3" in *"$2"*) echo "ok   - $1" ;; *) echo "FAIL - $1 (missing '$2' in '$3')" >&2; fails=$((fails + 1)) ;; esac; }
assert_not_has() { case "$3" in *"$2"*) echo "FAIL - $1 (unexpected '$2' in '$3')" >&2; fails=$((fails + 1)) ;; *) echo "ok   - $1" ;; esac; }

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin"
cat > "$tmp/bin/tea" <<'STUB'
#!/usr/bin/env bash
[ "$1" = api ] || exit 9
if [ -f "$STUB_FIX/tea.rc" ]; then
    [ -f "$STUB_FIX/tea.out" ] && cat "$STUB_FIX/tea.out"
    exit "$(cat "$STUB_FIX/tea.rc")"
fi
bump() { local n=0; [ -f "$1" ] && n=$(cat "$1"); n=$((n + 1)); echo "$n" > "$1"; echo "$n"; }
pick() { # pick PREFIX K -> highest existing PREFIX.<=K>.json
    local k=$2
    while [ "$k" -gt 0 ]; do
        [ -f "$1.$k.json" ] && { cat "$1.$k.json"; return; }
        k=$((k - 1))
    done
    exit 1
}
case "$2" in
    */pulls/*)
        k="$(bump "$STUB_FIX/pulls.count")"
        if [ -f "$STUB_FIX/pulls.always-moves" ]; then
            printf '{"state":"open","merged":false,"head":{"sha":"dyn%s"}}\n' "$k"
        else pick "$STUB_FIX/pulls" "$k"; fi ;;
    */commits/*/status)
        sha="${2#*/commits/}"; sha="${sha%/status}"
        k="$(bump "$STUB_FIX/status.count")"
        if [ -f "$STUB_FIX/status.$sha.json" ]; then cat "$STUB_FIX/status.$sha.json"; else pick "$STUB_FIX/status" "$k"; fi ;;
    *) exit 8 ;;
esac
STUB
chmod +x "$tmp/bin/tea"

COMPLETE='PR Validation / PR Validation Complete (pull_request)'

# pr_json STATE MERGED SHA
pr_json() { jq -n --arg s "$1" --argjson m "$2" --arg h "$3" '{state:$s,merged:$m,head:{sha:$h}}'; }
# st CONTEXT STATUS DESC -> one status object
st() { jq -n --arg c "$1" --arg s "$2" --arg d "$3" '{context:$c,status:$s,description:$d,target_url:"/runs/1/jobs/1"}'; }
# status_json OBJ... -> status document
status_json() { printf '%s\n' "$@" | jq -s '{statuses:.}'; }

# new_case NAME -> sets $fix
new_case() { fix="$tmp/$1"; mkdir -p "$fix"; }

# run_case ARGS... -> sets $out and $rc
run_case() {
    out="$(STUB_FIX="$fix" PATH="$tmp/bin:$PATH" CI_STATUS_INTERVAL="${INTERVAL:-0}" "$SCRIPT" "$@" 2>&1)"
    rc=$?
}

new_case success
cp "$REAL/pulls-229.json" "$fix/pulls.1.json"
cp "$REAL/status-229.json" "$fix/status.1.json"
run_case 229 --timeout 3s
assert_eq "success: exit" 0 "$rc"
assert_has "success: head sha" "39b30f67ead66c3beededf19d870af4d9ad78377" "$out"
assert_has "success: verdict" "PR #229: success" "$out"
assert_eq "green: judged line" "judged 39b30f67ead66c3beededf19d870af4d9ad78377" "$(tail -n 1 <<< "$out")"

new_case failure
pr_json open false sha1 > "$fix/pulls.1.json"
status_json "$(st "$COMPLETE" success ok)" "$(st 'PR Validation / Build X (pull_request)' failure 'Failing after 5s')" > "$fix/status.1.json"
run_case 5 --timeout 3s
assert_eq "failure: exit" 1 "$rc"
assert_has "failure: context" "PR Validation / Build X:" "$out"
assert_not_has "failure: suffix stripped" "Build X (pull_request)" "$out"
assert_eq "red: judged line" "judged sha1" "$(tail -n 1 <<< "$out")"

new_case cancelled-as-failure
pr_json open false sha1 > "$fix/pulls.1.json"
status_json "$(st "$COMPLETE" failure 'Has been cancelled')" > "$fix/status.1.json"
run_case 5 --timeout 3s
assert_eq "cancelled: exit" 1 "$rc"
assert_has "cancelled: desc" "Has been cancelled" "$out"

new_case pending-then-success
pr_json open false sha1 > "$fix/pulls.1.json"
status_json "$(st "$COMPLETE" pending 'Has started running')" > "$fix/status.1.json"
status_json "$(st "$COMPLETE" success ok)" > "$fix/status.2.json"
run_case 5 --timeout 3s
assert_eq "pending-then-success: exit" 0 "$rc"

new_case head-moved
pr_json open false aaa > "$fix/pulls.1.json"
pr_json open false bbb > "$fix/pulls.2.json"
status_json "$(st "$COMPLETE" failure red)" > "$fix/status.aaa.json"
status_json "$(st "$COMPLETE" success ok)" > "$fix/status.bbb.json"
run_case 5 --timeout 3s
assert_eq "head-moved: exit" 0 "$rc"
assert_has "head-moved: message" "head moved aaa → bbb" "$out"

new_case absent-complete-timeout
pr_json open false sha1 > "$fix/pulls.1.json"
status_json "$(st 'PR Validation / Build X (pull_request)' success ok)" > "$fix/status.1.json"
INTERVAL=1 run_case 5 --timeout 2s
assert_eq "absent-complete: exit" 2 "$rc"
assert_has "absent-complete: message" "timed out waiting for" "$out"

new_case push-event-ignored
pr_json open false sha1 > "$fix/pulls.1.json"
status_json "$(st "$COMPLETE" success ok)" "$(st 'PR Validation / Build X (push)' failure red)" > "$fix/status.1.json"
run_case 5 --timeout 3s
assert_eq "push-event-ignored: exit" 0 "$rc"

new_case pr-404
echo 1 > "$fix/tea.rc"
run_case 5 --timeout 3s
assert_eq "pr-404: exit" 2 "$rc"

new_case pr-404-json-body
echo '{"message":"not found"}' > "$fix/tea.out"
echo 0 > "$fix/tea.rc"
run_case 5 --timeout 3s
assert_eq "pr-404-json-body: exit" 2 "$rc"

new_case head-always-moves
touch "$fix/pulls.always-moves"
status_json "$(st "$COMPLETE" pending running)" > "$fix/status.1.json"
INTERVAL=1 run_case 5 --timeout 3s
assert_eq "head-always-moves: exit" 2 "$rc"
assert_has "head-always-moves: message" "timed out waiting for" "$out"

new_case non-numeric
run_case abc
assert_eq "non-numeric: exit" 2 "$rc"
assert_has "non-numeric: usage" "usage" "$out"

new_case bad-json
echo 'not json' > "$fix/tea.out"
echo 0 > "$fix/tea.rc"
run_case 5 --timeout 3s
assert_eq "bad-json: exit" 2 "$rc"

new_case closed-unmerged
pr_json closed false sha1 > "$fix/pulls.1.json"
run_case 5 --timeout 3s
assert_eq "closed-unmerged: exit" 2 "$rc"

if [ "$fails" -ne 0 ]; then echo "$fails failure(s)" >&2; exit 1; fi
echo "all passed"
