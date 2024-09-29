package commit

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func createNewBranch(repoPath string, branchName string, worktree *git.Worktree) error {
	// Stash changes
	cmd := exec.Command("git", "-C", repoPath, "stash", "push", "-u", "-m", "Temporary stash before branch switch")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stash changes: %w", err)
	}

	// Checkout a new branch
	newBranchRef := plumbing.NewBranchReferenceName(branchName)
	err := worktree.Checkout(&git.CheckoutOptions{
		Branch: newBranchRef,
		Create: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create and check out branch: %w", err)
	}

	// Apply the stash
	cmd = exec.Command("git", "-C", repoPath, "stash", "pop")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply stash: %w", err)
	}
	return nil
}
