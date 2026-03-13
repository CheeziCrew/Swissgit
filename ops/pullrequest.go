package ops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/CheeziCrew/swissgit/git"
	gogit "github.com/go-git/go-git/v5"
)

// PRRequest is the GitHub API payload for creating a PR.
type PRRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

// PRResult holds the outcome of a PR creation.
type PRResult struct {
	RepoName string
	PRURL    string
	Success  bool
	Error    string
}

// CommitAndCreatePR commits changes and creates a PR for a single repo.
func CommitAndCreatePR(repoPath, branchName, commitMessage, base string, changes []string, breakingChange bool) PRResult {
	commitResult := CommitAndPush(repoPath, branchName, commitMessage)
	if !commitResult.Success {
		return PRResult{
			RepoName: commitResult.RepoName,
			Error:    commitResult.Error,
		}
	}

	repo, err := plainOpen(repoPath)
	if err != nil {
		return PRResult{RepoName: commitResult.RepoName, Error: fmt.Sprintf("could not open repo: %s", err)}
	}

	url, err := CreatePullRequest(repo, commitMessage, branchName, base, changes, breakingChange)
	if err != nil {
		return PRResult{RepoName: commitResult.RepoName, Error: fmt.Sprintf("PR creation failed: %s", err)}
	}

	return PRResult{
		RepoName: commitResult.RepoName,
		PRURL:    url,
		Success:  true,
	}
}

// CreatePullRequest creates a new PR on GitHub. Returns the HTML URL.
func CreatePullRequest(repo *gogit.Repository, commitMessage, branch, base string, changes []string, breakingChange bool) (string, error) {
	owner, repoName, err := git.GetRepoOwnerAndName(repo)
	if err != nil {
		return "", fmt.Errorf("could not detect repository owner and name: %w", err)
	}

	accessToken := os.Getenv("GITHUB_TOKEN")
	if accessToken == "" {
		return "", fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repoName)

	body, err := BuildPullRequestBody(changes, breakingChange)
	if err != nil {
		return "", fmt.Errorf("failed to build pull request body: %w", err)
	}

	pr := PRRequest{
		Title: branch + ": " + commitMessage,
		Head:  branch,
		Body:  body,
		Base:  base,
	}

	jsonData, err := json.Marshal(pr)
	if err != nil {
		return "", fmt.Errorf("failed to encode pull request data: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create PR, status: %s, body: %s", resp.Status, string(respBody))
	}

	var result struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse PR response: %w", err)
	}
	return result.HTMLURL, nil
}
