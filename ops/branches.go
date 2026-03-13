package ops

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/CheeziCrew/swissgit/git"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// BranchInfo represents a single branch with metadata.
type BranchInfo struct {
	Name     string
	IsStale  bool // >120 days since last commit
	IsRemote bool
}

// BranchesResult holds the outcome of listing branches for a repo.
type BranchesResult struct {
	RepoName       string
	CurrentBranch  string
	DefaultBranch  string
	LocalBranches  []BranchInfo
	RemoteBranches []BranchInfo
	Error          string
}

// GetBranches fetches and lists branches for a single repo.
func GetBranches(repoPath string) BranchesResult {
	repoName := filepath.Base(repoPath)

	repo, err := plainOpen(repoPath)
	if err != nil {
		return BranchesResult{RepoName: repoName, Error: fmt.Sprintf("could not open repository: %s", err)}
	}

	if err := fetchRemote(repo); err != nil {
		return BranchesResult{RepoName: repoName, Error: fmt.Sprintf("could not fetch remote: %s", err)}
	}

	branch, err := git.GetBranchName(repo)
	if err != nil {
		return BranchesResult{RepoName: repoName, Error: fmt.Sprintf("could not get branch name: %s", err)}
	}

	defaultBranch := git.DefaultBranch(repoPath, "main")

	local, err := listBranches(repo, false)
	if err != nil {
		return BranchesResult{RepoName: repoName, Error: fmt.Sprintf("could not list local branches: %s", err)}
	}

	remote, err := listBranches(repo, true)
	if err != nil {
		return BranchesResult{RepoName: repoName, Error: fmt.Sprintf("could not list remote branches: %s", err)}
	}

	return BranchesResult{
		RepoName:       repoName,
		CurrentBranch:  branch,
		DefaultBranch:  defaultBranch,
		LocalBranches:  local,
		RemoteBranches: remote,
	}
}

func listBranches(repo *gogit.Repository, isRemote bool) ([]BranchInfo, error) {
	var branches []BranchInfo

	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("could not list branches: %w", err)
	}

	refs.ForEach(func(ref *plumbing.Reference) error {
		if isRemote && ref.Name().IsRemote() {
			name := ref.Name().Short()
			if name == "origin/HEAD" || name == "origin/main" {
				return nil
			}
			name = name[len("origin/"):]
			branches = append(branches, BranchInfo{
				Name:     name,
				IsStale:  isBranchStale(repo, ref),
				IsRemote: true,
			})
		} else if !isRemote && ref.Name().IsBranch() {
			branches = append(branches, BranchInfo{
				Name:    ref.Name().Short(),
				IsStale: isBranchStale(repo, ref),
			})
		}
		return nil
	})

	return branches, nil
}

func isBranchStale(repo *gogit.Repository, ref *plumbing.Reference) bool {
	c, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return false
	}
	age := time.Since(c.Committer.When).Hours() / 24
	return age > 120
}
