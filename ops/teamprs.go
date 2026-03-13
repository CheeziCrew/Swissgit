package ops

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// TeamPR represents an open PR in a team repo.
type TeamPR struct {
	Repo      string
	Number    int
	Author    string
	Title     string
	URL       string
	Draft     bool
	CreatedAt time.Time
}

type ghTeamPRResult struct {
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	URL       string    `json:"url"`
	IsDraft   bool      `json:"isDraft"`
	CreatedAt time.Time `json:"createdAt"`
}

// FetchTeamRepoNames returns non-archived repo names for an org/team, excluding prefixes.
func FetchTeamRepoNames(org, team string, excludePrefixes []string) ([]string, error) {
	out, err := ghRun("api",
		fmt.Sprintf("/orgs/%s/teams/%s/repos", org, team),
		"--paginate",
		"--jq", ".[].name",
	)
	if err != nil {
		return nil, fmt.Errorf("gh api failed: %s", err)
	}

	var names []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		if matchesAnyPrefix(name, excludePrefixes) {
			continue
		}
		names = append(names, name)
	}

	return names, nil
}

// FetchTeamPRs finds all open PRs in the org and filters to team repos only.
func FetchTeamPRs(org string, teamRepos []string) ([]TeamPR, error) {
	out, err := ghRun("search", "prs",
		"--state=open",
		"--owner="+org,
		"--limit=200",
		"--json", "repository,number,title,author,url,isDraft,createdAt",
	)
	if err != nil {
		return nil, fmt.Errorf("gh search failed: %s", err)
	}

	var results []ghTeamPRResult
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}

	repoSet := make(map[string]bool, len(teamRepos))
	for _, r := range teamRepos {
		repoSet[r] = true
	}

	var prs []TeamPR
	for _, r := range results {
		if !repoSet[r.Repository.Name] {
			continue
		}
		prs = append(prs, TeamPR{
			Repo:      r.Repository.Name,
			Number:    r.Number,
			Author:    r.Author.Login,
			Title:     r.Title,
			URL:       r.URL,
			Draft:     r.IsDraft,
			CreatedAt: r.CreatedAt,
		})
	}

	return prs, nil
}

func IsBot(author string) bool {
	return strings.HasSuffix(author, "[bot]") ||
		strings.HasSuffix(author, "-bot") ||
		author == "dependabot" ||
		author == "renovate"
}

func matchesAnyPrefix(name string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}
