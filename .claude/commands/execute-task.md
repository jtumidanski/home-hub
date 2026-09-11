---
description: Phase 4 — invoke superpowers:subagent-driven-development to implement a planned task in its existing worktree
argument-hint: Task identifier — accepts "task-044-superpowers-integration", "task-044", "044", or "44"
---

You are starting Phase 4 of the Home Hub four-phase development workflow. Argument: **$ARGUMENTS**

## Process

### Step 1 — Resolve the task

Same fuzzy-match algorithm as `/design-task` Step 1 and `/fix-pr-bug` Step 1:

1. Glob `docs/tasks/task-*` (main) and `.worktrees/*/docs/tasks/task-*` (sibling worktrees).
2. Match `$ARGUMENTS` against folder names — exact, number-only (`44`/`044`/`task-44`/`task-044`), or slug fragment.
3. Zero matches → ask for correction. Multiple matches → list and let the user pick.
4. If the task lives only on main with no worktree, stop and tell the user the task needs a worktree.
5. Resolve to `<worktree>/docs/tasks/<id>/`.

Home-hub has no single resolver/facts script for this — assemble it yourself
with the glob above rather than hand-scanning `.worktrees/` some other way
each time.

Do NOT create a new worktree — the worktree was created by `/spec-task` and
must be reused so phase artifacts stay co-located.

### Step 2 — Verify we're in the right worktree

Run `pwd`. If it does NOT match `<worktree>`, tell the user:

> Task `<id>` lives in `<worktree>`. Please `cd <worktree>` and re-run `/execute-task <id>`.

Do NOT auto-`cd` and do NOT create a new worktree — the worktree was created by `/spec-task` and must be reused so phase artifacts stay co-located.

### Step 3 — Validate inputs

Confirm `<worktree>/docs/tasks/<id>/plan.md` AND `context.md` exist. If either is missing, tell the user to complete `/plan-task` first.

### Step 4 — Invoke subagent-driven-development

Use the Skill tool to invoke `superpowers:subagent-driven-development` (default). Pass:

- Plan path: `<worktree>/docs/tasks/<id>/plan.md`
- Context path: `<worktree>/docs/tasks/<id>/context.md`
- Project conventions: `<worktree>/CLAUDE.md`
- **Worktree absolute path** (`<worktree>`) for every dispatched implementer subagent. Subagent prompts MUST follow cwd-discipline: every Bash call prefixed with `cd <worktree> && ...`, post-commit branch verification, no destructive git ops, no `git add -A` / `git add .`.

**Every implementer dispatch uses `subagent_type: task-implementer`** — not
`general-purpose`. That agent carries three contracts the generic implementer
template lacks (tool-call cap with `PARTIAL` hand-back, verification scope,
brief-first discovery). Its contracts override the plugin's
`implementer-prompt.md` wherever the two disagree — in particular, the
plugin template's "run the full suite once before committing" does NOT apply
here; see Step 4c.

Your dispatch prompt still supplies what the agent cannot know: where the
task fits, the brief path, the report path, interfaces and decisions from
earlier tasks, and your resolution of any ambiguity you noticed. Do not
restate the agent's contracts in the prompt.

**Before dispatching a second implementer at the same templated
transformation**, stop and check whether an AST codemod is cheaper than the
remaining manual dispatches — see
[docs/codemod-vs-agents.md](../../docs/codemod-vs-agents.md) for the
break-even arithmetic and the worked example.

If the user explicitly requests inline mode this session (rare), invoke
`superpowers:executing-plans` instead.

### Step 4a — Model discipline for every dispatch

Model selection for every dispatch — the job → model table, the `model: opus`
opt-in, and the escalation rule — is owned by
[`docs/agent-dispatch.md`](../../docs/agent-dispatch.md). Pass an explicit
`model` on every dispatch.

### Step 4b — Check the brief carries its file inventory

Generate each brief with the repo's own script:

```sh
tools/task-brief.sh <worktree>/docs/tasks/<id>/plan.md <N>
```

It extracts the plan's task section verbatim and prints the path it wrote
(default: `<repo-root>/.superpowers/sdd/<plan-basename>/task-<N>-brief.md`).
`<N>` must be the numeric task number — a bare slug exits 2.

