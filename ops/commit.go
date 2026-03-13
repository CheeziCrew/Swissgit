package ops

import (
	"fmt"

	"github.com/CheeziCrew/swissgit/git"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// CommitResult holds the outcome of a commit+push operation.
type CommitResult struct {
	RepoName string
	Branch   string
	Success  bool
	Error    string
}

// CommitAndPush stages, commits, and pushes changes for a single repo.
// Uses shell git add (to respect .gitignore) and shell git commit,
// then go-git for push.
func CommitAndPush(repoPath, branchName, commitMessage string) CommitResult {
	repo, err := plainOpen(repoPath)
	if err != nil {
		return CommitResult{Error: fmt.Sprintf("failed to open repository: %s", err)}
	}

	repoName, _ := git.GetRepoNameFromRepo(repo)
	result := CommitResult{RepoName: repoName}

	worktree, err := repo.Worktree()
	if err != nil {
		result.Error = fmt.Sprintf("failed to get worktree: %s", err)
		return result
	}

	// Create/checkout branch before staging
	if branchName != "" {
		if err := createOrCheckoutBranch(branchName, worktree); err != nil {
			result.Error = fmt.Sprintf("failed to switch to branch: %s", err)
			return result
		}
	} else {
		branchName, _ = git.GetBranchName(repo)
	}
	result.Branch = branchName

	// shell git to respect .gitignore
	if _, err := gitRunInDir(repoPath, "-C", repoPath, "add", "."); err != nil {
		result.Error = fmt.Sprintf("failed to add files: %s", err)
		return result
	}

	status, err := worktree.Status()
	if err != nil {
		result.Error = fmt.Sprintf("failed to get status: %s", err)
		return result
	}

	changes := git.CountChanges(status)
	if !changes.HasChanges() {
		result.Error = "no changes to commit"
		return result
	}

	fullMessage := fmt.Sprintf("%s: %s", branchName, commitMessage)
	if _, err := gitRunInDir(repoPath, "-C", repoPath, "commit", "-m", fullMessage); err != nil {
		result.Error = fmt.Sprintf("failed to commit changes: %s", err)
		return result
	}

	if err := pushChanges(repo); err != nil {
		result.Error = fmt.Sprintf("failed to push changes: %s", err)
		return result
	}

	result.Success = true
	return result
}

func createOrCheckoutBranch(branchName string, worktree *gogit.Worktree) error {
	ref := plumbing.NewBranchReferenceName(branchName)

	err := worktree.Checkout(&gogit.CheckoutOptions{
		Branch: ref,
		Keep:   true,
		Create: true,
	})
	if err == nil {
		return nil
	}

	err = worktree.Checkout(&gogit.CheckoutOptions{
		Branch: ref,
		Keep:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to check out branch %s: %w", branchName, err)
	}
	return nil
}
