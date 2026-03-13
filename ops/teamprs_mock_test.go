package ops

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestFetchTeamRepoNames(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	t.Run("returns filtered repo names", func(t *testing.T) {
		ghRun = func(args ...string) (string, error) {
			if args[0] != "api" {
				t.Errorf("expected api call, got %v", args)
			}
			return "service-one\nservice-two\nweb-app-frontend\nservice-three\n", nil
		}

		names, err := FetchTeamRepoNames("myorg", "backend", []string{"web-app-"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 3 {
			t.Fatalf("got %d names, want 3", len(names))
		}
		for _, n := range names {
			if n == "web-app-frontend" {
				t.Error("expected web-app-frontend to be excluded")
			}
		}
	})

	t.Run("no exclusions", func(t *testing.T) {
		ghRun = func(args ...string) (string, error) {
			return "repo-a\nrepo-b\n", nil
		}

		names, err := FetchTeamRepoNames("myorg", "team", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 2 {
			t.Errorf("got %d names, want 2", len(names))
		}
	})

	t.Run("api error", func(t *testing.T) {
		ghRun = func(args ...string) (string, error) {
			return "", fmt.Errorf("forbidden")
		}

		_, err := FetchTeamRepoNames("myorg", "team", nil)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("empty response", func(t *testing.T) {
		ghRun = func(args ...string) (string, error) {
			return "", nil
		}

		names, err := FetchTeamRepoNames("myorg", "team", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 0 {
			t.Errorf("got %d names, want 0", len(names))
		}
	})
}

func TestFetchTeamPRs(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	t.Run("filters to team repos only", func(t *testing.T) {
		now := time.Now()
		results := []ghTeamPRResult{
			{
				Repository: struct {
					Name string `json:"name"`
				}{Name: "service-one"},
				Number:    1,
				Title:     "Fix bug",
				Author:    struct{ Login string `json:"login"` }{Login: "alice"},
				URL:       "https://github.com/myorg/service-one/pull/1",
				IsDraft:   false,
				CreatedAt: now,
			},
			{
				Repository: struct {
					Name string `json:"name"`
				}{Name: "other-repo"},
				Number:    2,
				Title:     "Add feature",
				Author:    struct{ Login string `json:"login"` }{Login: "bob"},
				URL:       "https://github.com/myorg/other-repo/pull/2",
				IsDraft:   true,
				CreatedAt: now,
			},
		}
		jsonData, _ := json.Marshal(results)

		ghRun = func(args ...string) (string, error) {
			if args[0] != "search" {
				t.Errorf("expected search call, got %v", args)
			}
			return string(jsonData), nil
		}

		prs, err := FetchTeamPRs("myorg", []string{"service-one"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prs) != 1 {
			t.Fatalf("got %d PRs, want 1", len(prs))
		}
		if prs[0].Repo != "service-one" {
			t.Errorf("Repo = %q, want %q", prs[0].Repo, "service-one")
		}
		if prs[0].Author != "alice" {
			t.Errorf("Author = %q, want %q", prs[0].Author, "alice")
		}
		if prs[0].Number != 1 {
			t.Errorf("Number = %d, want 1", prs[0].Number)
		}
	})

	t.Run("search error", func(t *testing.T) {
		ghRun = func(args ...string) (string, error) {
			return "", fmt.Errorf("rate limited")
		}

		_, err := FetchTeamPRs("myorg", []string{"repo"})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("empty results", func(t *testing.T) {
		ghRun = func(args ...string) (string, error) {
			return "[]", nil
		}

		prs, err := FetchTeamPRs("myorg", []string{"repo"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prs) != 0 {
			t.Errorf("got %d PRs, want 0", len(prs))
		}
	})
}
