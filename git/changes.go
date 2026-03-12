package git

import (
	"os/exec"
	"strings"

	gogit "github.com/go-git/go-git/v5"
)

// Changes holds counts of different change types.
type Changes struct {
	Modified  int
	Added     int
	Deleted   int
	Untracked int
}

// HasChanges returns true if any changes exist.
func (c Changes) HasChanges() bool {
	return c.Modified > 0 || c.Added > 0 || c.Deleted > 0 || c.Untracked > 0
}

// Total returns the total number of changes.
func (c Changes) Total() int {
	return c.Modified + c.Added + c.Deleted + c.Untracked
}

// CountChanges counts changes using go-git's worktree status.
func CountChanges(status gogit.Status) Changes {
	var c Changes
	for _, state := range status {
		if state.Worktree == gogit.Untracked {
			c.Untracked++
		}
		if state.Staging == gogit.Modified || state.Worktree == gogit.Modified {
			c.Modified++
		}
		if state.Staging == gogit.Added || state.Worktree == gogit.Added {
			c.Added++
		}
		if state.Staging == gogit.Deleted || state.Worktree == gogit.Deleted {
			c.Deleted++
		}
	}
	return c
}

// CountChangesShell counts changes using shell git to respect .gitignore.
func CountChangesShell(repoPath string) (Changes, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return Changes{}, err
	}

	var c Changes
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		classifyStatusLine(line[0], line[1], &c)
	}
	return c, nil
}

// classifyStatusLine classifies a single porcelain status line by its X and Y indicators.
func classifyStatusLine(x, y byte, c *Changes) {
	if x == '?' || y == '?' {
		c.Untracked++
	}
	if x == 'M' || y == 'M' {
		c.Modified++
	}
	if x == 'A' || y == 'A' || x == 'R' || y == 'R' || x == 'C' || y == 'C' {
		c.Added++
	}
	if x == 'D' || y == 'D' {
		c.Deleted++
	}
}