This matters because the failure is silent-ish and expensive: when the brief
command fails, the fallback is assembling briefs by hand out of the full
`plan.md` — which is exactly the context bloat the brief exists to prevent.

Because the brief is the plan's task section verbatim, a plan task written per
`/plan-task` already carries a `### Files` block naming every file the task
touches plus the patterns to copy. That block is what removes the implementer's
discovery phase — the phase that inflates context before a single edit happens.

**Before each dispatch, check the generated brief for a `### Files` section.**
If it is missing (an older plan, or a task the planner under-specified),
produce the inventory yourself — once, in your own context — and append it to
the brief file before dispatching:

```markdown
### Files

- `path/to/file.go` — what this task changes here
- `path/to/other_test.go` — new test file

Patterns to copy: `path/to/reference.go:120` (handler shape)
```

For example: say "the JSON:API resource shape in `shared/go/server/response.go`"
and "the tenant scoping in `shared/go/tenant/context.go`", and the implementer
reaches for exactly those two files instead of surveying `shared/go/`.

Appending to the brief, not to the dispatch prompt, is deliberate: the brief
stays the single source of requirements, and a continuation dispatch inherits
it. One inventory pass in the controller costs a fraction of the same
discovery repeated inside a large implementer context.

`task-implementer` reports `NEEDS_CONTEXT` on a brief with no `### Files`
section rather than falling back to a repo sweep. If that comes back, this
step was skipped — do it and re-dispatch.

**Where the brief points at a large reference document** — a design doc, an
audit, a result matrix — name the *slice*, not the document. The brief should
say which section or heading matters, and the implementer reaches it with
`grep -n '^##' <path>` for an outline, then `sed -n '/^## <heading>/,/^## /p'
<path>` for the section, rather than reading the whole file. Measured: two
such documents were read whole 74 times across 25 agent streams and cost 7.5%
of a task's entire tool-result carry, when each agent needed one section.
See [`docs/slice-first.md`](../../docs/slice-first.md).

### Step 4c — Verification runs outside the implementer

`task-implementer` runs only module-local `go build ./... && go test ./...`.
It never runs `tools/verify.sh`, any flag included, `go test -race`, or
docker commands — those in a 400k-token implementer context cost a large
multiple of the same run in a clean 20k one, and their output is the biggest
avoidable consumer of an implementer's window.

**The gate runs concurrently with the next task. Never idle waiting for it.**
See [`docs/agent-dispatch.md`](../../docs/agent-dispatch.md) §Verification
split for why this split exists.

A gate checks a commit, and commits are immutable — task N's verdict is equally
valid whenever it lands. Blocking on it is pure wall clock.

After an implementer reports `DONE` / `DONE_WITH_CONCERNS`:

1. **Launch** the gate for the range `<last-gated-commit>..HEAD`:

   ```sh
   tools/verify.sh --quick --base <last-gated-commit> > /tmp/gate-<N>.log 2>&1
   ```

   with `run_in_background: true`. Pass `--base` — without it the whole-branch
   diff makes each run far slower once any `shared/go/` file has been touched
   (docs/verification.md, "Iteration gate"). Ledger the commit you gated from.

   **If a gate behaves unexpectedly — too broad, too slow, skipping something
   you expected — ask it what it selected before investigating the script:**

   ```sh
   tools/verify.sh --facts --quick --base <last-gated-commit>
   ```

   It prints the change base, changed services and shared modules, the
   fan-out reason, the module count, and every leg that would run, then exits
   without building. It is the same code path as a real run with the work
   removed, so it cannot disagree with one. Never reverse-engineer the
   selection from `verify.sh`'s source; ask it.
