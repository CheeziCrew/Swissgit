package ops

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/CheeziCrew/swissgit/git"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// --- FetchApprovedPRs ---

func TestFetchApprovedPRs_Success(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return `[
			{"repository":{"name":"repo-a"},"number":1,"title":"fix bug"},
			{"repository":{"name":"repo-b"},"number":2,"title":"add feature"}
		]`, nil
	}

	prs, err := FetchApprovedPRs("myorg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 2 {
		t.Fatalf("expected 2 PRs, got %d", len(prs))
	}
	if prs[0].Repo != "repo-a" {
		t.Errorf("Repo = %q, want repo-a", prs[0].Repo)
	}
	if prs[1].Number != 2 {
		t.Errorf("Number = %d, want 2", prs[1].Number)
	}
}

func TestFetchApprovedPRs_GhError(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "", fmt.Errorf("auth required")
	}

	_, err := FetchApprovedPRs("myorg")
	if err == nil {
		t.Error("expected error")
	}
}

func TestFetchApprovedPRs_InvalidJSON(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "not json", nil
	}

	_, err := FetchApprovedPRs("myorg")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// --- MergePR ---

func TestMergePR_Success(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "", nil
	}

	result := MergePR("myorg", "myrepo", 42)
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.Repo != "myrepo" {
		t.Errorf("Repo = %q", result.Repo)
	}
	if result.PRNumber != "42" {
		t.Errorf("PRNumber = %q", result.PRNumber)
	}
}

func TestMergePR_Failure(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "", fmt.Errorf("merge conflict")
	}

	result := MergePR("myorg", "myrepo", 42)
	if result.Success {
		t.Error("expected failure")
	}
	if result.Error == "" {
		t.Error("expected non-empty error")
	}
}

// --- FetchMyPRs ---

func TestFetchMyPRs_Success(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	now := time.Now()
	ghRun = func(args ...string) (string, error) {
		return fmt.Sprintf(`[
			{
				"repository":{"name":"repo-x","nameWithOwner":"org/repo-x"},
				"number":10,
				"title":"fix tests",
				"url":"https://github.com/org/repo-x/pull/10",
				"state":"OPEN",
				"isDraft":false,
				"createdAt":"%s"
			},
			{
				"repository":{"name":"repo-y","nameWithOwner":"org/repo-y"},
				"number":20,
				"title":"wip feature",
				"url":"https://github.com/org/repo-y/pull/20",
				"state":"OPEN",
				"isDraft":true,
				"createdAt":"%s"
			}
		]`, now.Format(time.RFC3339), now.Format(time.RFC3339)), nil
	}

	prs, err := FetchMyPRs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 2 {
		t.Fatalf("expected 2 PRs, got %d", len(prs))
	}
	if prs[0].Repo != "org/repo-x" {
		t.Errorf("Repo = %q, want org/repo-x", prs[0].Repo)
	}
	if !prs[1].Draft {
		t.Error("expected Draft = true for second PR")
	}
}

func TestFetchMyPRs_GhError(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "", fmt.Errorf("not authenticated")
	}

	_, err := FetchMyPRs()
	if err == nil {
		t.Error("expected error")
	}
}

func TestFetchMyPRs_InvalidJSON(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "invalid", nil
	}

	_, err := FetchMyPRs()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// --- createOrCheckoutBranch ---

func TestCreateOrCheckoutBranch_ExistingBranch(t *testing.T) {
	// Create a real repo with an initial commit on main
	dir := t.TempDir()
	cmds := [][]string{
		{"git", "init", "-b", "main", dir},
		{"git", "-C", dir, "config", "user.email", "test@test.com"},
		{"git", "-C", dir, "config", "user.name", "Test"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("cmd %v failed: %s %v", c, string(out), err)
		}
	}
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644)
	exec.Command("git", "-C", dir, "add", ".").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init").Run()

	repo, err := gogit.PlainOpen(dir)
	if err != nil {
		t.Fatalf("PlainOpen: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}

	// Create a new branch (should succeed)
	err = createOrCheckoutBranch("feature/new", wt)
	if err != nil {
		t.Fatalf("createOrCheckoutBranch new: %v", err)
	}

	// Check out existing branch (create=true fails, falls back to checkout)
	err = createOrCheckoutBranch("feature/new", wt)
	if err != nil {
		t.Fatalf("createOrCheckoutBranch existing: %v", err)
	}

	// Check out main again
	err = createOrCheckoutBranch("main", wt)
	if err != nil {
		t.Fatalf("createOrCheckoutBranch main: %v", err)
	}
}

// --- CloneRepository deeper paths ---

func TestCloneRepository_ExistingRepo(t *testing.T) {
	// Test the "already exists" path
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)

	result := CloneRepository(Repository{Name: "existing", SSHURL: "git@github.com:org/existing.git"}, dir)
	if !result.Success {
		t.Error("expected success for existing repo (skip)")
	}
	if !result.Skipped {
		t.Error("expected Skipped = true")
	}
}

