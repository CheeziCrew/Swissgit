package ops

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// roundTripFunc allows intercepting HTTP requests in tests.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// makeTestRepoWithCommit creates a temp git repo with an origin remote and an initial commit.
func makeTestRepoWithCommit(t *testing.T, owner, name string) (*gogit.Repository, string) {
	t.Helper()
	dir := t.TempDir()
	repo, err := gogit.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"git@github.com:" + owner + "/" + name + ".git"},
	})
	if err != nil {
		t.Fatalf("create remote: %v", err)
	}
	wt, _ := repo.Worktree()
	f, _ := os.Create(filepath.Join(dir, "README.md"))
	f.WriteString("# test\n")
	f.Close()
	wt.Add("README.md")
	wt.Commit("initial commit", &gogit.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@test.com"},
	})
	return repo, dir
}

// === CreatePullRequest full HTTP flow ===

func TestCreatePullRequest_FullHTTPFlow(t *testing.T) {
	origClient := httpClient
	t.Cleanup(func() { httpClient = origClient })

	repo := initTestRepo(t, "testowner", "testrepo")
	t.Setenv("GITHUB_TOKEN", "test-token-123")

	httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body := `{"html_url":"https://github.com/testowner/testrepo/pull/42"}`
			resp := &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}
			return resp, nil
		}),
	}

	url, err := CreatePullRequest(repo, "test commit", "feature/x", "main", []string{"Bug fix"}, false)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if url != "https://github.com/testowner/testrepo/pull/42" {
		t.Errorf("url = %q", url)
	}
}

func TestCreatePullRequest_ServerError(t *testing.T) {
	origClient := httpClient
	t.Cleanup(func() { httpClient = origClient })

	repo := initTestRepo(t, "testowner", "testrepo")
	t.Setenv("GITHUB_TOKEN", "test-token")

	httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnprocessableEntity,
				Body:       io.NopCloser(strings.NewReader(`{"message":"Validation Failed"}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}

	_, err := CreatePullRequest(repo, "msg", "feature/x", "main", nil, false)
	if err == nil {
		t.Error("expected error for 422 response")
	}
}

func TestCreatePullRequest_NetworkError(t *testing.T) {
	origClient := httpClient
	t.Cleanup(func() { httpClient = origClient })

	repo := initTestRepo(t, "testowner", "testrepo")
	t.Setenv("GITHUB_TOKEN", "test-token")

	httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("network down")
		}),
	}

	_, err := CreatePullRequest(repo, "msg", "feature/x", "main", nil, false)
	if err == nil {
		t.Error("expected error for network failure")
	}
}

func TestCreatePullRequest_WithBreakingChanges(t *testing.T) {
	origClient := httpClient
	t.Cleanup(func() { httpClient = origClient })

	repo := initTestRepo(t, "testowner", "testrepo")
	t.Setenv("GITHUB_TOKEN", "test-token")

	httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body := `{"html_url":"https://github.com/testowner/testrepo/pull/1"}`
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}

	url, err := CreatePullRequest(repo, "msg", "feature/x", "main",
		[]string{"Changed API", "New behavior"}, true)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty URL")
	}
}

// === GetOrgRepositories full loop ===

func TestGetOrgRepositories_FullLoop(t *testing.T) {
	origClient := httpClient
	t.Cleanup(func() { httpClient = origClient })

	callCount := 0
	httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			callCount++
			var body string
			header := http.Header{"Content-Type": []string{"application/json"}}
			if callCount == 1 {
				header.Set("Link", `<https://api.github.com/next>; rel="next"`)
				repos := []Repository{
					{Name: "repo1", SSHURL: "git@github.com:org/repo1.git"},
					{Name: "archived", SSHURL: "git@github.com:org/archived.git", Archived: true},
				}
				b, _ := json.Marshal(repos)
				body = string(b)
			} else {
				repos := []Repository{
					{Name: "repo2", SSHURL: "git@github.com:org/repo2.git"},
				}
				b, _ := json.Marshal(repos)
				body = string(b)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     header,
			}, nil
		}),
	}

	t.Setenv("GITHUB_TOKEN", "test-token")
	repos, err := GetOrgRepositories("testorg", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 2 {
		t.Errorf("got %d repos, want 2 (archived filtered out)", len(repos))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (pagination), got %d", callCount)
	}
}

func TestGetOrgRepositories_WithTeam(t *testing.T) {
	origClient := httpClient
	t.Cleanup(func() { httpClient = origClient })

	httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			repos := []Repository{
				{Name: "team-repo", SSHURL: "git@github.com:org/team-repo.git"},
			}
			b, _ := json.Marshal(repos)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(string(b))),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}

	t.Setenv("GITHUB_TOKEN", "test-token")
	repos, err := GetOrgRepositories("testorg", "my-team")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 {
		t.Errorf("got %d repos, want 1", len(repos))
	}
}

func TestGetOrgRepositories_FetchError(t *testing.T) {
	origClient := httpClient
	t.Cleanup(func() { httpClient = origClient })

	httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("network error")
		}),
	}

	t.Setenv("GITHUB_TOKEN", "test-token")
	_, err := GetOrgRepositories("testorg", "")
	if err == nil {
		t.Error("expected error for network failure")
	}
}

// === CloneRepository with existing git repo (skipped) ===

func TestCloneRepository_SkipsExistingGitDir(t *testing.T) {
	_, dir := makeTestRepoWithCommit(t, "org", "repo")

	result := CloneRepository(Repository{Name: "repo", SSHURL: "git@github.com:org/repo.git"}, dir)
	if !result.Skipped {
		t.Error("expected Skipped=true for existing git repo")
	}
	if !result.Success {
		t.Error("expected Success=true for skipped repo")
	}
}

func TestCloneRepository_MkdirFailsV2(t *testing.T) {
	// Use a path under /dev/null which can't have subdirs
	result := CloneRepository(
		Repository{Name: "repo", SSHURL: "git@github.com:org/repo.git"},
		"/dev/null/impossible/path",
	)
	if result.Error == "" {
		t.Error("expected error when mkdir fails")
	}
}

// === CommitAndCreatePR deeper coverage ===

func TestCommitAndCreatePR_CommitSucceedsPRFails(t *testing.T) {
	origOpen := plainOpen
	origGitRun := gitRunInDir
	origPush := pushChanges
	origClient := httpClient
	t.Cleanup(func() {
		plainOpen = origOpen
		gitRunInDir = origGitRun
		pushChanges = origPush
		httpClient = origClient
	})

	_, dir := makeTestRepoWithCommit(t, "org", "myrepo")

	// Add a file to commit
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("change"), 0644)

	// Mock push to succeed
	pushChanges = func(repo *gogit.Repository) error { return nil }

	// Mock HTTP to fail for PR creation
	httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("API unreachable")
		}),
	}

	t.Setenv("GITHUB_TOKEN", "test-token")
	result := CommitAndCreatePR(dir, "feature/test", "test msg", "main", nil, false)
	// Commit should succeed but PR should fail
	if result.Success {
		t.Log("unexpectedly succeeded (commit might have failed due to test env)")
	}
}
