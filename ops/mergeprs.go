package ops

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// PRInfo represents an approved PR found via gh search.
type PRInfo struct {
	Repo   string
	Number int
	Title  string
}

// MergePRResult holds the outcome of merging a single PR.
type MergePRResult struct {
	Repo     string
	PRNumber string
	Title    string
	Success  bool
	Error    string
}

type ghSearchResult struct {
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
	Number int    `json:"number"`
	Title  string `json:"title"`
}

// FetchApprovedPRs finds all open PRs authored by the current user with approved reviews.
func FetchApprovedPRs(org string) ([]PRInfo, error) {
	out, err := ghRun("search", "prs",
		"--author=@me",
		"--review=approved",
		"--state=open",
		"--owner="+org,
		"--limit=200",
		"--json", "repository,number,title",
	)
	if err != nil {
		return nil, fmt.Errorf("gh search failed: %s", err)
	}

	var results []ghSearchResult
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}

	var prs []PRInfo
	for _, r := range results {
		prs = append(prs, PRInfo{
			Repo:   r.Repository.Name,
			Number: r.Number,
			Title:  r.Title,
		})
	}

	return prs, nil
}

// MergePR merges a single PR via gh CLI.
func MergePR(org, repo string, number int) MergePRResult {
	result := MergePRResult{
		Repo:     repo,
		PRNumber: strconv.Itoa(number),
	}

	fullRepo := fmt.Sprintf("%s/%s", org, repo)
	_, err := ghRun("pr", "merge",
		strconv.Itoa(number),
		"--repo", fullRepo,
		"--merge",
		"--delete-branch",
	)
	if err != nil {
		result.Error = fmt.Sprintf("merge failed: %s", err)
		return result
	}

	result.Success = true
	return result
}
