package ops

import (
	"encoding/json"
	"fmt"
	"strings"
)

// EnableWorkflowResult holds the outcome of checking/enabling workflows for a repo.
type EnableWorkflowResult struct {
	Repo           string
	EnabledCount   int
	RetriggeredPRs int
	Success        bool
	Error          string
}

const (
	flagJSON = "--json"
	flagRepo = "--repo"
)

type ghWorkflow struct {
	Name  string `json:"name"`
	ID    int    `json:"id"`
	State string `json:"state"`
}

// FetchOrgRepoNames returns all non-archived repo names for an org.
func FetchOrgRepoNames(org string) ([]string, error) {
	out, err := ghRun("repo", "list", org,
		"--limit", "500",
		"--no-archived",
		flagJSON, "name",
		"-q", ".[].name",
	)
	if err != nil {
		return nil, fmt.Errorf("gh repo list failed: %s", err)
	}

	var names []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			names = append(names, line)
		}
	}

	return names, nil
}

// FindAndEnableWorkflows checks a repo for disabled-by-inactivity workflows and re-enables them.
// If prBranch is non-empty and workflows were enabled, it close/reopens PRs from that head branch to retrigger runs.
func FindAndEnableWorkflows(org, repo, workflowName, prBranch string) EnableWorkflowResult {
	result := EnableWorkflowResult{Repo: repo}
	fullRepo := fmt.Sprintf("%s/%s", org, repo)

	out, err := ghRun("workflow", "list",
		flagRepo, fullRepo,
		"--all",
		flagJSON, "name,id,state",
	)
	if err != nil {
		result.Error = fmt.Sprintf("list workflows failed: %s", err)
		return result
	}

	var workflows []ghWorkflow
	if strings.TrimSpace(out) == "" {
		result.Success = true
		return result
	}
	if err := json.Unmarshal([]byte(out), &workflows); err != nil {
		result.Error = fmt.Sprintf("failed to parse workflows: %s", err)
		return result
	}

	enabled := 0
	for _, wf := range workflows {
		if wf.State != "disabled_inactivity" {
			continue
		}
		if workflowName != "" && wf.Name != workflowName {
			continue
		}

		_, err := ghRun("workflow", "enable",
			fmt.Sprintf("%d", wf.ID),
			flagRepo, fullRepo,
		)
		if err != nil {
			result.Error = fmt.Sprintf("failed to enable workflow %q: %s", wf.Name, err)
			return result
		}
		enabled++
	}

	result.EnabledCount = enabled

	if enabled > 0 && prBranch != "" {
		retriggered, err := retriggerPRs(fullRepo, prBranch)
		if err != nil {
			result.Error = fmt.Sprintf("enabled %d workflow(s) but retrigger failed: %s", enabled, err)
			return result
		}
		result.RetriggeredPRs = retriggered
	}

	result.Success = true
	return result
}

type ghPR struct {
	Number int `json:"number"`
}

func retriggerPRs(fullRepo, headBranch string) (int, error) {
	out, err := ghRun("pr", "list",
		flagRepo, fullRepo,
		"--head", headBranch,
		"--state", "open",
		flagJSON, "number",
	)
	if err != nil {
		return 0, err
	}

	if strings.TrimSpace(out) == "" {
		return 0, nil
	}

	var prs []ghPR
	if err := json.Unmarshal([]byte(out), &prs); err != nil {
		return 0, err
	}

	retriggered := 0
	for _, pr := range prs {
		_, err := ghRun("pr", "close", fmt.Sprintf("%d", pr.Number), flagRepo, fullRepo)
		if err != nil {
			return retriggered, fmt.Errorf("close PR #%d: %s", pr.Number, err)
		}

		_, err = ghRun("pr", "reopen", fmt.Sprintf("%d", pr.Number), flagRepo, fullRepo)
		if err != nil {
			return retriggered, fmt.Errorf("reopen PR #%d: %s", pr.Number, err)
		}

		retriggered++
	}
	return retriggered, nil
}
