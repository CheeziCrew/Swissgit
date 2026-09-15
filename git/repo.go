package git

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	gogit "github.com/go-git/go-git/v5"
)

// SkippedRepo records a directory that looks like a git repo but could not be
// opened, together with a human-readable reason (and a fix, when we know one).
type SkippedRepo struct {
	Path   string
	Reason string
}

// IsGitRepository checks if a directory contains a .git entry. That entry is a
// directory for a normal clone and a file for a linked worktree; both count.
func IsGitRepository(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

// DiscoverRepos scans one level of subdirectories for usable git repos.
// Directories that look like repos but cannot be opened are skipped rather than
// failing the whole scan — use DiscoverReposWithSkipped to find out why.
func DiscoverRepos(rootPath string) ([]string, error) {
	repos, _, err := DiscoverReposWithSkipped(rootPath)
	return repos, err
}

// DiscoverReposWithSkipped is DiscoverRepos, but it also reports the repos it
// had to skip. Callers that can surface a warning should prefer this.
func DiscoverReposWithSkipped(rootPath string) ([]string, []SkippedRepo, error) {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var repos []string
	var skipped []SkippedRepo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sub := filepath.Join(rootPath, entry.Name())
		if !IsGitRepository(sub) {
			continue
		}
		if _, err := gogit.PlainOpen(sub); err != nil {
			skipped = append(skipped, SkippedRepo{Path: sub, Reason: explainOpenFailure(sub, err)})
			continue
		}
		repos = append(repos, sub)
	}
	return repos, skipped, nil
}

// explainOpenFailure turns a go-git open error into something actionable. The
// common case in practice is a dangling extensions.worktreeConfig: some git
// subcommands set that key without bumping core.repositoryformatversion, and
// removing the worktree never cleans it up. C git tolerates the mismatch;
// go-git refuses to open the repository at all.
func explainOpenFailure(path string, err error) string {
	if hasDanglingWorktreeConfig(path) {
		return fmt.Sprintf("%v — dangling extensions.worktreeConfig; repair with: git -C %s config --local --unset extensions.worktreeConfig", err, path)
	}
	return err.Error()
}

func hasDanglingWorktreeConfig(path string) bool {
	data, err := os.ReadFile(filepath.Join(path, ".git", "config"))
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "worktreeconfig")
}

// GetBranchName returns the current branch name for a repo.
func GetBranchName(repo *gogit.Repository) (string, error) {
	headRef, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("could not get head: %w", err)
	}
	return headRef.Name().Short(), nil
}

// GetRepoOwnerAndName extracts the owner and repo name from origin remote URL.
func GetRepoOwnerAndName(repo *gogit.Repository) (string, string, error) {
	remotes, err := repo.Remotes()
	if err != nil {
		return "", "", fmt.Errorf("failed to get remotes: %w", err)
	}
	if len(remotes) == 0 {
		return "", "", fmt.Errorf("no remotes configured")
	}

	urls := remotes[0].Config().URLs
	if len(urls) == 0 {
		return "", "", fmt.Errorf("remote has no URLs configured")
	}

	re := regexp.MustCompile(`(?:[:/])([^/]+)/([^/]+?)(?:\.git)?$`)
	matches := re.FindStringSubmatch(urls[0])
	if len(matches) < 3 {
		return "", "", fmt.Errorf("failed to parse remote URL: %s", urls[0])
	}
	return matches[1], matches[2], nil
}

// GetRepoName returns just the repo name for a path.
func GetRepoName(repoPath string) (string, error) {
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("could not open repository at %s: %w", repoPath, err)
	}
	_, name, err := GetRepoOwnerAndName(repo)
	return name, err
}

// GetRepoNameFromRepo extracts the repo name from an already-opened repo.
func GetRepoNameFromRepo(repo *gogit.Repository) (string, error) {
	_, name, err := GetRepoOwnerAndName(repo)
	return name, err
}
