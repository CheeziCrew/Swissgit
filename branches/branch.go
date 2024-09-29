package branches

import (
	"fmt"
	"path/filepath"

	"github.com/CheeziCrew/swissgit/utils"
	gc "github.com/CheeziCrew/swissgit/utils/gitCommands"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

// PrintBranches prints information about the main branch, local branches, and remote branches in the repository.
func PrintBranches(repoPath string, verbose bool) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	statusMessage := fmt.Sprintf("%s: Updating branches", repoPath)
	done := make(chan bool)
	go utils.ShowSpinner(statusMessage, done)

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		err := fmt.Sprintf("%s: Could not open repository:", filepath.Base(repoPath))
		fmt.Printf("\r%s [%s]: %s\n", err, red("x"), err)
		done <- true
		return
	}

	// Update the remote references
	if err := gc.FetchRemote(repo); err != nil {
		err := fmt.Sprintf("%s: could not fetch remote:", filepath.Base(repoPath))
		fmt.Printf("\r%s [%s]: %s\n", err, red("x"), err)
		done <- true
		return
	}

	branch, err := utils.GetBranchName(repo)
	if err != nil {
		err := fmt.Sprintf("%s: could not get branchName:", filepath.Base(repoPath))
		fmt.Printf("\r%s [%s]: %s\n", err, red("x"), err)
		done <- true
		return
	}

	// Get local branches
	localBranches, err := listBranches(repo, false)
	if err != nil {
		err := fmt.Sprintf("%s: could not list local branches:", filepath.Base(repoPath))
		fmt.Printf("\r%s [%s]: %s\n", err, red("x"), err)
		done <- true
		return
	}

	// Get remote branches
	remoteBranches, err := listBranches(repo, true)
	if err != nil {
		err := fmt.Sprintf("%s: could not list remote branches:", filepath.Base(repoPath))
		fmt.Printf("\r%s [%s]: %s\n", err, red("x"), err)
		done <- true
		return
	}

	isClean := (len(localBranches) == 1 && len(remoteBranches) == 0)
	if !verbose && isClean && branch == "main" {
		done <- true
		return
	}

	// Build and print the status line
	statusLine := buildStatus(branch, localBranches, remoteBranches)
	done <- true
	statusMessage = fmt.Sprintf("%s: Updated branches", filepath.Base(repoPath))
	fmt.Printf("\r%s [%s]: %s\n", statusMessage, green("✔"), statusLine)
}
