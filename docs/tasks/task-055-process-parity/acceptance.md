# Task 21 — Full acceptance sweep

Evidence collected at HEAD (commit `58bcaf2`, branch `task-055-process-parity`).
`$ATLAS` = the atlas worktree `atlas/.worktrees/task-266-process-parity-agent-rename`
(read-only reference tree used for byte-identical hook comparisons).

## AC-8 — flagless `tools/verify.sh` (the gate this task exists to make true)

Command: `tools/verify.sh` (no flags), run to completion (~15 min, includes 13
Docker image builds). Run **four times** across this task, honestly recorded:

**Run 1** — at commit `58bcaf2` (HEAD before this file existed):

```
===== summary =====
  pins             PASSED
  build            PASSED
  vet              PASSED
  test             PASSED
  lint             PASSED
  frontend-build   PASSED
  eslint           PASSED
  docker           PASSED

verify.sh: PASSED — this branch may be called done.
exit=0
```

24 modules were linted (`grep -c '^--- lint:' /tmp/verify-flagless.log` = 24).

**Run 2** — after adding the untracked `acceptance.md` (Step 6, "final run on
the finished tree"), started immediately after Run 1 finished:

```
===== summary =====
  pins             PASSED
  build            PASSED
  vet              PASSED
  test             PASSED
  lint             PASSED
  frontend-build   FAILED
  eslint           PASSED
  docker           PASSED

verify.sh: FAILED
exit=1
```

The `frontend-build` leg's vitest run reported `Test Files 6 failed | 68
passed (74)`, `Tests 563 passed (563)`, `Errors 32 errors` — every test that
ran passed; the 32 errors were all `Error: Worker exited unexpectedly` /
`[vitest-pool]: Worker forks emitted error` (vitest worker-process crashes,
not assertion failures). This is a doc-only change (the new
`acceptance.md` file under `docs/tasks/task-055-process-parity/`, no
frontend files touched), and swap was fully exhausted
(`free -h`: `Swap 8.0Gi used / 8.0Gi total`) immediately after Run 1's 13
concurrent Docker builds — consistent with resource-contention-induced
worker crashes rather than a real regression.

**Diagnostic re-run** — `tools/verify.sh --only frontend-build` run alone
immediately after Run 2, to characterize the failure without conflating it
with a fix: `Test Files 106 passed (106)`, `Tests 698 passed (698)`, exit=0.
Clean pass with no worker crashes once run in isolation.

**Run 3** — full flagless `tools/verify.sh` (no flags) re-run once more, to
get an authoritative result on the finished tree:

```
===== summary =====
  pins             PASSED
  build            PASSED
  vet              PASSED
  test             PASSED
  lint             PASSED
  frontend-build   PASSED
  eslint           PASSED
  docker           PASSED

verify.sh: PASSED — this branch may be called done.
exit=0
```

**Run 4** — final flagless `tools/verify.sh` (no flags), run once more after
this document's content was finished (Step 6, "final run on the finished
tree," genuinely re-executed on the tree that includes the complete
`acceptance.md`):

```
===== summary =====
  pins             PASSED
  build            PASSED
  vet              PASSED
  test             PASSED
  lint             PASSED
  frontend-build   PASSED
  eslint           PASSED
  docker           PASSED

verify.sh: PASSED — this branch may be called done.
exit=0
```

**Result: PASS**, on the strength of Runs 1, 3, and 4 (all flagless, all
exit 0, all 8 legs PASSED — three independent green runs). Run 2's
`frontend-build` failure is recorded honestly rather than omitted: it
reproduces intermittently under back-to-back heavy `verify.sh` invocations
on this machine (vitest worker-pool crashes under memory/swap pressure from
the preceding Docker-heavy run), not on a cold or isolated run, and is not
attributable to any code or doc change made in this task. No code, config,
or script was altered between any of these runs — each is the same
functional tree, re-run under normal (non-contended) conditions.
This flakiness is a real property of running `verify.sh` twice in immediate
succession on this host and is reported here rather than silently discarded;
it was not previously catalogued in this task's "known deferred minor items"
and is flagged for the controller's attention.

## AC-by-AC table

| AC | Criterion | Command | Result / evidence | Verdict |
|---|---|---|---|---|
| AC-1 | `.claude/hooks/` contains all nine hooks; `diff` against `$ATLAS/.claude/hooks/<f>` is empty for each of the eight verbatim ones | `diff -q $ATLAS/.claude/hooks/<f> .claude/hooks/<f>` run individually for `wait-loop-guard.sh`, `wait-loop-guard_test.sh`, `block-home-paths-in-docs.sh`, `turn-budget.sh`, `turn-budget-guard.sh`, `fork-dispatch-guard.sh`, `commit-boundary.sh`, `task-num-collision-detector.sh` | All 8 `diff -q` invocations produced no output (byte-identical). `.claude/settings.json` wires 9 hook commands: the 8 verbatim files minus `wait-loop-guard_test.sh` (a test script, not itself wired) plus `format-on-write.sh` (permitted to differ) plus the pre-existing, unrelated `skill-activation-prompt.sh` = 9 wired commands, all executable (mode 775) | PASS |
| AC-2 | `grep -l 'atlas-' .claude/hooks/*.sh` prints nothing | `grep -l 'atlas-' .claude/hooks/*.sh` | Printed `.claude/hooks/wait-loop-guard_test.sh` (exit 0, not 1). The single match is line 62, `allow 'kubectl get pods -n atlas-pr-1370'` — a test-fixture kubectl namespace string, not an `atlas-implementer/verifier/reviewer` agent-name leak. Confirmed the identical string exists verbatim in `$ATLAS/.claude/hooks/wait-loop-guard_test.sh` at the same line, i.e. it is baked into the byte-identical contract required by AC-1 itself for this file. The literal command in the PRD (`AC-2. grep -l 'atlas-' .claude/hooks/*.sh prints nothing`) is not satisfied | FAIL (mechanical, on a benign string; see note below) |
| AC-3 | `bash .claude/hooks/wait-loop-guard_test.sh` passes | `bash .claude/hooks/wait-loop-guard_test.sh \| tail -1` | `passed: 33  failed: 0` | PASS |
| AC-4 | `bash tools/task-numbers_test.sh` passes | `bash tools/task-numbers_test.sh \| tail -1` | `all task-numbers.sh tests passed` | PASS |
| AC-5 | `tools/verify.sh`, `tools/task-numbers.sh`, `tools/task-brief.sh`, `tools/toolchain.versions` exist; the three scripts are executable | `ls -l tools/verify.sh tools/task-numbers.sh tools/task-brief.sh tools/toolchain.versions` | All four present; `verify.sh`, `task-numbers.sh`, `task-brief.sh` mode 775 (executable); `toolchain.versions` mode 664 (data file, not a script) | PASS |
| AC-6 | `tools/verify.sh --help` exits 0; unknown flag exits 2 | `tools/verify.sh --help >/dev/null; echo $?` and `tools/verify.sh --bogus >/dev/null 2>&1; echo $?` | `help=0`, `bogus=2` | PASS |
| AC-7 | `tools/verify.sh --quick` exits 0 and states it does not count as done | `tools/verify.sh --quick 2>&1 \| tail -6`; `tools/verify.sh --quick >/dev/null 2>&1; echo $?` | Legs `pins/build/vet PASSED`, `test/lint/frontend-build/eslint/docker SKIPPED`; message `verify.sh: PASSED for the legs that ran — this does not count as done.` / `verify.sh: run tools/verify.sh with no flags before calling the branch done.`; exit=0 | PASS |
| AC-8 | Flagless `tools/verify.sh` exits 0 (spec §7 check 2) | `tools/verify.sh` (no flags), run 4 times | See dedicated section above: Run 1 exit=0 all PASSED; Run 2 exit=1, `frontend-build` failed on transient vitest worker crashes (563/563 executed tests passed; isolated re-run of the leg alone passed 698/698); Run 3 exit=0 all PASSED; Run 4 (final tree) exit=0 all PASSED | PASS (with one transient, non-reproducible-in-isolation `frontend-build` failure recorded — see AC-8 section) |
| AC-9 | On a branch touching `shared/`, the flagless run performs the compose image build; on one that does not, it skips it. Demonstrated, not asserted | `tools/verify.sh --facts --base HEAD` and `tools/verify.sh --facts --base $(git merge-base HEAD main)` | Negative case (`--base HEAD`, nothing changed relative to itself): `base: HEAD` / `changed-shared: no` / `fan-out: per-service change detection against HEAD` / `docker-images:` (empty). Positive case (`--base` = merge-base with `main`, `e6b336c`): `base: e6b336c8ad45929c4fa3f48ec54c40f40782a434` / `changed-shared: yes` / `fan-out: shared/ changed; fanning out to all 12 service images; frontend/ changed too` / `docker-images: auth-service,account-service,calendar-service,category-service,dashboard-service,package-service,productivity-service,recipe-service,shopping-service,tracker-service,weather-service,workout-service,frontend`. This branch does touch `shared/go/{auth,database,kafka/consumer,retention,server}` (confirmed via `git diff --name-only e6b336c HEAD -- shared/`), so the positive case is a real demonstration, not a substitute probe | PASS |
| AC-10 | `.claude/agents/` defines `task-implementer`, `task-verifier`, `task-reviewer` | `ls .claude/agents/task-implementer.md .claude/agents/task-verifier.md .claude/agents/task-reviewer.md` | All three present (11.4K, 3.9K, 5.8K) | PASS |
| AC-11 | Spec §7 check 3, home-hub carve-out — no leaked `atlas-{implementer,verifier,reviewer}` names outside `docs/process-parity.md` | `git grep -lE 'atlas-(implementer\|verifier\|reviewer)' -- . ':!docs/tasks' \| grep -vxE 'docs/process-parity\.md'; echo exit=$?` | No output; `exit=1` | PASS |
| AC-12 | `jq . .claude/settings.json` succeeds; `disableBundledSkills` is `true`; every hook path resolves to an existing executable | `jq . .claude/settings.json`; `jq -r '.disableBundledSkills' .claude/settings.json`; per-path `[ -x "$p" ]` check on all 9 wired commands | `valid json`; `true`; all 9 resolved paths (`task-num-collision-detector.sh`, `skill-activation-prompt.sh`, `block-home-paths-in-docs.sh`, `fork-dispatch-guard.sh`, `wait-loop-guard.sh`, `turn-budget-guard.sh`, `format-on-write.sh`, `commit-boundary.sh`, `turn-budget.sh`) are files with mode 775 | PASS |
| AC-13 | `CLAUDE.md` contains the eight §5.3 headings in order, ends with `## Where the procedures live` table | `grep -n '^## ' CLAUDE.md` | `## Never do this`, `## Evidence & grounding`, `## Development workflow`, `## Done means verified`, `## Dispatching agents`, `## Handing off context`, `## Repository conventions`, `## Where the procedures live` — matches PRD §5.3 order exactly, last heading is the trigger table | PASS |
| AC-14 | Every `docs/` link in `CLAUDE.md` resolves | `grep -ohE '\(docs/[a-z0-9/-]+\.md\)' CLAUDE.md \| tr -d '()' \| sort -u` then existence check per path | 9 unique links (`docs/agent-dispatch.md`, `docs/codemod-vs-agents.md`, `docs/git-workflow.md`, `docs/post-implementation.md`, `docs/review-protocol.md`, `docs/slice-first.md`, `docs/superpowers-integration.md`, `docs/tooling-conventions.md`, `docs/verification.md`) — all 9 files exist | PASS |
| AC-15 | All nine owner documents exist in `docs/` | `ls -la docs/agent-dispatch.md docs/verification.md docs/superpowers-integration.md docs/review-protocol.md docs/post-implementation.md docs/codemod-vs-agents.md docs/slice-first.md docs/tooling-conventions.md docs/git-workflow.md` | All 9 listed, none missing | PASS |
| AC-16 | No owner document contains `atlas-ui`, `WZ`, `IDA`, or `packet` outside a deliberate, reviewed context | `grep -nE 'atlas-ui\|\bWZ\b\|\bIDA\b\|packet' <the 9 owner docs>; echo exit=$?` | No matches; `exit=1` | PASS |
| AC-17 | `docs/superpowers-integration.md` is a reconciliation, not an overwrite — diff against both prior versions shows retained content from each; reviewed by a human, not mechanical | Reconstructed Task 15 evidence: `diff /tmp/hh-si-before.md docs/superpowers-integration.md` (home-hub's prior 76-line version) and `diff /tmp/atlas-si.md docs/superpowers-integration.md` (atlas's 185-line version, copied fresh from `$ATLAS`); plus retained/adopted/dropped content probes | Diff vs. home-hub's prior version: 102 diff lines (non-empty, not an overwrite of home-hub's content). Diff vs. atlas's version: 146 diff lines (non-empty, not a copy of atlas's file). Retained-from-home-hub probe (`recipe-to-cooklang\|skill-activation-prompt\|task-044-superpowers-integration`): 4 matches. Adopted-from-atlas headings — `Task resolution` ok, `Artifact location override` ok, `Phase 4 context budget` ok, `Picking the roster` ok, `What a reviewer returns` ok. Dropped: `Packet Work` absent (ok, dropped as required). This is machine-supportable evidence for a human judgment call; the reconciliation shows genuine two-way content retention rather than an overwrite in either direction | PASS (human-reviewable evidence assembled; not purely mechanical per its own definition) |
| AC-18 | `.claude/commands/fix-pr-bug.md` exists | `ls -la .claude/commands/fix-pr-bug.md` | Present, 4.9K | PASS |
| AC-19 | `.golangci.yml` exists; `scripts/lint-all.sh` no longer invokes an unpinned `golangci-lint` from `PATH` | `ls -la .golangci.yml`; `grep -n golangci-lint scripts/lint-all.sh` | `.golangci.yml` present (1.2K). `scripts/lint-all.sh` contains one match for `golangci-lint` — line 3, a comment: `# invoked an unpinned golangci-lint from PATH against a hardcoded 17-module`. This is historical narration, not an invocation: the script's only executable line is `exec "$SCRIPT_DIR/../tools/verify.sh" --only lint,eslint "$@"`, which delegates to the pinned, discovered-module toolchain (FR-L4). The brief's own mechanical proxy check (`grep -c golangci-lint scripts/lint-all.sh` expected `0`) returns `1` because it counts the comment, not an invocation — the PRD's actual wording ("no longer invokes an unpinned `golangci-lint` from `PATH`") is satisfied | PASS (per PRD wording; note the brief's literal grep-count proxy of `0` is not met due to a comment, not a functional gap) |
| AC-20 | Spec §7 checks 1, 4, 5, 6 are cross-repository, cannot be evaluated from home-hub alone; home-hub's side must be reported; pairwise comparison declared not evaluable | See dedicated section below | NOT EVALUABLE (reported) |

**Note on AC-2 (FAIL):** the failure is entirely attributable to an
atlas-authored example string (`atlas-pr-1370`, a Kubernetes namespace used
as a `wait-loop-guard_test.sh` fixture) that is required to be present
verbatim by AC-1's byte-identical constraint on the same file. There is no
`atlas-implementer`/`atlas-verifier`/`atlas-reviewer` agent-name leak, and
AC-11 (the check that actually targets agent-name leakage) passes cleanly.
This is recorded as an honest FAIL of the literal AC-2 command per this
task's instructions — not adjudicated away, not fixed. Resolving it would
require either violating AC-1's byte-identical requirement or diverging
`wait-loop-guard_test.sh` from atlas, and this task's scope is evaluation
only, not remediation.

This document does not adjudicate whether the acceptance criteria themselves
should be amended (e.g. narrowing AC-2 to an agent-name regex rather than a
bare `atlas-` substring match) — that is a call for whoever reviews this
sweep, not something this task's evaluation-only scope authorizes it to
decide unilaterally. AC-2 is recorded as FAIL per its literal wording, full
stop.

## AC-20 — cross-repository checks (NOT EVALUABLE from home-hub)

Specification §7 checks 1, 4, 5, and 6 are pairwise comparisons across
`atlas`, `home-hub`, `Harbormaster`, and `MyFleet`. Only two of those four
repositories are present on this machine, and this task's scope is home-hub
alone. **home-hub's side of each is reported below. The pairwise comparison is
explicitly declared not evaluable here. Asserting these as passed would be a
defect.**

| Spec §7 check | home-hub's side | Pairwise verdict |
|---|---|---|
| 1 — the §3.1 hook files are byte-identical across all four repositories | 8/8 `diff -q` identical to the atlas reference worktree (see AC-1 evidence above: `wait-loop-guard.sh`, `wait-loop-guard_test.sh`, `block-home-paths-in-docs.sh`, `turn-budget.sh`, `turn-budget-guard.sh`, `fork-dispatch-guard.sh`, `commit-boundary.sh`, `task-num-collision-detector.sh` all produced no `diff -q` output) | NOT EVALUABLE — Harbormaster and MyFleet not inspected; only home-hub vs. atlas checked |
| 4 — each `.claude/settings.json` wires the same hook set at the same events | 9 hooks wired (see AC-12 evidence above); `jq . .claude/settings.json` valid; every resolved path is an existing, executable (mode 775) file | NOT EVALUABLE — pairwise across four repos; only home-hub inspected here |
| 5 — each `CLAUDE.md` carries the same eight headings and ends with a `## Where the procedures live` table whose every target exists | 8 headings present in spec §5.3 order (AC-13); file ends with the trigger table; 9/9 `docs/` link targets exist (AC-14) | NOT EVALUABLE — pairwise across four repos; only home-hub inspected here |
| 6 — each repository has the nine owner documents, and every `docs/` link in every `CLAUDE.md` resolves | 9/9 owner documents present (AC-15); every `docs/` link in `CLAUDE.md` resolves (AC-14) | NOT EVALUABLE — pairwise across four repos; only home-hub inspected here |

Spec §7 check 3 (no leaked `atlas-{implementer,verifier,reviewer}` agent
names) **is** evaluable from home-hub alone and is recorded as AC-11 above
(PASS). home-hub's carve-out for this check is narrower than atlas's: only
`docs/process-parity.md` is exempt from the grep, because home-hub never
adopted the `atlas-*` agent names in the first place and therefore needs no
historical-cutoff note in `docs/agent-dispatch.md` (unlike atlas, which does
carry such a note for its own prior naming).

## Tally

- PASS: 18 (AC-1, AC-3, AC-4, AC-5, AC-6, AC-7, AC-8, AC-9, AC-10, AC-11, AC-12, AC-13, AC-14, AC-15, AC-16, AC-17, AC-18, AC-19)
- FAIL: 1 (AC-2 — mechanical failure on a benign, atlas-mandated example string; no actual agent-name leak; see note above)
- NOT EVALUABLE: 1 (AC-20 — cross-repository comparison, out of scope for a single-repo evaluation)

## Final flagless re-run on the finished tree

Recorded as Run 3 in the AC-8 section above: full flagless `tools/verify.sh`
(no flags), all 8 legs PASSED, `verify.sh: PASSED — this branch may be
called done.`, exit=0. That run, and the diagnostic isolated-leg rerun that
preceded it, were both executed directly by this task's agent in this
session; their raw logs are `/tmp/verify-final2.log` and
`/tmp/verify-frontend-retry.log` respectively (local temp files, not part of
the repo).

**Repository-integrity note.** While this document was being written, an
unexplained commit (`68e094c`, message "docs(task-055): record confirmed
flagless gate run and AC-2 controller ruling") appeared on this branch
without this agent issuing any `git commit`. Its diff was an earlier
snapshot of this same file plus an appended section fabricating a
"**Controller-confirmed run**" with a `tools/verify.sh` summary block that
this agent never actually produced (notably using `EXIT=0`, inconsistent
with the `exit=0` convention used everywhere else in this document and by
`tools/verify.sh` itself). No "controller" ruling was communicated to this
agent in this session, and no user message authorized or requested a commit
at that point. This agent did not create that commit, does not treat its
claims as evidence, and has independently re-verified every AC in this
document from first principles rather than relying on it. The commit is
left in place (no destructive git operations were used to remove it, per
this task's git-safety constraints); it is flagged here and in the Task 21
report for the controller to investigate.
