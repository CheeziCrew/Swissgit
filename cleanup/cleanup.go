package cleanup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/CheeziCrew/swissgit/utils"
	"github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// CleanupOptions holds the configuration for the cleanup process
type CleanupOptions struct {
	RepoPath    string
	AllFlag     bool
	DropChanges bool
}

// Cleanup performs the cleanup process on repositories in the specified directory
func Cleanup(opts CleanupOptions) error {
	if opts.AllFlag {
		return ProcessSubdirectories(opts)
	} else {
		return ProcessSingleRepository(opts)
	}
}

func ProcessSubdirectories(opts CleanupOptions) error {
	entries, err := os.ReadDir(opts.RepoPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subRepoPath := filepath.Join(opts.RepoPath, entry.Name())

			// Check if the subdirectory is a Git repository
			if !utils.IsGitRepository(subRepoPath) {
				continue
			}

			err := ProcessSingleRepository(CleanupOptions{RepoPath: subRepoPath, AllFlag: false, DropChanges: opts.DropChanges})
			if err != nil {
				// Log the error but continue processing other repositories
				fmt.Printf("Error processing repository %s: %s\n", subRepoPath, err)
			}
		}
	}
	return nil
}

// ProcessSingleRepository processes a single Git repository
func ProcessSingleRepository(opts CleanupOptions) error {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	repoName, err := utils.GetRepoName(opts.RepoPath)
	if err != nil {
		statusMessage := fmt.Sprintf("%s: getting repoName", opts.RepoPath)
		fmt.Printf("\r%s failed [%s]: %s \n", statusMessage, red("x"), red(err.Error()))
	}

	statusMessage := fmt.Sprintf("%s: Cleaning up repo", repoName)
	done := make(chan bool)
	go utils.ShowSpinner(statusMessage, done)

	changes := checkForChanges(opts.RepoPath, opts.DropChanges)

	currentBranch, prunedBranches, branchesCount := updateBranches(opts.RepoPath)

	statusLine := constructStatusLine(changes, currentBranch, prunedBranches, branchesCount)
	done <- true
	statusMessage = fmt.Sprintf("%s: Cleaned up repo", repoName)
	fmt.Printf("\r%s [%s] %s\n", statusMessage, green("✔"), statusLine)

	return nil
}

// checkForChanges handles resetting changes or listing uncommitted changes
func checkForChanges(repoPath string, dropChanges bool) string {
	var changes string

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return ""
	}

	wt, err := repo.Worktree()
	if err != nil {
		return ""
	}

	status, err := wt.Status()
	if err != nil {
		return ""
	}

	if !status.IsClean() {
		if dropChanges {
			_ = wt.Reset(&git.ResetOptions{Mode: git.HardReset})
			changes = color.RedString("Dropped all changes")
		} else {
			modified, added, deleted, untracked := CountChanges(status)
			changes = fmt.Sprintf("[%s] [%s] [%s] [%s]", color.YellowString("Modified: %d", modified), color.GreenString("Added: %d", added), color.RedString("Deleted: %d", deleted), color.BlueString("Untracked: %d", untracked))
		}
	}
	return changes
}

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

// updateBranches updates the main branch, prunes branches, and counts the branches
func updateBranches(repoPath string) (string, int, int) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", 0, 0
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", 0, 0
	}

	// Set up SSH authentication
	auth, err := utils.SshAuth()
	if err != nil {
		fmt.Printf("failed to set up SSH authentication: %s", err)
	}

	// Switch to main branch and update
	err = wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main"), Keep: true})
	if err != nil {
		return "", 0, 0
	}

	err = wt.Pull(&git.PullOptions{RemoteName: "origin", Progress: io.Discard, Auth: auth})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		fmt.Printf("failed to pull: %s", err)
		return "", 0, 0
	}

	// Get current branch
	head, err := repo.Head()
	if err != nil {
		return "", 0, 0
	}
	currentBranch := head.Name().Short()

	// Prune branches merged into main
	iter, _ := repo.Branches()
	prunedCount, branchesCount := 0, 1

	mainRef, _ := repo.Reference(plumbing.NewBranchReferenceName("main"), true)
	mainCommit, _ := repo.CommitObject(mainRef.Hash())

	iter.ForEach(func(ref *plumbing.Reference) error {
		branchName := ref.Name().Short()
		if branchName != "main" {
			branchCommit, _ := repo.CommitObject(ref.Hash())
			if isMerged, _ := mainCommit.IsAncestor(branchCommit); isMerged {
				// Remove the branch reference
				err := repo.Storer.RemoveReference(ref.Name())
				if err == nil {
					prunedCount++
				} else {
					fmt.Printf("failed to remove branch %s: %s\n", branchName, err)
				}
			} else {
				branchesCount++
			}
		}
		return nil
	})

	return currentBranch, prunedCount, branchesCount
}
