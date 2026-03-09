# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SwissGit is a Go CLI tool (Cobra) that automates common multi-repository Git workflows: status checks, cloning org repos, committing & pushing, creating PRs, branch listings, cleanup, and enabling auto-merge.

Module: `github.com/CheeziCrew/swissgit` — requires Go >= 1.23, toolchain go1.24.1.

## Build & Development Commands

```bash
go build                              # Build binary in current directory
go install                            # Install to $GOBIN/$GOPATH/bin
go vet ./...                          # Lint/typecheck
go test ./...                         # Run all tests
go test ./status -run TestName -v     # Run a single test
./swissgit --help                     # Verify CLI wiring
```

Releases are automated via GoReleaser v2 (`.github/workflows/release.yml`), triggered on git tags.

## Architecture

**Entry point:** `main.go` — registers Cobra commands and loads `.env` from the executable's directory (not CWD) via `godotenv`.

**Package-per-command structure:**
- `status/` — repo status with ahead/behind counting and colored output
- `branches/` — local, remote, and stale branch listing (>120 days)
- `clone/` — SSH-authenticated cloning (single repo or entire GitHub org)
- `commit/` — stage, commit, and push workflow
- `pull_request/` — PR creation via GitHub REST API with embedded template
- `cleanup/` — reset changes, update main, prune merged branches
- `automerge/` — enable auto-merge via `gh` CLI

**Shared utilities in `utils/`:**
- `gitCommands/` — fetch/pull/push wrappers using go-git + `SshAuth()`
- `ssh_auth.go` — reads `~/.ssh/$SSH_KEY`
- `spinner.go` — goroutine-driven spinner with `done chan bool` signaling
- `is_git_repo.go`, `repo_owner_name.go`, `branch_name.go`, `changes.go`, `validation/`

## Key Design Decisions

**Two Git approaches coexist intentionally:**
- **go-git** (`go-git/v5`) for most programmatic operations (fetch, pull, push, repo metadata)
- **Shell `git`** for `git add .` (to respect `.gitignore`) and `git commit` in `commit/commit.go`
- **`gh` CLI** for automerge functionality in `automerge/automerge.go`

**Subdirectory scanning:** Commands with `--all`/`-a` iterate exactly one level deep for `.git` directories. Do not assume deeper recursion.

**Spinner pattern:** `utils.ShowSpinner(message, done)` runs in a goroutine. Always signal `done <- true` to avoid goroutine leaks.

**Required flags:** `commit` requires `--message`; `pullrequest` requires `--message` and `--branch`.

## Environment & Auth

- `GITHUB_TOKEN` — required for PR creation and GitHub API calls
- `SSH_KEY` — private key filename under `~/.ssh/` used by go-git
- `.env` is loaded from the directory of the executable at runtime, not CWD. With `go run .`, env vars must be exported in the shell instead.
- `automerge` requires `gh` CLI installed and authenticated (`gh auth login`)

## Testing

- Disable color for deterministic output: `color.NoColor = true` or `NO_COLOR=1`
- Prefer unit tests for pure functions (formatting, parsing in `utils/`)
- Place tests in the same package to access unexported functions
- Avoid end-to-end tests requiring network/filesystem; use dependency injection for mocking
- Wrap `exec.Command` calls in helpers to enable mocking when testing shell interop

## Safe Modification Checklist

- Preserve SSH auth via `SSH_KEY` and `.env` loading in `main.init()`
- Preserve shell `git add` in commit workflow (bypassing go-git to respect `.gitignore`)
- Keep one-level subdirectory scanning behavior unless intentionally changing CLI semantics
- Keep CLI flags backward-compatible; update help text alongside any flag changes
- Reuse `GITHUB_TOKEN` from env for new network calls; never hard-code tokens
