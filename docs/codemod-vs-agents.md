# Codemod vs. agents — when a templated transformation earns a tool

This doc answers one question at dispatch time: **you are about to send
several implementers at the same mechanical, repeated change — should one of
them write a codemod instead?**

This document is the rule, a worked example drawn from this plan's own
upcoming linter-cleanup tasks, and the specification for the rewriter that
would apply it — the rewriter itself is **not built yet** (see "Current
status" below).

## The rule

> Evaluate whether an AST codemod is cheaper **before dispatching the second
> implementer** at the same templated transformation.

"Templated transformation" means: the same multi-step edit, repeated across
call sites/files/services, where most steps are syntactic (a rename, an
added import, a threaded parameter, a fixed call inserted at a fixed
location) and at most one step requires per-site judgment (a log message, a
comment, a domain-specific choice).

### The arithmetic that sets the trigger at the second dispatch

The threshold follows from two separate steps, not one doubled figure:

1. **Why not evaluate at the first dispatch — a precondition, not a cost
   argument.** A single site cannot tell you a transformation is templated;
   it could be a one-off. The first implementer dispatch is what reveals the
   shape (the same edit needed again elsewhere) — there is nothing to
   evaluate before that.
2. **Why the second dispatch is the trigger — a cost argument.** Writing a
   codemod is itself exactly the implementer's shape of task — a small,
   self-contained Go module (four or five files: `go.mod`, `analyzer.go`,
   `analyzer_test.go`, `cmd/`, `testdata/`) — so building and testing it once
   is bounded by the *same* 120-tool-call cap as any other `task-implementer`
   dispatch (`docs/agent-dispatch.md` "The implementer budget"). A second
   manual dispatch at the same transformation is bounded by that identical
   cap, because it is the same kind of dispatch — one implementer, one
   120-call ceiling. So the second manual dispatch is the first point that
   both (a) confirms the transformation is templated (step 1) and (b) has not
   yet cost more than the rewriter itself would in the worst case — one
   further manual dispatch already reaches the codemod's own worst-case
   build cost. That is the break-even: evaluate there, not later.

Every site the codemod covers beyond that point is a `--check`-verified
mechanical rewrite instead of another implementer turn.

This repo has not yet run a templated transformation large enough to measure
the actual token cost of ignoring this rule — task-232's 6,231-turn,
760M-token formatter/threading sweep, the incident that motivated this rule
in the sibling repo this process was ported from, has no home-hub analogue
on record. The rule is adopted here on the strength of the structural
argument above (both dispatches share the same 120-call cap, independent of
tokens-per-turn), not on a locally reproduced disaster.

## Worked example: this plan's own linter-cleanup wave (Tasks 17–19)

This plan's own linter-cleanup tasks are the closest available example of a
templated-transformation family, and by the time of this fix wave they have
been executed, so the figures below are the measured outcome, not a
planning-time estimate. The real numbers were larger than planned: **377
files** for the formatting sweep, **106 `errcheck`** sites, and **39** total
findings for the `staticcheck`/`unused`/`ineffassign`/`govet` group — split
across three tasks precisely because they sit at three different points on
the mechanical-to-judgment spectrum this document is about. The evidence is
the commit history, not a git-ignored progress file: `639cec7` (formatting),
`f41a4e7` (errcheck), and `34a8c19` (staticcheck/unused/ineffassign/govet).

- **Task 17 — `gofumpt`/`goimports` formatting sweep across 377 files**
  (`639cec7`). Every finding is a pure syntactic rewrite with one
  deterministic correct output; `tools/verify.sh --fix-fmt` already applies
  exactly this class of fix in place. This is the **canonical codemod
  case** — zero per-site judgment, and the tool to do it (`--fix-fmt`)
  already exists rather than needing to be built. If a transformation this
  templated required 377 separate implementer dispatches instead of one
  scripted pass, it would clear the second-dispatch threshold on the first
  two sites alone.
- **Task 18 — 106 `errcheck` sites, behaviour-preserving** (`f41a4e7`). Each
  site needs the same shape of fix (handle or explicitly discard a
  previously-ignored error) but the *right* handling — return it, log it,
  wrap it with a call-specific message — is a per-site judgment call even
  though the fix is mechanical in structure. This is the **canonical
  borderline case**: AST tooling can find every site and apply a templated
  fix, but a human or reviewing agent still has to confirm each choice was
  appropriate, so a scripted sweep plus a verification pass (Ruling 3) is
  the right shape rather than either a blind codemod or 106 individual
  dispatches.
- **Task 19 — 39 hand-judged `staticcheck`/`unused`/`ineffassign`/`govet`
  findings, including one explicitly authorised behaviour change**
  (`34a8c19`). These are heterogeneous findings across four different
  linters with no single templated edit shape, and at least one site
  requires a real behavioural decision, not just a mechanical fix. This is
  the **canonical agent case** — dispatch `task-implementer` per finding or
  per small batch, because there is no repeated shape for a codemod to
  exploit.

The split mirrors FR-2.2 from the sibling repo this document was ported
from: **rewrite what is derivable, list what is not, and never silently skip
a site.** Task 17 needs no residue list because nothing is left over. Task 18
needs one because the fix shape is templated but the choice per site is not.
Task 19 is residue from the start.

## The deferred rewriter's contract (specification only)

If a templated transformation clears the second-dispatch threshold, the
rewriter that gets written should follow this shape — this is a
specification for future work, not a description of an existing tool.
Nothing under `tools/` currently implements it, and home-hub has no existing
small analyzer module to mirror the layout on (unlike `tools/verify.sh`,
which is a single script, not a Go module).

**Module layout:**

- `tools/<name>/go.mod` — its own module, independent of the service/shared
  modules `go.work` already lists
- `tools/<name>/analyzer.go` — the AST rewrite logic
- `tools/<name>/analyzer_test.go` — table-driven tests over `testdata/`
- `tools/<name>/cmd/` — the CLI entry point
- `tools/<name>/testdata/` — before/after fixture pairs, built from real
  diffs of the transformation it targets

**Two contracts it must honor:**

- **Every site is rewritten or listed, never silently skipped.** A site the
  tool cannot safely rewrite (a judgment step, or any pattern it doesn't
  recognize) goes into a residue report with file:line and reason. Silent
  omission is the failure mode that makes a codemod untrustworthy — a human
  has to be able to trust that "not in the residue list" means "rewritten,"
  not "not looked at."
- **A `--check` mode for use as a guard afterward.** The same analyzer, run
  in a mode that exits non-zero if any un-migrated site remains, becomes the
  regression guard once the migration lands — replacing hand-maintained
  allowlists with a mechanical check. `tools/verify.sh --fix-fmt` already
  demonstrates the shape of this for the one class of finding (formatting)
  that already has first-class tooling.

## Current status — dormant

**No rewriter exists yet** beyond `tools/verify.sh --fix-fmt`'s formatter
pass, which covers Task 17's class of finding but nothing more general. This
document specifies what a general one would look like and the threshold at
which writing one pays for itself; it does not claim one is available to run
for arbitrary transformations. Because no general `--check` mode exists,
no batch beyond formatting can be verified as codemod-applied today — see the
review-step guidance in `.claude/commands/execute-task.md`: until a rewriter
with `--check` lands for a given transformation class, every batch of that
class is treated as judgment-bearing and gets the full per-task review agent.
