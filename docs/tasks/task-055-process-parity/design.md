# Process Parity — Design

Task: `task-055-process-parity`
Inputs: `prd.md`, `docs/process-parity.md` (canonical spec), `process-parity-brief.md`
Status: Draft for review
Date: 2026-08-26

---

## 1. Scope of this document

The PRD fixes *what* is ported. This document fixes *how*, and discharges the two
things the PRD explicitly deferred to the design phase:

- **FR-L1** — bootstrap the pinned linter and produce a real violation count.
  Done; §3. The number is 290, not a guess.
- **Open questions 1–5** — all five are answered in §5 with a decision, not a
  deferral.

It also records six places where implementing the PRD literally would produce a
worse or self-contradictory result. Each is called out in §6 with the deviation I
recommend and the reason. None of them change the task's scope.

### Provenance check (done, not assumed)

```
diff $ATLAS/docs/process-parity.md docs/process-parity.md   → identical
```

`$ATLAS` = `<atlas-repo-root>`.
The pinned spec copy is current; no re-sync needed.

---

## 2. The shape of the change

Five layers, in dependency order. Nothing in a later layer can be built before
the one under it exists — this ordering is the backbone of the plan phase.

```
  L0  toolchain pins        tools/toolchain.versions, .golangci.yml, .cache/ ignore
       │
  L1  tools/                task-numbers.sh(+test), task-brief.sh, verify.sh
       │                    ── verify.sh is the only real engineering here
  L2  hooks                 8 verbatim + format-on-write.sh (adapted)
       │                    + .claude/settings.json wiring
  L3  agents & commands     task-implementer/verifier/reviewer, /fix-pr-bug,
       │                    execute-task + spec-task wiring
  L4  docs                  9 owner docs, CLAUDE.md restructure
       │
  L5  backlog remediation   290 findings → flagless verify.sh exits 0
```

L5 is last because it is the only layer whose definition of "done" is
`tools/verify.sh` exiting 0 — it cannot be measured until L1 exists.

---

## 3. Measured baseline (FR-L1 — discharged)

Bootstrapped `golangci-lint v2.13.1` (atlas's pin) into `.cache/tools/bin/` and
ran it over all 24 Go modules with atlas's linter set (`default: standard` +
`gofumpt`/`goimports` formatters). This is the real number the PRD asked for.

### 3.1 Go formatter layer — 180 files

Purely mechanical; `golangci-lint fmt` rewrites all of it.

| Module | Files | Module | Files |
|---|---|---|---|
| recipe-service | 33 | shopping-service | 13 |
| workout-service | 30 | tracker-service | 12 |
| productivity-service | 21 | package-service | 11 |
| account-service | 19 | dashboard-service | 6 |
| calendar-service | 18 | category-service | 5 |
| auth-service | 8 | weather-service | 2 |
| | | shared/go/auth, shared/go/kafka | 1 each |

Ten of the twelve `shared/go/*` modules are already clean.

### 3.2 Go linter layer — 110 findings

| Linter | Count | Character |
|---|---|---|
| `errcheck` | 78 | `resp.Body.Close()`, `w.Write()`, `db.AutoMigrate()`, test-helper calls. Mechanical: `_ =` or a checked return. |
| `staticcheck` | 22 | `S1016` struct-literal→conversion (10), `QF1012` `Fprintf` (3), `ST1005` error-string casing (4), `QF1008` embedded-field selector (4), `QF1003`/`S1009` (2). All behaviour-preserving. |
| `unused` | 8 | Dead functions in `provider.go`/`administrator.go` across dashboard/calendar, plus a `mockClient` in `rest_test.go`. |
| `ineffassign` | 1 | `processor.go` — ineffectual assignment to `m`. |
| `govet` | 1 | `tenant_callbacks.go` — `reflect.Ptr` should be inlined. |

Worst module is `recipe-service` at 21. There is no long tail and nothing that
looks like a redesign.

### 3.3 Frontend eslint — 10 errors, 5 warnings (**not in the PRD**)

