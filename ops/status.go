package ops

import (
	"fmt"
	"path/filepath"

	"github.com/CheeziCrew/swissgit/git"
)

// StatusResult holds the outcome of checking a single repo's status.
type StatusResult struct {
	RepoName      string
	Branch        string
	DefaultBranch string
	Modified      int
	Added         int
	Deleted       int
	Untracked     int
	Ahead         int
	Behind        int
	Clean         bool
	Error         string
}

// GetRepoStatus fetches, counts changes, and returns status for a single repo.
func GetRepoStatus(repoPath string) StatusResult {
	repoName := filepath.Base(repoPath)

	repo, err := plainOpen(repoPath)
	if err != nil {
		return StatusResult{RepoName: repoName, Error: fmt.Sprintf("could not open repository: %s", err)}
	}

	if err := fetchRemote(repo); err != nil {
		return StatusResult{RepoName: repoName, Error: fmt.Sprintf("could not fetch remote: %s", err)}
	}

	branch, err := git.GetBranchName(repo)
	if err != nil {
		return StatusResult{RepoName: repoName, Error: fmt.Sprintf("could not get branch name: %s", err)}
	}

	changes, err := git.CountChangesShell(repoPath)
	if err != nil {
		return StatusResult{RepoName: repoName, Error: fmt.Sprintf("could not get status: %s", err)}
	}

	ahead, behind := git.AheadBehind(repoPath, branch)
	defaultBranch := git.DefaultBranch(repoPath, "main")

	isClean := !changes.HasChanges() && ahead == 0 && behind == 0

	return StatusResult{
		RepoName:      repoName,
		Branch:        branch,
		DefaultBranch: defaultBranch,
		Modified:      changes.Modified,
		Added:         changes.Added,
		Deleted:       changes.Deleted,
		Untracked:     changes.Untracked,
		Ahead:         ahead,
		Behind:        behind,
		Clean:         isClean,
	}
}
