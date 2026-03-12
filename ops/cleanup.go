package ops

import (
	"fmt"
	"os/exec"
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

	repo, err := gogit.PlainOpen(repoPath)
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

		cmd := exec.Command("git", "-C", repoPath, "checkout", ".")
		if out, err := cmd.CombinedOutput(); err != nil {
			result.Error = fmt.Sprintf("git checkout . failed: %s", strings.TrimSpace(string(out)))
			return result
		}
		cmd = exec.Command("git", "-C", repoPath, "clean", "-fd")
		if out, err := cmd.CombinedOutput(); err != nil {
			result.Error = fmt.Sprintf("git clean -fd failed: %s", strings.TrimSpace(string(out)))
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
	cmd := exec.Command("git", "-C", repoPath, "checkout", defaultBranch)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout %s: %s", defaultBranch, strings.TrimSpace(string(out)))
	}

	if err := git.FetchRemote(repo); err != nil {
		return fmt.Errorf("failed to fetch remote: %w", err)
	}

	pullCmd := exec.Command("git", "-C", repoPath, "pull")
	if out, err := pullCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to pull: %s", strings.TrimSpace(string(out)))
	}

	pruneCmd := exec.Command("git", "-C", repoPath, "remote", "prune", "origin")
	_ = pruneCmd.Run()

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
	cmd := exec.Command("git", "-C", repoPath, "branch", "--merged")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list merged branches: %w", err)
	}
	for _, raw := range strings.Split(string(output), "\n") {
		branch := parseBranchName(raw)
		if branch != "" && !protected[branch] {
			deleteSet[branch] = true
		}
	}
	return nil
}

// collectGoneBranches adds branches whose remote tracking branch is gone to deleteSet.
func collectGoneBranches(repoPath string, protected, deleteSet map[string]bool) {
	cmd := exec.Command("git", "-C", repoPath, "branch", "-vv")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	for _, raw := range strings.Split(string(output), "\n") {
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
	cmd := exec.Command("git", "-C", repoPath, "branch", "--format=%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	for _, raw := range strings.Split(string(output), "\n") {
		branch := strings.TrimSpace(raw)
		if branch == "" || protected[branch] || deleteSet[branch] {
			continue
		}
		check := exec.Command("git", "-C", repoPath, "rev-parse", "--verify", "origin/"+branch)
		if check.Run() != nil {
			deleteSet[branch] = true
		}
	}
}

// countLocalBranches returns the number of local branches.
func countLocalBranches(repoPath string) int {
	cmd := exec.Command("git", "-C", repoPath, "branch")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	count := 0
	for _, raw := range strings.Split(string(output), "\n") {
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
		args := append([]string{"-C", repoPath, "branch", "-D"}, toDelete...)
		cmd := exec.Command("git", args...)
		if _, err := cmd.CombinedOutput(); err != nil {
			return -1
		}
	}
	return len(toDelete)
}
