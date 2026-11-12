# SDD ledger — plan: docs/tasks/task-058-url-addressable-views/plan.md

Spec: docs/tasks/task-058-url-addressable-views/design.md (PRD prd.md) — reachable.

## Pre-flight scan

| Tasks | Shared file / interface | Produces → consumes | Finding |
|---|---|---|---|
| 1 ↔ 7 | `MealsPage.tsx` | T1 swaps `formatDateStr`→`formatLocalDateStr`; T7 replaces line-57 `useState` | Disjoint hunks; T7 line numbers may shift by ~6. OK |
| 1 → 7, 8 | `date-utils.ts` | `parseLocalDateStr`/`formatLocalDateStr` | Signatures agree. OK |
| 2 → 6, 7, 8 | `use-query-param-state.ts` | `QueryParamCodec`, `useQueryParamState`, `withParam` | Signatures agree. OK |
| 2 ↔ 3 | `LocationProbe` | T2 inlines its own probe (T3 creates shared one later) | Plan says so explicitly. OK |
| 3 → 4–8 | `app-routes.tsx`, `app-routes.test.tsx` harness | stub renders `name:{view ?? tab ?? ""}` | T4–T8 assertions match the stub format. OK |
| 3 ↔ 6 | `tracker` route + baseline test `/app/tracker?x=1 → /app/habits?x=1` | T6 retargets to `/app/habits/today` | T3 baseline assertion becomes stale in T6 — see Ruling 1 |
| 5 | ingredients routes | no `ingredients/*` | Self-consistent with `not-a-view → IngredientDetailPage`. OK |
| 8 | `calendar-url.ts` | desktop week = Sunday start | 2026-03-01 is a Sunday, 2026-03-18 → 03-15, 11-12 → 11-08: test values consistent. OK |
| 9 | repo gate | — | Duplicates /execute-task Step 5 item 2 — see Ruling 2 |
| each task 1–8 | self-consistency | tests vs code vs files | No internal contradictions found |

Ruling: Task 6 updates Task 3's baseline `/app/tracker?x=1` assertion to `/app/habits/today?x=1` — the plan's T6 retarget supersedes the T3 interim target — cost if wrong: one test line.
Ruling: Task 9 (repo gate) is executed as /execute-task Step 5 item 2 (flagless verify at branch end), not as a separate plan-task dispatch — same command, same contract — cost if wrong: none; gate still runs flagless.

Ruling: installed `frontend/node_modules` in the worktree once via `npm ci` (git-ignored) so implementers/reviewers run vitest/tsc locally instead of symlinking the main repo's node_modules — cost if wrong: a stale install; `npm ci` again.
Ruling: the per-task `--quick --base` gate selects only Go legs (pins/build/vet) for frontend-only ranges; frontend correctness per task rests on implementer vitest/tsc/eslint, and the frontend leg runs at the branch-end flagless verify — cost if wrong: a frontend break surfaces at branch end instead of per task.
Ruling: reverted gate-side-effect edits to `go.work.sum` (2 checksum lines added by the Go build) — the branch is frontend-only — cost if wrong: none; the Go toolchain re-adds them locally.

Gate: 331ed5c..cf6cc69 --quick PASS (legs pins build vet). last-gated=cf6cc69
Agent: Task 1 type=task-implementer model=sonnet status=DONE commit=cf6cc69
Agent: Task 1 type=task-reviewer model=sonnet status=DONE verdict=APPROVED caused-fix=no
Task 1: complete (commits 331ed5c..cf6cc69)
Gate: cf6cc69..6d58a19 --quick PASS. last-gated=6d58a19
Agent: Task 2 type=task-implementer model=sonnet status=DONE commit=6d58a19
Agent: Task 2 type=task-reviewer model=sonnet status=DONE verdict=APPROVED caused-fix=no
Task 2: minor (deferred): "normalizing codec rewrites once" test proves no-loop with only two Promise.resolve ticks + serialize count — weak but fails without the effect
Task 2: complete (commits cf6cc69..6d58a19)
Gate: 6d58a19..229da43 --quick PASS. last-gated=229da43
Agent: Task 3 type=task-implementer model=sonnet status=DONE commit=229da43
Agent: Task 3 type=task-reviewer model=sonnet status=DONE verdict=APPROVED caused-fix=no
Task 3: minor (deferred): redirect-preserving-query test was green on first run (component written before RED); app-routes test did show RED
Task 3: complete (commits 6d58a19..229da43)
Gate: 229da43..dfcb9f9 --quick PASS. last-gated=dfcb9f9
Agent: Task 4 type=task-implementer model=sonnet status=DONE commit=dfcb9f9
Agent: Task 4 type=task-reviewer model=sonnet status=DONE verdict=APPROVED caused-fix=no
Task 4: minor (deferred): app-routes.test.tsx:33 page Stub prefers props over query params — harmless
Task 4: complete (commits 229da43..dfcb9f9)