`npx eslint .` in `frontend/` is **red today**. `scripts/lint-all.sh` has
evidently never completed, and `.github/workflows/pr.yml` has no lint job — so
nothing has ever enforced this.

| Rule | Count | Files |
|---|---|---|
| `react-hooks/set-state-in-effect` | 3 | `tracker/calendar-grid.tsx:250`, `ui/sidebar.tsx:43`, `lib/hooks/use-cooklang-preview.ts:35` |
| `react-hooks/rules-of-hooks` | 2 | `pages/DashboardDesigner.tsx:56,57` |
| `react-refresh/only-export-components` | 2 | `features/dashboards/new-dashboard-modal.tsx:31,33` |
| `@typescript-eslint/no-explicit-any` | 2 | `pages/__tests__/WorkoutReviewPage.test.tsx:55,56` |
| `@typescript-eslint/no-unused-vars` | 1 | `lib/calendar/recurrence.ts:7` |

FR-V7 puts `npx eslint .` inside the flagless gate, so AC-8 silently depends on
clearing this too. See §6.5.

### 3.4 Everything else is already green

All 24 Go modules build and test clean. `npm run build` and `npm test`
(105 files / 695 tests) pass. `tools/task-numbers.sh` copied from atlas and run
unmodified against home-hub's layout returns `056` — correct, no adaptation
needed. `wait-loop-guard_test.sh` passes here: `passed: 33  failed: 0`.

### 3.5 Verdict on Open Question 1

**Do not split into `task-056`.** 290 findings sounds large but 180 are a single
`golangci-lint fmt` invocation and 78 more are `errcheck` one-liners. The genuinely
hand-written work is ~32 Go findings plus ~7 frontend files. That is well inside
one task, and splitting would ship `tools/verify.sh` with its lint leg behind a
flag — i.e. a flagless run that does not mean "done", which breaks the one
contract (FR-V1) the whole exercise exists to establish.

---

## 4. Architecture

### 4.1 `tools/verify.sh` — the only real engineering

Contract (fixed by spec §3.4, unchanged): **flagless exit 0 means the branch may
be called done.** `--quick`/`--no-docker` exit 0 too and explicitly do not.

**Legs, in order:**

| # | Leg | Flagless | `--no-docker` | `--quick` |
|---|---|---|---|---|
| 1 | pin consistency (§5.4) | ✅ | ✅ | ✅ |
| 2 | go build — 12 services + 12 shared modules | ✅ | ✅ | ✅ |
| 3 | go vet (via golangci-lint `govet`) | ✅ | ✅ | ✅ |
| 4 | go test `./... -count=1` per module | ✅ | ✅ | ⏭ |
| 5 | golangci-lint fmt `--diff` + run | ✅ | ✅ | ⏭ |
| 6 | frontend `npm ci` + `build` + `test` | ✅ | ✅ | ⏭ |
| 7 | frontend `eslint .` | ✅ | ✅ | ⏭ |
| 8 | docker per-service image build | ✅ | ⏭ | ⏭ |

`--no-ui` skips 6 and 7 (§5.5). `--base <rev>` narrows change detection for
leg 8. `-h`/`--help` → 0; unknown option → 2 (FR-V4).

**Module discovery is computed, never listed.** `find services shared/go -name
go.mod` is the single source of the module set. A hardcoded list is exactly how
`scripts/ci-build.sh` drifted three services behind `go.work` (§6.1); the design
removes the possibility rather than fixing the instance.

**Linter bootstrap.** `verify.sh` owns `ensure_golangci()`, ported from atlas's
`tools/lint.sh`: download the pinned prebuilt release, verify its published
SHA256, fall back to `go install` when curl/sha256sum are unavailable, cache at
`.cache/tools/bin/golangci-lint-$GOLANGCI_LINT_VERSION`. This is the only place
in the repository that touches the network (NFR-5). Carry atlas's two hard-won
settings verbatim, both of which apply here for the same reason:
`GOLANGCI_LINT_CACHE=$ROOT/.cache/golangci-lint` (per-worktree, or cached issues
replay against files from a sibling worktree that no longer exist) and
`--allow-parallel-runners` (the machine-global `$TMPDIR` flock makes two
concurrent worktree runs fail spuriously with exit 3 and no findings).

