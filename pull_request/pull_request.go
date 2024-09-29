package pull_request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/CheeziCrew/swissgo/commit"
	"github.com/CheeziCrew/swissgo/utils"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

// PullRequest represents the body of a pull request creation request
type PullRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

// CommitAndPull commits changes, creates a pull request, and optionally processes subdirectories if `allFlag` is true.
func CommitAndPull(repoPath, branchName, commitMessage, base string, changes []string, breakingChange, allFlag bool) error {
	// If allFlag is true, process all subdirectories
	if allFlag {
		return ProcessSubdirectories(repoPath, branchName, commitMessage, base, changes, breakingChange)
	}
	// Process a single repository
	return ProcessSingleRepository(repoPath, branchName, commitMessage, base, changes, breakingChange)
}

// ProcessSubdirectories iterates over subdirectories and calls CommitAndPull on each
func ProcessSubdirectories(repoPath, branchName, commitMessage, base string, changes []string, breakingChange bool) error {
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subRepoPath := filepath.Join(repoPath, entry.Name())

			// Check if the subdirectory is a Git repository
			if !utils.IsGitRepository(subRepoPath) {
				continue
			}

			err := ProcessSingleRepository(subRepoPath, branchName, commitMessage, base, changes, breakingChange)
			if err != nil {
			}
		}
	}
	return nil
}

// ProcessSingleRepository commits changes and creates a pull request for a single repository
func ProcessSingleRepository(repoPath, branchName, commitMessage, base string, changes []string, breakingChange bool) error {

	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	if err := commit.CommitChanges(repoPath, branchName, commitMessage); err != nil {
		return err
	}

	repoName, err := utils.GetRepoName(repoPath)
	if err != nil {
		statusMessage := fmt.Sprintf("%s: getting repoName", repoPath)
		fmt.Printf("\r%s failed [%s]: %s\n", statusMessage, red("x"), red(err.Error()))
		return err
	}

	statusMessage := fmt.Sprintf("%s: creating pull request", repoName)
	done := make(chan bool)
	go utils.ShowSpinner(statusMessage, done)

	if err := CreatePullRequest(repoPath, commitMessage, branchName, base, changes, breakingChange); err != nil {
		done <- true
		fmt.Printf("\r%s failed [✗]: %s \n", statusMessage, red(err.Error()))
		return err
	}

	done <- true
	fmt.Printf("\r%s [%s] \n", statusMessage, green("✔"))
	return nil
}

// CreatePullRequest creates a new pull request on GitHub
func CreatePullRequest(repoPath, commitMessage, branch, base string, changes []string, breakingChange bool) error {

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("could not open repository at %s: %w", repoPath, err)
	}

	owner, repoName, err := utils.GetRepoOwnerAndName(repo)
	if err != nil {
		return fmt.Errorf("could not detect repository owner and name: %w", err)
	}

	accessToken := os.Getenv("GITHUB_TOKEN")
	if accessToken == "" {
		return fmt.Errorf("GitHub access token not provided. Make sure to set the GITHUB_TOKEN environment variable")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repoName)

	body, err := BuildPullRequestBody(changes, breakingChange)
	if err != nil {
		return fmt.Errorf("failed to build pull request body: %w", err)
	}

	pr := PullRequest{
		Title: branch + ": " + commitMessage,
		Head:  branch,
		Body:  body,
		Base:  base,
	}

	jsonData, err := json.Marshal(pr)
	if err != nil {
		return fmt.Errorf("failed to encode pull request data: %w", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create pull request, status: %s", resp.Status)
	}
	return nil
}
