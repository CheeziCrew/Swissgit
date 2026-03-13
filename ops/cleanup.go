package ops

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/CheeziCrew/swissgit/git"
	gogit "github.com/go-git/go-git/v5"
)

// CleanupResult holds the outcome of cleaning up a single repo.
type CleanupResult struct {
	RepoName        string
	PrunedBranches  int
	RemainingBranch int
	CurrentBranch   string
	DefaultBranch   string
	Changes         git.Changes
	DroppedChanges  bool
	Success         bool
	Error           string
}

// CleanupRepo performs cleanup on a single repository:
// optionally drops changes, checks out default branch, fetches, pulls, prunes merged branches.
func CleanupRepo(repoPath string, dropChanges bool, defaultBranchOverride string) CleanupResult {
	repoName := filepath.Base(repoPath)

	repo, err := plainOpen(repoPath)
	if err != nil {
		return CleanupResult{RepoName: repoName, Error: fmt.Sprintf("could not open repository: %s", err)}
	}

	name, err := git.GetRepoNameFromRepo(repo)
	if err == nil {
		repoName = name
	}

	defaultBranch := git.DefaultBranch(repoPath, defaultBranchOverride)
	result := CleanupResult{RepoName: repoName, DefaultBranch: defaultBranch}

	hadChanges := false
	if dropChanges {
		pre, _ := git.CountChangesShell(repoPath)
		hadChanges = pre.HasChanges()

		if _, err := gitRunInDir(repoPath,"checkout", "."); err != nil {
			result.Error = fmt.Sprintf("git checkout . failed: %s", err)
			return result
		}
		if _, err := gitRunInDir(repoPath,"clean", "-fd"); err != nil {
			result.Error = fmt.Sprintf("git clean -fd failed: %s", err)
			return result
		}
		result.DroppedChanges = hadChanges
	}

	pruned, remaining, err := updateBranches(repo, repoPath, defaultBranch)
	if err != nil {
		result.Error = fmt.Sprintf("branch update failed: %s", err)
		return result
	}
	result.PrunedBranches = pruned
	result.RemainingBranch = remaining

	if !dropChanges {
		changes, err := git.CountChangesShell(repoPath)
		if err != nil {
			result.Error = fmt.Sprintf("failed to get status: %s", err)
			return result
		}
		result.Changes = changes
	}

	result.CurrentBranch, _ = git.GetBranchName(repo)
	result.Success = true
	return result
}

func updateBranches(repo *gogit.Repository, repoPath, defaultBranch string) (pruned, remaining int, err error) {
	if err := checkoutFetchPull(repo, repoPath, defaultBranch); err != nil {
		return 0, 0, err
	}

	protected := map[string]bool{defaultBranch: true}
	deleteSet := map[string]bool{}

	if err := collectMergedBranches(repoPath, protected, deleteSet); err != nil {
		return 0, 0, err
	}
	collectGoneBranches(repoPath, protected, deleteSet)
	collectOrphanedBranches(repoPath, protected, deleteSet)

	totalLocal := countLocalBranches(repoPath)

	toDelete := deleteBranchSet(repoPath, deleteSet)
	if toDelete < 0 {
		return 0, 0, fmt.Errorf("failed to delete branches")
	}

	rem := totalLocal - toDelete
	if rem < 0 {
		rem = 0
	}
	return toDelete, rem, nil
}

// checkoutFetchPull switches to the default branch, fetches, pulls, and prunes remote refs.
func checkoutFetchPull(repo *gogit.Repository, repoPath, defaultBranch string) error {
	if _, err := gitRunInDir(repoPath,"checkout", defaultBranch); err != nil {
		return fmt.Errorf("failed to checkout %s: %s", defaultBranch, err)
	}

	if err := fetchRemote(repo); err != nil {
		return fmt.Errorf("failed to fetch remote: %w", err)
	}

	if _, err := gitRunInDir(repoPath,"pull"); err != nil {
		return fmt.Errorf("failed to pull: %s", err)
	}

	gitRunInDir(repoPath,"remote", "prune", "origin")
	return nil
}

// parseBranchName strips leading whitespace and the current-branch marker (*) from a branch line.
func parseBranchName(raw string) string {
	branch := strings.TrimSpace(raw)
	if strings.HasPrefix(branch, "*") {
		branch = strings.TrimSpace(branch[1:])
	}
	return branch
}

// collectMergedBranches adds branches merged into the default branch to deleteSet.
func collectMergedBranches(repoPath string, protected, deleteSet map[string]bool) error {
	output, err := gitRunInDir(repoPath,"branch", "--merged")
	if err != nil {
		return fmt.Errorf("failed to list merged branches: %w", err)
	}
	for _, raw := range strings.Split(output, "\n") {
		branch := parseBranchName(raw)
		if branch != "" && !protected[branch] {
			deleteSet[branch] = true
		}
	}
	return nil
}

// collectGoneBranches adds branches whose remote tracking branch is gone to deleteSet.
func collectGoneBranches(repoPath string, protected, deleteSet map[string]bool) {
	output, err := gitRunInDir(repoPath,"branch", "-vv")
	if err != nil {
		return
	}
	for _, raw := range strings.Split(output, "\n") {
		line := parseBranchName(raw)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 1 || protected[parts[0]] {
			continue
		}
		if strings.Contains(raw, ": gone]") {
			deleteSet[parts[0]] = true
		}
	}
}

// collectOrphanedBranches adds local branches whose remote counterpart no longer exists to deleteSet.
func collectOrphanedBranches(repoPath string, protected, deleteSet map[string]bool) {
	output, err := gitRunInDir(repoPath,"branch", "--format=%(refname:short)")
	if err != nil {
		return
	}
	for _, raw := range strings.Split(output, "\n") {
		branch := strings.TrimSpace(raw)
		if branch == "" || protected[branch] || deleteSet[branch] {
			continue
		}
		_, err := gitRunInDir(repoPath,"rev-parse", "--verify", "origin/"+branch)
		if err != nil {
			deleteSet[branch] = true
		}
	}
}

// countLocalBranches returns the number of local branches.
func countLocalBranches(repoPath string) int {
	output, err := gitRunInDir(repoPath,"branch")
	if err != nil {
		return 0
	}
	count := 0
	for _, raw := range strings.Split(output, "\n") {
		if strings.TrimSpace(raw) != "" {
			count++
		}
	}
	return count
}

// deleteBranchSet deletes branches in the set. Returns the count deleted, or -1 on error.
func deleteBranchSet(repoPath string, deleteSet map[string]bool) int {
	var toDelete []string
	for b := range deleteSet {
		toDelete = append(toDelete, b)
	}

	if len(toDelete) > 0 {
		args := append([]string{"branch", "-D"}, toDelete...)
		_, err := gitRunInDir(repoPath, args...)
		if err != nil {
			return -1
		}
	}
	return len(toDelete)
}