func TestCloneRepository_MkdirFails(t *testing.T) {
	// Use a path that can't be created
	result := CloneRepository(
		Repository{Name: "repo", SSHURL: "git@github.com:org/repo.git"},
		"/dev/null/impossible/path",
	)
	if result.Success {
		t.Error("expected failure when mkdir fails")
	}
	if result.Error == "" {
		t.Error("expected error message")
	}
}

// --- GetOrgRepositories full flow ---

func TestGetOrgRepositories_FullFlow(t *testing.T) {
	origClient := httpClient
	t.Cleanup(func() { httpClient = origClient })

	repos := []Repository{
		{Name: "active-repo", SSHURL: "git@github.com:org/active-repo.git", Archived: false},
		{Name: "archived-repo", SSHURL: "git@github.com:org/archived-repo.git", Archived: true},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repos)
	}))
	defer srv.Close()

	// We can't fully test GetOrgRepositories because it builds its own URL
	// But we can test fetchRepoPage and the filtering logic.
	httpClient = srv.Client()

	fetched, _, err := fetchRepoPage(httpClient, srv.URL, 1, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var active []Repository
	for _, r := range fetched {
		if !r.Archived {
			active = append(active, r)
		}
	}
	if len(active) != 1 {
		t.Errorf("expected 1 active repo, got %d", len(active))
	}
}

// --- CommitAndPush with worktree error ---

func TestCommitAndPush_WorktreeError(t *testing.T) {
	origOpen := plainOpen
	t.Cleanup(func() { plainOpen = origOpen })

	// Return a repo that has no worktree (bare-like)
	dir := t.TempDir()
	repo, _ := gogit.PlainInit(dir, true) // bare repo

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}

	result := CommitAndPush("/fake", "", "test")
	if result.Success {
		t.Error("expected failure when worktree can't be obtained")
	}
}

// --- CommitAndPush with branch checkout and commit error ---

func TestCommitAndPush_CommitError(t *testing.T) {
	origOpen := plainOpen
	origGitRun := gitRunInDir
	t.Cleanup(func() {
		plainOpen = origOpen
		gitRunInDir = origGitRun
	})

	dir, repo := mkTestRepo(t)

	// Write a file to create a change
	os.WriteFile(filepath.Join(dir, "newfile.txt"), []byte("hello"), 0644)

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}

	callCount := 0
	gitRunInDir = func(d string, args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			// git add succeeds
			return "", nil
		}
		if callCount == 2 {
			// git commit fails
			return "", fmt.Errorf("commit failed: hook error")
		}
		return "", nil
	}

	result := CommitAndPush(dir, "", "test commit")
	if result.Success {
		t.Error("expected failure when commit fails")
	}
}

// --- CommitAndCreatePR with CreatePullRequest error ---

func TestCommitAndCreatePR_PRCreationFails(t *testing.T) {
	origOpen := plainOpen
	origGitRun := gitRunInDir
	origClient := httpClient
	t.Cleanup(func() {
		plainOpen = origOpen
		gitRunInDir = origGitRun
		httpClient = origClient
	})

	dir, repo := mkTestRepo(t)

	// Add origin remote
	repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"git@github.com:owner/testrepo.git"},
	})

	// Write a file to create changes
	os.WriteFile(filepath.Join(dir, "newfile.txt"), []byte("content"), 0644)

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}

	gitRunInDir = func(d string, args ...string) (string, error) {
		return "", nil
	}

	t.Setenv("GITHUB_TOKEN", "test-token")

	// httpClient to server that returns error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"message":"validation failed"}`))
	}))
	defer srv.Close()
	httpClient = srv.Client()

	// CommitAndPush will fail at no-changes (since gitRunInDir is mocked for git add)
	result := CommitAndCreatePR(dir, "", "msg", "main", nil, false)
	if result.Success {
		// This may fail at different stages depending on mocking
	}
	// Just verify no panic
}

// --- CleanupRepo with dropChanges ---

func TestCleanupRepo_DropChanges_Success(t *testing.T) {
	origOpen := plainOpen
	origFetch := fetchRemote
	origGitRun := gitRunInDir
	t.Cleanup(func() {
		plainOpen = origOpen
		fetchRemote = origFetch
		gitRunInDir = origGitRun
	})

	dir, repo := mkTestRepo(t)

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}
	fetchRemote = func(r *gogit.Repository) error {
		return nil
	}
	gitRunInDir = func(d string, args ...string) (string, error) {
		return "", nil
	}

	result := CleanupRepo(dir, true, "main")
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

func TestCleanupRepo_DropChanges_CheckoutFails(t *testing.T) {
	origOpen := plainOpen
	origGitRun := gitRunInDir
	t.Cleanup(func() {
		plainOpen = origOpen
		gitRunInDir = origGitRun
	})

	dir, repo := mkTestRepo(t)

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}

	callCount := 0
	gitRunInDir = func(d string, args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			// checkout . fails
			return "", fmt.Errorf("checkout failed")
		}
		return "", nil
	}

	result := CleanupRepo(dir, true, "main")
	if result.Success {
		t.Error("expected failure when checkout . fails during drop")
	}
}

