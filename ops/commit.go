package ops

import (
	"fmt"
	"io"
	"os/exec"

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
	repo, err := gogit.PlainOpen(repoPath)
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

	// Shell git add . (respects .gitignore)
	cmd := exec.Command("git", "-C", repoPath, "add", ".")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		result.Error = fmt.Sprintf("failed to add files: %s", err)
		return result
	}

	// Pull latest
	if err := git.PullChanges(worktree); err != nil {
		result.Error = fmt.Sprintf("failed to pull changes: %s", err)
		return result
	}

	// Check for changes
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

	// Create/checkout branch
	if branchName != "" {
		if err := createOrCheckoutBranch(branchName, worktree); err != nil {
			result.Error = fmt.Sprintf("failed to switch to branch: %s", err)
			return result
		}
	} else {
		branchName, _ = git.GetBranchName(repo)
	}
	result.Branch = branchName

	// Shell git commit
	fullMessage := fmt.Sprintf("%s: %s", branchName, commitMessage)
	commitCmd := exec.Command("git", "-C", repoPath, "commit", "-m", fullMessage)
	commitCmd.Stdout = io.Discard
	commitCmd.Stderr = io.Discard
	if err := commitCmd.Run(); err != nil {
		result.Error = fmt.Sprintf("failed to commit changes: %s", err)
		return result
	}

	// go-git push
	if err := git.PushChanges(repo); err != nil {
		result.Error = fmt.Sprintf("failed to push changes: %s", err)
		return result
	}

	result.Success = true
	return result
}

func createOrCheckoutBranch(branchName string, worktree *gogit.Worktree) error {
	ref := plumbing.NewBranchReferenceName(branchName)

	// Try create+checkout
	err := worktree.Checkout(&gogit.CheckoutOptions{
		Branch: ref,
		Keep:   true,
		Create: true,
	})
	if err == nil {
		return nil
	}

	// Branch exists — just switch
	err = worktree.Checkout(&gogit.CheckoutOptions{
		Branch: ref,
		Keep:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to check out branch %s: %w", branchName, err)
	}
	return nil
}
