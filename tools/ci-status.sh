#!/usr/bin/env bash
# ci-status.sh — wait for the PR Validation verdict on a PR's current head.
#
# Usage: tools/ci-status.sh <pr-number> [--timeout <N>[s|m|h]]
# Env:   CI_STATUS_INTERVAL  poll interval in seconds (default 30)
# Exit:  0 verdict green; 1 a kept status is failing;
#        2 usage/API/JSON error, closed-unmerged PR, or timeout.
#
# Only `(pull_request)` contexts are judged. The verdict is pending until the
# `PR Validation / PR Validation Complete (pull_request)` context exists and no
# kept status is `pending`. The PR head is re-read every iteration and after
# each judgement, so a push mid-wait restarts judgement on the new head.
set -uo pipefail

COMPLETE_CTX='PR Validation / PR Validation Complete (pull_request)'

usage() {
    echo "usage: ci-status.sh <pr-number> [--timeout <N>[s|m|h]]" >&2
    exit 2
}

[ $# -ge 1 ] || usage
pr="$1"
shift
case "$pr" in '' | *[!0-9]*) usage ;; esac

timeout_arg="90m"
while [ $# -gt 0 ]; do
    case "$1" in
        --timeout)
            [ $# -ge 2 ] || usage
            timeout_arg="$2"
            shift 2
            ;;
        *) usage ;;
    esac
done

case "$timeout_arg" in
    *[0-9]s) mult=1; num="${timeout_arg%s}" ;;
    *[0-9]m) mult=60; num="${timeout_arg%m}" ;;
    *[0-9]h) mult=3600; num="${timeout_arg%h}" ;;
    *[0-9]) mult=1; num="$timeout_arg" ;;
    *) usage ;;
esac
case "$num" in '' | *[!0-9]*) usage ;; esac
timeout_s=$((num * mult))

interval="${CI_STATUS_INTERVAL:-30}"
case "$interval" in '' | *[!0-9]*) echo "ci-status: invalid CI_STATUS_INTERVAL '$interval'" >&2; exit 2 ;; esac

command -v jq > /dev/null 2>&1 || { echo "ci-status: jq not found" >&2; exit 2; }
command -v tea > /dev/null 2>&1 || { echo "ci-status: tea not found" >&2; exit 2; }

# fetch_head -> prints head sha; returns 2 on API/JSON error or closed-unmerged.
fetch_head() {
    local body state merged sha
    if ! body="$(tea api "repos/{owner}/{repo}/pulls/$pr" 2> /dev/null)"; then
        echo "ci-status: failed to fetch PR #$pr" >&2
        return 2
    fi
    if ! state="$(jq -er '.state' <<< "$body" 2> /dev/null)"; then
        echo "ci-status: invalid JSON for PR #$pr" >&2
        return 2
    fi
    merged="$(jq -r '.merged // false' <<< "$body")"
    if [ "$state" != "open" ] && [ "$merged" != "true" ]; then
        echo "ci-status: PR #$pr is $state and not merged" >&2
        return 2
    fi
    if ! sha="$(jq -er '.head.sha' <<< "$body" 2> /dev/null)"; then
        echo "ci-status: no head sha for PR #$pr" >&2
        return 2
    fi
    printf '%s\n' "$sha"
}

deadline=$((SECONDS + timeout_s))
last_complete_seen="no"

# check_deadline SHA -> exits 2 with a timeout message once the deadline passed.
check_deadline() {
    [ "$SECONDS" -ge "$deadline" ] || return 0
    if [ "$last_complete_seen" = "true" ]; then
        echo "ci-status: timed out waiting for pending statuses on $1" >&2
    else
        echo "ci-status: timed out waiting for '$COMPLETE_CTX' on $1" >&2
    fi
    exit 2
}

while :; do
    head="$(fetch_head)" || exit 2

    if ! body="$(tea api "repos/{owner}/{repo}/commits/$head/status" 2> /dev/null)"; then
        echo "ci-status: failed to fetch status for $head" >&2
        exit 2
    fi
    if ! kept="$(jq -ce '[.statuses[]? | select(.context | endswith(" (pull_request)"))]' <<< "$body" 2> /dev/null)"; then
        echo "ci-status: invalid JSON status for $head" >&2
        exit 2
    fi

    # Re-read the head: if it moved while we read statuses, the result is stale.
    head2="$(fetch_head)" || exit 2
    if [ "$head2" != "$head" ]; then
        echo "head moved $head → $head2"
        check_deadline "$head2"
        sleep "$interval"
        continue
    fi

    has_complete="$(jq -r --arg c "$COMPLETE_CTX" 'any(.[]; .context == $c)' <<< "$kept")"
    any_pending="$(jq -r 'any(.[]; .status == "pending")' <<< "$kept")"
    last_complete_seen="$has_complete"

    if [ "$has_complete" = "true" ] && [ "$any_pending" = "false" ]; then
        failing="$(jq -r '.[] | select(.status != "success" and .status != "skipped") | "\(.context | sub(" \\(pull_request\\)$"; "")): \(.status) — \(.description) \(.target_url)"' <<< "$kept")"
        if [ -z "$failing" ]; then
            echo "PR #$pr: success"
            # Consumed contract: claude-drive's ci_wait_outcome parses "judged <sha>". Do not change the format.
            echo "judged $head"
            exit 0
        fi
        echo "PR #$pr: failing"
        printf '%s\n' "$failing"
        # Consumed contract: claude-drive's ci_wait_outcome parses "judged <sha>". Do not change the format.
        echo "judged $head"
        exit 1
    fi

    check_deadline "$head"
    sleep "$interval"
done