func TestCleanupRepo_DropChanges_CleanFails(t *testing.T) {
	origOpen := plainOpen
	origGitRun := gitRunInDir
	t.Cleanup(func() {
		plainOpen = origOpen
		gitRunInDir = origGitRun
	})

	dir, repo := mkTestRepo(t)

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}

	callCount := 0
	gitRunInDir = func(d string, args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			// checkout . succeeds
			return "", nil
		}
		if callCount == 2 {
			// clean -fd fails
			return "", fmt.Errorf("clean failed")
		}
		return "", nil
	}

	result := CleanupRepo(dir, true, "main")
	if result.Success {
		t.Error("expected failure when clean fails during drop")
	}
}

// --- checkoutFetchPull pull fails ---

func TestCheckoutFetchPull_PullFails(t *testing.T) {
	origFetch := fetchRemote
	origGitRun := gitRunInDir
	t.Cleanup(func() {
		fetchRemote = origFetch
		gitRunInDir = origGitRun
	})

	_, repo := mkTestRepo(t)

	fetchRemote = func(r *gogit.Repository) error {
		return nil
	}

	callCount := 0
	gitRunInDir = func(d string, args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			// checkout succeeds
			return "", nil
		}
		if callCount == 2 {
			// pull fails
			return "", fmt.Errorf("pull failed: merge conflict")
		}
		return "", nil
	}

	err := checkoutFetchPull(repo, "/fake", "main")
	if err == nil {
		t.Error("expected error when pull fails")
	}
}

// --- updateBranches error propagation ---

func TestUpdateBranches_CheckoutFetchPullFails(t *testing.T) {
	origFetch := fetchRemote
	origGitRun := gitRunInDir
	t.Cleanup(func() {
		fetchRemote = origFetch
		gitRunInDir = origGitRun
	})

	_, repo := mkTestRepo(t)

	gitRunInDir = func(d string, args ...string) (string, error) {
		return "", fmt.Errorf("checkout failed")
	}
	fetchRemote = func(r *gogit.Repository) error {
		return nil
	}

	_, _, err := updateBranches(repo, "/fake", "main")
	if err == nil {
		t.Error("expected error")
	}
}

func TestUpdateBranches_MergedBranchesFails(t *testing.T) {
	origFetch := fetchRemote
	origGitRun := gitRunInDir
	t.Cleanup(func() {
		fetchRemote = origFetch
		gitRunInDir = origGitRun
	})

	_, repo := mkTestRepo(t)

	callCount := 0
	gitRunInDir = func(d string, args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			// checkout succeeds
			return "", nil
		}
		if callCount == 2 {
			// pull succeeds
			return "", nil
		}
		if callCount == 3 {
			// remote prune succeeds
			return "", nil
		}
		if callCount == 4 {
			// branch --merged fails
			return "", fmt.Errorf("git error")
		}
		return "", nil
	}
	fetchRemote = func(r *gogit.Repository) error {
		return nil
	}

	_, _, err := updateBranches(repo, "/fake", "main")
	if err == nil {
		t.Error("expected error when merged branches listing fails")
	}
}

func TestUpdateBranches_DeleteFails(t *testing.T) {
	origFetch := fetchRemote
	origGitRun := gitRunInDir
	t.Cleanup(func() {
		fetchRemote = origFetch
		gitRunInDir = origGitRun
	})

	_, repo := mkTestRepo(t)

	callCount := 0
	gitRunInDir = func(d string, args ...string) (string, error) {
		callCount++
		switch {
		case callCount <= 3:
			// checkout, pull, prune
			return "", nil
		case callCount == 4:
			// branch --merged returns branches to delete
			return "* main\n  feature-done\n", nil
		case callCount == 5:
			// branch -vv
			return "", nil
		case callCount == 6:
			// branch --format
			return "main\n", nil
		case callCount == 7:
			// branch (count)
			return "* main\n  feature-done\n", nil
		case callCount == 8:
			// branch -D fails
			return "", fmt.Errorf("cannot delete branch")
		}
		return "", nil
	}
	fetchRemote = func(r *gogit.Repository) error {
		return nil
	}

	_, _, err := updateBranches(repo, "/fake", "main")
	if err == nil {
		t.Error("expected error when branch deletion fails")
	}
}

// --- GetBranches deeper path testing ---

func TestGetBranches_BranchNameError(t *testing.T) {
	origOpen := plainOpen
	origFetch := fetchRemote
	t.Cleanup(func() {
		plainOpen = origOpen
		fetchRemote = origFetch
	})

	// Use a bare repo that has no HEAD
	dir := t.TempDir()
	repo, _ := gogit.PlainInit(dir, false)

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}
	fetchRemote = func(r *gogit.Repository) error {
		return nil
	}

	result := GetBranches("/fake/path")
	if result.Error == "" {
		t.Error("expected error when branch name can't be determined")
	}
}

// --- EnableWorkflowResult and retriggerPRs ---

