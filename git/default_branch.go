package git

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DefaultBranch detects the default branch for a repo.
// Tries origin/HEAD first, then common branch names, then the current branch.
func DefaultBranch(repoPath, fallback string) string {
	// Try origin/HEAD (set by regular git clone)
	output, err := gitRun(repoPath, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		return filepath.Base(strings.TrimSpace(output))
	}

	// Check common default branch names
	for _, candidate := range []string{fallback, "main", "master"} {
		_, err := gitRun(repoPath, "show-ref", "--verify", "--quiet", "refs/heads/"+candidate)
		if err == nil {
			return candidate
		}
	}

	// Fall back to current branch
	output, err = gitRun(repoPath, "branch", "--show-current")
	if err == nil {
		if branch := strings.TrimSpace(output); branch != "" {
			return branch
		}
	}

	return fallback
}

// AheadBehind uses git rev-list to count commits ahead/behind the remote tracking branch.
func AheadBehind(repoPath, branch string) (ahead, behind int) {
	output, err := gitRun(repoPath, "rev-list", "--left-right", "--count", "HEAD...origin/"+branch)
	if err != nil {
		return 0, 0
	}

	parts := strings.Fields(strings.TrimSpace(output))
	if len(parts) != 2 {
		return 0, 0
	}

	var a, b int
	fmt.Sscanf(parts[0], "%d", &a)
	fmt.Sscanf(parts[1], "%d", &b)
	return a, b
}
