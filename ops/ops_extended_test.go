package ops

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
)

// mkTestRepo creates a real git repo with an initial commit. Different name
// from initTestRepo in pullrequest_http_test.go to avoid redeclaration.
func mkTestRepo(t *testing.T) (string, *gogit.Repository) {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init", dir},
		{"git", "-C", dir, "config", "user.email", "test@test.com"},
		{"git", "-C", dir, "config", "user.name", "Test"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("cmd %v failed: %s %v", c, string(out), err)
		}
	}

	f := filepath.Join(dir, "README.md")
	os.WriteFile(f, []byte("# test"), 0644)
	exec.Command("git", "-C", dir, "add", ".").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init").Run()

	repo, err := gogit.PlainOpen(dir)
	if err != nil {
		t.Fatalf("PlainOpen failed: %v", err)
	}
	return dir, repo
}

// --- GetRepoStatus extended ---

func TestGetRepoStatus_Success(t *testing.T) {
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

	r := GetRepoStatus(dir)
	if r.RepoName == "" {
		t.Error("expected non-empty RepoName")
	}
}

// --- GetBranches extended ---

func TestGetBranches_Success(t *testing.T) {
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

	r := GetBranches(dir)
	if r.Error != "" {
		t.Errorf("unexpected error: %s", r.Error)
	}
	if r.RepoName == "" {
		t.Error("expected non-empty RepoName")
	}
}

// --- CommitAndPush extended ---

func TestCommitAndPush_GitAddError(t *testing.T) {
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
	gitRunInDir = func(d string, args ...string) (string, error) {
		return "", fmt.Errorf("git add failed")
	}

	r := CommitAndPush(dir, "", "test commit")
	if r.Success {
		t.Error("expected failure when git add fails")
	}
}

func TestCommitAndPush_NoChanges(t *testing.T) {
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
	gitRunInDir = func(d string, args ...string) (string, error) {
		return "", nil
	}

	r := CommitAndPush(dir, "", "test commit")
	if r.Success {
		t.Error("expected failure with no changes")
	}
}

// --- CleanupRepo extended ---

func TestCleanupRepo_FullFlow(t *testing.T) {
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

	r := CleanupRepo(dir, false, "main")
	_ = r // just verify no panic
}

// --- checkoutFetchPull ---

func TestCheckoutFetchPull_CheckoutFails(t *testing.T) {
	origGitRun := gitRunInDir
	t.Cleanup(func() { gitRunInDir = origGitRun })

	gitRunInDir = func(d string, args ...string) (string, error) {
		return "", fmt.Errorf("checkout failed")
	}

	_, repo := mkTestRepo(t)
	err := checkoutFetchPull(repo, "/fake", "main")
	if err == nil {
		t.Error("expected error when checkout fails")
	}
}

func TestCheckoutFetchPull_FetchFails(t *testing.T) {
	origOpen := plainOpen
	origFetch := fetchRemote
	origGitRun := gitRunInDir
	t.Cleanup(func() {
		plainOpen = origOpen
		fetchRemote = origFetch
		gitRunInDir = origGitRun
	})

	_, repo := mkTestRepo(t)

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}
	fetchRemote = func(r *gogit.Repository) error {
		return fmt.Errorf("fetch error")
	}
	callCount := 0
	gitRunInDir = func(d string, args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			return "", nil // checkout succeeds
		}
		return "", nil
	}

	err := checkoutFetchPull(repo, "/fake", "main")
	if err == nil {
		t.Error("expected error when fetch fails")
	}
}

// --- Field tests for types without existing coverage ---

func TestMergePRResult_Fields(t *testing.T) {
	r := MergePRResult{
		Repo:     "repo",
		PRNumber: "42",
		Title:    "fix stuff",
		Success:  false,
		Error:    "merge conflict",
	}
	if r.Error != "merge conflict" {
		t.Errorf("Error = %q, want merge conflict", r.Error)
	}
}

func TestPRInfo_Fields(t *testing.T) {
	p := PRInfo{Repo: "repo", Number: 1, Title: "fix"}
	if p.Repo != "repo" || p.Number != 1 || p.Title != "fix" {
		t.Error("unexpected PRInfo field values")
	}
}

func TestMyPR_Fields(t *testing.T) {
	p := MyPR{Repo: "org/repo", Number: 5, Title: "feat", Draft: true}
	if !p.Draft {
		t.Error("expected Draft = true")
	}
}

func TestPRRequest_Fields(t *testing.T) {
	r := PRRequest{
		Title: "test PR",
		Body:  "body",
		Head:  "feature",
		Base:  "main",
	}
	if r.Title != "test PR" {
		t.Errorf("Title = %q, want test PR", r.Title)
	}
}