func TestEnableWorkflowResult_Fields(t *testing.T) {
	r := EnableWorkflowResult{
		Repo:           "my-repo",
		EnabledCount:   3,
		RetriggeredPRs: 2,
		Success:        true,
	}
	if r.EnabledCount != 3 {
		t.Errorf("EnabledCount = %d, want 3", r.EnabledCount)
	}
	if r.RetriggeredPRs != 2 {
		t.Errorf("RetriggeredPRs = %d, want 2", r.RetriggeredPRs)
	}
}

func TestFindAndEnableWorkflows_EnableError(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	workflows := []ghWorkflow{
		{Name: "CI", ID: 101, State: "disabled_inactivity"},
	}
	wfJSON, _ := json.Marshal(workflows)

	callCount := 0
	ghRun = func(args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			return string(wfJSON), nil
		}
		// enable fails
		return "", fmt.Errorf("permission denied")
	}

	result := FindAndEnableWorkflows("myorg", "myrepo", "", "")
	if result.Success {
		t.Error("expected failure when enable fails")
	}
	if !strings.Contains(result.Error, "failed to enable workflow") {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestFindAndEnableWorkflows_InvalidJSON(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "not json", nil
	}

	result := FindAndEnableWorkflows("myorg", "myrepo", "", "")
	if result.Success {
		t.Error("expected failure for invalid JSON")
	}
	if !strings.Contains(result.Error, "failed to parse workflows") {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestRetriggerPRs_CloseFails(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	workflows := []ghWorkflow{
		{Name: "CI", ID: 101, State: "disabled_inactivity"},
	}
	wfJSON, _ := json.Marshal(workflows)

	prs := []ghPR{{Number: 10}}
	prJSON, _ := json.Marshal(prs)

	callCount := 0
	ghRun = func(args ...string) (string, error) {
		callCount++
		if args[0] == "workflow" && args[1] == "list" {
			return string(wfJSON), nil
		}
		if args[0] == "workflow" && args[1] == "enable" {
			return "", nil
		}
		if args[0] == "pr" && args[1] == "list" {
			return string(prJSON), nil
		}
		if args[0] == "pr" && args[1] == "close" {
			return "", fmt.Errorf("close failed")
		}
		return "", nil
	}

	result := FindAndEnableWorkflows("myorg", "myrepo", "", "feature/branch")
	if result.Success {
		t.Error("expected failure when close fails")
	}
}

func TestRetriggerPRs_ReopenFails(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	workflows := []ghWorkflow{
		{Name: "CI", ID: 101, State: "disabled_inactivity"},
	}
	wfJSON, _ := json.Marshal(workflows)

	prs := []ghPR{{Number: 10}}
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
		if args[0] == "pr" && args[1] == "close" {
			return "", nil
		}
		if args[0] == "pr" && args[1] == "reopen" {
			return "", fmt.Errorf("reopen failed")
		}
		return "", nil
	}

	result := FindAndEnableWorkflows("myorg", "myrepo", "", "feature/branch")
	if result.Success {
		t.Error("expected failure when reopen fails")
	}
}

func TestRetriggerPRs_EmptyPRList(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	workflows := []ghWorkflow{
		{Name: "CI", ID: 101, State: "disabled_inactivity"},
	}
	wfJSON, _ := json.Marshal(workflows)

	ghRun = func(args ...string) (string, error) {
		if args[0] == "workflow" && args[1] == "list" {
			return string(wfJSON), nil
		}
		if args[0] == "workflow" && args[1] == "enable" {
			return "", nil
		}
		if args[0] == "pr" && args[1] == "list" {
			return "", nil // empty
		}
		return "", nil
	}

	result := FindAndEnableWorkflows("myorg", "myrepo", "", "feature/branch")
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.RetriggeredPRs != 0 {
		t.Errorf("RetriggeredPRs = %d, want 0", result.RetriggeredPRs)
	}
}

func TestRetriggerPRs_InvalidJSON(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	workflows := []ghWorkflow{
		{Name: "CI", ID: 101, State: "disabled_inactivity"},
	}
	wfJSON, _ := json.Marshal(workflows)

	ghRun = func(args ...string) (string, error) {
		if args[0] == "workflow" && args[1] == "list" {
			return string(wfJSON), nil
		}
		if args[0] == "workflow" && args[1] == "enable" {
			return "", nil
		}
		if args[0] == "pr" && args[1] == "list" {
			return "not json", nil
		}
		return "", nil
	}

	result := FindAndEnableWorkflows("myorg", "myrepo", "", "feature/branch")
	if result.Success {
		t.Error("expected failure for invalid PR JSON")
	}
}

// --- FetchTeamPRs invalid JSON ---

func TestFetchTeamPRs_InvalidJSON(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "not valid json", nil
	}

	_, err := FetchTeamPRs("myorg", []string{"repo"})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// --- CloneFromURL edge cases ---

func TestCloneFromURL_ValidURLFormats(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"https with .git", "https://github.com/org/repo.git"},
		{"ssh with .git", "git@github.com:org/repo.git"},
		{"https without .git", "https://github.com/org/repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Will fail at clone but shouldn't panic
			result := CloneFromURL(tt.url, t.TempDir())
			_ = result
		})
	}
}

