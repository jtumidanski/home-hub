# Agent Dispatch

This document owns *how* to dispatch any agent — model, budget, isolation,
handoff — for every dispatch in every session, including ad-hoc ones outside
the four-phase workflow. `docs/superpowers-integration.md` owns *which*
command, agent, or skill to reach for in a given situation.

---

## Model selection

Pass an explicit `model` on **every** Agent/Task dispatch. Never rely on
inheritance: an unspecified model inherits the main-loop model (Opus), and an
Opus subagent turn costs ~7x a Sonnet one.

The pin is chosen by the **job the agent is doing**, not by its
`subagent_type`. Named reviewer agents carry a Sonnet pin in their
frontmatter, but an ad-hoc `general-purpose` dispatch carrying a review
prompt does not — that is the hole this rule closes.

| Job | Model | Notes |
|---|---|---|
| Review, verify, audit, re-review, whole-branch review | **`sonnet`** — always | No exceptions. Reviewing is reading against a checklist; Opus buys nothing and these run long |
| Scan, inventory, doc sweep, file-finding | `haiku` | `todo-scanner`'s frontmatter pin |
| Run the verification gate (`task-verifier`) | `haiku` | Frontmatter pin; it runs one command and quotes the output |
| Implement a plan task (`task-implementer`) | `sonnet` | Default; frontmatter pin |
| Backend/frontend/plan-adherence audit (`backend-guidelines-reviewer`, `frontend-guidelines-reviewer`, `plan-adherence-reviewer`) | `sonnet` | Frontmatter pin |
| Service documentation refresh (`service-documentation`) | `sonnet` | Frontmatter pin |
| Implement a plan task tagged `model: opus` in plan.md | `opus` | Opt-in only — see below; pass `model: opus` on the dispatch to override the frontmatter |

A plan task may be tagged `model: opus` in `plan.md` when it is genuinely
derivation-heavy: a cross-service contract change, a saga-shaped multi-step
orchestration across services, or a data-model derivation with no existing
pattern to copy. `/plan-task` should apply that tag sparingly and justify it
in one line. Everything else — REST surfaces, GORM entities, Kafka consumers,
React components, tests, template routing — runs Sonnet.

If an implementer comes back wrong twice on Sonnet, escalate that one task to
Opus and note it, rather than raising the default.

Never use Fable for background or review workflows.

## The implementer budget

The implementer budget is **120 tool calls, warned at 100**, counted by
`.claude/hooks/turn-budget.sh` and contracted in
`.claude/agents/task-implementer.md`. At the cap the implementer commits what
works and reports `PARTIAL`; the controller dispatches a continuation. The
number is changed in the counting hook only.

The cap is **binding**: `.claude/hooks/turn-budget-guard.sh` (PreToolUse)
denies subagent calls past CAP+5, exempting the commit-and-report path —
controllers are never blocked by it.

The underlying arithmetic: context grows with turn count and every turn
re-reads all of it, so one 600-turn agent costs far more than the same work
split across fresh contexts. Splitting is the designed outcome, not a
failure.

Before dispatching a second implementer at the same templated transformation,
check whether an AST codemod is cheaper than the remaining manual dispatches
— see [docs/codemod-vs-agents.md](codemod-vs-agents.md).

## Verification split

Implementers never run `tools/verify.sh`, `-race`, or docker — a `--quick`
run inside a 400k-token implementer costs a large multiple of the same run in
a clean 20k one, and its output is the biggest avoidable consumer of an
implementer's window. Implementers run module-local
`go build ./... && go test ./...` and nothing more; the repo gate belongs to
`task-verifier`, in its own clean context, invoked as
`tools/verify.sh --quick --base <last-gated-commit>`.

For the concurrency procedure that runs that gate against the rest of the
plan — launch, keep going, reconcile, at most one gate in flight — see
`/execute-task` Step 4c. That is command mechanics and stays there.

## Inline vs delegate

Delegation is strongly preferred when it replaces a meaningful sequence of
expensive turns in an already-large context. It is a loss when it replaces one
or two cheap ones.

A fresh subagent carries a **dispatch floor** before it has done anything —
the standing prompt, the tool listing, the custom agent roster — well before
its first useful tool call. So the decision is arithmetic, not conceptual:

> **If you can answer the question in roughly one or two targeted tool calls,
> answer it yourself.** A fresh dispatch only pays for itself against several
> turns of your own work at a large context, not one or two.

Decide on *expected turns and context*, not on how big the task sounds. "Audit
the recipe-import consumer's error handling" sounds like a delegation and is
a grep. "Fix this one-line bug" sounds trivial and is a fresh implementer,
because your own context is 300k.

What this rules out, measured elsewhere: one guideline-reviewer agent
dispatched six children for individual checklist questions — "does a
Dockerfile exist for this service", "is this constant already defined
shared-side", and four similar. Together the six cost **4.32M billed input
tokens for 4,669 output tokens**; four of the six returned fewer than 40
output tokens, and one made a single tool call and cost 52,857 tokens on its
own. Their *returns* were maximally compact — the cost was the dispatch floor
plus each child's own context growth, spent on questions the parent could
have answered with a path glob.

The parent then had nothing to do while its async children ran and emitted
**30 `Bash true` no-op turns** — 33% of its tool calls, roughly a third of
the whole agent's spend, for zero information.
`.claude/hooks/wait-loop-guard.sh` now refuses those calls and the polling
equivalents; this rule removes the reason to make them.