Controller notes for next session: briefs use `**Files:**` from plan — append a `### Files` block via tools/task-brief.sh output before dispatch (done for T2–T4). Harness exports `renderAt`/`shellMounts` from `frontend/src/__tests__/app-routes.test.tsx`; page stub renders `name:{view ?? tab ?? ""}` from props. After each --quick gate, `git checkout -- go.work.sum` (gate side effect). Task 6 must apply Ruling 1 (update `/app/tracker?x=1` baseline to `/app/habits/today?x=1`).
Handoff after Task 4 (4 tasks completed this session)
Gate: dfcb9f9..de8f412 --quick PASS. last-gated=de8f412
Agent: Task 5 type=task-implementer model=sonnet status=DONE commit=de8f412
Agent: Task 5 type=task-reviewer model=sonnet status=DONE verdict=APPROVED_WITH_FINDINGS caused-fix=no
Task 5: minor (deferred): `?category=` preservation only tested for list↔categories; Bulk and "New" navigation tested without a query (code passes location.search in both)
Task 5: minor (deferred): `page`/`search` state persists across views since the page does not remount — matches old behaviour
Task 5: note: "New" kept the form via plain navigate+setShowCreate (no remount) — `state.openCreate` fallback not used; no D1 remount note needed in PR
Task 5: note: reviewer returned its review inline and did not write task-5-review.md
Task 5: complete (commits 1dd448d..de8f412, review clean)
Gate: de8f412..6e6a6be --quick PASS. last-gated=6e6a6be
Agent: Task 6 type=task-implementer model=sonnet status=DONE commit=6e6a6be
Agent: Task 6 type=task-reviewer model=sonnet status=DONE verdict=APPROVED_WITH_FINDINGS caused-fix=no
Task 6: minor (deferred): empty `?month=` not pinned by a TrackerPage test (only `?month=2026-13`); hook code handles it in one replace (use-query-param-state.ts:42-47)
Task 6: note: implementer did not write task-6-report.md; controller transcribed the hand-back into it
Task 6: complete (commits de8f412..6e6a6be, review clean)
Gate: 6e6a6be..73fb8cd --quick PASS. last-gated=73fb8cd
Agent: Task 7 type=task-implementer model=sonnet status=DONE commit=73fb8cd
Agent: Task 7 type=task-reviewer model=sonnet status=DONE verdict=APPROVED_WITH_FINDINGS caused-fix=no
Task 7: minor (deferred): Wednesday/invalid `week` tests assert final URL + REPLACE but do not count navigations — "exactly one rewrite" pinned only by hook code (effect keyed on `raw`)
Task 7: minor (deferred): WeekSelector stub emits only Mondays; non-Monday from selector untested (real selector emits Mondays)
Task 7: minor (deferred): implementer wrote implementation before first test run — no RED observed
Task 7: complete (commits 6e6a6be..73fb8cd, review clean)
Gate: 73fb8cd..1f50068 --quick PASS. last-gated=1f50068
Agent: Task 8 type=task-implementer model=sonnet status=DONE commit=1f50068
Agent: Task 8 type=task-reviewer model=sonnet status=DONE verdict=APPROVED_WITH_FINDINGS caused-fix=no
Task 8: minor (deferred): no page-level mobile (matchMedia false) CalendarPage test; mobile anchor math only unit-tested in calendar-url.test.ts
Task 8: minor (deferred): fixed-point test runs on canonical values only (implementer concern); reviewer judged this matches design §5, raw inputs covered by parse assertions
Task 8: minor (deferred): navigateToView carries unrelated params (e.g. `connected`) through on view change — harmless
Task 8: note: implementer ran full vitest (175 files / 1120 tests) + tsc -b + `npm run lint` (0 errors); `npx eslint src` errored "could not determine executable to run" — watch the branch-end frontend lint leg
Task 8: complete (commits 73fb8cd..1f50068, review clean)

Controller notes for next session: all plan tasks 1–8 complete; Task 9 is the branch end per Ruling 2. Resume at /execute-task Step 5 item 1 (merge-main), then flagless verify, then code review (plan-adherence + frontend-guidelines reviewers; no Go changes → no backend reviewer). Point reviewers at the `minor (deferred)` lines above. After any gate, `git checkout -- go.work.sum`.
Handoff after Task 8 (4 tasks completed this session)
Branch end: merge-main up-to-date
Agent: Branch end type=plan-adherence-reviewer model=sonnet status=DONE verdict=APPROVED_WITH_FINDINGS caused-fix=no
Agent: Branch end type=frontend-guidelines-reviewer model=sonnet status=DONE verdict=APPROVED_WITH_FINDINGS caused-fix=no
Branch end: review APPROVED_WITH_FINDINGS (0 blocking; FE-02 warn TrackerPage.tsx:63 template-literal class; "unrelated tooling" finding was a stale local-main diff — vs fetched main the branch is frontend + task docs only)

Controller notes for next session: merge-main up-to-date and review APPROVED_WITH_FINDINGS (no fix round) are both done at head b2d6f1b + ledger-only commits. Flagless verify on b2d6f1b was still running (log /tmp/verify-task-058-url-addressable-views-b2d6f1b.log). If that log has `EXIT=`, reconcile from it, since no non-ledger commit has landed since. Otherwise re-run it. Then `git checkout -- go.work.sum` (gate side effect, never commit it), push, and open the PR (item 4).
Handoff after branch end (verify unreconciled)
Branch end: verify PASS (flagless tools/verify.sh at b2d6f1b EXIT=0, all legs PASSED: pins build vet test lint frontend-build eslint docker; only task docs changed since)
Branch end: pr-open PR #36
CI wait: PR #36 head 9ae7723
