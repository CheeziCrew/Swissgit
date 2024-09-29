package commit

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/CheeziCrew/swissgo/utils"
	gc "github.com/CheeziCrew/swissgo/utils/gitCommands"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

// CommitChanges adds all changes, checks out a new branch, commits with the provided message, and pushes to the remote repository
func CommitChanges(repoPath, branchName, commitMessage string) error {

	repoName, err := utils.GetRepoName(repoPath)
	if err != nil {
		return fmt.Errorf("could not get repository name: %w", err)
	}

	statusMessage := fmt.Sprintf("%s: committing and pushing", repoName)
	done := make(chan bool)
	go utils.ShowSpinner(statusMessage, done)

	if err := commitAndPush(repoPath, branchName, commitMessage); err != nil {
		done <- true
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf("\r%s failed [%s]: %s", statusMessage, red("x"), red(err.Error()))
		return err
	}

	// Success
	done <- true
	green := color.New(color.FgGreen).SprintFunc()
	statusMessage = fmt.Sprintf("%s: committed and pushed to remote", repoName)
	fmt.Printf("\r%s [%s]", statusMessage, green("✔"))
	return nil
}

func commitAndPush(repoPath, branchName, commitMessage string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Add all changes (equivalent to `git add .`)
	if err := worktree.AddGlob("."); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status for repository: %w", err)
	}

	modified, added, deleted, untracked := utils.CountChanges(status)
	if modified == 0 && added == 0 && deleted == 0 && untracked == 0 {
		return fmt.Errorf("no changes to commit")
	}

	// Create a new branch
	if branchName != "" {
		createNewBranch(repoPath, branchName, worktree)
	}

	commit(branchName, commitMessage, repoPath)
	gc.PushChanges(repo)
	return nil
}

func commit(branchName, commitMessage, repoPath string) error {
	fullCommitMessage := fmt.Sprintf("%s: %s", branchName, commitMessage)

	cmd := exec.Command("git", "-C", repoPath, "commit", "-m", fullCommitMessage)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	return nil
}
