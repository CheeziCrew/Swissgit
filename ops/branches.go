package ops

import (
	"fmt"
	"path/filepath"
	"strings"
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
	IsMerged bool // merged into default branch
}

// BranchesResult holds the outcome of listing branches for a repo.
type BranchesResult struct {
	RepoName       string
	Path           string
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

	merged := getMergedBranches(repoPath, defaultBranch)
	markMerged(local, merged)
	markMerged(remote, merged)

	return BranchesResult{
		RepoName:       repoName,
		Path:           repoPath,
		CurrentBranch:  branch,
		DefaultBranch:  defaultBranch,
		LocalBranches:  local,
		RemoteBranches: remote,
	}
}

// getMergedBranches returns a set of branch names merged into the target branch.
func getMergedBranches(repoPath, target string) map[string]bool {
	merged := make(map[string]bool)
	out, err := gitRunInDir(repoPath, "branch", "--merged", target)
	if err != nil {
		return merged
	}
	for _, line := range strings.Split(out, "\n") {
		name := strings.TrimSpace(strings.TrimPrefix(line, "*"))
		name = strings.TrimSpace(name)
		if name != "" && name != target {
			merged[name] = true
		}
	}
	return merged
}

func markMerged(branches []BranchInfo, merged map[string]bool) {
	for i := range branches {
		if merged[branches[i].Name] {
			branches[i].IsMerged = true
		}
	}
}

// DeleteMergedBranches deletes the given local branches. Returns count deleted.
func DeleteMergedBranches(repoPath string, branches []string) (int, error) {
	deleted := 0
	for _, b := range branches {
		_, err := gitRunInDir(repoPath, "branch", "-d", b)
		if err != nil {
			return deleted, fmt.Errorf("failed to delete %s: %w", b, err)
		}
		deleted++
	}
	return deleted, nil
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
