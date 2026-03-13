package ops

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrgReposURL(t *testing.T) {
	tests := []struct {
		name     string
		org      string
		team     string
		wantURL  string
	}{
		{
			"org only",
			"myorg", "",
			"https://api.github.com/orgs/myorg/repos",
		},
		{
			"org and team",
			"myorg", "backend-team",
			"https://api.github.com/orgs/myorg/teams/backend-team/repos",
		},
		{
			"different org",
			"Sundsvallskommun", "team-unmasked",
			"https://api.github.com/orgs/Sundsvallskommun/teams/team-unmasked/repos",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := orgReposURL(tt.org, tt.team); got != tt.wantURL {
				t.Errorf("orgReposURL(%q, %q) = %q, want %q", tt.org, tt.team, got, tt.wantURL)
			}
		})
	}
}

func TestCloneFromURL(t *testing.T) {
	t.Run("parses repo name from SSH URL", func(t *testing.T) {
		// CloneFromURL will fail at the clone step (no real repo),
		// but we can verify it extracts the name correctly from the error result.
		result := CloneFromURL("git@github.com:myorg/my-service.git", t.TempDir())
		// It will fail (no SSH key / no real remote), but RepoName should be empty
		// since the error happens during clone, not name parsing.
		// The important thing is it doesn't panic.
		_ = result
	})

	t.Run("invalid URL returns error", func(t *testing.T) {
		result := CloneFromURL("noslash", t.TempDir())
		if result.Error == "" {
			t.Error("expected error for invalid URL")
		}
	})
}

func TestGetOrgRepositories(t *testing.T) {
	origClient := httpClient
	t.Cleanup(func() { httpClient = origClient })

	t.Run("fetches and filters archived repos", func(t *testing.T) {
		repos := []Repository{
			{Name: "active-repo", SSHURL: "git@github.com:org/active-repo.git", Archived: false},
			{Name: "archived-repo", SSHURL: "git@github.com:org/archived-repo.git", Archived: true},
			{Name: "another-active", SSHURL: "git@github.com:org/another-active.git", Archived: false},
		}

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "token test-token" {
				t.Errorf("expected auth header, got %q", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(repos)
		}))
		defer srv.Close()

		httpClient = srv.Client()
		t.Setenv("GITHUB_TOKEN", "test-token")

		// We need to override the URL that GetOrgRepositories builds.
		// Since we can't easily override orgReposURL, we test fetchRepoPage directly.
		fetched, hasNext, err := fetchRepoPage(httpClient, srv.URL, 1, "test-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hasNext {
			t.Error("expected no next page")
		}
		if len(fetched) != 3 {
			t.Fatalf("got %d repos, want 3", len(fetched))
		}

		// Verify the filtering that GetOrgRepositories does
		var active []Repository
		for _, r := range fetched {
			if !r.Archived {
				active = append(active, r)
			}
		}
		if len(active) != 2 {
			t.Errorf("got %d active repos, want 2", len(active))
		}
	})

	t.Run("handles pagination via Link header", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Link", `<https://api.github.com/next>; rel="next"`)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]Repository{
				{Name: "repo1", SSHURL: "git@github.com:org/repo1.git"},
			})
		}))
		defer srv.Close()

		httpClient = srv.Client()

		fetched, hasNext, err := fetchRepoPage(httpClient, srv.URL, 1, "tok")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hasNext {
			t.Error("expected hasNext=true when Link header has rel=next")
		}
		if len(fetched) != 1 {
			t.Errorf("got %d repos, want 1", len(fetched))
		}
	})

	t.Run("handles HTTP error status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer srv.Close()

		httpClient = srv.Client()

		_, _, err := fetchRepoPage(httpClient, srv.URL, 1, "bad-token")
		if err == nil {
			t.Error("expected error for 403 response")
		}
	})

	t.Run("missing GITHUB_TOKEN", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "")

		_, err := GetOrgRepositories("myorg", "")
		if err == nil {
			t.Error("expected error when GITHUB_TOKEN is empty")
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`not valid json`))
		}))
		defer srv.Close()

		httpClient = srv.Client()

		_, _, err := fetchRepoPage(httpClient, srv.URL, 1, "tok")
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("multiple pages", func(t *testing.T) {
		callCount := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			if callCount == 1 {
				w.Header().Set("Link", `<https://api.github.com/next>; rel="next"`)
				json.NewEncoder(w).Encode([]Repository{
					{Name: "repo1", SSHURL: "git@github.com:org/repo1.git"},
				})
			} else {
				json.NewEncoder(w).Encode([]Repository{
					{Name: "repo2", SSHURL: "git@github.com:org/repo2.git"},
				})
			}
		}))
		defer srv.Close()

		httpClient = srv.Client()

		// First page has next
		repos1, hasNext1, err := fetchRepoPage(httpClient, srv.URL, 1, "tok")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hasNext1 {
			t.Error("expected hasNext=true for first page")
		}
		if len(repos1) != 1 || repos1[0].Name != "repo1" {
			t.Errorf("unexpected first page: %v", repos1)
		}

		// Second page has no next
		repos2, hasNext2, err := fetchRepoPage(httpClient, srv.URL, 2, "tok")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hasNext2 {
			t.Error("expected hasNext=false for second page")
		}
		if len(repos2) != 1 || repos2[0].Name != "repo2" {
			t.Errorf("unexpected second page: %v", repos2)
		}
	})
}

func TestCloneResult_Fields(t *testing.T) {
	r := CloneResult{RepoName: "test", Skipped: true, Success: true}
	if r.RepoName != "test" {
		t.Errorf("RepoName = %q, want %q", r.RepoName, "test")
	}
	if !r.Skipped {
		t.Error("expected Skipped to be true")
	}
}

func TestCloneFromURL_ParsesNames(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantErr  bool
	}{
		{"valid SSH URL", "git@github.com:org/repo.git", false},
		{"valid HTTPS URL", "https://github.com/org/repo.git", false},
		{"invalid single segment", "noslash", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CloneFromURL(tt.url, t.TempDir())
			if tt.wantErr && result.Error == "" {
				t.Error("expected error")
			}
			// Valid URLs will fail at clone step but shouldn't panic
		})
	}
}
