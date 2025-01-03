package commit

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/CheeziCrew/swissgit/utils"
	gc "github.com/CheeziCrew/swissgit/utils/gitCommands"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

// CommitChanges adds all changes, checks out a new branch, commits with the provided message, and pushes to the remote repository
func CommitChanges(repoPath, branchName, commitMessage string) error {

	repoName, err := utils.GetRepoName(repoPath)
	if err != nil {
		return fmt.Errorf("could not get repository name: %w", err)
	}

	statusMessage := fmt.Sprintf("%s: committing and pushing", repoName)
	done := make(chan bool)
	go utils.ShowSpinner(statusMessage, done)

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	if err := commitAndPush(repo, repoPath, branchName, commitMessage); err != nil {
		done <- true
		fmt.Printf("\r%s failed [%s]: %s\n", statusMessage, red("x"), red(err.Error()))
		return err
	}

	done <- true
	statusMessage = fmt.Sprintf("%s: committed and pushed to remote", repoName)
	fmt.Printf("\r%s [%s] \n", statusMessage, green("✔"))
	return nil
}

func commitAndPush(repo *git.Repository, repoPath, branchName, commitMessage string) error {

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	// Use git command to add files, respecting .gitignore
	cmd := exec.Command("git", "-C", repoPath, "add", ".")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	gc.PullChanges(worktree)

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
		createNewBranch(branchName, worktree)
	} else if branchName == "" {
		branchName, _ = utils.GetBranchName(repo)
	}

	commit(branchName, commitMessage, repoPath)
	err = gc.PushChanges(repo)
	if err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}
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