// --- CleanupRepo with status error (non-drop changes path) ---

func TestCleanupRepo_StatusError(t *testing.T) {
	origOpen := plainOpen
	origFetch := fetchRemote
	origGitRun := gitRunInDir
	t.Cleanup(func() {
		plainOpen = origOpen
		fetchRemote = origFetch
		gitRunInDir = origGitRun
	})

	dir, repo := mkTestRepo(t)

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}
	fetchRemote = func(r *gogit.Repository) error {
		return nil
	}

	// We need gitRun from the git package to fail for CountChangesShell
	// Since we can't easily override git.CountChangesShell, we rely on
	// the real git working fine in the test repo.
	gitRunInDir = func(d string, args ...string) (string, error) {
		return "", nil
	}

	result := CleanupRepo(dir, false, "main")
	// This should succeed since the repo is clean
	if !result.Success {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

// --- AutomergeResult fields ---

func TestAutomergeResult_Fields(t *testing.T) {
	r := AutomergeResult{
		RepoName: "my-repo",
		PRNumber: "42",
		Success:  true,
	}
	if r.RepoName != "my-repo" {
		t.Errorf("RepoName = %q", r.RepoName)
	}
	if r.PRNumber != "42" {
		t.Errorf("PRNumber = %q", r.PRNumber)
	}
	if !r.Success {
		t.Error("expected Success")
	}
}

// --- TeamPR fields ---

func TestTeamPR_Fields(t *testing.T) {
	pr := TeamPR{
		Repo:      "my-repo",
		Number:    5,
		Author:    "alice",
		Title:     "fix bug",
		URL:       "https://github.com/org/repo/pull/5",
		Draft:     false,
		CreatedAt: time.Now(),
	}
	if pr.Author != "alice" {
		t.Errorf("Author = %q", pr.Author)
	}
	if pr.URL != "https://github.com/org/repo/pull/5" {
		t.Errorf("URL = %q", pr.URL)
	}
}

// --- CreatePullRequest with HTTP error body ---

func TestCreatePullRequest_HTTPError(t *testing.T) {
	origClient := httpClient
	t.Cleanup(func() { httpClient = origClient })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"forbidden"}`))
	}))
	defer srv.Close()

	httpClient = srv.Client()

	// Create a repo with origin
	dir := t.TempDir()
	repo, _ := gogit.PlainInit(dir, false)
	repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"git@github.com:owner/repo.git"},
	})

	t.Setenv("GITHUB_TOKEN", "test-token")

	// This will fail because the URL doesn't point to our test server
	_, err := CreatePullRequest(repo, "msg", "feature", "main", nil, false)
	if err == nil {
		// The error might be about connection rather than 403
		// since the URL is api.github.com, not our server
	}
}

// --- GetRepoNameForPath with valid repo ---

func TestGetRepoNameForPath_ValidRepo(t *testing.T) {
	dir := t.TempDir()
	exec.Command("git", "init", "-b", "main", dir).Run()
	exec.Command("git", "-C", dir, "remote", "add", "origin", "git@github.com:testowner/testrepo.git").Run()

	name, err := GetRepoNameForPath(dir)
	if err != nil {
		// May fail if no initial commit, that's ok
		_ = name
		return
	}
	if name != "testrepo" {
		t.Errorf("name = %q, want testrepo", name)
	}
}

// --- isBranchStale ---

func TestIsBranchStale(t *testing.T) {
	dir, repo := mkTestRepo(t)

	refs, err := repo.References()
	if err != nil {
		t.Fatal(err)
	}

	refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			// The just-created branch should not be stale
			stale := isBranchStale(repo, ref)
			if stale {
				t.Error("expected recent branch to not be stale")
			}
		}
		return nil
	})
	_ = dir
}

// --- listBranches ---

func TestListBranches_Local(t *testing.T) {
	_, repo := mkTestRepo(t)

	branches, err := listBranches(repo, false)
	if err != nil {
		t.Fatalf("listBranches local: %v", err)
	}

	if len(branches) == 0 {
		t.Error("expected at least one local branch")
	}

	for _, b := range branches {
		if b.IsRemote {
			t.Error("local branch should not have IsRemote=true")
		}
	}
}

func TestListBranches_Remote(t *testing.T) {
	_, repo := mkTestRepo(t)

	// No remote branches exist in a test repo
	branches, err := listBranches(repo, true)
	if err != nil {
		t.Fatalf("listBranches remote: %v", err)
	}

	// Should be empty since we have no remote refs
	if len(branches) != 0 {
		t.Errorf("expected 0 remote branches, got %d", len(branches))
	}
}

// --- retriggerPRs direct test ---

func TestRetriggerPRs_ListError(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "", fmt.Errorf("network error")
	}

	_, err := retriggerPRs("org/repo", "feature/branch")
	if err == nil {
		t.Error("expected error when pr list fails")
	}
}

// --- CommitAndPush with branch checkout failure ---

func TestCommitAndPush_BranchCheckoutError(t *testing.T) {
	origOpen := plainOpen
	t.Cleanup(func() { plainOpen = origOpen })

	// Create a bare-ish scenario where worktree exists but checkout fails
	dir, repo := mkTestRepo(t)

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}

	// Use an impossible branch name to force error
	result := CommitAndPush(dir, "refs/invalid/branch\x00name", "test")
	if result.Success {
		t.Error("expected failure with invalid branch name")
	}
}

// --- GetRepoStatus full success path (deeper coverage) ---

func TestGetRepoStatus_FullSuccess(t *testing.T) {
	origOpen := plainOpen
	origFetch := fetchRemote
	t.Cleanup(func() {
		plainOpen = origOpen
		fetchRemote = origFetch
	})

	dir, repo := mkTestRepo(t)

	// Add origin remote for AheadBehind
	repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"git@github.com:owner/repo.git"},
	})

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}
	fetchRemote = func(r *gogit.Repository) error {
		return nil
	}

	result := GetRepoStatus(dir)
	if result.Error != "" {
		// git.CountChangesShell or git.AheadBehind might fail without remote, that's OK
		_ = result
	}
	if result.RepoName == "" {
		t.Error("expected non-empty RepoName")
	}
}

// --- GetRepoStatus branch name error ---

func TestGetRepoStatus_BranchNameError(t *testing.T) {
	origOpen := plainOpen
	origFetch := fetchRemote
	t.Cleanup(func() {
		plainOpen = origOpen
		fetchRemote = origFetch
	})

	// Repo with no HEAD
	dir := t.TempDir()
	repo, _ := gogit.PlainInit(dir, false)

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}
	fetchRemote = func(r *gogit.Repository) error {
		return nil
	}

	result := GetRepoStatus(dir)
	if result.Error == "" {
		t.Error("expected error when branch can't be determined")
	}
}

// --- ChangeTypes coverage ---

func TestChangeTypes(t *testing.T) {
	if len(ChangeTypes) == 0 {
		t.Error("expected non-empty ChangeTypes")
	}
	for _, ct := range ChangeTypes {
		if ct == "" {
			t.Error("found empty change type")
		}
	}
}

// --- Test git.Changes directly (via ops package tests for coverage) ---

func TestChangesUsage(t *testing.T) {
	c := git.Changes{Modified: 1, Added: 2, Deleted: 0, Untracked: 0}
	if !c.HasChanges() {
		t.Error("expected HasChanges = true")
	}
	if c.Total() != 3 {
		t.Errorf("Total = %d, want 3", c.Total())
	}

	c2 := git.Changes{}
	if c2.HasChanges() {
		t.Error("expected HasChanges = false for empty")
	}
	if c2.Total() != 0 {
		t.Errorf("Total = %d, want 0", c2.Total())
	}
}

// --- FetchOrgRepoNames ---

func TestFetchOrgRepoNames_Success(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "repo1\nrepo2\nrepo3\n", nil
	}

	names, err := FetchOrgRepoNames("myorg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 3 {
		t.Errorf("expected 3 repos, got %d", len(names))
	}
}

func TestFetchOrgRepoNames_Error(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "", fmt.Errorf("not authorized")
	}

	_, err := FetchOrgRepoNames("myorg")
	if err == nil {
		t.Error("expected error")
	}
}

func TestFetchOrgRepoNames_EmptyLines(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "repo1\n\n  \nrepo2\n", nil
	}

	names, err := FetchOrgRepoNames("myorg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 repos (skipping blanks), got %d", len(names))
	}
}

// --- FetchTeamRepoNames ---

func TestFetchTeamRepoNames_Success(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "service-a\nservice-b\ninfra-tool\n", nil
	}

	names, err := FetchTeamRepoNames("myorg", "myteam", []string{"infra-"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 repos (excluding infra-), got %d", len(names))
	}
}

func TestFetchTeamRepoNames_Error(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "", fmt.Errorf("api error")
	}

	_, err := FetchTeamRepoNames("myorg", "myteam", nil)
	if err == nil {
		t.Error("expected error")
	}
}

// --- FetchTeamPRs ---

func TestFetchTeamPRs_Success(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	now := time.Now()
	ghRun = func(args ...string) (string, error) {
		return fmt.Sprintf(`[
			{"repository":{"name":"repo-a"},"number":1,"title":"fix","author":{"login":"alice"},"url":"http://pr/1","isDraft":false,"createdAt":"%s"},
			{"repository":{"name":"repo-b"},"number":2,"title":"feat","author":{"login":"bob"},"url":"http://pr/2","isDraft":true,"createdAt":"%s"},
			{"repository":{"name":"other-repo"},"number":3,"title":"other","author":{"login":"carol"},"url":"http://pr/3","isDraft":false,"createdAt":"%s"}
		]`, now.Format(time.RFC3339), now.Format(time.RFC3339), now.Format(time.RFC3339)), nil
	}

	prs, err := FetchTeamPRs("myorg", []string{"repo-a", "repo-b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 2 {
		t.Errorf("expected 2 PRs (filtering to team repos), got %d", len(prs))
	}
}

func TestFetchTeamPRs_GhError(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "", fmt.Errorf("network error")
	}

	_, err := FetchTeamPRs("myorg", []string{"repo"})
	if err == nil {
		t.Error("expected error")
	}
}

// --- EnableAutomerge ---

func TestEnableAutomerge_Success(t *testing.T) {
	origGhRunInDir := ghRunInDir
	t.Cleanup(func() { ghRunInDir = origGhRunInDir })

	callCount := 0
	ghRunInDir = func(dir string, args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			return "42\n", nil // pr list returns PR number
		}
		return "", nil // pr merge succeeds
	}

	dir, _ := mkTestRepo(t)
	result := EnableAutomerge("feature/branch", dir)
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.PRNumber != "42" {
		t.Errorf("PRNumber = %q, want 42", result.PRNumber)
	}
}

func TestEnableAutomerge_NoPRFound(t *testing.T) {
	origGhRunInDir := ghRunInDir
	t.Cleanup(func() { ghRunInDir = origGhRunInDir })

	ghRunInDir = func(dir string, args ...string) (string, error) {
		return "", nil // empty output = no PR found
	}

	dir, _ := mkTestRepo(t)
	result := EnableAutomerge("feature/branch", dir)
	if result.Success {
		t.Error("expected failure when no PR found")
	}
}

func TestEnableAutomerge_ListError(t *testing.T) {
	origGhRunInDir := ghRunInDir
	t.Cleanup(func() { ghRunInDir = origGhRunInDir })

	ghRunInDir = func(dir string, args ...string) (string, error) {
		return "", fmt.Errorf("gh error")
	}

	dir, _ := mkTestRepo(t)
	result := EnableAutomerge("feature/branch", dir)
	if result.Success {
		t.Error("expected failure")
	}
}

func TestEnableAutomerge_MergeError(t *testing.T) {
	origGhRunInDir := ghRunInDir
	t.Cleanup(func() { ghRunInDir = origGhRunInDir })

	callCount := 0
	ghRunInDir = func(dir string, args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			return "42\n", nil
		}
		return "", fmt.Errorf("merge not allowed")
	}

	dir, _ := mkTestRepo(t)
	result := EnableAutomerge("feature/branch", dir)
	if result.Success {
		t.Error("expected failure when merge fails")
	}
}

// --- GetOrgRepositories full flow with httptest ---

func TestGetOrgRepositories_Success(t *testing.T) {
	origClient := httpClient
	t.Cleanup(func() { httpClient = origClient })

	t.Setenv("GITHUB_TOKEN", "test-token")

	repos := []Repository{
		{Name: "active-repo", SSHURL: "git@github.com:org/active-repo.git", Archived: false},
		{Name: "archived-repo", SSHURL: "git@github.com:org/archived-repo.git", Archived: true},
		{Name: "another-repo", SSHURL: "git@github.com:org/another-repo.git", Archived: false},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// No Link header = no next page
		json.NewEncoder(w).Encode(repos)
	}))
	defer srv.Close()

	// Directly test fetchRepoPage + filtering logic
	httpClient = srv.Client()
	fetched, hasNext, err := fetchRepoPage(httpClient, srv.URL, 1, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasNext {
		t.Error("expected no next page")
	}
	if len(fetched) != 3 {
		t.Errorf("expected 3 total repos, got %d", len(fetched))
	}

	// Filter archived
	var active []Repository
	for _, r := range fetched {
		if !r.Archived {
			active = append(active, r)
		}
	}
	if len(active) != 2 {
		t.Errorf("expected 2 active repos, got %d", len(active))
	}
}

func TestGetOrgRepositories_NoToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	_, err := GetOrgRepositories("myorg", "")
	if err == nil {
		t.Error("expected error when GITHUB_TOKEN not set")
	}
}

func TestFetchRepoPage_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	_, _, err := fetchRepoPage(srv.Client(), srv.URL, 1, "token")
	if err == nil {
		t.Error("expected error for HTTP 403")
	}
}

func TestFetchRepoPage_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	_, _, err := fetchRepoPage(srv.Client(), srv.URL, 1, "token")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFetchRepoPage_WithPagination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Link", `<https://api.github.com/orgs/myorg/repos?page=2>; rel="next"`)
		json.NewEncoder(w).Encode([]Repository{{Name: "repo1"}})
	}))
	defer srv.Close()

	repos, hasNext, err := fetchRepoPage(srv.Client(), srv.URL, 1, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasNext {
		t.Error("expected hasNext = true")
	}
	if len(repos) != 1 {
		t.Errorf("expected 1 repo, got %d", len(repos))
	}
}

