package branches

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// listBranches returns a list of branches in the repository, checking for staleness if required.
func listBranches(repo *git.Repository, isRemote bool) ([]string, error) {
	var branches []string

	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("could not list branches: %w", err)
	}

	refs.ForEach(func(ref *plumbing.Reference) error {
		if isRemote && ref.Name().IsRemote() {
			// Filter out "origin/HEAD" and "origin/main"
			branchName := ref.Name().Short()
			if branchName == "origin/HEAD" || branchName == "origin/main" {
				return nil
			}

			branchName = branchName[len("origin/"):]

			isStale := checkIfBranchIsStale(repo, ref)
			if isStale {
				branchName = color.New(color.FgRed).Sprint(branchName) // Color stale branches red
			} else {
				branchName = color.New(color.FgYellow).Sprint(branchName)
			}

			branches = append(branches, branchName)
		} else if !isRemote && ref.Name().IsBranch() {
			// Handle local branches
			branchName := ref.Name().Short()

			// Check if the branch is stale
			isStale := checkIfBranchIsStale(repo, ref)
			if isStale {
				branchName = color.New(color.FgRed).Sprint(branchName) // Color stale branches red
			} else if branchName == "main" {
				branchName = color.New(color.FgGreen).Sprint(branchName)
			}

			branches = append(branches, branchName)
		}
		return nil
	})

	return branches, nil
}

// checkIfBranchIsStale checks if the branch has not been updated in the last 120 days.
func checkIfBranchIsStale(repo *git.Repository, ref *plumbing.Reference) bool {
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return false // Unable to determine, treat as not stale
	}

	// Calculate the age of the branch in days
	age := time.Since(commit.Committer.When).Hours() / 24
	return age > 120
}
