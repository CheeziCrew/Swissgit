package ops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// EnableWorkflowResult holds the outcome of checking/enabling workflows for a repo.
type EnableWorkflowResult struct {
	Repo            string
	EnabledCount    int
	RetriggeredPRs  int
	Success         bool
	Error           string
}

type ghWorkflow struct {
	Name  string `json:"name"`
	ID    int    `json:"id"`
	State string `json:"state"`
}

// FetchOrgRepoNames returns all non-archived repo names for an org.
func FetchOrgRepoNames(org string) ([]string, error) {
	cmd := exec.Command("gh", "repo", "list", org,
		"--limit", "500",
		"--no-archived",
		"--json", "name",
		"-q", ".[].name",
	)

	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gh repo list failed: %s", strings.TrimSpace(errBuf.String()))
	}

	var names []string
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
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

	// List all workflows
	cmd := exec.Command("gh", "workflow", "list",
		"--repo", fullRepo,
		"--all",
		"--json", "name,id,state",
	)

	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		result.Error = fmt.Sprintf("list workflows failed: %s", strings.TrimSpace(errBuf.String()))
		return result
	}

	var workflows []ghWorkflow
	if strings.TrimSpace(out.String()) == "" {
		result.Success = true
		return result
	}
	if err := json.Unmarshal(out.Bytes(), &workflows); err != nil {
		result.Error = fmt.Sprintf("failed to parse workflows: %s", err)
		return result
	}

	// Find and enable matching disabled workflows
	enabled := 0
	for _, wf := range workflows {
		if wf.State != "disabled_inactivity" {
			continue
		}
		if workflowName != "" && wf.Name != workflowName {
			continue
		}

		enableCmd := exec.Command("gh", "workflow", "enable",
			fmt.Sprintf("%d", wf.ID),
			"--repo", fullRepo,
		)
		if err := enableCmd.Run(); err != nil {
			result.Error = fmt.Sprintf("failed to enable workflow %q: %s", wf.Name, err)
			return result
		}
		enabled++
	}

	result.EnabledCount = enabled

	// Close/reopen matching PRs to retrigger workflow runs
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
	cmd := exec.Command("gh", "pr", "list",
		"--repo", fullRepo,
		"--head", headBranch,
		"--state", "open",
		"--json", "number",
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, err
	}

	if strings.TrimSpace(out.String()) == "" {
		return 0, nil
	}

	var prs []ghPR
	if err := json.Unmarshal(out.Bytes(), &prs); err != nil {
		return 0, err
	}

	retriggered := 0
	for _, pr := range prs {
		closeCmd := exec.Command("gh", "pr", "close", fmt.Sprintf("%d", pr.Number), "--repo", fullRepo)
		if err := closeCmd.Run(); err != nil {
			return retriggered, fmt.Errorf("close PR #%d: %s", pr.Number, err)
		}

		reopenCmd := exec.Command("gh", "pr", "reopen", fmt.Sprintf("%d", pr.Number), "--repo", fullRepo)
		if err := reopenCmd.Run(); err != nil {
			return retriggered, fmt.Errorf("reopen PR #%d: %s", pr.Number, err)
		}

		retriggered++
	}
	return retriggered, nil
}
