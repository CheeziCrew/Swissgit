package utils

import (
	"github.com/go-git/go-git/v5"
)

// countChanges counts the number of modified, added, deleted, and untracked files in the repository.
func CountChanges(status git.Status) (int, int, int, int) {
	var modified, added, deleted, untracked int
	for _, state := range status {
		if state.Worktree == git.Untracked {
			untracked++
		}
		if state.Staging == git.Modified || state.Worktree == git.Modified {
			modified++
		}
		if state.Staging == git.Added || state.Worktree == git.Added {
			added++
		}
		if state.Staging == git.Deleted || state.Worktree == git.Deleted {
			deleted++
		}
	}
	return modified, added, deleted, untracked
}