`.cache/` must be added to `.gitignore` — it is not there today.

### 4.2 `format-on-write.sh` — the adaptation

Structure and every fail-open branch stay byte-for-byte; only the two bindings
change.

| Half | atlas | home-hub |
|---|---|---|
| Go | `source tools/toolchain.versions`; `.cache/tools/bin/golangci-lint-$VER fmt` | identical — the file it sources now exists here (FR-F1) |
| Frontend | `*/services/atlas-ui/*.ts{,x}` → `npx --no-install prettier --write` | `*/frontend/*.ts{,x}` → `npx --no-install eslint --fix` |

`--no-install` is what keeps FR-F5 true for the frontend half: no
`node_modules`, no formatter, silent exit 0. The Go half is already fail-open on
a missing cached binary, and the hook still never bootstraps — `verify.sh` does
(FR-F3), so a first Write never eats a multi-minute download.

Confirmed against `frontend/package.json`: **no prettier dependency**, eslint 10
only. FR-F4's premise holds.

### 4.3 Hooks — verbatim set

All eight contain zero `atlas` occurrences at the pinned commit (verified by
`grep -c atlas` per file: all 0), so FR-H1/FR-H2/NFR-4 are satisfied by a plain
`cp`. Their two external dependencies — `tools/task-brief.sh` for
`commit-boundary.sh`, `tools/task-numbers.sh` for
`task-num-collision-detector.sh` — are created in L1, so no reference is ever
dangling (FR-H4, FR-H5). `turn-budget.sh` and `turn-budget-guard.sh` reference
`task-implementer` only in operator-facing text, and that name is the one this
repository will use, so nothing needs rewriting.

### 4.4 Agents — genericization mapping

The rules port unchanged; only illustrations move. FR-A3/FR-D2 require a real
home-hub example wherever one exists and a neutral one where it does not
(FR-D3 — a rule is never dropped for want of an example).

| atlas illustration | home-hub replacement |
|---|---|
| packet encode/decode, opcode tables | JSON:API resource/attribute shapes |
| WZ data as ground truth | service source + migrations as ground truth |
| IDA / client binary reading | *no analogue* → neutral: "the authoritative artifact" |
| `libs/atlas-constants` lookup | `shared/go/model`, `shared/go/tenant` lookup |
| `atlas-ui` prettier | `frontend/` eslint |
| `verify.sh --all/--facts/--no-ui` | home-hub's flag set from §4.1 |
| cross-service event fan-out | recipe → shopping-list ingredient flow; meal-plan → tracker |

`task-implementer` keeps its 120-call budget and `PARTIAL` hand-back — the cap is
enforced by `turn-budget-guard.sh`, which is one of the verbatim files, so
changing it would desynchronize the hook (FR-A4).

### 4.5 Owner docs

Nine files, 1,527 atlas lines. Expect meaningfully fewer here: `agent-dispatch.md`
loses its `atlas-*` rename cutoff note (FR-D6), `verification.md` loses atlas's
bake/`--facts` machinery, and `superpowers-integration.md` drops its Packet Work
section entirely.

`superpowers-integration.md` reconciliation (FR-D4): the two share an identical
eight-heading skeleton. atlas's 185 lines are a superset of home-hub's 76 plus
`### Task resolution`, `### Artifact location override`, `### Phase 4 context
budget`, `### Picking the roster`, `### What a reviewer returns`, and
`## Packet Work`. So the merge is mechanical and safe: **adopt atlas's structure,
drop `## Packet Work`, and keep home-hub's existing body text inside every shared
heading**, adding atlas's subsections around it. Home-hub content survives by
construction, which is what AC-17 asks a human to confirm.

### 4.6 `CLAUDE.md`

Eight headings in spec §5.3 order. Home-hub's current file is 69 lines of prose
carrying nine operative rules; every one has a destination:

