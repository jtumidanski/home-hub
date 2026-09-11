# Phase 5 — After Implementation

The four-phase flow ends at `/execute-task`. The work does not: PR validation,
live-environment testing, bug reproduction, regression investigation, and
follow-up fixes all happen afterwards, and nothing told those sessions to
delegate. This document is the missing phase.

It introduces no new context-clearing rule and requires nothing of the user. It
generalizes the handoff principle the execute loop already applies — see
[`docs/agent-dispatch.md`](agent-dispatch.md) §Context handoff — to the work
that follows implementation.

`/fix-pr-bug` mechanizes this document's loop end to end; if anything below is
ambiguous, that command's steps are the concrete procedure.

---

## Why

Measured elsewhere, across one task's post-PR phase — several sessions of live
testing and bug fixing after the plan itself was already done:

| | Execute phase | Post-PR phase |
|---|---|---|
| Billed input | the large majority of the task's total | a small fraction of the task's total |
| Main-thread share | roughly a fifth | **the large majority** |
| Subagents | many, across the plan | **almost none** |
| Peak context | bounded per-session by the phase-4 handoff discipline | **large, solo, unbounded** |

Main-thread tokens are the expensive kind: the context grows monotonically and
every turn re-reads all of it. A turn at 200–330k costs several times the same
investigation inside a fresh subagent starting near-empty. Some of that
task's post-PR sessions burned tens of millions of tokens between them with no
subagents dispatched at all.

The habit was half-formed already in that measurement: the task folder
already contained several `bug-<slug>.md` diagnosis files from live-testing
sessions. **The artifact habit existed; the delegation habit did not.** The
session that came closest to the right shape opened by reading a bug file and
dispatched agents against it rather than fixing inline.

The rediscovery cost is visible too: one post-PR session re-grepped the repo
to relocate code the same branch had written days earlier, and another
re-read `plan.md` repeatedly to re-establish what the plan had already said.
Both are exactly what a written diagnosis file and a fresh dispatch would have
avoided.

---

## The loop

**Reproduce inline. Diagnose into a file. Delegate the fix. Verify fresh.**

### 1. Reproduce — stay in your own context

Reproduction is interactive: round-trip time against the running stack
matters more than tokens, and each step depends on what the last one showed.
Do this yourself. Do not delegate it.

Read service logs first — `deploy/compose/logs.sh <service>` — scoped to the
service(s) you actually name, never a whole-stack dump. Confirm the tenant and
exact build you are reproducing against before anything else; the wrong
tenant or a stale build sends the whole investigation down the wrong path.

### 2. Write the diagnosis to `docs/tasks/<task>/bug-<slug>.md`

Before dispatching anything. This is the boundary: everything after it must be
resumable from repository state plus this file.

```markdown
# bug: <one-line symptom>

**Reproduced:** <tenant, build, exact steps>
**Observed:** <what happens, with the log line / error verbatim>
**Expected:** <what should happen, and where that is specified — PRD/FR, plan task>
**Root cause:** <what you established, with file:line>  — or: "not yet established; <what is ruled out>"

## Fix

- `path/to/file.go:120` — <what changes here>
- `path/to/other_test.go` — <the test that must fail before and pass after>

## Not yet answered

- <anything the fix agent must decide, and what it should do if unsure>
```

The `## Fix` section is a `### Files` inventory by another name, and it does the
same job: it removes the implementer's discovery phase — the phase that inflates
context before a single edit happens. You already know these paths from
reproducing; the fix agent would otherwise pay to rediscover them, at its own
context depth, on top of what you already paid at yours.

If the root cause is not established, say so explicitly and name what is ruled
out. An honest "not yet established" is a fine brief; a guessed root cause is
not.

### 3. Delegate the fix to a fresh agent

```text
subagent_type: task-implementer
model: sonnet
brief: docs/tasks/<task>/bug-<slug>.md
```

The bug file is the brief. Do not restate it in the dispatch prompt — add only
what the file cannot carry: the worktree absolute path, and any ruling you have
made since writing it.

This is the step most often skipped. It is also where the saving is: the fix
agent starts near-empty instead of inheriting your 300k.

### 4. Verify in a clean context

`task-verifier` (`model: haiku`) for the gate:

```sh
tools/verify.sh --quick --base <last-gated-commit>
```

`task-reviewer` (`model: sonnet`) if the fix crosses a service boundary or
touches a contract — the gate cannot see a seam defect.

Launch the gate backgrounded and keep going; do not poll it.
`.claude/hooks/wait-loop-guard.sh` will refuse the poll anyway.

### 5. Record the outcome

Home-hub has no separate ledger tool. Update `docs/tasks/<task>/bug-<slug>.md`
itself with the outcome — the commit that fixed it, which agents ran
(implementer, and verifier or reviewer if dispatched) with their verdicts, and
whether live testing confirmed it — and append one line to
`.superpowers/sdd/<plan-basename>/progress.md`:

```
Task <N>: bug-<slug> fixed — agent=task-implementer model=sonnet commit=<sha>
```

So the next audit can answer "what did the post-PR phase actually cost"
without reconstructing transcripts, and so a bug file that never records its
resolution doesn't become the next session's rediscovery.

---

## When to hand off your own context

The same question as every other durable boundary: **does the next unit depend
materially on this conversation's history, or only on repository state plus a
short written diagnosis?**

After a bug file is written, the answer is almost always "repository state".
Once you have written the diagnosis, the reproduction conversation that produced
it is no longer load-bearing.

Concretely, in a debugging session:

- **After each bug file is written and its fix dispatched**, ask the question.
  If the next bug is unrelated to the one you just fixed, it is a fresh unit —
  run `/fix-pr-bug` again against its own bug file rather than continuing to
  accumulate.
- **Past ~150k**, stop starting new investigations in this context. Write the
  remaining leads into the task folder and hand off. The `/execute-task` ceiling
  applies here for the same reason it applies there: the marker on disk is
  meaningless if the session that wrote it keeps going.

`/fix-pr-bug` mechanizes steps 2–5 for a single bug.

---

## What does not change

- **Reproduction stays inline.** An over-delegated interactive debugging session
  is worse than an expensive one.
- **The bug-file habit is already right** — this document adds the delegation
  step after it, not a new artifact format.
- **`/execute-task`'s ceiling is unchanged.** Phase 5 borrows it; it does not
  redefine it.