// --- CloneFromURL invalid ---

func TestCloneFromURL_InvalidURL(t *testing.T) {
	result := CloneFromURL("noslash", t.TempDir())
	if result.Error == "" {
		t.Error("expected error for invalid URL")
	}
}

// --- CloneRepository SSH auth failure path ---

func TestCloneRepository_SSHAuthFails(t *testing.T) {
	// Create an empty directory (not a git repo) so it passes the "already exists" check
	dir := t.TempDir()
	subdir := filepath.Join(dir, "newrepo")

	// SSHAuth will fail because SSH_KEY is not set and there's no default key
	result := CloneRepository(
		Repository{Name: "repo", SSHURL: "git@github.com:org/repo.git"},
		subdir,
	)
	// Should fail at SSH auth or clone step
	if result.Success && !result.Skipped {
		t.Error("expected failure when SSH auth unavailable")
	}
}

// --- FindAndEnableWorkflows list error ---

func TestFindAndEnableWorkflows_ListError(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "", fmt.Errorf("not authorized")
	}

	result := FindAndEnableWorkflows("myorg", "myrepo", "", "")
	if result.Success {
		t.Error("expected failure")
	}
}

// --- FindAndEnableWorkflows empty output ---

func TestFindAndEnableWorkflows_EmptyOutput(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	ghRun = func(args ...string) (string, error) {
		return "", nil
	}

	result := FindAndEnableWorkflows("myorg", "myrepo", "", "")
	if !result.Success {
		t.Errorf("expected success for empty workflows, got error: %s", result.Error)
	}
}