| Current prose | Lands in |
|---|---|
| Docker builds on shared-library change | `## Done means verified` — as "flagless `tools/verify.sh`", which enforces it (FR-V9) |
| `scripts/local-up.sh` | `## Repository conventions` |
| plan≠implement | `## Development workflow` |
| four-phase flow + worktree discipline | `## Development workflow` |
| artifact-location override | `## Development workflow` |
| code review before PR | `## Done means verified` |
| design/plan output style | `## Development workflow` |
| verification over memory | `## Evidence & grounding` |
| shared-type refactor / no cross-layer internals | `## Repository conventions` |

The "Code Review Pattern" list becomes a `## Dispatching agents` bullet plus a
row pointing at `docs/review-protocol.md` — that is precisely the compression the
owner-doc table buys.

---

## 5. Open questions — answered

### 5.1 Q1, linter backlog → fix it here (§3.5)

### 5.2 Q2, `.golangci.yml` content → adopt atlas's wholesale

`version: "2"`, `linters.default: standard`, `formatters: [gofumpt, goimports]`
with `local-prefixes: [github.com/jtumidanski/home-hub]`. Drop atlas's single
`exclusions.rules` entry — it names an atlas test file.

**Do not port atlas's `--new-from-rev` gating.** atlas gates linter findings to
new code because its backlog is too large to clear; home-hub's is 110 and can be
zero from day one. A clean tree with no rev-gate is strictly stronger, needs no
burn-down document, and makes the flagless contract literally true instead of
true-modulo-a-baseline. This is a deliberate divergence from atlas and should be
noted in `docs/verification.md`.

### 5.3 Q3, future of `scripts/*.sh` → thin delegates

`tools/verify.sh` owns all leg logic and gains `--only <leg>[,<leg>]`. Then:

```sh
scripts/ci-build.sh  → exec tools/verify.sh --only build,frontend-build
scripts/ci-test.sh   → exec tools/verify.sh --only test
scripts/lint-all.sh  → exec tools/verify.sh --only lint,eslint
```

Rationale: the three scripts' granularity does not match the flag set —
`ci-build.sh` bundles the frontend build into the Go build leg, `lint-all.sh`
bundles eslint into the Go lint leg — so composing *from* them (FR-V10's first
preference) cannot express `--quick`. Inverting the dependency keeps exactly one
module list and one command per leg in the repository, which permanently removes
the drift class that §6.1 documents. It also discharges FR-L4 by construction:
`lint-all.sh` stops invoking an unpinned `golangci-lint` because it stops
invoking `golangci-lint` at all. `--only` is additive and does not weaken
FR-V1–V4.

### 5.4 Q4, toolchain pins → `GO_VERSION=1.27.0` + a cheap consistency leg

The PRD's premise is stale: `pr.yml` already pins `go-version: '1.27'` in all 13
Go jobs, not 1.26. Current state is `go.work` → `go 1.26.1`, CI → `1.27`, local
→ `1.27.0`.

Set `GO_VERSION=1.27.0`, `GOLANGCI_LINT_VERSION=v2.13.1` in
`tools/toolchain.versions`. **Do not port `tools/toolchain-pin-guard.sh`** —
atlas's version reconciles `go.mod`, `go.work`, `Dockerfile` ARGs and
`docker-bake.hcl` across 80 modules; home-hub has neither bake nor that surface
area. Instead, `verify.sh` leg 1 is ~20 lines comparing the major.minor of
`GO_VERSION` against `go.work`'s `go` directive and every `go-version:` in
`.github/workflows/*.yml`, failing on mismatch. The `go.work` `1.26.1` vs `1.27`
gap is a real (benign) divergence and this leg will surface it — bumping `go.work`
to `1.27.0` is the intended fix and is behaviour-neutral.

### 5.5 Q5, `--no-ui` → yes

The frontend legs are `npm ci` + `tsc -b && vite build` + 695 vitest tests,
around a minute even warm. Backend-only iteration should not pay it. ~4 lines.

---

## 6. Deviations from the PRD

Each is a place where the literal requirement is self-defeating. I recommend all
six; none change scope. Flagging them rather than silently complying.

