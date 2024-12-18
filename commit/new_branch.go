package commit

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func createNewBranch(branchName string, worktree *git.Worktree) error {

	// Checkout a new branch
	newBranchRef := plumbing.NewBranchReferenceName(branchName)
	err := worktree.Checkout(&git.CheckoutOptions{
		Branch: newBranchRef,
		Keep:   true,
		Create: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create and check out branch: %w", err)
	}
	return nil
}