2. **Keep going immediately** — do not poll, do not wait. Run the task review,
   then Step 4b's inventory for task N+1, then dispatch task N+1's implementer.
   The gate runs underneath all of it.

   **The per-task review agent is `task-reviewer` (`model: sonnet`), never a
   bare `general-purpose` dispatch.** `task-reviewer` carries both halves the
   contract needs — the durable artifact and the verdict-first return of
   [`docs/review-protocol.md`](../../docs/review-protocol.md).

   Read the review artifact only when the verdict is not `APPROVED`. On
   `CHANGES_REQUIRED` the enumerated `blocking` lines are the fix brief; open
   the artifact when a line is not actionable as written.

   **Right-sizing the task review agent.** A task whose diff was
   codemod-produced and `--check`-confirmed
   ([`docs/codemod-vs-agents.md`](../../docs/codemod-vs-agents.md)) may take
   a reduced or skipped per-task review agent. Every other task — including
   a hand-applied "mechanical" batch with no `--check` PASS behind it — is
   judgment-bearing and gets the full `task-reviewer` dispatch; that is the
   safe default. No rewriter exists yet, so this reduced path is dormant and
   every task takes full review today. This governs the per-task review
   agent only; `tools/verify.sh` and the guideline reviewers
   (`backend-guidelines-reviewer`, `frontend-guidelines-reviewer`) still run
   unconditionally before a PR, alongside `plan-adherence-reviewer`.
3. **Reconcile when it lands**, at the next natural pause (the notification
   from the next subagent). Read the log, ledger PASS or the failing block, and
   record the new last-gated commit.
4. **At most one gate in flight.** If task N+1 finishes while N's gate is still
   running, do not start a second — its commits join the next gate's range.
   The gate covers a *range*, not a task, which is why `--base` is the last
   gated commit rather than `HEAD~1`.

On a verdict:

- **PASS** → ledger it; the range is clean.
- **FAIL** → the quoted failing block becomes a review finding. Feed it into
  the existing fix loop (resume the implementer for rounds 1-3). Never fix it
  yourself in the controller session. **Do not gate the fix round separately** —
  the fix commit joins the next gate's range.
- **ERROR** → the gate did not run. Resolve it (wrong tree, timeout) and
  re-launch. Never treat ERROR as PASS.

A FAIL means one already-dispatched task started on a base that was not yet
green. That is cheap and expected: the fix is a scoped commit on top, the same
as any review finding. It is not a reason to serialize.

The flagless `tools/verify.sh` still runs exactly once, at branch end, in
`superpowers:finishing-a-development-branch`, against the full merge base.
`--quick --base` per task is the inner loop, not the gate — per
`tools/verify.sh`'s own contract, only the flagless run counts as done.

You may dispatch `task-verifier` (`model: haiku`) instead of launching the
gate yourself when you want the verdict summarized rather than reading the log.
The concurrency rule is unchanged: dispatch it and move on, reconcile later.

### Step 4d — Handle `PARTIAL`

`task-implementer` adds a fifth status to the plugin's four. `PARTIAL` means
the tool-call cap (120, warned at 100) was reached with work remaining: the
implementer committed what works and handed back the remainder. **This is the
designed outcome, not a failure — do not scold it, and do not re-dispatch the
same agent to "finish the job" in its now-large context.**

On `PARTIAL`:

1. Ledger it: `Task <N>: partial (commits <a7>..<b7>, cap reached — <what remains>)`.
2. Write a continuation brief beside the original — `task-N-brief-cont.md`,
   or `-cont2` for a second — containing: the remaining work file by file
   (from the report), the `### Files` inventory for just those files, and
   the interfaces and decisions the first implementer recorded.
3. Dispatch a **fresh** `task-implementer` with the continuation brief, the
   same report file path (it is the persistent memory across the split), and
   one line of framing: "A prior implementer completed part of this task and
   hit the tool-call cap. Read the report file for what was done."
4. The task review and `task-verifier` run once over the whole task range
   (BASE from before the first dispatch through the final HEAD) — not once
   per segment.

Two `PARTIAL`s on one task means the plan under-decomposed it. Rule on a
split, ledger the ruling, and carry it forward — the plan task was too big,
which is information `/plan-task` sizing should have caught.

`/fix-pr-bug` Step 4 follows this same protocol for a bug fix's continuation
— that anchor is this step.

### Step 4e — Hand off your own context

Steps 4a-4d bound every subagent's context. Nothing bounds yours, and you are
the one context that survives the entire plan: every implementer report, every
review, every fix ruling, every task-notification wake-up accumulates in it,
and each wake-up re-reads all of it. By the twelfth plan task that is 300k+
tokens billed to tick a checkbox.

The measured cost arithmetic behind this — a real multi-task controller run,
what it cost in tokens and tool calls, and the fresh-session comparison —
lives in [`docs/agent-dispatch.md`](../../docs/agent-dispatch.md) §Context
handoff.