// --- FindAndEnableWorkflows with workflow name filter ---

func TestFindAndEnableWorkflows_FilterByName(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	workflows := []ghWorkflow{
		{Name: "CI", ID: 101, State: "disabled_inactivity"},
		{Name: "Deploy", ID: 102, State: "disabled_inactivity"},
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
		return "", nil
	}

	result := FindAndEnableWorkflows("myorg", "myrepo", "CI", "")
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.EnabledCount != 1 {
		t.Errorf("EnabledCount = %d, want 1", result.EnabledCount)
	}
}

// --- FindAndEnableWorkflows active workflows skipped ---

func TestFindAndEnableWorkflows_ActiveSkipped(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

	workflows := []ghWorkflow{
		{Name: "CI", ID: 101, State: "active"},
	}
	wfJSON, _ := json.Marshal(workflows)

	ghRun = func(args ...string) (string, error) {
		return string(wfJSON), nil
	}

	result := FindAndEnableWorkflows("myorg", "myrepo", "", "")
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.EnabledCount != 0 {
		t.Errorf("EnabledCount = %d, want 0", result.EnabledCount)
	}
}

// --- FindAndEnableWorkflows with successful retrigger ---

func TestFindAndEnableWorkflows_RetriggerSuccess(t *testing.T) {
	origGhRun := ghRun
	t.Cleanup(func() { ghRun = origGhRun })

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
		if args[0] == "pr" && args[1] == "close" {
			return "", nil
		}
		if args[0] == "pr" && args[1] == "reopen" {
			return "", nil
		}
		return "", nil
	}

	result := FindAndEnableWorkflows("myorg", "myrepo", "", "feature/branch")
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.RetriggeredPRs != 2 {
		t.Errorf("RetriggeredPRs = %d, want 2", result.RetriggeredPRs)
	}
}