### 6.1 Coverage: match CI, not `scripts/ci-build.sh` (FR-V5, FR-V6)

FR-V5/V6 pin `verify.sh` to `ci-build.sh`'s "eight `shared/go/*` packages and
nine services". That set is **stale**. `go.work` declares 12 services and 12
shared modules; `pr.yml` builds and tests all 24. `ci-build.sh` and `ci-test.sh`
are missing `dashboard-service`, `tracker-service`, `workout-service`,
`shared/go/dashboard`, `events`, `kafka`, `retention`.

Encoding that gap into `verify.sh` would mean a flagless exit 0 that never
compiled three shipped services — the exact false "done" FR-V1 exists to prevent.
CI is the authority (atlas's own header says so). **`verify.sh` covers all 24
modules by discovery.** Free: §3.4 confirms all 24 already build and test green.

### 6.2 Docker leg: mirror CI's matrix, not `docker compose build` (FR-V9)

FR-V9 says "builds the compose images". Three problems: `docker compose` in
`deploy/compose` requires `--env-file .env` (NFR-3 forbids reading `.env`); it
includes an `nginx` service CI never builds; and CI builds each service with
`docker build -f services/<svc>/Dockerfile -t home-hub-<svc>:pr .` from the repo
root. Mirroring CI's per-service matrix needs no secrets and is what actually
gates the PR.

**Trigger, also widened:** FR-V9 fires the leg "when — and only when" the diff
touches `shared/`. CI's `paths-filter` fires a service's image build when *that
service* changes too. Under FR-V9's literal rule, a broken `Dockerfile` in a
service-only change passes flagless `verify.sh` and fails CI. Implement CI's
rule — `shared/**` fans out to all 13 images, a service change builds that
image — which contains FR-V9's condition as its fan-out case. `--no-docker`
remains the escape hatch and `--base <rev>` narrows the diff.

### 6.3 Failure handling: run every leg (FR-V11)

FR-V11 says stop at the first failing block. atlas deliberately does the
opposite — "every check runs even after an earlier one fails, so one pass gives
the complete picture" — because a 5–15 minute gate that reveals one problem per
run turns remediation into serial round-trips. §5's L5 layer, with 290 findings
across 24 modules, is precisely that situation.

**Run all legs; print a `PASSED / FAILED / SKIPPED` summary; reproduce the first
failing leg's output verbatim in the summary.** That satisfies FR-V11's intent
(the first failure is intact and unburied) without its cost.

### 6.4 `execute-task.md` and `spec-task.md` must change (FR-C2)

FR-C2 says the existing nine commands are unchanged. But home-hub's
`execute-task.md` invokes `superpowers:subagent-driven-development` with generic
implementers and never names an agent, and `spec-task.md` picks task numbers by
hand. Ship the trio and `task-numbers.sh` without touching these two and both are
inert — directly contradicting two of the PRD's own user stories ("`/execute-task`
dispatches a budget-capped implementer", "collision-safe task numbering so two
concurrent `/spec-task` runs cannot claim the same NNN").

Read FR-C2 as **"add no command beyond `/fix-pr-bug`"**, and make the minimal
wiring edits:

- `spec-task.md` Step 1 → `tools/task-numbers.sh next` as the only number source.
- `execute-task.md` Step 4 → `subagent_type: task-implementer` per unit,
  `task-reviewer` (sonnet) per unit, `task-verifier` (haiku) for the gate, plus
  the `PARTIAL` continuation protocol.

Porting atlas's whole 376-line `execute-task.md` is the alternative; I'd rather
graft the trio wiring onto home-hub's existing 60-line command and keep its
worktree-discipline language, which atlas's version does not have.

### 6.5 Frontend eslint backlog is undisclosed scope (FR-V7 / AC-8)

§3.3: 10 errors today, and FR-V7 puts eslint in the flagless gate, so AC-8
cannot pass without clearing them. Seven of the ten are mechanical
(`no-unused-vars`, two `no-explicit-any` in a test, two
`react-refresh/only-export-components` → move the non-component exports to a
sibling module, three `set-state-in-effect` → derive during render or
`useSyncExternalStore`).

The two `react-hooks/rules-of-hooks` at `DashboardDesigner.tsx:56-57` are
different: a conditional hook call is a latent correctness bug, and fixing it
means hoisting hooks above an early return — which **changes behaviour** and so
collides with FR-L3's behaviour-preserving constraint. Recommendation: fix it
properly (it is a real bug, and the file has a test path), and if the fix proves
non-trivial during execution, fall back to a scoped `eslint-disable-next-line`
carrying a TODO rather than weakening `.golangci.yml`'s frontend counterpart or
dropping the leg. That decision belongs to whoever has the file open.

### 6.6 No `--new-from-rev` (implicit in FR-L3, and a divergence from atlas)

Covered in §5.2. Recorded here because §7 check 1 of the spec is about hook
byte-identity, not lint policy — this divergence does not affect parity.

---

## 7. Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| `golangci-lint fmt` over 180 files produces a diff too large to review | high | Land it as **one isolated commit** touching nothing else, generated by a single tool invocation, with the command in the message. Reviewable by re-running the tool, not by reading. |
| An `errcheck` fix changes behaviour (checking an error that was ignored on purpose) | medium | Default to `_ =` for genuinely-ignored returns; only surface a real error where the surrounding function already returns one. Never introduce a new error path (FR-L3). |
| Deleting `unused` functions removes something reflectively referenced | low | GORM/JSON:API code is not reflective over Go func names here; `go build ./...` per module plus the existing test suite is sufficient evidence. |
| `DashboardDesigner.tsx` hook fix regresses the designer | medium | §6.5. Vitest covers 105 files; check whether the designer is among them before editing. |
| Docker leg makes flagless `verify.sh` unusably slow | medium | 13 images from cold. Change-gated by default (§6.2); `--no-docker`/`--quick` for the inner loop; NFR-2 satisfied because a non-`shared/` change builds only affected images. |
| The eight verbatim hooks drift from atlas via a well-meaning edit | low | AC-1's `diff` is the guard; state in `docs/verification.md` that these files are copies and are fixed upstream. |
| `.cache/` gets committed | low | Add to `.gitignore` in L0, before the linter is ever bootstrapped. |

---

## 8. What the plan phase must sequence

Ordering constraints that matter, beyond the L0→L5 spine:

1. **`.gitignore` `.cache/` before any bootstrap.** Otherwise a 60 MB binary is
   staged by the first `git add`.
2. **`tools/verify.sh` before L5.** Remediation is defined by what the gate
   reports; measuring it any other way invites a second, different number.
3. **Formatter commit isolated and first within L5.** It touches 180 files; every
   later lint fix must be readable against it, not inside it.
4. **`CLAUDE.md` last within L4.** FR-M2/AC-14 require every table target to
   exist; writing the table before the nine docs land produces broken links that
   nothing catches until the end.
5. **`.claude/settings.json` after all nine hooks exist.** AC-12 checks that every
   wired path resolves to an existing executable; wiring first means a broken
   session for whoever `/clear`s next.
6. **AC-20 is a reporting obligation, not a check.** Spec §7 checks 1, 4, 5 and 6
   are pairwise across four repositories. Report home-hub's side and state plainly
   that the comparison is not evaluable from here. Asserting them passed is a
   defect, per the PRD's own wording.

---

## 9. Decisions requiring the user's confirmation

Everything above is decided. These three are the ones where I have chosen but the
call is arguably the user's — raised now so the plan does not have to stop:

1. **§6.4** — editing `execute-task.md` and `spec-task.md` despite FR-C2. Without
   it the agent trio and collision-safe numbering ship dead.
2. **§6.2** — Docker leg fires on service changes too, not only `shared/`. Slower
   than FR-V9's literal rule; catches Dockerfile breakage that FR-V9 would miss.
3. **§6.5** — the frontend eslint backlog is real, undisclosed scope. If it should
   be deferred, the honest form is `--no-ui` on the eslint leg with a follow-up
   task, and AC-8 stated as passing only under `--no-ui`.
