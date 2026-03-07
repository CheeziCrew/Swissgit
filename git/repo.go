package git

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	gogit "github.com/go-git/go-git/v5"
)

// IsGitRepository checks if a directory contains a .git folder.
func IsGitRepository(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

// DiscoverRepos scans one level of subdirectories for git repos.
func DiscoverRepos(rootPath string) ([]string, error) {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var repos []string
	for _, entry := range entries {
		if entry.IsDir() {
			sub := filepath.Join(rootPath, entry.Name())
			if IsGitRepository(sub) {
				repos = append(repos, sub)
			}
		}
	}
	return repos, nil
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
