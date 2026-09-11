# Git Workflow

This document owns the mechanics of branch safety, pushing after history
rewrites, what triggers a PR build, and `gh` authentication in this repo.

## Branch safety

Never commit or push directly to `main`. Branch protection blocks the push,
so a commit made on local `main` is stranded and never reaches the remote.
Check the branch before every `git commit`.

Setup work that must precede a feature branch still goes *on* the feature
branch — create it first; it branches from the same HEAD.

Recovery from a stray `main` commit: preserve the content on a branch
(cherry-pick if needed), then:

```sh
git fetch origin main && git reset --hard origin/main
```

## Worktree and branch naming

Task work lives in a git worktree at `.worktrees/task-NNN-slug/` on a branch
named `task-NNN-slug`, created by `/spec-task`. All four workflow phases
(`/spec-task` → `/design-task` → `/plan-task` → `/execute-task`) run inside
that same worktree so the PRD, design, plan, and code land as one unit. PRs
target `main`.

## Pushing and history rewrites

After completing a rebase/merge/history-rewrite, always push (force-push
when history was rewritten) so the PR reflects the resolved state. Do not
stop at local-only completion — a rebase resolved only locally leaves the PR
still showing conflicts.

## Build triggering and the conflict exception

A plain push to a task branch DOES trigger PR Validation
(`.github/workflows/pr.yml`). Do not merge `origin/main` as a routine
build-triggering ritual.

The one exception: when the branch conflicts with `main`, merge
`origin/main`, resolve, and push the merge commit — the merge is the
conflict resolution, not the trigger.

## `gh` authentication

Confirm the account `gh` will use before relying on it:

```sh
gh auth status
```

If a stale or unexpected token is in effect, clear the environment overrides
explicitly rather than guessing which credential source wins:

```sh
env -u GH_TOKEN -u GITHUB_TOKEN gh …
```

Never echo a token to stdout or into a committed file.
