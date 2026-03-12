# Swissgit — Multi-Repo Git Automation

Module: `github.com/CheeziCrew/swissgit`

## What This Is

Multi-repo Git automation tool. TUI by default (no args). CLI mode for scripting (with args).

## Build & Run

```sh
go build -o swissgit .
./swissgit              # TUI mode
./swissgit pr --help    # CLI mode

go test ./...
go vet ./...
```

## Architecture

### Entry Point: main.go

Loads `.env` from the binary directory. Routes to TUI (no args) or CLI (with args).

### TUI: tui/app.go

Root bubbletea model. Screen enum drives routing:

```
Screen (enum) → Update/View → routes to screen logic
```

### CLI: cli/root.go

Cobra command tree. One subcommand per operation:
- `pr` — Create PRs
- `commit` — Automated commits
- `cleanup` — Clean local branches
- `status` — Repo status check
- `branches` — Manage branches
- `clone` — Clone repos
- `automerge` — Auto-merge PRs
- `merge-prs` — Batch merge
- `enable-workflows` — Enable GitHub Actions
- `team-prs` — Team PR summary

### Screen Structure: tui/screens/

One file per screen (13 total). Each follows the **step-enum pattern**:

```go
type Step int
const (
	StepMenu Step = iota
	StepRepoSelect
	StepProgress
	StepResults
)

type Model struct {
	step Step
	// step-specific state
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.step {
	case StepMenu: // handle menu
	case StepRepoSelect: // handle repo selection
	case StepProgress: // handle execution
	case StepResults: // show results
	}
}
```

Screens available:
- `menu.go` — App menu
- `reposelect.go` — Repo picker
- `pullrequest.go`, `commit.go`, `cleanup.go`, etc. — One per operation

### styles.go (Bridge)

Re-exports curd colors and styles for all screens. Import once, use everywhere:

```go
import "github.com/CheeziCrew/swissgit/tui/screens"

// In Update/View:
styles := screens.Styles()
color := styles.AccentColor
```

### Components: tui/components/

Thin wrappers around curd models:

- `progress.go` — Wraps `curd.ProgressModel`, injects `SwissgitPalette`
- `result.go` — Wraps `curd.ResultModel`, injects `SwissgitPalette`

### git/: Low-Level Git Operations

Hybrid approach (intentional):

- **go-git** — Programmatic ops (branch creation, reset, etc.)
- **shell git** — Add/commit (respects `.gitignore`)
- **gh CLI** — GitHub API (PR creation, team queries)

Respecting `.gitignore` is why we shell out for add/commit.

### ops/: High-Level Operations

Each returns a `Result` struct:

- `pullrequest.go` — PR creation logic
- `commit.go` — Commit logic
- `cleanup.go` — Branch cleanup
- etc.

Operations are reusable by both TUI and CLI.

### history.go

Persists recent user inputs to `~/.swissgit/history.json`. Load/save for re-using inputs.

## Adding a New Command

1. **Add operation** in `ops/newcommand.go` (pure function, returns `Result`)
2. **Add TUI screen** in `tui/screens/newcommand.go` (follow step-enum pattern)
3. **Register screen** in `tui/app.go`:
   - Add `Screen` constant
   - Handle in `Update()` and `View()`
   - Add transition logic
4. **Add menu item** in `tui/screens/menu.go`
5. **Optionally add CLI command** in `cli/root.go` (use the operation from step 1)

## Environment

- `GITHUB_TOKEN` — Required for GitHub API calls (gh CLI)
- `SSH_KEY` — SSH key for go-git auth (if needed)
- `.env` — Loaded from binary directory at startup

## Palette

`SwissgitPalette` (magenta/cyan) from curd. Use `styles.Styles()` everywhere.

## Testing

```sh
go test ./...
```

Prefer unit tests for pure functions. Mock `exec.Command` for shell interop:

```go
import "github.com/golang/mock/gomock"

// Mock exec.Command in tests
```

## Common Patterns

### Multi-Step Workflow

TUI screen with step enum:

```go
Menu → RepoSelect → Progress → Results
```

Each step has its own event handling. Transition via `m.step = NextStep`.

### Sending Shell Commands

```go
cmd := exec.Command("git", "add", ".")
output, err := cmd.CombinedOutput()
```

### Using go-git

```go
import "github.com/go-git/go-git/v5"

repo, _ := git.PlainOpen(path)
ref, _ := repo.Head()
```

### GitHub API (via gh CLI)

```go
cmd := exec.Command("gh", "pr", "create", "--body", "...", "--title", "...")
```

Relies on `GITHUB_TOKEN`.

## Troubleshooting

- **go-git SSH fails:** Check `SSH_KEY` env var and key permissions
- **gh CLI fails:** Ensure `GITHUB_TOKEN` is set and valid
- **Shell commands fail:** Check that git/gh are in PATH
- **Missing .env:** Create in binary directory
