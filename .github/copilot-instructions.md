<!-- .github/copilot-instructions.md for Swissgit -->
# Copilot / AI agent instructions — Swissgit

These notes are a brief, actionable reference to help an AI coding agent be productive in this repository.
Keep suggestions anchored to the code and avoid proposing undetectable changes (for example: CI secrets, external API tokens) without explicit user direction.

What this project is
- Swissgit is a Go CLI (Cobra) that automates common multi-repository Git workflows: status checks, cloning org repos, committing & pushing, creating PRs, branch listings, cleanup and enabling auto-merge.
- Main entrypoint: `main.go` which wires Cobra commands to packages: `status/`, `branches/`, `clone/`, `commit/`, `pull_request/`, `cleanup/`, `automerge/`.

Key architectural points
- Language & tooling: Go (module `github.com/CheeziCrew/swissgit`, Go >= 1.23). See `go.mod`.
- Two Git approaches in the codebase:
  - Programmatic git operations use `github.com/go-git/go-git/v5` (fetch, pull, push, reading repo metadata). See `utils/gitCommands/*.go`, `clone`, `commit`.
  - Shell/CLI interop for actions not implemented with go-git: `git` shell commands for `git add`/`commit` in `commit/commit.go` and `gh` CLI usage in `automerge/automerge.go` to enable auto-merge.
- Authentication: SSH keys + GitHub token. `utils/ssh_auth.go` expects $SSH_KEY (file name, e.g. `id_ed25519`) in the user's home `.ssh` folder. `pull_request` and other HTTP calls require `GITHUB_TOKEN` environment variable (read from `.env` if present; `main.init()` loads `.env` next to the executable via `github.com/joho/godotenv`).

Developer workflows and commands
- Build locally (from repo root): `go build` (or `go install` to place binary in GOPATH/bin). The `README.md` documents this.
- Run help to explore commands: `./swissgit --help` (or after install: `swissgit --help`).
- Environment setup:
  - Create a `.env` in the executable directory (or set env vars in your shell):
    - `GITHUB_TOKEN` — required for creating PRs via the GitHub API.
    - `SSH_KEY` — file name of your SSH private key under `~/.ssh/` used by go-git.
- PR & automerge: the `automerge` command uses the GitHub CLI `gh`. Tests or changes that touch `automerge` should account for `gh` being present, or mock `exec.Command` calls.

Project-specific conventions and patterns
- Subdirectory scanning: many commands support an `--all`/`-a` flag which causes the command to iterate one level of subdirectories and operate on any subdirectory that contains a `.git` folder (see `utils/is_git_repo.go` and `pull_request.ProcessSubdirectories`). Do not assume recursion beyond one level.
- Output UX: most long-running operations show a spinner with `utils.ShowSpinner` which is driven by a `done chan bool` signal. When modifying these flows, ensure the spinner channel is closed/signalled reliably to avoid orphaned goroutines or hanging output.
- Error handling style: functions typically return an error that is printed by the caller. Some areas swallow errors intentionally (e.g., in `ProcessSubdirectories` errors are logged but not returned). Follow existing style when editing.
- Git add/commit strategy: `commit` uses `exec.Command("git", "-C", repoPath, "add", ".")` to respect `.gitignore` and then uses go-git for other operations. When changing add/commit logic, preserve the shell command to avoid inadvertently bypassing .gitignore handling.

Integration points & external dependencies to be aware of
- go-git (programmatic Git operations): `utils/gitCommands/*.go`, `clone/clone.go`, `commit/commit.go`.
- GitHub API vs `gh` CLI:
  - Creating PRs uses the REST API directly in `pull_request/CreatePullRequest` and requires `GITHUB_TOKEN`.
  - Automerge uses the `gh` CLI (`exec.Command("gh", ...)`) — tests or CI must install/ mock `gh` or adapt to using the API.
- Local SSH key: `utils.SshAuth()` reads `~/.ssh/$SSH_KEY`. Make sure code changes preserve that lookup or update README accordingly.

Files to inspect for concrete examples
- `main.go` — wiring of cobra commands and flags.
- `pull_request/pull_request.go` — building PR bodies and HTTP POST to GitHub.
- `commit/commit.go` — how commits, `git add`, and push are performed (mix of shell + go-git).
- `utils/gitCommands/*` — fetch/pull/push wrappers using go-git and `SshAuth()`.
- `automerge/automerge.go` — `gh` CLI usage and expectations.

Safe modification checklist (when editing behavior)
- Preserve how SSH auth picks up `SSH_KEY` and how `.env` is loaded in `main.init()`.
- If adding network calls, reuse `GITHUB_TOKEN` from env; do not hard-code tokens.
- When modifying subdirectory scanning, keep the one-level behavior unless explicitly changing the CLI semantics and updating help text.
- Keep CLI UX backward-compatible: commands, flags and required flags (`commit` and `pullrequest` mark `message` and `branch` required) should remain stable or be updated together with usage/help text.

Testing notes for AI suggested changes
- Unit tests are not present in the repo. For small edits add focused tests that exercise utilities (for example, parse functions in `utils/repo_owner_name.go`) and any changeable logic.
- When changing code that runs external commands (`git`, `gh`) prefer to wrap `exec.Command` in a small helper that can be mocked in tests.

When you are unsure
- Ask the user whether you may modify behavior that requires secrets (GITHUB_TOKEN) or local environment changes (SSH keys, installing `gh`).
- Prefer small, incremental changes and open a PR with a clear description referencing this guidance.

If you make edits, run these quick local checks
- Build: `go build` (root)
- Lint/typecheck: `go vet ./...` and `go list` (no separate linter configured in repo)
- Basic run of CLI help: `./swissgit --help` or `go run ./main.go --help`

Ask me for feedback if any part of the repository behavior is unclear (for example: exactly where `.env` should live when building with `go install`).
