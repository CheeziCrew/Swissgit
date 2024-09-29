package status

import (
	"fmt"
	"path/filepath"

	"github.com/CheeziCrew/swissgit/utils"
	gc "github.com/CheeziCrew/swissgit/utils/gitCommands"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

// CheckStatus checks the status of a Git repository at the given path and prints a summarized status
func CheckStatus(repoPath string, verbose bool) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	statusMessage := fmt.Sprintf("%s: Updating status", repoPath)
	done := make(chan bool)
	go utils.ShowSpinner(statusMessage, done)

	// Open the existing repository at the specified path
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		errMsg := fmt.Sprintf("%s: Could not open repository:", filepath.Base(repoPath))
		fmt.Printf("\r%s [%s]: %s\n", errMsg, red("x"), err)
		done <- true
		return
	}

	gc.FetchRemote(repo)

	// Get the branch name
	branch, err := getBranchName(repo)
	if err != nil {
		errMsg := fmt.Sprintf("%s: Could not get branch name:", filepath.Base(repoPath))
		fmt.Printf("\r%s [%s]: %s\n", errMsg, red("x"), err)
		done <- true
		return
	}

	// Get the head reference
	headRef, err := repo.Head()
	if err != nil {
		errMsg := fmt.Sprintf("%s: Could not get head", filepath.Base(repoPath))
		fmt.Printf("\r%s [%s]: %s\n", errMsg, red("x"), err)
		done <- true
		return
	}

	// Get the working directory of the repository
	worktree, err := repo.Worktree()
	if err != nil {
		errMsg := fmt.Sprintf("%s: Could not get worktree", filepath.Base(repoPath))
		fmt.Printf("\r%s [%s]: %s\n", errMsg, red("x"), err)
		done <- true
		return
	}

	// Get the status of the repository
	status, err := worktree.Status()
	if err != nil {
		errMsg := fmt.Sprintf("%s: Could not get status of repo", filepath.Base(repoPath))
		fmt.Printf("\r%s [%s]: %s\n", errMsg, red("x"), err)
		done <- true
		return
	}

	// Count different types of changes
	modified, added, deleted, untracked := utils.CountChanges(status)

	// Determine ahead/behind count
	ahead, behind := calculateAheadBehind(repo, headRef, branch)

	// Determine if we should skip output based on verbose flag
	isClean := (modified == 0 && added == 0 && deleted == 0 && untracked == 0 && ahead == 0 && behind == 0)
	if !verbose && isClean && branch == "main" {
		done <- true
		return

	}

	statusLine := buildStatusLine(branch, ahead, behind, modified, added, deleted, untracked)
	done <- true
	statusMessage = fmt.Sprintf("%s: Updated status:", filepath.Base(repoPath))
	fmt.Printf("\r%s [%s] %s\n", statusMessage, green("✔"), statusLine)
}
