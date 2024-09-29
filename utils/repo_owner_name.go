package utils

import (
	"fmt"
	"regexp"

	"github.com/go-git/go-git/v5"
)

// getRepoOwnerAndName retrieves the repository owner and name from the local repository's remote URL
func GetRepoOwnerAndName(repo *git.Repository) (string, string, error) {

	// Get the remote URL (default to 'origin')
	remotes, err := repo.Remotes()
	if err != nil || len(remotes) == 0 {
		return "", "", fmt.Errorf("failed to get remotes: %w", err)
	}

	// Extract the first remote's URL (usually 'origin')
	remoteURL := remotes[0].Config().URLs[0]

	// Parse the remote URL to extract the owner and repo name using regex
	re := regexp.MustCompile(`(?:[:/])([^/]+)/([^/]+?)(?:\.git)?$`)
	matches := re.FindStringSubmatch(remoteURL)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("failed to parse remote URL: %s", remoteURL)
	}

	return matches[1], matches[2], nil
}

func GetRepoName(repoPath string) (string, error) {

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("could not open repository at %s: %w", repoPath, err)
	}
	_, repoName, err := GetRepoOwnerAndName(repo)
	if err != nil {
		return "", err
	}
	return repoName, nil
}