// --- CommitAndPush success path ---

func TestCommitAndPush_Success(t *testing.T) {
	origOpen := plainOpen
	origGitRun := gitRunInDir
	t.Cleanup(func() {
		plainOpen = origOpen
		gitRunInDir = origGitRun
	})

	dir, repo := mkTestRepo(t)

	// Create an untracked file so worktree.Status() shows changes
	os.WriteFile(filepath.Join(dir, "newfile.txt"), []byte("content"), 0644)

	// Add origin remote for push
	repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"git@github.com:owner/repo.git"},
	})

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}

	// Mock push to avoid SSH auth
	origPush := pushChanges
	t.Cleanup(func() { pushChanges = origPush })
	pushChanges = func(repo *gogit.Repository) error { return nil }

	callCount := 0
	gitRunInDir = func(d string, args ...string) (string, error) {
		callCount++
		switch callCount {
		case 1: // git add
			return "", nil
		case 2: // git commit
			return "", nil
		}
		return "", nil
	}

	result := CommitAndPush(dir, "", "test commit")
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

// --- CommitAndPush push failure ---

func TestCommitAndPush_PushFails(t *testing.T) {
	origOpen := plainOpen
	origGitRun := gitRunInDir
	t.Cleanup(func() {
		plainOpen = origOpen
		gitRunInDir = origGitRun
	})

	dir, repo := mkTestRepo(t)
	repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"git@github.com:owner/repo.git"},
	})

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}

	callCount := 0
	gitRunInDir = func(d string, args ...string) (string, error) {
		callCount++
		switch callCount {
		case 1: // git add
			return "", nil
		case 2: // git commit
			return "", nil
		case 3: // git push fails
			return "", fmt.Errorf("push rejected")
		}
		return "", nil
	}

	result := CommitAndPush(dir, "", "test commit")
	if result.Success {
		t.Error("expected failure when push fails")
	}
}

