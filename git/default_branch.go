package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// DefaultBranch detects the default branch for a repo.
// Tries origin/HEAD first, then common branch names, then the current branch.
func DefaultBranch(repoPath, fallback string) string {
	// Try origin/HEAD (set by regular git clone)
	cmd := exec.Command("git", "-C", repoPath, "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()
	if err == nil {
		return filepath.Base(strings.TrimSpace(string(output)))
	}

	// Check common default branch names
	for _, candidate := range []string{fallback, "main", "master"} {
		cmd = exec.Command("git", "-C", repoPath, "show-ref", "--verify", "--quiet", "refs/heads/"+candidate)
		if cmd.Run() == nil {
			return candidate
		}
	}

	// Fall back to current branch
	cmd = exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err = cmd.Output()
	if err == nil {
		if branch := strings.TrimSpace(string(output)); branch != "" {
			return branch
		}
	}

	return fallback
}

// AheadBehind uses git rev-list to count commits ahead/behind the remote tracking branch.
func AheadBehind(repoPath, branch string) (ahead, behind int) {
	cmd := exec.Command("git", "-C", repoPath, "rev-list", "--left-right", "--count", "HEAD...origin/"+branch)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) != 2 {
		return 0, 0
	}

	var a, b int
	fmt.Sscanf(parts[0], "%d", &a)
	fmt.Sscanf(parts[1], "%d", &b)
	return a, b
}
