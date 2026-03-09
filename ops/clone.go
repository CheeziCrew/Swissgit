package ops

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/CheeziCrew/swissgit/git"
	gogit "github.com/go-git/go-git/v5"
)

// Repository represents a GitHub repository.
type Repository struct {
	Name     string `json:"name"`
	SSHURL   string `json:"ssh_url"`
	Archived bool   `json:"archived"`
}

// CloneResult holds the outcome of a clone operation.
type CloneResult struct {
	RepoName string
	Skipped  bool
	Success  bool
	Error    string
}

// CloneRepository clones a single repo via SSH.
func CloneRepository(repo Repository, destPath string) CloneResult {
	result := CloneResult{RepoName: repo.Name}

	if git.IsGitRepository(destPath) {
		result.Skipped = true
		result.Success = true
		return result
	}

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		if err := os.MkdirAll(destPath, 0755); err != nil {
			result.Error = fmt.Sprintf("failed to create directory: %s", err)
			return result
		}
	}

	auth, err := git.SSHAuth()
	if err != nil {
		result.Error = fmt.Sprintf("SSH auth failed: %s", err)
		return result
	}

	_, err = gogit.PlainClone(destPath, false, &gogit.CloneOptions{
		URL:      repo.SSHURL,
		Progress: io.Discard,
		Auth:     auth,
	})
	if err != nil {
		result.Error = fmt.Sprintf("clone failed: %s", err)
		return result
	}

	result.Success = true
	return result
}

// CloneFromURL parses a repo URL and clones it.
func CloneFromURL(repoURL, destPath string) CloneResult {
	parts := strings.Split(repoURL, "/")
	if len(parts) < 2 {
		return CloneResult{Error: fmt.Sprintf("invalid repository URL: %s", repoURL)}
	}
	repo := Repository{
		Name:   strings.TrimSuffix(filepath.Base(parts[len(parts)-1]), ".git"),
		SSHURL: repoURL,
	}
	return CloneRepository(repo, destPath)
}

// GetOrgRepositories fetches repos for a GitHub org (optionally filtered by team).
func GetOrgRepositories(orgName, teamName string) ([]Repository, error) {
	accessToken := os.Getenv("GITHUB_TOKEN")
	if accessToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	var url string
	if teamName != "" {
		url = fmt.Sprintf("https://api.github.com/orgs/%s/teams/%s/repos", orgName, teamName)
	} else {
		url = fmt.Sprintf("https://api.github.com/orgs/%s/repos", orgName)
	}

	var allRepos []Repository
	client := &http.Client{}
	page := 1

	for {
		paginatedURL := fmt.Sprintf("%s?page=%d&per_page=100", url, page)
		req, err := http.NewRequest("GET", paginatedURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "token "+accessToken)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to fetch repositories, status: %s", resp.Status)
		}

		var repos []Repository
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		hasNext := strings.Contains(resp.Header.Get("Link"), `rel="next"`)
		resp.Body.Close()

		for _, r := range repos {
			if !r.Archived {
				allRepos = append(allRepos, r)
			}
		}
		if !hasNext {
			break
		}
		page++
	}

	return allRepos, nil
}
