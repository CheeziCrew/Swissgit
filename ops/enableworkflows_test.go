package ops

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestFetchOrgRepoNames(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	t.Run("returns repo names", func(t *testing.T) {
		ghRun = func(args ...string) (string, error) {
			if args[0] != "repo" || args[1] != "list" {
				t.Errorf("expected repo list, got %v", args)
			}
			return "repo-alpha\nrepo-beta\nrepo-gamma\n", nil
		}

		names, err := FetchOrgRepoNames("myorg")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 3 {
			t.Fatalf("got %d names, want 3", len(names))
		}
		if names[0] != "repo-alpha" || names[1] != "repo-beta" || names[2] != "repo-gamma" {
			t.Errorf("unexpected names: %v", names)
		}
	})

	t.Run("empty output", func(t *testing.T) {
		ghRun = func(args ...string) (string, error) {
			return "", nil
		}

		names, err := FetchOrgRepoNames("myorg")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 0 {
			t.Errorf("got %d names, want 0", len(names))
		}
	})

	t.Run("gh error", func(t *testing.T) {
		ghRun = func(args ...string) (string, error) {
			return "", fmt.Errorf("not authorized")
		}

		_, err := FetchOrgRepoNames("myorg")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestFindAndEnableWorkflows(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	t.Run("enables disabled_inactivity workflows", func(t *testing.T) {
		workflows := []ghWorkflow{
			{Name: "CI", ID: 101, State: "active"},
			{Name: "Deploy", ID: 102, State: "disabled_inactivity"},
			{Name: "Lint", ID: 103, State: "disabled_inactivity"},
		}
		wfJSON, _ := json.Marshal(workflows)

		enabledIDs := []string{}
		ghRun = func(args ...string) (string, error) {
			if args[0] == "workflow" && args[1] == "list" {
				return string(wfJSON), nil
			}
			if args[0] == "workflow" && args[1] == "enable" {
				enabledIDs = append(enabledIDs, args[2])
				return "", nil
			}
			return "", fmt.Errorf("unexpected call: %v", args)
		}

		result := FindAndEnableWorkflows("myorg", "myrepo", "", "")
		if !result.Success {
			t.Errorf("expected success, got error: %s", result.Error)
		}
		if result.EnabledCount != 2 {
			t.Errorf("EnabledCount = %d, want 2", result.EnabledCount)
		}
		if len(enabledIDs) != 2 {
			t.Errorf("expected 2 enable calls, got %d", len(enabledIDs))
		}
	})

	t.Run("filters by workflow name", func(t *testing.T) {
		workflows := []ghWorkflow{
			{Name: "Deploy", ID: 102, State: "disabled_inactivity"},
			{Name: "Lint", ID: 103, State: "disabled_inactivity"},
		}
		wfJSON, _ := json.Marshal(workflows)

		enabledIDs := []string{}
		ghRun = func(args ...string) (string, error) {
			if args[0] == "workflow" && args[1] == "list" {
				return string(wfJSON), nil
			}
			if args[0] == "workflow" && args[1] == "enable" {
				enabledIDs = append(enabledIDs, args[2])
				return "", nil
			}
			return "", fmt.Errorf("unexpected call: %v", args)
		}

		result := FindAndEnableWorkflows("myorg", "myrepo", "Deploy", "")
		if !result.Success {
			t.Errorf("expected success, got error: %s", result.Error)
		}
		if result.EnabledCount != 1 {
			t.Errorf("EnabledCount = %d, want 1", result.EnabledCount)
		}
	})

	t.Run("no disabled workflows", func(t *testing.T) {
		workflows := []ghWorkflow{
			{Name: "CI", ID: 101, State: "active"},
		}
		wfJSON, _ := json.Marshal(workflows)

		ghRun = func(args ...string) (string, error) {
			if args[0] == "workflow" && args[1] == "list" {
				return string(wfJSON), nil
			}
			return "", fmt.Errorf("unexpected call: %v", args)
		}

		result := FindAndEnableWorkflows("myorg", "myrepo", "", "")
		if !result.Success {
			t.Errorf("expected success, got error: %s", result.Error)
		}
		if result.EnabledCount != 0 {
			t.Errorf("EnabledCount = %d, want 0", result.EnabledCount)
		}
	})

	t.Run("empty workflow list", func(t *testing.T) {
		ghRun = func(args ...string) (string, error) {
			if args[0] == "workflow" && args[1] == "list" {
				return "", nil
			}
			return "", fmt.Errorf("unexpected call: %v", args)
		}

		result := FindAndEnableWorkflows("myorg", "myrepo", "", "")
		if !result.Success {
			t.Errorf("expected success, got error: %s", result.Error)
		}
	})

	t.Run("list error", func(t *testing.T) {
		ghRun = func(args ...string) (string, error) {
			return "", fmt.Errorf("network error")
		}

		result := FindAndEnableWorkflows("myorg", "myrepo", "", "")
		if result.Success {
			t.Error("expected failure")
		}
		if !strings.Contains(result.Error, "list workflows failed") {
			t.Errorf("unexpected error: %s", result.Error)
		}
	})

	t.Run("enables workflows and retriggers PRs", func(t *testing.T) {
		workflows := []ghWorkflow{
			{Name: "CI", ID: 101, State: "disabled_inactivity"},
		}
		wfJSON, _ := json.Marshal(workflows)

		prs := []ghPR{{Number: 10}, {Number: 20}}
		prJSON, _ := json.Marshal(prs)

		ghRun = func(args ...string) (string, error) {
			if args[0] == "workflow" && args[1] == "list" {
				return string(wfJSON), nil
			}
			if args[0] == "workflow" && args[1] == "enable" {
				return "", nil
			}
			if args[0] == "pr" && args[1] == "list" {
				return string(prJSON), nil
			}
			if args[0] == "pr" && (args[1] == "close" || args[1] == "reopen") {
				return "", nil
			}
			return "", fmt.Errorf("unexpected call: %v", args)
		}

		result := FindAndEnableWorkflows("myorg", "myrepo", "", "feature/branch")
		if !result.Success {
			t.Errorf("expected success, got error: %s", result.Error)
		}
		if result.EnabledCount != 1 {
			t.Errorf("EnabledCount = %d, want 1", result.EnabledCount)
		}
		if result.RetriggeredPRs != 2 {
			t.Errorf("RetriggeredPRs = %d, want 2", result.RetriggeredPRs)
		}
	})
}
