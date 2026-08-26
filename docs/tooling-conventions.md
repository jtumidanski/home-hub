# Tooling Conventions

This document owns three conventions for how commands and edits are run in
this repo: locating Go module source, waiting on long-running processes, and
general shell/editing hygiene.

## Locating Go module source

Never sweep the filesystem to locate a Go dependency's source. Ask the
toolchain instead:

```sh
go list -m -f '{{.Dir}}' <module>
```

This prints the directory in a fraction of a second, whether the module
resolves to the module cache or to a local `replace` — every service's
`go.mod` in this repo `replace`s `github.com/jtumidanski/home-hub/shared/go/<name>`
with `../../shared/go/<name>`. The same applies to `go doc <pkg>` for a
symbol and `go list -m all` for the version set.

`find /` takes minutes on a large filesystem and is the wrong tool here in
any case: a `replace`d module's real source is a `go.mod` fact, not
something a filesystem sweep can discover reliably — `find` would need to
guess the target path, and guessing at it is the tell that you should have
asked `go list` instead. `find` is for paths you own, rooted at a directory
you name — never at `/`.

## Waiting on processes

Never spend inference turns waiting for a process. Launch it once with a
bound — `run_in_background: true`, or `Monitor` with an until-loop — and do
something else or hand back.

Repeated `sleep` / `ps aux | grep` / `echo waiting` / `for i in $(seq …); do
sleep` calls are the anti-pattern: each one re-reads the whole context to
learn nothing, and they cluster late in a session where that is most
expensive. If the process exceeds its bound, kill it and fall back; do not
keep polling. When a tool has a known hang mode, the fallback belongs in
that tool's agent doc, not in a longer wait.

The same holds for **waiting on a child agent**. There is no wait primitive
because none is needed: completions arrive as notifications, so do other work
or end the turn and be re-invoked. Emitting `true` to stay alive is the worst
version of this — measured at 30 such calls inside one agent, 36% of its entire
cost, for zero information.

`.claude/hooks/wait-loop-guard.sh` makes this machine-checked rather than
advisory, the way `.claude/hooks/fork-dispatch-guard.sh` did for forks. It
refuses bare no-ops, sleep-driven polls, and broad `ps`/`pgrep` sweeps. It
deliberately allows real process debugging — `ps -p <pid>`, `kill`/`pkill`,
`docker ps`, `top -b -n1` — and anything prefixed `POLL-JUSTIFIED: <reason>`,
mirroring `FORK-JUSTIFIED:`. A considered wait costs one sentence; the
reflexive one is blocked.

## Ask for a fact rather than deriving it

Mechanical repository facts have deterministic sources. Use them:

| Question | Ask |
|---|---|
| Which task-NNN numbers are already assigned, and where | `tools/task-numbers.sh list` |
| The next unused task number | `tools/task-numbers.sh next` |
| Whether any task-NNN number collides across sources | `tools/task-numbers.sh check` |
| What will the gate build, and why is it fanning out | `tools/verify.sh --facts --quick --base <sha>` |
| Which services exist | `go.work`'s `use` block |
| Which images CI (and the docker leg) build | `.github/workflows/pr.yml`'s matrix, mirrored by `DOCKER_IMAGES` in `tools/verify.sh` |
| The toolchain pin | `tools/toolchain.versions` |
| One task out of a plan | `tools/task-brief.sh <plan> <N>` |
| One section or a few lines of a large document | `grep -n` / `sed -n` slicing — see [slice-first.md](slice-first.md) |

Do not probe for toolchain availability (`command -v`, `--version`, `which`)
when a script already reports it — e.g. `tools/verify.sh`'s `pins` leg and
`--facts` output state the resolved Go version directly. Asking the
environment to state a fact beats deriving it by hand.

**A deterministic tool defeated by a wrapper is a net loss.** When a
token-optimizing shell wrapper swallows a script's stdout, the saved bytes cost
a whole extra turn to recover the output. Any such wrapper must pass `tools/*.sh`
output through unfiltered — `tools/verify.sh --facts` and `tools/task-numbers.sh`
already emit compact, line-oriented output and have nothing to trim.

## Shell and editing conventions

Prefer portable POSIX shell; avoid zsh/direnv-specific constructs and batch
patch loops that can produce garbled or unapplied output. For a multi-file
edit, prefer per-file Edit/Write over a shell patch loop.

Quote glob arguments in shell tool calls — `--include='*.go'`, not
`--include=*.go` — zsh expands an unquoted glob before `grep` sees it,
producing `no matches found` and a wasted retry.

Preserve line endings when editing — do not normalize CRLF→LF as a side
effect; it inflates diffs with spurious changes.

Always use repo-relative paths or placeholders in committed files; never
literal home or absolute paths like `/Users/<name>/...` or
`/home/<name>/...` — a committed absolute path is not reproducible on
another machine. `.claude/hooks/block-home-paths-in-docs.sh` machine-checks
this for every write under `docs/`.
