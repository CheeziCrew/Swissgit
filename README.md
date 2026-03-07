# SwissGit

![Untitled_Artwork_15](https://github.com/CheeziCrew/Swissgit/assets/110965999/0edfe55f-38a2-4d06-9c39-5b60ff7f5441)

Multi-repo Git workflows without the pain. Interactive TUI or fast CLI — your call.

## Quick Start

Grab a binary from [Releases](https://github.com/CheeziCrew/swissgit/releases), or build it yourself:

```sh
go install github.com/CheeziCrew/swissgit@latest
```

Drop a `.env` next to the binary:

```env
GITHUB_TOKEN=ghp_...
SSH_KEY=id_ed25519
```

Run `swissgit` for the TUI, or pass a command for CLI mode.

## TUI

```sh
swissgit
```

Full interactive terminal UI — navigate commands, pick repos, fill forms, watch progress. Adapts to your terminal's color theme (base16).

## CLI

Every command works with `--all` / `-a` to hit all repos in subdirectories.

```sh
swissgit status -a                      # What's dirty?
swissgit branches -a                    # What branches exist?
swissgit commit -m "fix stuff" -a       # Stage, commit, push everywhere
swissgit pr -m "Title" -b feat-1 -a     # Create PRs across repos
swissgit clone -o MyOrg -t my-team      # Clone org (skips archived repos)
swissgit cleanup -a -d                  # Reset, pull main, prune branches
swissgit automerge -t "PR title" -a     # Enable auto-merge (needs gh CLI)
```

## Prerequisites

| What | Why | Required |
|---|---|---|
| Git | duh | Yes |
| SSH key on GitHub | Clone/push over SSH | Yes |
| `GITHUB_TOKEN` | PR creation, org clone, GitHub API | Yes |
| [`gh` CLI](https://cli.github.com/) | Automerge only | Only for `automerge` |

The `.env` file is loaded from the binary's directory, not your CWD. If using `go run .`, export the vars in your shell instead.

## Acknowledgements

**Theo the Cat** — moral support department, head of naps division.

## License

[MIT](LICENSE)
