package ops

import (
	"encoding/json"
	"fmt"
	"time"
)

// MyPR represents an open PR authored by the current user.
type MyPR struct {
	Repo      string
	Number    int
	Title     string
	URL       string
	State     string
	Draft     bool
	CreatedAt time.Time
}

type ghMyPRResult struct {
	Repository struct {
		Name          string `json:"name"`
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	State  string `json:"state"`
	IsDraft   bool      `json:"isDraft"`
	CreatedAt time.Time `json:"createdAt"`
}

// FetchMyPRs returns all open PRs authored by the currently authenticated gh user.
func FetchMyPRs() ([]MyPR, error) {
	out, err := ghRun("search", "prs",
		"--author=@me",
		"--state=open",
		"--limit=100",
		"--json", "repository,number,title,url,state,isDraft,createdAt",
	)
	if err != nil {
		return nil, fmt.Errorf("gh search failed: %s", err)
	}

	var results []ghMyPRResult
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}

	var prs []MyPR
	for _, r := range results {
		prs = append(prs, MyPR{
			Repo:      r.Repository.NameWithOwner,
			Number:    r.Number,
			Title:     r.Title,
			URL:       r.URL,
			State:     r.State,
			Draft:     r.IsDraft,
			CreatedAt: r.CreatedAt,
		})
	}

	return prs, nil
}
