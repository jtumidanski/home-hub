# Slice First; Escalate When Necessary

How to read a large artifact inside an agent or a controller.

**This is a default, not a prohibition.** If semantic correctness requires the
whole document, read the whole document. Under-reading and getting it wrong
costs a fix round, which is far more than any read it saved. The rule is only
that a whole-file read should be a *decision*, not the reflex opening move.

---

## The measurement

Cost is not bytes; it is bytes × the turns that re-read them. A result entering
at turn 5 of a 200-turn agent is re-billed ~195 times; the same result at turn
190 is re-billed 10 times.

That multiplier lands where it hurts most:

> **The median tool result over 12 KB arrives at position 0.10 of its stream,
> and 75% of them arrive in the first quarter.**

So the heavy tail is not merely large — it is placed exactly where it is most
expensive. This pattern has been measured directly against agent sessions
elsewhere: large documents read whole and early, re-billed dozens of times
over a long session, accounted for a large share of total input cost even
though they were a small fraction of tool calls. The costliest objects in
that analysis were reference documents, not source code — read many times
each because an agent needed one section or a handful of rows, not the whole
file.

**The documents are not the problem.** A complete reference document — a
service-wiring guide, a large architecture doc — is often the reason a
pattern gets applied consistently across many callers, and a thinner version
would just be re-derived by every agent that needs it, at far greater cost.
Change the access pattern; leave the content complete.

---

## The rule

**For any file you expect to exceed ~20 KB, lead with a slice.**

In this repo that includes `docs/architecture.md` (~22 KB) and
`docs/process-parity.md` (~15 KB), any `docs/tasks/<task>/plan.md` for a
large task, and a repo-wide formatter diff (`tools/verify.sh --fix-fmt` can
touch many files at once across the 24 Go modules).

Escalate to a full read when the slice is insufficient — and when you do, say so
in your report, because a document that is repeatedly escalated is a document
that needs restructuring.

This repo has no dedicated slicing tool (no `doc-slice.sh` equivalent) — use
standard `grep`/`sed`/`git` directly:

```sh
# What is in here, and where? Cheap first move on an unfamiliar document.
grep -n '^#' docs/architecture.md

# The one section this batch needs.
sed -n '/^## Data model/,/^## /p' docs/architecture.md

# The rows for the services in scope, table header preserved.
grep -n -A0 '^| auth-service \|^| recipe-service \|^|---' docs/tasks/<task>/audit.md

# A needle in an offloaded tool result, rather than re-reading it whole.
grep -n -C3 'PANIC' <session>/tool-results/<file>.txt
```

Each of these prints only the source path plus the slice, so escalation to a
full `Read` is always one deliberate step away.

## Where each case lands

| Situation | Slice-first move | Escalate when |
|---|---|---|
| One plan task out of a large `plan.md` | `tools/task-brief.sh <plan> <N>` — the brief IS the extract | the task references a decision recorded elsewhere in the plan |
| Auditing many plan tasks (`plan-adherence-reviewer`) | one `tools/task-brief.sh` extract per task under audit | never read the whole plan repeatedly for each task in turn |
| A reference document (e.g. `docs/architecture.md`) | `grep -n '^#'` for an outline, then `sed -n '/^## X/,/^## /p'` for one section | the section cross-references another section you also need |
| An audit or result table | `grep -n` for the rows in scope, header included | a row's meaning depends on the document's preamble — read the preamble once, then slice rows |
| A review diff | `git diff --stat <range>` first, then `git diff <range> -- <file>` per flagged file | a change's correctness genuinely spans files |
| An offloaded `tool-results/*.txt` | `grep`/`sed` from disk | you need the whole log, which is rare |
| A config or routing table | `grep -n` for the entries you need | you are changing its structure, not one entry |
| Source code | targeted `grep -n` / `sed -n` for the symbol, then read the file that owns it | reading a file you are about to edit — do read it |

That last row matters: **targeted read slices are not the problem.** Choosing
what to look at is semantic work, done well. This document is about
whole-document reads as a discovery reflex, not about reading less.

## Front-load the cheap thing, not the expensive one

If you need both an inventory and a detail, take the inventory first:
`git diff --stat` before hunks, `grep -n '^#'` before a section, `ls` before
`Read`. The inventory is small and tells you which expensive read is actually
required — and it arrives at the position where a large result would have
been worst.

## What already works — do not "improve" it

Build and test output is consistently bounded when piped through a filter —
`go build ./... 2>&1 | head -60` and friends. `tools/verify.sh`'s own summary
already reproduces only the first failing leg's output, not every leg's full
log. The tool-result spill mechanism works too — large payloads land as small
stubs and get sliced from disk. Both are the model this document asks
documentation reads to copy, not costs to trim.