**Reviewers do not fan out at all.** A reviewer answers its own checklist. See
[docs/review-protocol.md](review-protocol.md).

### Shrinking the floor itself

Two parts of the floor are ours, not the harness's:

- **The agent roster.** The custom-agent listing ("Available agent types for
  the Agent tool") is real overhead on every dispatch and every turn of a
  session that has one running — it is delivered whether or not the current
  turn needs it. This repo's roster is deliberately small: the trio
  (`task-implementer`, `task-verifier`, `task-reviewer`) plus
  `backend-guidelines-reviewer`, `frontend-guidelines-reviewer`,
  `plan-adherence-reviewer`, `todo-scanner`, `service-documentation`. Do not
  add an agent to `.claude/agents/` for a job an existing one already covers.
- **Bundled skills.** Claude Code's built-in skills (`dataviz`, `claude-api`,
  `design`, `update-config`, …) are listing overhead this repo never invokes.
  `disableBundledSkills: true` is set in `.claude/settings.json` for exactly
  this reason. Plugin skills (`superpowers:*`) and the four-phase commands
  are unaffected; the built-in `/code-review` skill goes with the bundled
  set, and repo code review runs through the reviewer agents anyway (see
  `CLAUDE.md`'s "Code Review Pattern").

**Never idle waiting on a child.** Agent completions arrive as notifications —
do other work, or end the turn and be re-invoked. There is no wait primitive
because none is needed.

## Fork vs fresh context

Fan out with **fresh-context agents, not `subagent_type: "fork"`** — a named
agent type plus an explicit brief. Fork only to continue an interactive
debugging thread, and say why inline. `.claude/hooks/fork-dispatch-guard.sh`
denies an unjustified fork and states the cost.

## Context handoff

The unit of work is a **briefable task, not a conversation.** Context cost
scales with turn count × context size, so 50 turns carried at 190k cost
roughly ten times the same 50 turns at 19k — regardless of what they
accomplish.

At every durable boundary — a commit landing, a verification gate returning,
a fan-out of agents reporting — the decision criterion is whether the next
unit of work depends materially on this conversation's history, or only on
repository state. If it can be resumed from repo state + the task's own
reports + a short written diagnosis, hand off.

Handing off means delegating, not clearing — `/clear` is a user action, and
no agent can clear itself. The diagnosis is written down *before* the
handoff, not carried in your head: one paragraph into the task folder, so
the handoff is lossless even though the reasoning does not survive in
conversation.

**The floor.** Below roughly 40 tool calls a fresh agent re-discovers files
you already hold, and you pay for that discovery twice; prefer continuing.
`.claude/hooks/commit-boundary.sh` encodes this floor (`FLOOR=40`) and raises
the question at commits past it.

**The backstop.** ~150k tokens for a controller, or 4 completed plan tasks
in one session, whichever comes first — the one context that lives for a
whole plan, where every wake-up re-reads it, and the second trigger exists
for a controller that cannot read its own context size. Apply the ceiling
unconditionally: past the threshold, the controller does not start another
plan task, however many remain — a carve-out for "only a couple left" is
exactly the shape of the failure below. `.claude/hooks/commit-boundary.sh`
derives its second-tier `ESCALATE=60` tool-call threshold from this same
150k figure (context growing roughly 2.1k tokens per tool call over a ~23k
standing-prompt floor), so a controller past ~60 calls since its last
handoff is already near the backstop even before it checks token counts
directly. A handoff the same context then works past is not a handoff — a
marker on disk is meaningless if the session that wrote it keeps going.

Generate briefs with `tools/task-brief.sh <plan> <N>`, never by hand out of
`plan.md` — assembling them by hand is exactly the context bloat the brief
exists to prevent. The durable artifacts that make a handoff resumable are
`task-N-report.md` per task and the workspace ledger
`.superpowers/sdd/<plan-basename>/progress.md` for the whole plan.

`/execute-task` Steps 4d–4e are this handoff in its two concrete procedural
forms — `PARTIAL` handling and controller handoff — each keyed to one of
those durable artifacts. Apply the same shape in any session; where a
canonical ledger already exists, write there rather than inventing a second
artifact.

**The rule does not stop when implementation does.** PR validation, live
testing, debugging, regression investigation, and follow-up fixes are the same
question at the same boundaries, and are exactly where discipline tends to be
dropped: a controller that handled a whole plan correctly can still let a
post-PR debugging session run solo in one enormous main-thread context because
nothing about "just fixing one bug" looks like a dispatch decision. The
concrete loop — reproduce inline, diagnose into
`docs/tasks/<task>/bug-<slug>.md`, dispatch a fresh implementer against that
file, verify in a clean context — is
[docs/post-implementation.md](post-implementation.md), mechanized as
`/fix-pr-bug`.

## Recording what a dispatch cost

Home-hub has no separate ledger tool. When you reconcile an agent
(implementer, verifier, reviewer), append one line by hand to the task's
`.superpowers/sdd/<plan-basename>/progress.md`:

```
Task <N>: <status> — agent=<type> model=<model> commit=<sha>
```

Reviewer rows add `verdict=<verdict> caused-fix=<yes|no>`; a handoff records
`handoff after Task <N>, context~<n>k`, which is the only marker anywhere for
a handoff that was written and then worked past.

**Record only what you actually know.** If you cannot state a token count or
a commit sha, leave the field off rather than estimating — a fabricated
number poisons the next session's read of the ledger worse than a gap would.
