package ops

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

// initTestRepo creates a git repo with a GitHub-style origin remote
// so that GetRepoOwnerAndName can extract owner/name from it.
func initTestRepo(t *testing.T, owner, name string) *gogit.Repository {
	t.Helper()
	dir := t.TempDir()
	repo, err := gogit.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"git@github.com:" + owner + "/" + name + ".git"},
	})
	if err != nil {
		t.Fatalf("failed to create remote: %v", err)
	}
	return repo
}

func TestCreatePullRequest_HTTPMock(t *testing.T) {
	origClient := httpClient
	t.Cleanup(func() { httpClient = origClient })

	t.Run("request body and response parsing", func(t *testing.T) {
		var receivedReq PRRequest
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("Authorization") != "token test-token" {
				t.Errorf("unexpected auth header: %q", r.Header.Get("Authorization"))
			}

			json.NewDecoder(r.Body).Decode(&receivedReq)

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"html_url": "https://github.com/myorg/myrepo/pull/42",
			})
		}))
		defer srv.Close()

		httpClient = srv.Client()

		body, err := BuildPullRequestBody([]string{"Bug fix"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pr := PRRequest{
			Title: "feature/test: Fix the thing",
			Head:  "feature/test",
			Body:  body,
			Base:  "main",
		}

		jsonData, err := json.Marshal(pr)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		req, err := http.NewRequest("POST", srv.URL+"/repos/myorg/myrepo/pulls", strings.NewReader(string(jsonData)))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Authorization", "token test-token")
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
		}

		var result struct {
			HTMLURL string `json:"html_url"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		if result.HTMLURL != "https://github.com/myorg/myrepo/pull/42" {
			t.Errorf("html_url = %q, want %q", result.HTMLURL, "https://github.com/myorg/myrepo/pull/42")
		}

		if receivedReq.Title != "feature/test: Fix the thing" {
			t.Errorf("Title = %q, want %q", receivedReq.Title, "feature/test: Fix the thing")
		}
		if receivedReq.Head != "feature/test" {
			t.Errorf("Head = %q, want %q", receivedReq.Head, "feature/test")
		}
		if receivedReq.Base != "main" {
			t.Errorf("Base = %q, want %q", receivedReq.Base, "main")
		}
		if !strings.Contains(receivedReq.Body, "- [x] Bug fix") {
			t.Error("expected body to contain checked Bug fix")
		}
	})

	t.Run("server returns error status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"message":"Validation Failed"}`))
		}))
		defer srv.Close()

		httpClient = srv.Client()

		resp, err := httpClient.Post(srv.URL, "application/json", strings.NewReader("{}"))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			t.Error("expected non-201 status")
		}
	})

	t.Run("CreatePullRequest fails without GITHUB_TOKEN", func(t *testing.T) {
		repo := initTestRepo(t, "testowner", "testrepo")
		t.Setenv("GITHUB_TOKEN", "")

		_, err := CreatePullRequest(repo, "commit msg", "feature/x", "main", nil, false)
		if err == nil {
			t.Error("expected error when GITHUB_TOKEN is empty")
		}
		if !strings.Contains(err.Error(), "GITHUB_TOKEN") {
			t.Errorf("expected GITHUB_TOKEN error, got: %v", err)
		}
	})
}