**After completing any plan task, if your context exceeds ~150k tokens — or
4 plan tasks have completed in this controller session, whichever comes
first — hand off. This applies unconditionally, however many plan tasks
remain; there is no carve-out for "only one or two left."**

1. Confirm `<workspace>/progress.md` records every finished task, its commit
   range, and any ruling you made that is not already in a `task-N-report.md`.
2. Tell the user: "Controller context is ~<N>k with <M> tasks remaining.
   `/clear` and re-run `/execute-task <task-id>` — it resumes from the ledger."
3. Stop: no further tool calls after the handoff line. Do not start the next
   task, and do not use the handoff message as a lead-in to one more action —
   a handoff the same context then works past is not a handoff.

This is the controller-shaped case `docs/agent-dispatch.md`'s context-handoff
guidance describes: in `/execute-task` the controller *is* the loop — it
dispatches implementers, reconciles gates, keeps the ledger — so "the next
unit" here is the loop itself, which a dispatched subagent cannot take over.
That is why step 2 above is a `/clear` instruction to the user rather than a
fresh-agent dispatch.

This is safe because the ledger is already the recovery map the skill designs
for: it resumes at the first task with no `Task <N>: complete` line, and the
workspace (briefs, reports, review packages) lives on disk, git-ignored, not in
your context. Handing off mid-plan is cheaper than finishing it large, and it
costs nothing in implementation quality — implementer contexts are untouched
either way.

Hand off unconditionally, regardless of remaining task count, when the next
task is a self-contained detour from the rest of the plan — a tooling
investigation, a spike, a docs sweep. Those share no state with what you are
carrying, so they pay full freight for none of it.

### Step 4f — Record what each agent cost

Home-hub has no separate ledger tool. When you reconcile an agent
(implementer, verifier, reviewer), append one line by hand to
`.superpowers/sdd/<plan-basename>/progress.md`:

```
Task <N>: <status> — agent=<type> model=<model> commit=<sha>
```

Reviewer rows add `verdict=<verdict> caused-fix=<yes|no>`; when you hand off
at Step 4e, add a `handoff after Task <N>, context~<n>k` line.

Record only what you actually know — an omitted field stays out of the line.
**Do not estimate.** A fabricated turn count or token figure poisons the next
session's read of the ledger worse than a gap would.

Batch this with the ledger edit and the next dispatch — it is one more line in a
turn that is already happening, not a turn of its own.

**Batch the ledger update with the next dispatch.** Editing `progress.md` and
the following brief/dispatch call are independent — issue them in one message.
A standalone turn for a small checkbox edit costs the same as a turn that does
real work, and there are one or two of them per plan task.

### Step 5 — On completion

After all plan tasks complete and verify, the chosen skill hands off to `superpowers:finishing-a-development-branch`. Honor that handoff. Then suggest:

> All plan tasks complete. Recommend running `superpowers:requesting-code-review` next, which dispatches the appropriate reviewer agents (plan-adherence, backend-guidelines, frontend-guidelines) in parallel.

Whatever the plan's size, plan adherence is `plan-adherence-reviewer`'s job and
nothing else's. On a long plan, prefer scoping that agent to a task range over
dispatching ad-hoc `general-purpose` "audit plan tasks N-M vs code" agents,
and never run two overlapping-range dispatches of it at once.

## Important Rules

- The worktree was created by `/spec-task`. NEVER create a new one here.
- Implementers are `task-implementer`, never `general-purpose`.
- Per-task reviewers are `task-reviewer`, never `general-purpose` (Step 4c).
- Never poll a backgrounded gate — `.claude/hooks/wait-loop-guard.sh` refuses it.
- Never reverse-engineer the gate's selection from its source; ask
  `tools/verify.sh --facts` (Step 4c).
- Never run `tools/verify.sh` inside an implementer — that is `task-verifier`'s job (Step 4c).
- Never dispatch a brief with no `### Files` section (Step 4b).
- Never carry the controller past ~150k tokens, or 4 completed plan tasks in one session — hand off to a fresh session via the ledger, unconditionally, regardless of tasks remaining (Step 4e).
- Never start implementation outside the task worktree.
- Follow plan steps exactly; stop and ask when blocked rather than guessing.
- Run the verification commands the plan specifies; don't claim completion based on assumption.
