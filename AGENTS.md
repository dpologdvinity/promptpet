# AGENTS.md

Instructions for AI coding agents (Claude Code, Codex, etc.) working in this
repository. `CLAUDE.md` imports this file, so this is the single source of truth.

## What this is

`promptpet` is a single-binary Go CLI that puts a small pet in your shell
prompt (zsh, bash, fish) and an optional Ebitengine desktop companion window
(`promptpet gui`). User-facing docs: [README.md](README.md). Package boundaries
and data flow: [ARCHITECTURE.md](ARCHITECTURE.md).

## Setup and commands

Go version comes from `go.mod` (`go 1.26`). Shell-hook tests need `zsh` and
`fish` installed; tests for a missing shell are skipped.

```sh
go build -o bin/promptpet .                                # build (bin/ is gitignored)
go test ./...                                              # all tests
go test -race ./...                                        # what CI runs (plus coverage)
go vet ./...                                               # static analysis
test -z "$(gofmt -s -l .)"                                 # format check (CI fails on diff)
go test -run '^$' -bench BenchmarkPrompt ./internal/cli/   # prompt budget benchmark
go test -run '^$' -bench . ./test/...                      # prompt + event-logging benchmark
```

Git hooks run through [lefthook](https://lefthook.dev) (`lefthook install`):
`gitleaks` on staged changes, `gofmt` on staged Go files and `typos`.

## Hard constraints

- **Prompt path stays under 10 ms.** `promptpet prompt` must not spawn `git`
  or touch the network. Repo detection uses `os.Stat` only (`internal/repo`).
  All git work happens in the detached `promptpet _feed` process (`internal/feed`).
- **Hooks must preserve the user's exit status** (`$?` / `$status`) and must be
  idempotent when an rc file is sourced twice. `internal/hooks` tests check both.
- **The prompt segment must be shell-safe.** Never emit `%`, `$`, `` ` ``, `\`,
  `"` or `'` in it (`internal/pet` rendering).
- **State is local only.** Dir `0700`, files `0600`, atomic writes and `flock`
  (`internal/store`). No telemetry.
- `internal/pet` is pure: clock and RNG are injected, so keep it free of I/O.

## Conventions

- Conventional Commits (`feat(scope): ...`, `fix:`, `docs:`, `test:`,
  `refactor:`, `perf:`, `chore:`, `ci:`), small and atomic.
- Bugs get a failing test first, then the fix.
- Before you call work done, run `go test ./...`, `go vet ./...` and the gofmt check.
- Default branch is `master`. Releases are tag-driven: pushing `v*` runs
  `.github/workflows/release.yml` (GoReleaser, then an optional Homebrew tap
  update that only runs when the `HOMEBREW_TAP_GITHUB_TOKEN` secret is set).
- No env vars are needed for development. Runtime overrides (`PROMPTPET_HOME`,
  etc.) are listed in the README.

## Agent workflow

This project was built spec-first with the
[superpowers](https://github.com/obra/superpowers) skills for Claude Code:

1. **Brainstorm → spec.** Specs live in [`docs/agents/specs/`](docs/agents/specs/).
   When a spec and a plan disagree, the spec wins.
2. **Spec → plan.** Plans in [`docs/agents/plans/`](docs/agents/plans/) are
   bite-sized TDD tasks with checkboxes, exact file paths and commit messages.
3. **Execute.** Plans run with subagent-driven development (one fresh subagent
   per task, with review between tasks) or inline plan execution.

Add new specs and plans to those directories with a `YYYY-MM-DD-topic.md` name.

Project-scoped Claude Code config lives in `.claude/settings.json` (it enables
the `gopls-lsp` and `commit-commands` plugins). Personal overrides go in
`.claude/settings.local.json` or `CLAUDE.local.md`, both gitignored.
