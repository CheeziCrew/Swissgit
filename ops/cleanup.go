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

	// Drop changes
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

	// Update branches
	pruned, remaining, err := updateBranches(repo, repoPath, defaultBranch)
	if err != nil {
		result.Error = fmt.Sprintf("branch update failed: %s", err)
		return result
	}
	result.PrunedBranches = pruned
	result.RemainingBranch = remaining

	// Check remaining changes
	if !dropChanges {
		changes, err := git.CountChangesShell(repoPath)
		if err != nil {
			result.Error = fmt.Sprintf("failed to get status: %s", err)
			return result
		}
		result.Changes = changes
	}

	// Get current branch
	result.CurrentBranch, _ = git.GetBranchName(repo)
	result.Success = true
	return result
}

func updateBranches(repo *gogit.Repository, repoPath, defaultBranch string) (pruned, remaining int, err error) {
	// Shell checkout default branch
	cmd := exec.Command("git", "-C", repoPath, "checkout", defaultBranch)
	if out, err := cmd.CombinedOutput(); err != nil {
		return 0, 0, fmt.Errorf("failed to checkout %s: %s", defaultBranch, strings.TrimSpace(string(out)))
	}

	if err := git.FetchRemote(repo); err != nil {
		return 0, 0, fmt.Errorf("failed to fetch remote: %w", err)
	}

	// Shell pull
	pullCmd := exec.Command("git", "-C", repoPath, "pull")
	if out, err := pullCmd.CombinedOutput(); err != nil {
		return 0, 0, fmt.Errorf("failed to pull: %s", strings.TrimSpace(string(out)))
	}

	// Prune remote tracking refs so we detect gone upstreams
	pruneCmd := exec.Command("git", "-C", repoPath, "remote", "prune", "origin")
	_ = pruneCmd.Run()

	protected := map[string]bool{defaultBranch: true}
	deleteSet := map[string]bool{}

	// 1. Branches merged into default (covers regular merges)
	cmd = exec.Command("git", "-C", repoPath, "branch", "--merged")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to list merged branches: %w", err)
	}
	for _, raw := range strings.Split(string(output), "\n") {
		branch := strings.TrimSpace(raw)
		if branch == "" {
			continue
		}
		if strings.HasPrefix(branch, "*") {
			branch = strings.TrimSpace(branch[1:])
		}
		if !protected[branch] {
			deleteSet[branch] = true
		}
	}

	// 2. Branches whose remote tracking branch is gone (covers squash/rebase merges with -u tracking)
	cmd = exec.Command("git", "-C", repoPath, "branch", "-vv")
	output, err = cmd.Output()
	if err == nil {
		for _, raw := range strings.Split(string(output), "\n") {
			line := strings.TrimSpace(raw)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "*") {
				line = strings.TrimSpace(line[1:])
			}
			parts := strings.Fields(line)
			if len(parts) < 1 {
				continue
			}
			branch := parts[0]
			if protected[branch] {
				continue
			}
			if strings.Contains(raw, ": gone]") {
				deleteSet[branch] = true
			}
		}
	}

	// 3. Local branches whose remote counterpart no longer exists (covers squash/rebase merges without -u tracking)
	// After prune, if origin/<branch> doesn't exist, the remote was deleted (i.e. merged via PR)
	cmd = exec.Command("git", "-C", repoPath, "branch", "--format=%(refname:short)")
	output, err = cmd.Output()
	if err == nil {
		for _, raw := range strings.Split(string(output), "\n") {
			branch := strings.TrimSpace(raw)
			if branch == "" || protected[branch] || deleteSet[branch] {
				continue
			}
			// Check if origin/<branch> exists
			check := exec.Command("git", "-C", repoPath, "rev-parse", "--verify", "origin/"+branch)
			if check.Run() != nil {
				// Remote branch doesn't exist — it was deleted (merged via PR)
				deleteSet[branch] = true
			}
		}
	}

	// Count total local branches
	cmd = exec.Command("git", "-C", repoPath, "branch")
	output, err = cmd.Output()
	totalLocal := 0
	if err == nil {
		for _, raw := range strings.Split(string(output), "\n") {
			if strings.TrimSpace(raw) != "" {
				totalLocal++
			}
		}
	}

	var toDelete []string
	for b := range deleteSet {
		toDelete = append(toDelete, b)
	}

	if len(toDelete) > 0 {
		args := append([]string{"-C", repoPath, "branch", "-D"}, toDelete...)
		cmd = exec.Command("git", args...)
		if _, err := cmd.CombinedOutput(); err != nil {
			return 0, 0, fmt.Errorf("failed to delete branches: %w", err)
		}
	}

	rem := totalLocal - len(toDelete)
	if rem < 0 {
		rem = 0
	}
	return len(toDelete), rem, nil
}
