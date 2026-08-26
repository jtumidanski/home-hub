# Brief — Process Parity Phase 4 (home-hub)

**Read `docs/process-parity.md` first.** It is the canonical specification for this
work and was written to be self-contained: you do not need any prior conversation.
This brief only tells you which parts apply to home-hub and in what order.

## Provenance

`docs/process-parity.md` here is a verbatim copy pinned at atlas commit
`e83f59e61` (re-synced 2026-08-26 from the original pin `e75c2a168`; the only
intervening change was an evidence note added to §7 check 3), on the unmerged
branch `task-266-process-parity-agent-rename`. Atlas
phase 1 is complete and gated green but not yet merged.

Set `ATLAS` to the path of the atlas worktree
`.worktrees/task-266-process-parity-agent-rename` on your machine. Before copying
anything, confirm the copy is still current:

```sh
diff "$ATLAS/docs/process-parity.md" docs/process-parity.md
```

If that diff is non-empty, the atlas PR changed the spec after this brief was
written. Stop and re-sync before proceeding — do not merge the two by hand.

## Your task

Execute `docs/process-parity.md` §6 step 2 for home-hub, using this repo's own
four-phase flow (`/spec-task` → `/design-task` → `/plan-task` → `/execute-task`).

Full parity is the scope: the portable hooks, the agent trio, the verify
entrypoint, the owner documents, `/fix-pr-bug`, the `.claude/settings.json` hook
wiring, and the `CLAUDE.md` restructure.

**Do not skip `/design-task`.** home-hub is the harder of the two repos needing a
verify entrypoint: unlike Harbormaster it has no consolidated command list to work
from, and unlike MyFleet it has no `make ci`. `tools/verify.sh` has to be composed
from the `scripts/` directory, and there is an unresolved linter question below.

## What home-hub already has

Do not re-create these:

- the four phase commands, `/audit-plan`, `/review-todos`, `/service-doc`
- `backend-guidelines-reviewer`, `frontend-guidelines-reviewer`,
  `plan-adherence-reviewer`, `todo-scanner`, `service-documentation`
- `backend-dev-guidelines` / `frontend-dev-guidelines` skills, `skill-rules.json`,
  the `skill-activation-prompt` hook
- `docs/superpowers-integration.md` — **this one already exists.** Reconcile it
  against `$ATLAS/docs/superpowers-integration.md` rather than overwriting.
  Atlas's version is 185 lines and has diverged; diff them and merge deliberately.

## What home-hub is missing that atlas has

- `tools/` does not exist at all. Create it, and port `task-numbers.sh` (plus its
  test `task-numbers_test.sh`) and `task-brief.sh` from `$ATLAS/tools/`, along
  with the `task-num-collision-detector.sh` SessionStart hook. Without these there
  is no collision-safe task numbering and `commit-boundary.sh` has a dangling
  reference.

## home-hub's binding row (`docs/process-parity.md` §4)

| Binding | Value |
|---|---|
| Verify entrypoint | **create** `tools/verify.sh` composed from `scripts/ci-build.sh`, `scripts/ci-test.sh`, `scripts/lint-all.sh` |
| Go layout | `go.work`, `services/*` + `shared/go/*` |
| Frontend path | `frontend/` |
| Go formatter | **unresolved** — see below |
| Local deployment | `scripts/local-up.sh` (existing `CLAUDE.md` rule; preserve it) |
| Docker rule | existing `CLAUDE.md` requires verifying Docker builds when shared libraries change — **fold into `verify.sh`, do not drop** |

`tools/verify.sh` must match atlas's contract: **flagless run exits 0 means the
branch may be called done**; `--quick` / `--no-docker` also exit 0 but skip the
slow gates and do NOT count as done.

The Docker rule matters here more than elsewhere. `shared/go/*` is consumed by a
dozen services, and the existing `CLAUDE.md` calls out Docker verification on
shared-library changes specifically. That belongs in the flagless run, gated on
whether `shared/` changed — not in prose where it can be skipped.

## Open question you must resolve, not defer

`format-on-write.sh` needs a pinned Go formatter. Atlas sources
`tools/toolchain.versions` and uses a pinned `golangci-lint`; MyFleet uses
`tools/lint.sh` + `tools/lint.versions`. **No pinned linter configuration was
found in home-hub.** `scripts/lint-all.sh` exists — read it and determine what it
actually invokes.

Decide this during `/design-task` and implement it. Do not port
`format-on-write.sh` with a dangling reference to a file that does not exist, and
do not silently drop the hook — either would leave the parity claim false.

## Copying the portable files

From `$ATLAS`, copy verbatim into `.claude/hooks/`:

`wait-loop-guard.sh`, `wait-loop-guard_test.sh`, `block-home-paths-in-docs.sh`,
`turn-budget.sh`, `turn-budget-guard.sh`, `fork-dispatch-guard.sh`,
`commit-boundary.sh`, `task-num-collision-detector.sh`

These contain no atlas-specific strings as of `e75c2a168`. Verify after copying:

```sh
grep -l 'atlas-' .claude/hooks/*.sh   # must print nothing
```

`format-on-write.sh` must NOT be copied verbatim — see the open question above.
Its frontend half rebinds from `services/atlas-ui` to `frontend/`.

The agent trio (`task-implementer`, `task-verifier`, `task-reviewer`) and the owner
documents copy from `$ATLAS/.claude/agents/` and `$ATLAS/docs/`. The owner docs
need the §5.2 genericization pass — replace atlas-specific examples (packet work,
WZ data, IDA) with home-hub equivalents or neutral ones. **Do not delete a rule
because its example does not transfer; find a new example.**

Do not port: anything under `docs/packets/`, `docs/reverse-engineering.md`.

## The one carve-out that differs from atlas

`docs/process-parity.md` §7 check 3 exempts two files. In home-hub **only
`docs/process-parity.md` is exempt** — the `docs/agent-dispatch.md` exemption is
atlas-only, because atlas is the only repo that ever used the `atlas-*` names.
Confirmed 2026-08-26: this repo has zero such references today. Your check is:

```sh
git grep -lE 'atlas-(implementer|verifier|reviewer)' -- . ':!docs/tasks' \
  | grep -vxE 'docs/process-parity\.md'
```

It must print nothing.

## Done

`docs/process-parity.md` §7 lists the six checks. Checks 1, 4, 5, and 6 are
cross-repo and cannot be fully evaluated from home-hub alone — report your side of
them and say plainly that the pairwise comparison is not evaluable here. Checks 2
and 3 are fully checkable in this repo and must pass.

Report back what you could not verify rather than asserting it.
